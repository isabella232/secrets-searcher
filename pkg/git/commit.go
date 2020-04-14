package git

import (
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    "github.com/pantheon-systems/search-secrets/pkg/structures"
    gitobject "gopkg.in/src-d/go-git.v4/plumbing/object"
    gitstorer "gopkg.in/src-d/go-git.v4/plumbing/storer"
    "time"
)

type (
    Commit struct {
        Hash        string
        Message     string
        Time        time.Time
        AuthorEmail string
        AuthorFull  string
        gitCommit   *gitobject.Commit
    }
    FileChangeFilter struct {
        ExcludeFileDeletions bool
        ExcludeMatchingPaths structures.RegexpSet
        ExcludeBinaryOrEmpty         bool
        ExcludeOnesWithNoCodeChanges bool
    }
)

func newCommit(gitCommit *gitobject.Commit) (result *Commit) {
    result = &Commit{
        Hash:        gitCommit.Hash.String(),
        Message:     gitCommit.Message,
        Time:        gitCommit.Committer.When,
        AuthorEmail: gitCommit.Author.Name,
        AuthorFull:  gitCommit.Author.String(),
        gitCommit:   gitCommit,
    }

    return
}

func (c *Commit) FirstParent() (result *Commit, err error) {
    var parents []*Commit
    parents, err = c.Parents(1)
    if err != nil {
        return
    }
    if len(parents) == 0 {
        err = errors.New("no first parent")
    }
    result = parents[0]
    return
}

func (c *Commit) Parents(limit int) (result []*Commit, err error) {
    i := 0
    err = c.gitCommit.Parents().ForEach(func(gitCommit *gitobject.Commit) (err error) {
        result = append(result, newCommit(gitCommit))
        i += 1
        if limit > 0 && i == limit {
            return gitstorer.ErrStop
        }
        return
    })
    return
}

func (c *Commit) HasParents() bool {
    return len(c.gitCommit.ParentHashes) > 0
}

func (c *Commit) IsMergeCommit() bool {
    return len(c.gitCommit.ParentHashes) < 2
}

func (c *Commit) FileChanges(filter *FileChangeFilter) (result []*FileChange, err error) {
    if !c.HasParents() {
        err = errors.New("unable to get file changes with no parent commits")
        return
    }
    if c.IsMergeCommit() {
        err = errors.New("unable to get file changes for a merge commit")
        return
    }

    var parentCommit *Commit
    parentCommit, err = c.FirstParent()
    if err != nil {
        return
    }

    var commitTree *gitobject.Tree
    commitTree, err = c.gitCommit.Tree()
    if err != nil {
        return
    }

    var parentCommitTree *gitobject.Tree
    parentCommitTree, err = parentCommit.gitCommit.Tree()
    if err != nil {
        return
    }

    var gitFileChanges gitobject.Changes
    gitFileChanges, err = parentCommitTree.Diff(commitTree)
    if err != nil {
        return
    }

    for _, gitFileChange := range gitFileChanges {
        fileChange := NewFileChange(gitFileChange)

        // Filter out deletions
        if filter.ExcludeFileDeletions && fileChange.IsDeletion() {
            continue
        }

        // Filter by path name
        if filter.ExcludeMatchingPaths.MatchAny(gitFileChange.To.Name) {
            continue
        }

        // Filter out ones with no code changes
        if filter.ExcludeOnesWithNoCodeChanges {
            var hasCodeChanges bool
            hasCodeChanges, err = fileChange.HasCodeChanges()
            if err != nil {
                return
            }
            if !hasCodeChanges {
                continue
            }
        }

        // Filter out empty or binary files
        if filter.ExcludeBinaryOrEmpty {
            var isBinary bool
            isBinary, err = fileChange.IsBinaryOrEmpty()
            if err != nil {
                return
            }
            if isBinary {
                continue
            }
        }

        result = append(result, fileChange)
    }

    return
}

func (c *Commit) FileContents(fileChange *FileChange) (result string, err error) {
    var file *gitobject.File
    file, err = c.gitCommit.File(fileChange.Path)
    if err != nil {
        return
    }

    result, err = file.Contents()

    return
}
