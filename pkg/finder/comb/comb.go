package comb

import (
    "bytes"
    "github.com/pantheon-systems/search-secrets/pkg/finder"
    rulepkg "github.com/pantheon-systems/search-secrets/pkg/rule"
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

type (
    Comb struct {
        log *logrus.Logger
    }
    searchState struct {
        repoID         string
        repoName       string
        refs           []string
        gitRepo        *git.Repository
        rules          []*rulepkg.Rule
        out            chan *finder.DriverResult
        searched       structures.Set
        earliestTime   time.Time
        latestTime     time.Time
        earliestCommit string
        latestCommit   string
        whitelistPath  structures.RegexpSet
    }
)

func New(log *logrus.Logger) *Comb {
    return &Comb{
        log: log,
    }
}

func (c *Comb) Find(repoID, repoName, cloneDir string, refs []string, rules []*rulepkg.Rule, earliestTime, latestTime time.Time, earliestCommit, latestCommit string, whitelistPath structures.RegexpSet, out chan *finder.DriverResult) {
    var err error
    defer func() {
        if err != nil {
            out <- &finder.DriverResult{Err: err}
        }
    }()

    var gitRepo *git.Repository
    gitRepo, err = git.PlainOpen(cloneDir)
    if err != nil {
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
    }

    var branchIter gitstorer.ReferenceIter
    branchIter, err = gitRepo.Branches()
    if err != nil {
        return
    }

    var branches []*gitplumbing.Reference
    err = branchIter.ForEach(func(branch *gitplumbing.Reference) (err error) {
        branches = append(branches, branch)
        return
    })

    for _, branch := range branches {
        err = c.findInBranch(search, branch)
        if err != nil {
            return
        }
    }

    return
}

func (c *Comb) findInBranch(search *searchState, branch *gitplumbing.Reference) (err error) {
    var history gitobject.CommitIter
    history, err = search.gitRepo.Log(&git.LogOptions{From: branch.Hash(), Order: git.LogOrderCommitterTime})
    if err != nil {
        return
    }

    latestCommitReached := false

    var commits []*gitobject.Commit
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

        commits = append(commits, commit)

        return
    })

    for _, commit := range commits {
        err = c.findInCommit(search, commit)
        if err != nil {
            return
        }
    }

    return
}

func (c *Comb) findInCommit(search *searchState, commit *gitobject.Commit) (err error) {
    log := c.log.WithFields(logrus.Fields{
        "commit": commit.Hash.String(),
        "repo":   search.repoName,
    })
    log.Debug("Searching commit")

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

    var fileChanges []*finder.DriverFileChange
    for _, change := range changes {
        var driverFileChange *finder.DriverFileChange
        driverFileChange, err = c.findInFileChange(search, commit, change)
        if err != nil {
            return
        }
        if driverFileChange != nil {
            fileChanges = append(fileChanges, driverFileChange)
        }
    }

    if fileChanges != nil {
        search.out <- &finder.DriverResult{Commit: &finder.DriverCommit{
            RepoID:      search.repoID,
            Commit:      commit.Message,
            CommitHash:  commit.Hash.String(),
            Date:        commit.Committer.When,
            AuthorEmail: commit.Author.Email,
            FileChanges: fileChanges,
        }}
    }

    return
}

func (c *Comb) findInFileChange(search *searchState, commit *gitobject.Commit, fileChange *gitobject.Change) (result *finder.DriverFileChange, err error) {

    // Get file patch
    var patch *gitobject.Patch
    patch, err = fileChange.Patch()
    if err != nil {
        return
    }
    filePatch := patch.FilePatches()[0]

    // Get file
    _, changedFile := filePatch.Files()
    if changedFile == nil {
        // Deleted file
        return
    }
    filePath := changedFile.Path()

    if search.whitelistPath.MatchStringAny(filePath) {
        c.log.WithField("filePath", filePath).Debug("file whitelisted and skipped")
        return
    }

    chunks := filePatch.Chunks()

    // Get diff
    buf := bytes.NewBuffer(nil)
    encoder := gitdiff.NewUnifiedEncoder(buf, 3)
    if err = encoder.Encode(patch); err != nil {
        return
    }
    diffString := buf.String()

    var file *gitobject.File
    file, err = commit.File(filePath)
    if err != nil {
        return
    }

    var fileContents string
    fileContents, err = file.Contents()
    if err != nil {
        return
    }

    var findings []*finder.DriverFinding

    for _, rule := range search.rules {
        var fileChangeFindings []*rulepkg.FileChangeFinding
        fileChangeFindings, err = rule.Processor.FindInFileChange(fileChange, chunks, diffString)
        if err != nil {
            return
        }

        for _, fileChangeFinding := range fileChangeFindings {
            findings = append(findings, &finder.DriverFinding{
                Rule:             rule,
                FileRange:        fileChangeFinding.FileRange,
                SecretsProcessed: fileChangeFinding.SecretsProcessed,
                SecretValues:     fileChangeFinding.SecretValues,
            })
        }
    }

    currentFileLineNumber := 1

    for _, chunk := range chunks {
        var ff []*finder.DriverFinding
        ff, err = c.findInChunk(search, chunk, &currentFileLineNumber)
        if err != nil {
            return
        }

        findings = append(findings, ff...)
    }

    // Remove overlapping findings
    var findingsNew []*finder.DriverFinding
    for _, finding := range findings {
        if overlapsWithAny(finding, findingsNew) {
            continue
        }

        findingsNew = append(findingsNew, finding)
    }
    findings = findingsNew

    if findings != nil {
        result = &finder.DriverFileChange{
            Path:         filePath,
            FileContents: fileContents,
            Diff:         diffString,
            Findings:     findings,
        }
    }

    return
}

func (c *Comb) findInChunk(search *searchState, chunk gitdiff.Chunk, currentFileLineNumber *int) (result []*finder.DriverFinding, err error) {
    chunkString := chunk.Content()

    // Remove the trailing line break
    chunkLen := len(chunkString)
    if chunkLen > 0 && chunkString[chunkLen-1:] == "\n" {
        chunkString = chunkString[:chunkLen-1]
    }

    switch chunk.Type() {

    case gitdiff.Equal:

        // Advance to the first line of the next chunk
        *currentFileLineNumber += countRunes(chunkString, '\n') + 1

    case gitdiff.Add:

        // For each line in chunk
        lines := strings.Split(chunkString, "\n")
        for _, line := range lines {
            if line == "" {
                *currentFileLineNumber += 1
                continue
            }

            for _, rule := range search.rules {
                var lineFindings []*rulepkg.LineFinding
                lineFindings, err = rule.Processor.FindInLine(line)
                if err != nil {
                    return
                }

                for _, lineFinding := range lineFindings {

                    secrets := lineFinding.SecretValues
                    if ! lineFinding.SecretsProcessed {
                        secrets = []string{lineFinding.LineRange.GetStringFrom(line)}
                    }

                    result = append(result, &finder.DriverFinding{
                        Rule: rule,
                        FileRange: &structures.FileRange{
                            StartLineNum: *currentFileLineNumber,
                            StartIndex:   lineFinding.LineRange.StartIndex,
                            EndLineNum:   *currentFileLineNumber,
                            EndIndex:     lineFinding.LineRange.EndIndex,
                        },
                        SecretsProcessed: true,
                        SecretValues:     secrets,
                    })
                }
            }

            // Advance to the next line
            *currentFileLineNumber += 1
        }
    }

    return
}

func overlapsWithAny(input *finder.DriverFinding, others []*finder.DriverFinding) bool {
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
