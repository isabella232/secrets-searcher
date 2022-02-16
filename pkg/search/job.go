package search

import (
	"fmt"

	"github.com/pantheon-systems/secrets-searcher/pkg/stats"

	"github.com/pantheon-systems/secrets-searcher/pkg/database"
	"github.com/pantheon-systems/secrets-searcher/pkg/errors"
	"github.com/pantheon-systems/secrets-searcher/pkg/git"
	"github.com/pantheon-systems/secrets-searcher/pkg/interact/progress"
	"github.com/pantheon-systems/secrets-searcher/pkg/logg"
	"github.com/pantheon-systems/secrets-searcher/pkg/manip"
	"github.com/pantheon-systems/secrets-searcher/pkg/search/contract"
)

type (
	Job struct {
		name         string
		repoID       string
		repoName     string
		repository   *git.Repository
		commitHashes []string
		oldest       string
		bar          *progress.Bar
		log          logg.Logg
		jobState
	}
	jobState struct {
		startedFlag bool

		// Results
		results []*contract.JobResult
		ignores map[*git.FileChange][]*manip.FileRange

		scope    *Scope
		scopeLog logg.Logg

		// Stats
		secretTracker manip.Set
	}
)

func NewJob(name, repoID, repoName string, repository *git.Repository, commitHashes []string, oldest string, enableProfiling bool, bar *progress.Bar, log logg.Logg, stats *stats.Stats) (result *Job) {
	return &Job{
		name:         name,
		repoID:       repoID,
		repoName:     repoName,
		repository:   repository,
		commitHashes: commitHashes,
		oldest:       oldest,
		bar:          bar,
		log:          log,
		jobState: jobState{
			scope:         NewScope(enableProfiling, stats),
			secretTracker: manip.NewEmptyBasicSet(),
			ignores:       make(map[*git.FileChange][]*manip.FileRange),
		},
	}
}

func (j *Job) Start() (result []*git.Commit, err error) {
	if j.startedFlag {
		panic("already started")
	}
	j.startedFlag = true

	j.log.Debug("started job")

	if j.bar != nil {
		j.bar.Start()
	}

	result, err = j.commits()

	j.scope.StartRepo(j.repoName)
	j.scope.StartRepoJob(j.name)

	return
}

func (j *Job) SearchingCommit(commit *git.Commit) {
	j.scope.StartCommit(commit)
}

func (j *Job) SearchingFileChange(fileChange *git.FileChange) {
	j.scope.StartFileChange(fileChange)
}

func (j *Job) SearchingWithProcessor(proc contract.NamedProcessorI) {
	j.scope.StartProc(proc)
}

func (j *Job) SearchingLine(line int) {
	j.scope.StartLine(line)
}

func (j *Job) Commit() (commit *git.Commit) {
	return j.scope.Commit
}

func (j *Job) Processor() (proc contract.NamedProcessorI) {
	return j.scope.Proc
}

func (j *Job) FileChange() (fileChange *git.FileChange) {
	return j.scope.FileChange
}

func (j *Job) Line() (line int) {
	return j.scope.Line
}

func (j *Job) Diff() (result *git.Diff) {
	if j.scope.FileChange == nil {
		panic("file change not set, cannot get diff")
	}

	var err error
	if result, err = j.scope.FileChange.Diff(); err != nil {
		j.Log(nil).Error("unable to get diff")
		panic("unable to get diff")
	}

	return
}

func (j *Job) Log(prefixLog logg.Logg) (result logg.Logg) {
	if j.scopeLog != nil {
		return j.scopeLog
	}

	result = j.log

	if prefixLog != nil {
		if prefix, ok := prefixLog.Data()["prefix"]; ok {
			result = result.WithField("prefix", prefix)
		}
	}

	// Add scope fields
	result = result.WithFields(j.scope.Fields())
	j.scopeLog = result

	return
}

func (j *Job) SubmitResult(result *contract.Result) {
	log := j.Log(j.log).WithField("fileRange", result.FileRange)
	log.Info("a secret was found!")
	log = log.WithField("secretValue", result.SecretValue)

	// Validate
	j.checkSubmission(result.FileRange)
	if result.SecretValue == "" {
		panic("missing secret value")
	}

	// Check if the secret has been ignored by another processor
	if ignores, ok := j.ignores[j.scope.FileChange]; ok {
		for _, ignore := range ignores {
			if ignore.Overlaps(result.FileRange) {
				log.WithField("existingIgnore", ignore).
					Tracef("finding overlaps with %v, ignored", ignore)
				return
			}
		}
	}

	// Add result to ignore list
	j.ignores[j.scope.FileChange] = append(j.ignores[j.scope.FileChange], result.FileRange)

	// Build result
	j.results = append(j.results, &contract.JobResult{
		RepoID:           j.repoID,
		Processor:        j.scope.Proc,
		FileChange:       j.scope.FileChange,
		SecretValue:      result.SecretValue,
		FileRange:        result.FileRange,
		ContextFileRange: result.ContextFileRange,
		FileBaseName:     result.FileBasename,
		SecretExtras:     result.SecretExtras,
		FindingExtras:    result.FindingExtras,
	})

	// For stats
	j.secretTracker.Add(database.CreateHashID(result.SecretValue))
}

func (j *Job) SubmitIgnore(fileRange *manip.FileRange) {
	log := j.Log(j.log).WithField("lineRange", fileRange)
	log.Trace("ignore submitted")

	// Validate
	j.checkSubmission(fileRange)

	j.ignores[j.scope.FileChange] = append(j.ignores[j.scope.FileChange], fileRange)
}

func (j *Job) checkSubmission(fileRange *manip.FileRange) {
	if j.scope.Proc == nil {
		panic("no processor")
	}
	if j.scope.FileChange == nil {
		panic("no file change")
	}
	if fileRange == nil {
		panic("no file range submitted")
	}
	procName := j.scope.Proc.GetName()
	if procName == "" {
		panic("missing proc name")
	}
	if fileRange.StartLineNum != j.scope.Line {
		j.Log(j.log).WithField("submittedLine", fileRange.StartLineNum).
			Warn("you didn't tell the job you were searching this line")
	}
}

func (j *Job) LogError(errInput error, message string) {
	err := errors.WithStack(errInput)

	if j.bar == nil {
		errors.ErrLog(j.log, err).Error(message)
		return
	}

	j.bar.BustThrough(func() {
		errors.ErrLog(j.log, err).Error(message)
	})
}

func (j *Job) Increment() {
	if j.bar != nil {
		j.bar.Incr()
	}
}

func (j *Job) Finish() {
	j.scope.FinishRepo()

	if j.bar != nil {
		secretsFound := j.secretTracker.Len()
		message := fmt.Sprintf("%d commits searched, %d secrets found",
			len(j.commitHashes), secretsFound)

		j.bar.Finished(message)
	}

	j.log.Debug("ended job")
}

func (j *Job) GetJobResults() []*contract.JobResult {
	results := *&j.results

	// Self-destruct for GC
	j.repository = nil
	j.results = nil
	j.ignores = nil
	j.scope = nil

	return results
}

func (j *Job) commits() (result []*git.Commit, err error) {
	result = []*git.Commit{}
	for _, commitHash := range j.commitHashes {
		var commit *git.Commit
		commit, err = j.repository.Commit(commitHash)
		if err != nil {
			err = errors.WithMessagev(err, "unable to get commit", commitHash)
			return
		}
		commit.Oldest = commit.Hash == j.oldest
		result = append(result, commit)
	}

	return
}
