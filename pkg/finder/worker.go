package finder

import (
    "fmt"
    "github.com/pantheon-systems/search-secrets/pkg/database"
    "github.com/pantheon-systems/search-secrets/pkg/dev"
    diffpkg "github.com/pantheon-systems/search-secrets/pkg/diff"
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    rulepkg "github.com/pantheon-systems/search-secrets/pkg/finder/rule"
    "github.com/pantheon-systems/search-secrets/pkg/interact/progress"
    "github.com/pantheon-systems/search-secrets/pkg/structures"
    "github.com/sirupsen/logrus"
    "github.com/vbauerster/mpb/v5"
    "github.com/vbauerster/mpb/v5/decor"
    "gopkg.in/src-d/go-git.v4"
    gitplumbing "gopkg.in/src-d/go-git.v4/plumbing"
    gitdiff "gopkg.in/src-d/go-git.v4/plumbing/format/diff"
    gitobject "gopkg.in/src-d/go-git.v4/plumbing/object"
    gitstorer "gopkg.in/src-d/go-git.v4/plumbing/storer"
    "io"
    "strings"
    "time"
)

type worker struct {
    repo           *database.Repo
    cloneDir       string
    refs           []string
    rules          []rulepkg.Rule
    earliestTime   time.Time
    latestTime     time.Time
    earliestCommit string
    latestCommit   string
    whitelistPath  structures.RegexpSet
    prog           *progress.Progress
    out            chan *DriverResult
    db             *database.Database
    searched       structures.Set
    progressing    bool
    log            *logrus.Entry
}

func (w worker) Perform() {
    defer w.handlePanic(w.log)

    if dev.Enabled {
        w.latestCommit = dev.Commit
        w.earliestCommit = dev.Commit
    }

    w.log.Trace("worker started")

    if err := w.perform(); err != nil {
        errors.ErrorLogForEntry(w.log, err).Error("unable to perform search of repo")
    }

    w.log.Trace("worker finished")
}

func (w worker) perform() (err error) {
    var gitRepo *git.Repository
    gitRepo, err = git.PlainOpen(w.cloneDir)
    if err != nil {
        err = errors.Wrapv(err, "unable to open directory", w.cloneDir)
        return
    }

    logOptions := &git.LogOptions{Order: git.LogOrderCommitterTime}
    if w.latestCommit != "" {
        logOptions.From = gitplumbing.NewHash(w.latestCommit)
    } else {
        logOptions.All = true
    }

    var fromCommits gitobject.CommitIter
    fromCommits, err = gitRepo.Log(logOptions)
    if err != nil {
        err = errors.Wrap(err, "unable to get git log")
        return
    }

    var commits []*gitobject.Commit
    err = w.appendCommitsFromCommit(fromCommits, &commits)
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

    // BAR
    var bar *progress.Bar
    if w.prog != nil {
        bar = w.prog.AddBar(w.repo.Name, commitTotal)
        w.progressing = true
    }

    hashesIndex := structures.NewSet(nil)
    for _, c := range commits {
        hashesIndex.Add(c.Hash.String())
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

            newLog := w.log.WithFields(logrus.Fields{"commit": commit.Hash.String()})

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

func (w worker) findInCommit(commit *gitobject.Commit, log *logrus.Entry) (err error) {
    defer w.handlePanic(log)

    log.WithField("date", commit.Committer.When.Format("2006-01-02")).Debug("searching commit")

    if len(commit.ParentHashes) == 0 {
        log.Debug("No parent commits found")
        return
    }
    if len(commit.ParentHashes) > 1 {
        log.Debug("skipping merge commit with multiple parents")
        return
    }

    var parentCommit *gitobject.Commit
    parentCommit, err = commit.Parents().Next()
    if err != nil {
        return
    }

    var commitTree *gitobject.Tree
    commitTree, err = commit.Tree()
    if err != nil {
        return
    }

    var parentCommitTree *gitobject.Tree
    parentCommitTree, err = parentCommit.Tree()
    if err != nil {
        return
    }

    var changes gitobject.Changes
    changes, err = parentCommitTree.Diff(commitTree)
    if err != nil {
        return
    }

    var fileChanges []*DriverFileChange
    for _, change := range changes {
        var driverFileChange *DriverFileChange
        newLog := log.WithField("file", change.To.Name)
        driverFileChange, err = w.findInFileChange(commit, change, newLog)
        if err != nil {
            return
        }
        if driverFileChange != nil {
            fileChanges = append(fileChanges, driverFileChange)
        }
    }

    if fileChanges != nil {
        w.out <- &DriverResult{Commit: &DriverCommit{
            RepoID:      w.repo.ID,
            RepoName:    w.repo.Name,
            Commit:      commit.Message,
            CommitHash:  commit.Hash.String(),
            Date:        commit.Committer.When,
            AuthorEmail: commit.Author.Email,
            AuthorFull:  commit.Author.String(),
            FileChanges: fileChanges,
        }}
    }

    return
}

func (w worker) findInFileChange(commit *gitobject.Commit, fileChange *gitobject.Change, log *logrus.Entry) (result *DriverFileChange, err error) {
    defer w.handlePanic(log)

    // Deleted file?
    if fileChange.To.Name == "" {
        log.Trace("file deletion skipped")
        return
    }

    if w.whitelistPath.MatchAny(fileChange.To.Name) {
        log.Debug("file whitelisted by path and skipped")
        return
    }

    if dev.Enabled && fileChange.To.Name != dev.Path {
        return
    }

    context := rulepkg.NewFileChangeContext(w.repo.Name, commit, fileChange, log)

    var isBinary bool
    isBinary, err = context.IsBinaryOrEmpty()
    if err != nil {
        return
    }
    if isBinary {
        log.Debug("empty or binary file skipped")
        return
    }

    var hasCodeChanges bool
    hasCodeChanges, err = context.HasCodeChanges()
    if err != nil || !hasCodeChanges {
        return
    }

    var findings []*DriverFinding
    var ignore []*structures.FileRange

    for _, rule := range w.rules {
        var fileChangeFindings []*rulepkg.FileChangeFinding
        var ign []*structures.FileRange
        fileChangeFindings, ign, err = rule.Processor.FindInFileChange(context, log)
        if err != nil {
            return
        }

        if fileChangeFindings != nil {
            for _, fileChangeFinding := range fileChangeFindings {

                var secrets []*DriverSecret
                for _, fileChangeSecret := range fileChangeFinding.Secrets {
                    secrets = append(secrets, &DriverSecret{
                        Value:   fileChangeSecret.Value,
                        Decoded: fileChangeSecret.Decoded,
                    })
                }

                findings = append(findings, &DriverFinding{
                    RuleName:  rule.Name,
                    FileRange: fileChangeFinding.FileRange,
                    Secrets:   secrets,
                })
            }
        }
        if ign != nil {
            ignore = append(ignore, ign...)
        }
    }

    // Find in each chunk
    var chunks []gitdiff.Chunk
    chunks, err = context.Chunks()
    if err != nil {
        return
    }

    currentFileLineNumber := 1
    currentDiffLineNumber := 1
    for _, chunk := range chunks {
        var ff []*DriverFinding
        var ign []*structures.FileRange
        ff, ign, err = w.findInChunk(chunk, &currentFileLineNumber, &currentDiffLineNumber, log)
        if err != nil {
            return
        }

        if ff != nil {
            findings = append(findings, ff...)
        }
        if ign != nil {
            ignore = append(ignore, ign...)
        }
    }

    // Remove overlapping and ignored findings
    findings = removeIgnored(findings, ignore)
    findings = removeOverlapping(findings)

    if findings != nil {
        var file *gitobject.File
        file, err = commit.File(fileChange.To.Name)
        if err != nil {
            return
        }

        var fileContents string
        fileContents, err = file.Contents()
        if err != nil {
            return
        }

        var diff *diffpkg.Diff
        diff, err = context.Diff()
        if err != nil {
            return
        }

        result = &DriverFileChange{
            Path:         fileChange.To.Name,
            FileContents: fileContents,
            Diff:         diff.String(),
            Findings:     findings,
        }
    }

    return
}

func (w worker) findInChunk(chunk gitdiff.Chunk, currentFileLineNumber, currentDiffLineNumber *int, log *logrus.Entry) (result []*DriverFinding, ignore []*structures.FileRange, err error) {
    chunkString := chunk.Content()

    // Remove the trailing line break
    chunkLen := len(chunkString)
    if chunkLen > 0 && chunkString[chunkLen-1:] == "\n" {
        chunkString = chunkString[:chunkLen-1]
    }

    switch chunk.Type() {
    case gitdiff.Delete:
        lineCount := countRunes(chunkString, '\n') + 1
        *currentDiffLineNumber += lineCount
    case gitdiff.Equal:
        lineCount := countRunes(chunkString, '\n') + 1
        *currentFileLineNumber += lineCount
        *currentDiffLineNumber += lineCount
    case gitdiff.Add:

        // For each line in chunk
        lines := strings.Split(chunkString, "\n")
        for _, line := range lines {
            if line == "" {
                *currentFileLineNumber += 1
                continue
            }

            for _, rule := range w.rules {
                if dev.Enabled && strings.Contains(dev.Rule, rule.Name) {
                    fmt.Print("")
                }

                var ff []*DriverFinding
                var ign []*structures.FileRange
                newLog := log.WithField("rule", rule.Name).WithField("line", *currentFileLineNumber)
                ff, ign, err = w.evaluateLineWithRule(currentFileLineNumber, currentDiffLineNumber, line, rule, newLog)
                if err != nil {
                    return
                }

                if ff != nil {
                    result = append(result, ff...)
                }
                if ign != nil {
                    ignore = append(ignore, ign...)
                }
            }

            // Advance to the next line
            *currentFileLineNumber += 1
            *currentDiffLineNumber += 1
        }
    }

    return
}

func (w worker) appendCommitsFromCommit(fromCommits gitobject.CommitIter, commits *[]*gitobject.Commit) (err error) {
    latestCommitReached := false

    err = fromCommits.ForEach(func(commit *gitobject.Commit) (err error) {
        commitTime := commit.Committer.When
        if commitTime.After(w.latestTime) {
            return
        }
        if commitTime.Before(w.earliestTime) {
            return gitstorer.ErrStop
        }

        if w.latestCommit != "" {
            if commit.Hash.String() == w.latestCommit {
                latestCommitReached = true
            }
            if !latestCommitReached {
                return
            }
        }

        if w.earliestCommit != "" && w.searched.Contains(w.earliestCommit) {
            return gitstorer.ErrStop
        }
        if w.searched.Contains(commit.Hash.String()) {
            return
        }
        w.searched.Add(commit.Hash.String())

        *commits = append(*commits, commit)

        return
    })

    return
}

func (w worker) evaluateLineWithRule(currentFileLineNumber, currentDiffLineNumber *int, line string, rule rulepkg.Rule, log *logrus.Entry) (result []*DriverFinding, ignore []*structures.FileRange, err error) {
    var lineFindings []*rulepkg.LineFinding
    var ign []*structures.LineRange

    if dev.Enabled && strings.Contains(dev.Rule, rule.Name) {
        fmt.Print("")
    }

    if dev.Enabled && strings.Contains(dev.Rule, rule.Name) &&
        ((dev.DiffLine > 0 && *currentDiffLineNumber == dev.DiffLine) ||
            (dev.LineContains != "" && strings.Contains(line, dev.LineContains))) {
        fmt.Print("")
    }

    lineFindings, ign, err = rule.Processor.FindInLine(line, log)
    if err != nil {
        return
    }

    if lineFindings != nil {
        for _, lineFinding := range lineFindings {
            fileRange := &structures.FileRange{
                StartLineNum:     *currentFileLineNumber,
                StartIndex:       lineFinding.LineRange.StartIndex,
                EndLineNum:       *currentFileLineNumber,
                EndIndex:         lineFinding.LineRange.EndIndex,
                StartDiffLineNum: *currentDiffLineNumber,
                EndDiffLineNum:   *currentDiffLineNumber,
            }

            if len(lineFinding.Secrets) == 0 {
                ignore = append(ignore, fileRange)
                continue
            }

            var secrets []*DriverSecret
            for _, fileChangeSecret := range lineFinding.Secrets {
                secrets = append(secrets, &DriverSecret{
                    Value:   fileChangeSecret.Value,
                    Decoded: fileChangeSecret.Decoded,
                })
            }

            result = append(result, &DriverFinding{
                RuleName:  rule.Name,
                FileRange: fileRange,
                Secrets:   secrets,
            })
        }
    }

    if ign != nil {
        for _, ignRange := range ign {
            ignore = append(ignore, &structures.FileRange{
                StartLineNum:     *currentFileLineNumber,
                StartIndex:       ignRange.StartIndex,
                EndLineNum:       *currentFileLineNumber,
                EndIndex:         ignRange.EndIndex,
                StartDiffLineNum: *currentDiffLineNumber,
                EndDiffLineNum:   *currentDiffLineNumber,
            })
        }
    }

    return
}

func (w worker) handlePanic(log *logrus.Entry) {
    if recovered := recover(); recovered != nil {
        message := "panic during find in commit"

        if w.prog != nil {
            w.prog.BustThrough(func() { errors.PanicLogEntryError(log, recovered).Error(message) })
        } else {
            errors.PanicLogEntryError(log, recovered).Error(message)
        }
    }
}

func removeOverlapping(findings []*DriverFinding) (result []*DriverFinding) {
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

func removeIgnored(findings []*DriverFinding, ignore []*structures.FileRange) (result []*DriverFinding) {
    for _, finding := range findings {
        keep := true
        for _, ignoreRange := range ignore {
            if ignoreRange.StartLineNum == 260 && finding.FileRange.StartLineNum == 260 {
                fmt.Print("")
            }
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
