package finder

import (
    "fmt"
    "github.com/pantheon-systems/search-secrets/pkg/database"
    "github.com/pantheon-systems/search-secrets/pkg/dbug"
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    gitpkg "github.com/pantheon-systems/search-secrets/pkg/git"
    "github.com/pantheon-systems/search-secrets/pkg/git/diff_operation"
    "github.com/pantheon-systems/search-secrets/pkg/interact/progress"
    "github.com/pantheon-systems/search-secrets/pkg/structures"
    "github.com/sirupsen/logrus"
    "strings"
)

type (
    worker struct {
        name             string
        payload          *payload
        processors       []Processor
        whitelistPath    structures.RegexpSet
        fileChangeFilter *gitpkg.FileChangeFilter
        bar              *progress.Bar
        out              chan *workerResult
        db               *database.Database
        log              *logrus.Entry
    }
    workerResult struct {
        RepoID         string
        Commit         *gitpkg.Commit
        FindingResults []*workerFindingResult
    }
    workerFindingResult struct {
        FileChange *gitpkg.FileChange
        Findings   []*Finding
    }

    // Processor
    Processor interface {
        FindInFileChange(fileChange *gitpkg.FileChange, commit *gitpkg.Commit, log *logrus.Entry) (result []*Finding, ignore []*structures.FileRange, err error)
        FindInLine(line string, log *logrus.Entry) (result []*FindingInLine, ignore []*structures.LineRange, err error)
        Name() string
    }
    Finding struct {
        ProcessorName string
        FileRange     *structures.FileRange
        Secret        *Secret
        SecretExtras  []*Extra
        FindingExtras []*Extra
    }
    FindingInLine struct {
        ProcessorName string
        LineRange     *structures.LineRange
        Secret        *Secret
        SecretExtras  []*Extra
        FindingExtras []*Extra
    }
    Secret struct {
        Value string
    }
    Extra struct {
        Key    string
        Header string
        Value  string
        Code   bool
        URL    string
    }
)

func NewWorker(name string, payload *payload, processors []Processor, fileChangeFilter *gitpkg.FileChangeFilter, bar *progress.Bar, out chan *workerResult, db *database.Database, log *logrus.Entry) worker {
    return worker{
        name:             name,
        payload:          payload,
        processors:       processors,
        fileChangeFilter: fileChangeFilter,
        bar:              bar,
        out:              out,
        db:               db,
        log:              log,
    }
}

func (w worker) Perform() {
    defer w.handlePanic(w.log)

    w.log.Debug("start worker")

    var commits []*gitpkg.Commit
    var err error
    commits, err = w.payload.getCommits()
    if err != nil {
        errors.ErrorLogForEntry(w.log, err).Error("error retrieving commits from payload")
        return
    }

    // Start bar
    if w.bar != nil {
        w.bar.Start()
    }

    for _, commit := range commits {
        func() {
            if w.bar != nil {
                defer func() { w.bar.Incr() }()
            }

            newLog := w.log.WithField("commit", commit.Hash)
            newLog.WithField("date", commit.Time.Format("2006-01-02")).Debug("searching commit")

            if err := w.findInCommit(commit, newLog); err != nil {
                errors.ErrorLogForEntry(w.log, err).Error("error while processing commit")
            }
            newLog.Debug("searched commit")
        }()
    }

    w.log.Debug("ending worker")
}

func (w worker) findInCommit(comm *gitpkg.Commit, log *logrus.Entry) (err error) {
    defer w.handlePanic(log)

    var fileChanges []*gitpkg.FileChange
    fileChanges, err = comm.FileChanges(w.fileChangeFilter)
    if err != nil {
        return
    }

    var results []*workerFindingResult
    for _, change := range fileChanges {
        newLog := log.WithField("file", change.Path)

        var result *workerFindingResult
        result, err = w.findInFileChange(comm, change, newLog)
        if err != nil {
            return
        }
        if result != nil {
            results = append(results, result)
        }
    }

    if results == nil {
        return
    }

    w.out <- &workerResult{
        RepoID:         w.payload.repo.ID,
        Commit:         comm,
        FindingResults: results,
    }

    // Stats
    for _, result := range results {
        for _, finding := range result.Findings {
            w.bar.SecretTracker.Add(database.CreateHashID(finding.Secret.Value))
        }
    }

    return
}

func (w worker) findInFileChange(comm *gitpkg.Commit, fileChange *gitpkg.FileChange, log *logrus.Entry) (result *workerFindingResult, err error) {
    defer w.handlePanic(log)

    var findings []*Finding
    var ignore []*structures.FileRange

    for _, proc := range w.processors {
        log := log.WithField("processor", proc.Name())
        if err = w.findInFileChangeWithProcessor(comm, fileChange, proc, log, &findings, &ignore); err != nil {
            return
        }

        // Find in each chunk
        var chunks []gitpkg.Chunk
        chunks, err = fileChange.Chunks()
        if err != nil {
            return
        }

        currFileLineNum := 1
        currDiffLineNum := 1
        for _, chunk := range chunks {
            chunkString := chunk.Content

            // Remove the trailing line break
            chunkLen := len(chunkString)
            if chunkLen > 0 && chunkString[chunkLen-1:] == "\n" {
                chunkString = chunkString[:chunkLen-1]
            }

            switch chunk.Type {
            case diff_operation.Equal{}.New():
                lineCount := countRunes(chunkString, '\n') + 1
                currFileLineNum += lineCount
                currDiffLineNum += lineCount
            case diff_operation.Delete{}.New():
                lineCount := countRunes(chunkString, '\n') + 1
                currDiffLineNum += lineCount
            case diff_operation.Add{}.New():

                // For each line in chunk
                lines := strings.Split(chunkString, "\n")
                for _, line := range lines {
                    if line == "" {
                        currFileLineNum += 1
                        continue
                    }

                    log := log.WithField("line", currFileLineNum)

                    if err = w.findInLineWithProcessor(line, proc, currFileLineNum, log, &findings, &ignore); err != nil {
                        return
                    }

                    // Advance to the next line
                    currFileLineNum += 1
                    currDiffLineNum += 1
                }
            }
        }
    }

    if findings != nil {
        log.Debugf("findings: %d", len(findings))
        result = &workerFindingResult{
            FileChange: fileChange,
            Findings:   findings,
        }
    }

    return
}

func (w worker) findInFileChangeWithProcessor(commit *gitpkg.Commit, fileChange *gitpkg.FileChange, processor Processor, log *logrus.Entry, findings *[]*Finding, ignore *[]*structures.FileRange) (err error) {
    var fileFindings []*Finding
    var ign []*structures.FileRange
    fileFindings, ign, err = processor.FindInFileChange(fileChange, commit, log)
    if err != nil {
        return
    }

    for _, finding := range fileFindings {
        if !shouldKeep(finding, *findings, *ignore) {
            continue
        }
        finding.ProcessorName = processor.Name()
        *findings = append(*findings, finding)
    }
    *ignore = append(*ignore, ign...)
    return
}

func (w worker) findInLineWithProcessor(line string, processor Processor, currFileLineNum int, log *logrus.Entry, findings *[]*Finding, ignore *[]*structures.FileRange) (err error) {
    if dbug.Cnf.Enabled && strings.Contains(dbug.Cnf.Filter.Processor, processor.Name()) &&
        dbug.Cnf.Filter.Line > -1 && currFileLineNum == dbug.Cnf.Filter.Line {
        fmt.Print("") // For breakpoint
    }

    var lineFindings []*FindingInLine
    var ign []*structures.LineRange
    lineFindings, ign, err = processor.FindInLine(line, log)
    if err != nil {
        return
    }

    for _, lineFinding := range lineFindings {
        finding := NewFindingFromLineFinding(lineFinding, currFileLineNum)
        if !shouldKeep(finding, *findings, *ignore) {
            continue
        }
        finding.ProcessorName = processor.Name()
        *findings = append(*findings, finding)
    }
    for _, ignRange := range ign {
        *ignore = append(*ignore, structures.NewFileRangeFromLineRange(ignRange, currFileLineNum))
    }

    return
}

func (w worker) handlePanic(log *logrus.Entry) {
    if recovered := recover(); recovered != nil {
        message := "panic during find in commit"

        if w.bar != nil {
            w.bar.BustThrough(func() {
                errors.PanicLogEntryError(log, recovered).Error(message)
            })
            return
        }

        errors.PanicLogEntryError(log, recovered).Error(message)
    }
}

func NewFindingFromLineFinding(finding *FindingInLine, fileLineNum int) *Finding {
    return &Finding{
        ProcessorName: finding.ProcessorName,
        FileRange:     structures.NewFileRangeFromLineRange(finding.LineRange, fileLineNum),
        Secret:        finding.Secret,
        SecretExtras:  finding.SecretExtras,
        FindingExtras: finding.FindingExtras,
    }
}

func shouldKeep(finding *Finding, otherFindings []*Finding, ignore []*structures.FileRange) (result bool) {
    for _, ignoreRange := range ignore {
        if ignoreRange.Overlaps(finding.FileRange) {
            return false
        }
    }
    for _, otherFinding := range otherFindings {
        if otherFinding.FileRange.Overlaps(finding.FileRange) {
            return false
        }
    }

    return true
}

func countRunes(input string, r rune) (result int) {
    for _, c := range input {
        if c == r {
            result++
        }
    }
    return
}
