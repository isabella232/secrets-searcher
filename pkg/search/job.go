package search

import (
	"fmt"

	"github.com/pantheon-systems/search-secrets/pkg/database"
	"github.com/pantheon-systems/search-secrets/pkg/errors"
	gitpkg "github.com/pantheon-systems/search-secrets/pkg/git"
	"github.com/pantheon-systems/search-secrets/pkg/interact/progress"
	"github.com/pantheon-systems/search-secrets/pkg/logg"
	"github.com/pantheon-systems/search-secrets/pkg/manip"
	"github.com/pantheon-systems/search-secrets/pkg/search/contract"
)

type (
	Job struct {
		name         string
		repoID       string
		repoName     string
		repository   *gitpkg.Repository
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
		ignores map[*gitpkg.FileChange][]*manip.FileRange

		// Stats
		secretTracker manip.Set

		// Logging concerns
		commit     *gitpkg.Commit
		fileChange *gitpkg.FileChange
		proc       contract.NamedProcessorI
		line       int

		scopedLogCache logg.Logg
	}
)

func NewJob(name, repoID, repoName string, repository *gitpkg.Repository, commitHashes []string, oldest string,
	bar *progress.Bar, log logg.Logg) (result *Job) {

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
			secretTracker: manip.NewEmptyBasicSet(),
			ignores:       make(map[*gitpkg.FileChange][]*manip.FileRange),
		},
	}
}

func (j *Job) Start() (result []*gitpkg.Commit, err error) {
	if j.startedFlag {
		panic("already started")
	}
	j.startedFlag = true

	j.log.Debug("started job")

	if j.bar != nil {
		j.bar.Start()
	}

	result, err = j.commits()

	return
}

func (j *Job) SearchingCommit(commit *gitpkg.Commit) {
	j.commit = commit
	j.fileChange = nil
	j.proc = nil
	j.line = 0

	j.scopedLogCache = nil
}

func (j *Job) SearchingFileChange(fileChange *gitpkg.FileChange) {
	if j.commit == nil {
		panic("commit not set, cannot set file change")
	}
	j.fileChange = fileChange
	j.proc = nil
	j.line = 0

	j.scopedLogCache = nil
}

func (j *Job) Diff() (result *gitpkg.Diff) {
	if j.fileChange == nil {
		panic("file change not set, cannot get diff")
	}

	var err error
	if result, err = j.fileChange.Diff(); err != nil {
		j.Log(nil).Error("unable to get diff")
		panic("unable to get diff")
	}

	//result.SetLineHook = func(line *gitpkg.Line) {
	//	j.SearchingLine(line.NumInFile)
	//}

	return
}

func (j *Job) SearchingWithProcessor(proc contract.NamedProcessorI) {
	if j.commit == nil {
		panic("commit not set, cannot set processor")
	}
	if j.fileChange == nil {
		panic("file change not set, cannot set processor")
	}
	j.proc = proc
	j.line = 0

	j.scopedLogCache = nil
}

func (j *Job) SearchingLine(line int) {
	if j.commit == nil {
		panic("commit not set, cannot set line")
	}
	if j.fileChange == nil {
		panic("file change not set, cannot set line")
	}
	if j.proc == nil {
		panic("processor not set, cannot set line")
	}
	j.line = line

	j.scopedLogCache = nil
}

func (j *Job) Commit() (commit *gitpkg.Commit) {
	return j.commit
}

func (j *Job) Processor() (proc contract.NamedProcessorI) {
	return j.proc
}

func (j *Job) FileChange() (fileChange *gitpkg.FileChange) {
	return j.fileChange
}

func (j *Job) Line() (line int) {
	return j.line
}

func (j *Job) Log(prefixLog logg.Logg) (result logg.Logg) {
	if j.scopedLogCache != nil {
		result = j.scopedLogCache
		return
	}

	result = j.log

	if prefixLog != nil {
		if prefix, ok := prefixLog.Data()["prefix"]; ok {
			result = result.WithField("prefix", prefix)
		}
	}

	if j.commit != nil {
		result = result.WithField("commit", j.commit.Hash)
	}
	if j.fileChange != nil {
		result = result.WithField("path", j.fileChange.Path)
	}
	if j.proc != nil {
		result = result.WithField("processor", j.proc.GetName())
	}
	if j.line != 0 {
		result = result.WithField("line", j.line)
	}

	j.scopedLogCache = result

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
	if ignores, ok := j.ignores[j.fileChange]; ok {
		for _, ignore := range ignores {
			if ignore.Overlaps(result.FileRange) {
				log.WithField("existingIgnore", ignore).
					Tracef("finding overlaps with %v, ignored", ignore)
				return
			}
		}
	}

	// Add result to ignore list
	j.ignores[j.fileChange] = append(j.ignores[j.fileChange], result.FileRange)

	// Build result
	j.results = append(j.results, &contract.JobResult{
		RepoID:           j.repoID,
		Processor:        j.proc,
		FileChange:       j.fileChange,
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

	j.ignores[j.fileChange] = append(j.ignores[j.fileChange], fileRange)
}

func (j *Job) checkSubmission(fileRange *manip.FileRange) {
	if j.proc == nil {
		panic("no processor")
	}
	if j.fileChange == nil {
		panic("no file change")
	}
	if fileRange == nil {
		panic("no file range submitted")
	}
	procName := j.proc.GetName()
	if procName == "" {
		panic("missing proc name")
	}
	if fileRange.StartLineNum != j.line {
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
	if j.bar != nil {
		secretsFound := j.secretTracker.Len()
		message := fmt.Sprintf("%d commits searched", len(j.commitHashes))
		if secretsFound > 0 {
			message += fmt.Sprintf(", %d SECRETS FOUND", secretsFound)
		}

		j.bar.Finished(message)
	}

	j.log.Debug("ended job")
}

func (j *Job) GetJobResults() []*contract.JobResult {

	results := *&j.results

	// FIXME
	j.commit = nil
	j.fileChange = nil
	j.proc = nil
	j.scopedLogCache = nil
	j.ignores = nil
	// FIXME Should these be in the state?
	j.repository = nil
	j.commitHashes = nil

	return results
}

func (j *Job) commits() (result []*gitpkg.Commit, err error) {
	result = []*gitpkg.Commit{}
	for _, commitHash := range j.commitHashes {
		var commit *gitpkg.Commit
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
