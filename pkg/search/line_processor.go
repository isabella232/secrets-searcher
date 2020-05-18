package search

import (
	"strings"

	"github.com/pantheon-systems/search-secrets/pkg/dev"
	"github.com/pantheon-systems/search-secrets/pkg/errors"
	gitpkg "github.com/pantheon-systems/search-secrets/pkg/git"
	"github.com/pantheon-systems/search-secrets/pkg/logg"
	"github.com/pantheon-systems/search-secrets/pkg/manip"
	"github.com/pantheon-systems/search-secrets/pkg/search/contract"
)

//
// LineProcessor

type LineProcessorJob struct {
	job contract.ProcessorJobI
}

func NewLineProcessorJob(job contract.ProcessorJobI) *LineProcessorJob {
	return &LineProcessorJob{
		job: job,
	}
}

func (l *LineProcessorJob) SubmitLineRangeIgnore(lineRange *manip.LineRange) {
	fileRange := manip.NewFileRangeFromLineRange(lineRange, l.job.Line())
	l.job.SubmitIgnore(fileRange)
}

func (l *LineProcessorJob) SubmitLineResult(lineResult *contract.LineResult) {
	line := l.Line()

	fileRange := manip.NewFileRangeFromLineRange(lineResult.LineRange, line)

	var contextFileRange *manip.FileRange
	if lineResult.ContextLineRange != nil {
		contextFileRange = manip.NewFileRangeFromLineRange(lineResult.ContextLineRange, line)
	}

	l.job.SubmitResult(&contract.Result{
		FileRange:        fileRange,
		ContextFileRange: contextFileRange,
		SecretValue:      lineResult.SecretValue,
		SecretExtras:     lineResult.SecretExtras,
		FindingExtras:    lineResult.FindingExtras,
	})
}

func (l *LineProcessorJob) FileChange() (fileChange *gitpkg.FileChange) {
	return l.job.FileChange()
}

func (l *LineProcessorJob) Diff() (result *gitpkg.Diff) {
	return l.job.Diff()
}

func (l *LineProcessorJob) Line() int {
	return l.job.Line()
}

func (l *LineProcessorJob) Log(prefixLog logg.Logg) (result logg.Logg) {
	return l.job.Log(prefixLog)
}

//
// LineProcessorWrapper

type LineProcessorWrapper struct {
	proc contract.LineProcessorI
	log  logg.Logg
}

func NewLineProcessorWrapper(proc contract.LineProcessorI, log logg.Logg) *LineProcessorWrapper {
	return &LineProcessorWrapper{
		proc: proc,
		log:  log,
	}
}

func (l *LineProcessorWrapper) GetName() string {
	return l.proc.GetName()
}

func (l *LineProcessorWrapper) FindResultsInFileChange(job contract.ProcessorJobI) (err error) {
	job.SearchingWithProcessor(l.proc)
	fileChange := job.FileChange()

	// Find in each chunk
	chunks := fileChange.Chunks

	currFileLineNum := 1
	currDiffLineNum := 1

	// Decorate the job in a struct that can convert line ranges to file ranges
	lineJob := NewLineProcessorJob(job)

	for _, chunk := range chunks {
		chunkString := chunk.Content

		// Remove the trailing line break
		chunkLen := len(chunkString)
		if chunkLen > 0 && chunkString[chunkLen-1:] == "\n" {
			chunkString = chunkString[:chunkLen-1]
		}

		switch chunk.Operation {
		case gitpkg.Equal:
			lineCount := manip.CountRunes(chunkString, '\n') + 1
			currFileLineNum += lineCount
			currDiffLineNum += lineCount
		case gitpkg.Delete:
			lineCount := manip.CountRunes(chunkString, '\n') + 1
			currDiffLineNum += lineCount
		case gitpkg.Add:

			// For each line in chunk
			lines := strings.Split(chunkString, "\n")
			for _, line := range lines {
				if line == "" {
					currFileLineNum += 1
					continue
				}

				job.SearchingLine(currFileLineNum)
				dev.BreakpointInProcessor(fileChange.Path, l.proc.GetName(), currFileLineNum)

				// Run inner processor
				err = l.proc.FindResultsInLine(lineJob, line)

				if err != nil {
					err = errors.WithMessagev(err, "unable to find in line using processor", l.proc.GetName())
					return
				}

				// Advance to the next line
				currFileLineNum += 1
				currDiffLineNum += 1
			}
		}
	}

	return
}
