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
    "time"
)

type (
    search struct {
        name                string
        payload             *searchParameters
        processors          []Processor
        whitelistPath       structures.RegexpSet
        fileChangeFilter    *gitpkg.FileChangeFilter
        commitSearchTimeout time.Duration
        bar                 *progress.Bar
        out                 chan *searchResult
        db                  *database.Database
        log                 logrus.FieldLogger
    }
    searchResult struct {
        RepoID         string
        Commit         *gitpkg.Commit
        FindingResults []*searchFindingResult
        Err            error
    }
    searchFindingResult struct {
        FileChange *gitpkg.FileChange
        Findings   []*ProcFinding
    }

    // Processor
    Processor interface {
        FindInFileChange(fileChange *gitpkg.FileChange, commit *gitpkg.Commit, log logrus.FieldLogger) (result []*ProcFinding, ignore []*structures.FileRange, err error)
        FindInLine(line string, log logrus.FieldLogger) (result []*ProcFindingInLine, ignore []*structures.LineRange, err error)
        Name() string
    }
    ProcFinding struct {
        ProcessorName string
        FileRange     *structures.FileRange
        Secret        *ProcSecret
        SecretExtras  []*ProcExtra
        FindingExtras []*ProcExtra
    }
    ProcFindingInLine struct {
        ProcessorName string
        LineRange     *structures.LineRange
        Secret        *ProcSecret
        SecretExtras  []*ProcExtra
        FindingExtras []*ProcExtra
    }
    ProcSecret struct {
        Value string
    }
    ProcExtra struct {
        Key    string
        Header string
        Value  string
        Code   bool
        URL    string
    }
)

func NewSearch(out chan *searchResult, name string, payload *searchParameters, processors []Processor, fileChangeFilter *gitpkg.FileChangeFilter, commitSearchTimeout time.Duration, bar *progress.Bar, db *database.Database, log logrus.FieldLogger) search {
    return search{
        out:                 out,
        name:                name,
        payload:             payload,
        processors:          processors,
        fileChangeFilter:    fileChangeFilter,
        commitSearchTimeout: commitSearchTimeout,
        bar:                 bar,
        db:                  db,
        log:                 log,
    }
}

func (s search) Perform() {
    defer errors.CatchPanicDo(func(err error) {s.logError(err, s.log, "error during search job") })

    s.log.Debug("start worker")

    var err error

    var commits []*gitpkg.Commit
    commits, err = s.payload.getCommits()
    if err != nil {
        errors.ErrorLogger(s.log, err).Error("error retrieving commits from payload")
        return
    }

    // Start bar
    if s.bar != nil {
        s.bar.Start()
    }

    for _, commit := range commits {
        commitLog := s.log.WithField("commit", commit.Hash)

        func() {
            if s.bar != nil {
                defer func() { s.bar.Incr() }()
            }

            var findingResults []*searchFindingResult
            if findingResults, err = s.findInCommit(commit, commitLog); err != nil {
                errors.ErrorLogger(s.log, err).Error("error while processing commit")
                return
            }

            s.out <- &searchResult{
                RepoID:         s.payload.repo.ID,
                Commit:         commit,
                FindingResults: findingResults,
            }
        }()
    }

    s.log.Debug("ending worker")
}

func (s search) findInCommitTimeout(commit *gitpkg.Commit, log logrus.FieldLogger) (result []*searchFindingResult, err error) {
    retChan := make(chan []*searchFindingResult, 1)
    errChan := make(chan error, 1)

    go func() {
        var ret []*searchFindingResult
        ret, err = s.findInCommit(commit, log)
        if err != nil {
            errChan <- err
            return
        }
        if ret == nil {
            return
        }
        retChan <- ret
    }()

    select {
    case result = <-retChan:
    case err = <-errChan:
    case <-time.After(s.commitSearchTimeout):
        err = errors.Errorf("timed out while searching commit after %s", s.commitSearchTimeout)
    }

    return
}

func (s search) findInCommit(commit *gitpkg.Commit, log logrus.FieldLogger) (result []*searchFindingResult, err error) {
    defer errors.CatchPanicDo(func(err error) {s.logError(err, log, "error during commit search") })

    log.WithField("commitDate", commit.Date.Format("2006-01-02")).Debug("searching commit")

    var fileChanges []*gitpkg.FileChange
    fileChanges, err = commit.FileChanges(s.fileChangeFilter)
    if err != nil {
        err = errors.WithMessage(err, "unable to get file changes")
        return
    }

    for _, change := range fileChanges {
        newLog := log.WithField("file", change.Path)

        var findingResult *searchFindingResult
        findingResult, err = s.findInFileChange(commit, change, newLog)
        if err != nil {
            err = errors.WithMessage(err, "unable to find in file change")
            return
        }
        if findingResult != nil {
            result = append(result, findingResult)
        }
    }

    if result == nil {
        return
    }

    // Stats
    for _, result := range result {
        for _, finding := range result.Findings {
            s.bar.SecretTracker.Add(database.CreateHashID(finding.Secret.Value))
        }
    }

    return
}

func (s search) findInFileChange(comm *gitpkg.Commit, fileChange *gitpkg.FileChange, log logrus.FieldLogger) (result *searchFindingResult, err error) {
    defer errors.CatchPanicDo(func(err error) {s.logError(err, log, "error during file change search") })

    var findings []*ProcFinding
    var ignore []*structures.FileRange

    for _, proc := range s.processors {
        log := log.WithField("processor", proc.Name())
        if err = s.findInFileChangeWithProcessor(comm, fileChange, proc, log, &findings, &ignore); err != nil {
            return
        }

        // Find in each chunk
        var chunks []gitpkg.Chunk
        chunks, err = fileChange.Chunks()
        if err != nil {
            err = errors.WithMessage(err, "unable to get chunks for file change")
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

                    if err = s.findInLineWithProcessor(line, proc, currFileLineNum, log, &findings, &ignore); err != nil {
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
        result = &searchFindingResult{
            FileChange: fileChange,
            Findings:   findings,
        }
    }

    return
}

func (s search) findInFileChangeWithProcessor(commit *gitpkg.Commit, fileChange *gitpkg.FileChange, processor Processor, log logrus.FieldLogger, findings *[]*ProcFinding, ignore *[]*structures.FileRange) (err error) {
    var fileFindings []*ProcFinding
    var ign []*structures.FileRange
    fileFindings, ign, err = processor.FindInFileChange(fileChange, commit, log)
    if err != nil {
        err = errors.WithMessagev(err, "unable to search in file change using processor", processor.Name())
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

func (s search) findInLineWithProcessor(line string, processor Processor, currFileLineNum int, log logrus.FieldLogger, findings *[]*ProcFinding, ignore *[]*structures.FileRange) (err error) {
    if dbug.Cnf.Enabled && strings.Contains(dbug.Cnf.Filter.Processor, processor.Name()) &&
        dbug.Cnf.Filter.Line > -1 && currFileLineNum == dbug.Cnf.Filter.Line {
        fmt.Print("") // For breakpoint
    }

    var lineFindings []*ProcFindingInLine
    var ign []*structures.LineRange
    lineFindings, ign, err = processor.FindInLine(line, log)
    if err != nil {
        err = errors.WithMessagev(err, "unable to find in line using processor", processor.Name())
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

func (s search) logError(err error, log logrus.FieldLogger, message string) {
    log = errors.ErrorLogger(log, err)
    if s.bar != nil {
        s.bar.BustThrough(func() { log.Error(message) })
        return
    }
    log.Error(message)
}

func NewFindingFromLineFinding(finding *ProcFindingInLine, fileLineNum int) *ProcFinding {
    return &ProcFinding{
        ProcessorName: finding.ProcessorName,
        FileRange:     structures.NewFileRangeFromLineRange(finding.LineRange, fileLineNum),
        Secret:        finding.Secret,
        SecretExtras:  finding.SecretExtras,
        FindingExtras: finding.FindingExtras,
    }
}

func shouldKeep(finding *ProcFinding, otherFindings []*ProcFinding, ignore []*structures.FileRange) (result bool) {
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
