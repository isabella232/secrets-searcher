package contract

import (
	"github.com/pantheon-systems/secrets-searcher/pkg/git"
	"github.com/pantheon-systems/secrets-searcher/pkg/logg"
	"github.com/pantheon-systems/secrets-searcher/pkg/manip"
)

type (

	//
	// Result writer (for runner to use)

	ResultWriter interface {
		WriteResult(result *JobResult) (err error)
	}

	//
	// Job

	AnyJobI interface {
		Log(yourLog logg.Logg) (result logg.Logg)
	}

	IsManaged interface {
		Start() (result []*git.Commit, err error)
		Increment() // FIXME Deprecate for SearchingCommit()
		Finish()
		GetJobResults() []*JobResult
	}

	AcceptsContext interface {
		SearchingCommit(commit *git.Commit)
		SearchingWithProcessor(proc NamedProcessorI)
		SearchingFileChange(fileChange *git.FileChange)
	}
	AcceptsLineContext interface {
		SearchingLine(line int)
	}
	HasContext interface {
		Commit() (commit *git.Commit)
		Processor() (proc NamedProcessorI)
	}
	HasFileChangeContext interface {
		FileChange() (fileChange *git.FileChange)
		Diff() (result *git.Diff)
		Line() (line int)
	}

	AcceptsIgnoreI interface {
		SubmitIgnore(fileRange *manip.FileRange)
	}
	AcceptsLineRangeIgnoreI interface {
		SubmitLineRangeIgnore(lineRange *manip.LineRange)
	}

	AcceptsResultI interface {
		SubmitResult(result *Result)
	}
	AcceptsLineResultI interface {
		SubmitLineResult(lineResult *LineResult)
	}

	DealsWithProcessor interface {
		AnyJobI
		AcceptsResultI
		AcceptsIgnoreI
		AcceptsContext
		HasContext
		AcceptsLineContext
		HasFileChangeContext
	}

	WorkerJobI interface {
		DealsWithProcessor
		IsManaged
	}
	ProcessorJobI interface {
		DealsWithProcessor
	}
	LineProcessorJobI interface {
		AnyJobI
		AcceptsLineRangeIgnoreI
		AcceptsLineResultI
		HasFileChangeContext
	}

	//
	// Processor

	FindsResultsI interface {
		FindResultsInFileChange(job ProcessorJobI) (err error)
	}
	FindsResultsInLineI interface {
		FindResultsInLine(job LineProcessorJobI, line string) (err error)
	}
	NamedProcessorI interface {
		manip.Named
	}
	ProcessorI interface {
		NamedProcessorI
		FindsResultsI
	}
	LineProcessorI interface {
		NamedProcessorI
		FindsResultsInLineI
	}

	//
	// Result

	Result struct {
		FileRange        *manip.FileRange
		ContextFileRange *manip.FileRange
		SecretValue      string
		SecretExtras     []*ResultExtra
		FindingExtras    []*ResultExtra
		FileBasename     string
	}
	LineResult struct {
		LineRange        *manip.LineRange
		ContextLineRange *manip.LineRange
		SecretValue      string
		SecretExtras     []*ResultExtra
		FindingExtras    []*ResultExtra
	}
	ResultExtra struct {
		Key    string
		Header string
		Value  string
		Code   bool
		URL    string
		Debug  bool
	}

	JobResult struct {
		RepoID           string
		Processor        NamedProcessorI
		FileChange       *git.FileChange
		Line             string
		FileRange        *manip.FileRange
		ContextFileRange *manip.FileRange
		FileBaseName     string
		SecretValue      string
		SecretExtras     []*ResultExtra
		FindingExtras    []*ResultExtra
	}
)
