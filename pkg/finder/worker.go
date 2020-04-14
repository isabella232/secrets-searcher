package finder

import (
    "fmt"
    "github.com/pantheon-systems/search-secrets/pkg/database"
    "github.com/pantheon-systems/search-secrets/pkg/dev"
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    gitpkg "github.com/pantheon-systems/search-secrets/pkg/git"
    "github.com/pantheon-systems/search-secrets/pkg/interact/progress"
    "github.com/pantheon-systems/search-secrets/pkg/structures"
    "github.com/sirupsen/logrus"
    "github.com/vbauerster/mpb/v5"
    "github.com/vbauerster/mpb/v5/decor"
    "io"
    "strings"
    "time"
)

type worker struct {
    repo             *database.Repo
    processors       []Processor
    whitelistPath    structures.RegexpSet
    commitFilter     *gitpkg.CommitFilter
    fileChangeFilter *gitpkg.FileChangeFilter
    prog             *progress.Progress
    out              chan *result
    db               *database.Database
    log              *logrus.Entry
}

func NewWorker(repo *database.Repo, processors []Processor, commitFilter *gitpkg.CommitFilter, fileChangeFilter *gitpkg.FileChangeFilter, prog *progress.Progress, out chan *result, db *database.Database, log *logrus.Entry) worker {
    return worker{
        repo:             repo,
        processors:       processors,
        commitFilter:     commitFilter,
        fileChangeFilter: fileChangeFilter,
        prog:             prog,
        out:              out,
        db:               db,
        log:              log,
    }
}

func (w worker) Perform() {
    defer w.handlePanic(w.log)

    if dev.Enabled && dev.Commit != "" {
        w.commitFilter.LatestCommit = dev.Commit
        w.commitFilter.EarliestCommit = dev.Commit
    }

    w.log.Trace("worker started")

    if err := w.perform(); err != nil {
        errors.ErrorLogForEntry(w.log, err).Error("unable to perform search of repo")
    }

    w.log.Trace("worker finished")
}

func (w worker) perform() (err error) {
    git := gitpkg.New(w.log)

    var gitRepo *gitpkg.Repository
    gitRepo, err = git.NewRepository(w.repo.CloneDir)
    if err != nil {
        err = errors.Wrapv(err, "unable to open git repository", w.repo.CloneDir)
        return
    }

    var commits []*gitpkg.Commit
    commits, err = gitRepo.Log(w.commitFilter)
    if err != nil {
        err = errors.Wrap(err, "unable to find in ancestor commits of commit")
        return
    }

    commitTotal := len(commits)
    if commitTotal == 0 {
        err = errors.New("no commits found in repo")
        return
    }
    w.log.Debugf("%d commits found for repo", commitTotal)

    var bar *progress.Bar
    if w.prog != nil {
        bar = w.prog.AddBar(w.repo.Name, commitTotal)
    }

    hashesIndex := structures.NewSet(nil)
    for _, c := range commits {
        hashesIndex.Add(c.Hash)
    }
    dupeCount := commitTotal - len(hashesIndex.Values())
    if dupeCount > 0 {
        w.log.Warnf("%d duplicate commits detected", dupeCount)
    }

    for _, commit := range commits {
        func() {
            start := time.Now()
            defer func() {
                if bar != nil {
                    bar.Incr()
                    bar.DecoratorEwmaUpdate(time.Since(start))
                }
            }()

            newLog := w.log.WithFields(logrus.Fields{"commit": commit.Hash})

            if err := w.findInCommit(commit, newLog); err != nil {
                errors.ErrorLogForEntry(w.log, err).Error("error while processing commit")
            }
        }()
    }

    if w.prog != nil {
        w.prog.Add(0, mpb.BarFillerFunc(func(writer io.Writer, width int, st *decor.Statistics) {
            fmt.Fprintf(writer, "- search of %s is complete", w.repo.Name)
        })).SetTotal(0, true)
    }

    return
}

func (w worker) findInCommit(comm *gitpkg.Commit, log *logrus.Entry) (err error) {
    defer w.handlePanic(log)

    log.WithField("date", comm.Time.Format("2006-01-02")).Debug("searching commit")

    var fileChanges []*gitpkg.FileChange
    fileChanges, err = comm.FileChanges(w.fileChangeFilter)
    if err != nil {
        return
    }

    var results []*findingResult
    for _, change := range fileChanges {
        newLog := log.WithField("file", change.Path)

        var result *findingResult
        result, err = w.findInFileChange(change, newLog)
        if err != nil {
            return
        }
        if result != nil {
            results = append(results, result)
        }
    }

    if results != nil {
        w.out <- &result{
            RepoID:         w.repo.ID,
            Commit:         comm,
            FindingResults: results,
        }
    }

    return
}

func (w worker) findInFileChange(fileChange *gitpkg.FileChange, log *logrus.Entry) (result *findingResult, err error) {
    defer w.handlePanic(log)

    if dev.Enabled && fileChange.Path != dev.Path {
        return
    }
    var findings []*Finding
    var ignore []*structures.FileRange

    for _, proc := range w.processors {
        if err = w.findInFileChangeWithProcessor(fileChange, proc, log, &findings, &ignore); err != nil {
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
            case gitpkg.Equal{}.New():
                lineCount := countRunes(chunkString, '\n') + 1
                currFileLineNum += lineCount
                currDiffLineNum += lineCount
            case gitpkg.Delete{}.New():
                lineCount := countRunes(chunkString, '\n') + 1
                currDiffLineNum += lineCount
            case gitpkg.Add{}.New():

                // For each line in chunk
                lines := strings.Split(chunkString, "\n")
                for _, line := range lines {
                    if line == "" {
                        currFileLineNum += 1
                        continue
                    }

                    log := log.WithField("processor", proc.Name()).WithField("line", currFileLineNum)

                    if dev.Enabled && strings.Contains(dev.Processor, proc.Name()) &&
                        ((dev.DiffLine > 0 && currDiffLineNum == dev.DiffLine) ||
                            (dev.LineContains != "" && strings.Contains(line, dev.LineContains))) {
                        fmt.Print("")
                    }

                    if err = w.findInLineWithProcessor(line, proc, currFileLineNum, currDiffLineNum, log, &findings, &ignore); err != nil {
                        return
                    }

                    // Advance to the next line
                    currFileLineNum += 1
                    currDiffLineNum += 1
                }
            }
        }
    }

    // Remove overlapping and ignored findings
    findings = removeIgnored(findings, ignore)
    findings = removeOverlapping(findings)

    if findings != nil {
        result = &findingResult{
            FileChange: fileChange,
            Findings:   findings,
        }
    }

    return
}

func (w worker) findInFileChangeWithProcessor(fileChange *gitpkg.FileChange, processor Processor, log *logrus.Entry, findings *[]*Finding, ignore *[]*structures.FileRange) (err error) {
    var fileChangeFindings []*Finding
    var ign []*structures.FileRange
    fileChangeFindings, ign, err = processor.FindInFileChange(fileChange, log)
    if err != nil {
        return
    }
    *findings = append(*findings, fileChangeFindings...)
    *ignore = append(*ignore, ign...)
    return
}

func (w worker) findInLineWithProcessor(line string, processor Processor, currFileLineNum, currDiffLineNum int, log *logrus.Entry, findings *[]*Finding, ignore *[]*structures.FileRange) (err error) {
    var lineFindings []*FindingInLine
    var ign []*structures.LineRange
    lineFindings, ign, err = processor.FindInLine(line, log)
    if err != nil {
        return
    }
    for _, lineFinding := range lineFindings {
        *findings = append(*findings, NewFindingFromLineFinding(lineFinding, currFileLineNum, currDiffLineNum))
    }
    for _, ignRange := range ign {
        *ignore = append(*ignore, structures.NewFileRangeFromLineRange(ignRange, currFileLineNum, currDiffLineNum))
    }
    return
}

func (w worker) handlePanic(log *logrus.Entry) {
    if recovered := recover(); recovered != nil {
        message := "panic during find in commit"

        if w.prog != nil {
            w.prog.BustThrough(func() {
                errors.PanicLogEntryError(log, recovered).Error(message)
            })
            return
        }

        errors.PanicLogEntryError(log, recovered).Error(message)
    }
}

func removeOverlapping(findings []*Finding) (result []*Finding) {
    for _, finding := range findings {
        keep := true
        for _, resultFinding := range result {
            if resultFinding.FileRange.Overlaps(finding.FileRange) {
                keep = false
                break
            }
        }

        if !keep {
            continue
        }

        result = append(result, finding)
    }

    return
}

func removeIgnored(findings []*Finding, ignore []*structures.FileRange) (result []*Finding) {
    for _, finding := range findings {
        keep := true
        for _, ignoreRange := range ignore {
            if ignoreRange.Overlaps(finding.FileRange) {
                keep = false
                break
            }
        }

        if !keep {
            continue
        }

        result = append(result, finding)
    }

    return
}

func countRunes(input string, r rune) (result int) {
    for _, c := range input {
        if c == r {
            result++
        }
    }
    return
}
