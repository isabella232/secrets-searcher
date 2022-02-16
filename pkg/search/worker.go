package search

import (
	"github.com/pantheon-systems/secrets-searcher/pkg/dev"
	"github.com/pantheon-systems/secrets-searcher/pkg/errors"
	gitpkg "github.com/pantheon-systems/secrets-searcher/pkg/git"
	"github.com/pantheon-systems/secrets-searcher/pkg/logg"
	"github.com/pantheon-systems/secrets-searcher/pkg/search/contract"
)

type Worker struct {
	processors       []contract.ProcessorI
	targets          []contract.ProcessorI
	fileChangeFilter *gitpkg.FileChangeFilter
	log              logg.Logg
}

func NewWorker(processors []contract.ProcessorI, fileChangeFilter *gitpkg.FileChangeFilter, log logg.Logg) *Worker {

	return &Worker{
		processors:       processors,
		fileChangeFilter: fileChangeFilter,
		log:              log,
	}
}

func (w *Worker) Do(job contract.WorkerJobI) {
	var commits []*gitpkg.Commit
	var err error
	commits, err = job.Start()
	if err != nil {
		return
	}

	for _, commit := range commits {
		job.SearchingCommit(commit)

		if err = w.findInCommit(job, commit); err != nil {
			errors.ErrLog(job.Log(w.log), err).Error("error while processing commit")
		}
	}

	job.Finish()

	return
}

func (w *Worker) findInCommit(job contract.WorkerJobI, commit *gitpkg.Commit) (err error) {
	defer job.Increment()
	defer errors.CatchPanicDo(func(err error) { job.Log(w.log).Error(err, "error during commit search") })

	job.Log(w.log).WithField("commitDate", commit.Date.Format("2006-01-02")).
		Debug("searching commit")

	var fileChanges []*gitpkg.FileChange
	fileChanges, err = w.getFileChangesForCommit(job, commit)
	if err != nil {
		err = errors.WithMessage(err, "unable to get file changes")
		return
	}

	for _, change := range fileChanges {
		job.SearchingFileChange(change)

		err = w.findInFileChange(job)
		if err != nil {
			err = errors.WithMessage(err, "unable to find in file change")
			return
		}
	}

	return
}

func (w *Worker) findInFileChange(job contract.WorkerJobI) (err error) {
	defer errors.CatchPanicDo(func(err error) { job.Log(w.log).Error(err, "error during file change search") })

	for _, proc := range w.processors {
		procName := proc.GetName()
		path := job.FileChange().Path

		job.SearchingWithProcessor(proc)
		job.Diff().SetLine(1)
		dev.BreakpointInProcessor(path, procName, -1)

		err = proc.FindResultsInFileChange(job)
		if err != nil {
			err = errors.WithMessagev(err, "unable to search in file change using processor", procName)
			return
		}
	}

	return
}

func (w *Worker) getFileChangesForCommit(job contract.WorkerJobI, commit *gitpkg.Commit) (result []*gitpkg.FileChange, err error) {

	// The git.Change.Patch() function is too panicky so we'll just log it here
	defer errors.CatchPanicAndLogWarning(job.Log(w.log), "got a panic while getting file changes, probably from git.Change.Patch()")

	result, err = commit.FileChanges(w.fileChangeFilter)

	return
}
