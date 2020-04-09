package finder

import (
    diffpkg "github.com/pantheon-systems/search-secrets/pkg/diff"
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    rulepkg "github.com/pantheon-systems/search-secrets/pkg/finder/rule"
    "github.com/pantheon-systems/search-secrets/pkg/progress"
    "github.com/pantheon-systems/search-secrets/pkg/structures"
    "github.com/sirupsen/logrus"
    "gopkg.in/src-d/go-git.v4"
    gitplumbing "gopkg.in/src-d/go-git.v4/plumbing"
    gitdiff "gopkg.in/src-d/go-git.v4/plumbing/format/diff"
    gitobject "gopkg.in/src-d/go-git.v4/plumbing/object"
    gitstorer "gopkg.in/src-d/go-git.v4/plumbing/storer"
    "strings"
    "time"
)

var checker = structures.NewSet(nil)

const (
    commitProcessingTimeout = 30
)

type (
    Comb struct {
        log *logrus.Logger
    }
    searchState struct {
        repoID         string
        repoName       string
        refs           []string
        gitRepo        *git.Repository
        rules          []rulepkg.Rule
        out            chan *DriverResult
        searched       structures.Set
        earliestTime   time.Time
        latestTime     time.Time
        earliestCommit string
        latestCommit   string
        whitelistPath  structures.RegexpSet
        bar            *progress.Bar
    }
)

func NewComb(log *logrus.Logger) *Comb {
    return &Comb{
        log: log,
    }
}

func (c *Comb) Find(repoID, repoName, cloneDir string, refs []string, rules []rulepkg.Rule, earliestTime, latestTime time.Time, earliestCommit, latestCommit string, whitelistPath structures.RegexpSet, bar *progress.Bar, out chan *DriverResult) {
    var err error

    log := c.log.WithField("repo", repoName)
    defer func() {
    }()
    // Send an error through the channel
    defer func() {
        if panicErr := recover(); panicErr != nil {
            err = errors.Errorv("find method panic", panicErr)
            errors.LogEntryError(log, err)
            return
        }
        if err != nil {
            err = errors.WithMessage(err, "find method exiting with an error")
            errors.LogEntryError(log, err)
            out <- &DriverResult{Err: err}
        }
    }()

    var gitRepo *git.Repository
    gitRepo, err = git.PlainOpen(cloneDir)
    if err != nil {
        err = errors.Wrapv(err, "unable to clone into directory", cloneDir)
        return
    }

    var search = &searchState{
        repoID:         repoID,
        repoName:       repoName,
        refs:           refs,
        rules:          rules,
        gitRepo:        gitRepo,
        searched:       structures.NewSet(nil),
        earliestTime:   earliestTime,
        latestTime:     latestTime,
        earliestCommit: earliestCommit,
        latestCommit:   latestCommit,
        whitelistPath:  whitelistPath,
        out:            out,
        bar:            bar,
    }

    var hashes []gitplumbing.Hash
    if latestCommit != "" {
        hashes = []gitplumbing.Hash{gitplumbing.NewHash(latestCommit)}
    } else {
        var branchIter gitstorer.ReferenceIter
        branchIter, err = gitRepo.Branches()
        if err != nil {
            err = errors.Wrapv(err, "unable to get branches")
            return
        }

        var branches []*gitplumbing.Reference
        err = branchIter.ForEach(func(branch *gitplumbing.Reference) (err error) {
            branches = append(branches, branch)
            return
        })

        hashes = []gitplumbing.Hash{}
        for _, branch := range branches {
            hashes = append(hashes, branch.Hash())
        }
    }

    var commits []*gitobject.Commit
    for _, hash := range hashes {
        err = c.appendCommitsFromCommit(search, hash, &commits)
        if err != nil {
            err = errors.Wrapv(err, "unable to find in ancestor commits of commit", hash)
            return
        }
    }

    hashesIndex := structures.NewSet(nil)
    for _, c := range commits {
        hashesIndex.Add(c.Hash.String())
    }
    dupeCount := len(commits) - len(hashesIndex.Values())
    if dupeCount > 0 {
        c.log.Warnf("%d duplicate commits detected", dupeCount)
    }

    bar.Start(len(commits))
    for _, commit := range commits {
        newLog := log.WithFields(logrus.Fields{"commit": commit.Hash.String()})
        timer := time.NewTimer(commitProcessingTimeout * time.Second)

        errs := make(chan error, 1)
        go func() { errs <- c.findInCommit(search, commit, newLog) }()

        select {
        case err = <-errs:
            if err != nil {
                errors.LogEntryError(newLog, errors.WithMessage(err, "error while processing commit"))
            }
        case <-timer.C:
            errors.LogEntryError(newLog, errors.WithMessagev(err, "timeout while processing commit (secs)", commitProcessingTimeout))
        }

        timer.Stop()
        bar.Incr()
    }

    return
}

func (c *Comb) findInCommit(search *searchState, commit *gitobject.Commit, log *logrus.Entry) (err error) {
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
        driverFileChange, err = c.findInFileChange(search, commit, change, newLog)
        if err != nil {
            return
        }
        if driverFileChange != nil {
            fileChanges = append(fileChanges, driverFileChange)
        }
    }

    if fileChanges != nil {
        search.out <- &DriverResult{Commit: &DriverCommit{
            RepoID:      search.repoID,
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

func (c *Comb) findInFileChange(search *searchState, commit *gitobject.Commit, fileChange *gitobject.Change, log *logrus.Entry) (result *DriverFileChange, err error) {
    // Deleted file?
    if fileChange.To.Name == "" {
        log.Trace("file deletion skipped")
        return
    }

    if search.whitelistPath.MatchStringAny(fileChange.To.Name, "") {
        log.Debug("file whitelisted by path and skipped")
        return
    }

    context := rulepkg.NewFileChangeContext(search.repoName, commit, fileChange, log)

    var isBinary bool
    isBinary, err = context.IsBinaryOrEmpty()
    if err != nil {
        return
    }
    if isBinary {
        log.Trace("empty or binary file skipped")
        return
    }

    var hasCodeChanges bool
    hasCodeChanges, err = context.HasCodeChanges()
    if err != nil || ! hasCodeChanges {
        return
    }

    var findings []*DriverFinding
    for _, rule := range search.rules {
        var fileChangeFindings []*rulepkg.FileChangeFinding
        fileChangeFindings, err = rule.Processor.FindInFileChange(context, log)
        if err != nil {
            return
        }

        for _, fileChangeFinding := range fileChangeFindings {
            findings = append(findings, &DriverFinding{
                RuleName:     rule.Name,
                FileRange:    fileChangeFinding.FileRange,
                SecretValues: fileChangeFinding.SecretValues,
            })
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
        ff, err = c.findInChunk(search, chunk, &currentFileLineNumber, &currentDiffLineNumber, log)
        if err != nil {
            return
        }

        findings = append(findings, ff...)
    }

    // Remove overlapping findings
    var findingsNew []*DriverFinding
    for _, finding := range findings {
        if overlapsWithAny(finding, findingsNew) {
            continue
        }

        findingsNew = append(findingsNew, finding)
    }
    findings = findingsNew

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

func (c *Comb) findInChunk(search *searchState, chunk gitdiff.Chunk, currentFileLineNumber, currentDiffLineNumber *int, log *logrus.Entry) (result []*DriverFinding, err error) {
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

            for _, rule := range search.rules {
                var ff []*DriverFinding
                newLog := log.WithField("rule", rule.Name)
                ff, err = c.evaluateLineWithRule(currentFileLineNumber, currentDiffLineNumber, line, rule, newLog)
                if err != nil {
                    return
                }

                result = append(result, ff...)
            }

            // Advance to the next line
            *currentFileLineNumber += 1
            *currentDiffLineNumber += 1
        }
    }

    return
}

func (c *Comb) appendCommitsFromCommit(search *searchState, hash gitplumbing.Hash, commits *[]*gitobject.Commit) (err error) {
    var history gitobject.CommitIter
    history, err = search.gitRepo.Log(&git.LogOptions{From: hash, Order: git.LogOrderCommitterTime})
    if err != nil {
        return
    }

    latestCommitReached := false

    err = history.ForEach(func(commit *gitobject.Commit) (err error) {
        commitTime := commit.Committer.When
        if commitTime.After(search.latestTime) {
            return
        }
        if commitTime.Before(search.earliestTime) {
            return gitstorer.ErrStop
        }

        if search.latestCommit != "" {
            if commit.Hash.String() == search.latestCommit {
                latestCommitReached = true
            }
            if ! latestCommitReached {
                return
            }
        }

        if search.earliestCommit != "" && search.searched.Contains(search.earliestCommit) {
            return gitstorer.ErrStop
        }
        if search.searched.Contains(commit.Hash.String()) {
            return
        }
        search.searched.Add(commit.Hash.String())

        *commits = append(*commits, commit)

        return
    })

    return
}

func (c *Comb) evaluateLineWithRule(currentFileLineNumber, currentDiffLineNumber *int, line string, rule rulepkg.Rule, log *logrus.Entry) (result []*DriverFinding, err error) {
    var lineFindings []*rulepkg.LineFinding
    lineFindings, err = rule.Processor.FindInLine(line, log)
    if err != nil {
        return
    }

    for _, lineFinding := range lineFindings {
        result = append(result, &DriverFinding{
            RuleName: rule.Name,
            FileRange: &structures.FileRange{
                StartLineNum:     *currentFileLineNumber,
                StartIndex:       lineFinding.LineRange.StartIndex,
                EndLineNum:       *currentFileLineNumber,
                EndIndex:         lineFinding.LineRange.EndIndex,
                StartDiffLineNum: *currentDiffLineNumber,
                EndDiffLineNum:   *currentDiffLineNumber,
            },
            SecretValues: lineFinding.SecretValues,
        })
    }
    return
}

func overlapsWithAny(input *DriverFinding, others []*DriverFinding) bool {
    for _, other := range others {
        if other.FileRange.Overlaps(input.FileRange) {
            return true
        }
    }
    return false
}

func countRunes(input string, r rune) (result int) {
    for _, c := range input {
        if c == r {
            result++
        }
    }
    return
}
