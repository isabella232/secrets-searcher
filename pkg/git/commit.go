package git

import (
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    gitobject "gopkg.in/src-d/go-git.v4/plumbing/object"
    "sync"
    "time"
)

type (
    Commit struct {
        repository  *Repository
        Hash        string
        Message     string
        Date        time.Time
        AuthorEmail string
        AuthorFull  string
        Oldest      bool
        gitCommit   *gitobject.Commit
        memo        commitMemo
        mutex       *sync.Mutex
    }
    commitMemo struct {
        fileContentIndex map[string]string
        parents          []*Commit
        tree             *Tree
    }
)

func newCommit(repository *Repository, gitCommit *gitobject.Commit) (result *Commit, err error) {
    result = &Commit{
        repository:  repository,
        Hash:        gitCommit.Hash.String(),
        Message:     gitCommit.Message,
        Date:        gitCommit.Committer.When,
        AuthorEmail: gitCommit.Author.Name,
        AuthorFull:  gitCommit.Author.String(),
        gitCommit:   gitCommit,
        memo: commitMemo{
            fileContentIndex: map[string]string{},
        },
        mutex: &sync.Mutex{},
    }

    return
}

func (c *Commit) Parents() (result []*Commit, err error) {
    defer errors.CatchPanicDo(func(err error) {
        err = errors.WithMessage(err, "unable to retrieve parent commits")
    })

    if c.memo.parents != nil {
        result = c.memo.parents
        return
    }

    // Get parents
    result, err = c.repository.newCommitsFromItem(c.parents())
    if err != nil {
        err = errors.WithMessage(err, "unable to get parent commits of commit")
        return
    }

    c.memo.parents = result

    return
}

func (c *Commit) HasParents() (result bool, err error) {
    var parents []*Commit
    parents, err = c.Parents()
    result = err == nil && len(parents) > 0
    return
}

func (c *Commit) IsMergeCommit() (result bool, err error) {
    var parents []*Commit
    parents, err = c.Parents()
    result = err == nil && len(parents) > 1
    return
}

func (c *Commit) CanDiff() (result bool) {
    if c.Oldest {
        return true
    }

    var parents []*Commit
    parents, _ = c.Parents()
    result = len(parents) == 1
    return
}

func (c *Commit) FileChanges(filter *FileChangeFilter) (result []*FileChange, err error) {

    // Are we going to get
    if !c.CanDiff() {
        return
    }

    // Get parent
    var parents []*Commit
    parents, err = c.Parents()
    if err != nil {
        err = errors.WithMessage(err, "unable to get parents")
        return
    }
    parentsLen := len(parents)

    // Get this tree
    var commitTree *Tree
    commitTree, err = c.tree()
    if err != nil {
        err = errors.WithMessage(err, "unable to get tree")
        return
    }

    // Get parent tree
    var parentCommitTree *Tree
    if parentsLen == 1 {
        parentCommitTree, err = parents[0].tree()
        if err != nil {
            err = errors.WithMessage(err, "unable to get parent tree")
            return
        }
    } else if c.Oldest {
        parentCommitTree = c.repository.EmptyTree()
    } else {
        err = errors.New("unable to find parent (we shouldn't be here since we're calling CanDiff above)")
        return
    }

    // Diff them together
    var gitFileChanges gitobject.Changes
    gitFileChanges, err = parentCommitTree.wrapDiff(commitTree)
    if err != nil {
        err = errors.WithMessage(err, "unable to diff")
        return
    }

    // For each file change, build a FileChange object
    for _, gitFileChange := range gitFileChanges {
        fileChange := NewFileChange(c, gitFileChange)

        if filter != nil {

            // Filter out deletions
            if filter.ExcludeFileDeletions && fileChange.IsDeletion() {
                continue
            }

            // Filter by path name
            if filter.IncludeMatchingPaths != nil && !filter.IncludeMatchingPaths.MatchAny(gitFileChange.To.Name) {
                continue
            }
            if filter.ExcludeMatchingPaths != nil && filter.ExcludeMatchingPaths.MatchAny(gitFileChange.To.Name) {
                continue
            }

            // Filter out ones with no code changes
            if filter.ExcludeOnesWithNoCodeChanges {
                var hasCodeChanges bool
                hasCodeChanges, err = fileChange.HasCodeChanges()
                if err != nil {
                    err = errors.WithMessage(err, "unable to detect if file change has code changes")
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
                    err = errors.WithMessage(err, "unable to detect if file is binary/empty")
                    return
                }
                if isBinary {
                    continue
                }
            }
        }

        result = append(result, fileChange)
    }

    return
}

func (c *Commit) FileContents(path string) (result string, err error) {
    var ok bool
    result, ok = c.memo.fileContentIndex[path]
    if ok {
        return
    }

    var file *gitobject.File
    file, err = c.gitCommit.File(path)
    if err != nil {
        err = errors.WithMessage(err, "unable to get file at commit")
        return
    }

    result, err = file.Contents()

    c.memo.fileContentIndex[path] = result

    return
}

func (c *Commit) tree() (result *Tree, err error) {
    c.repository.mutex.Lock()
    defer c.repository.mutex.Unlock()

    if c.memo.tree != nil {
        result = c.memo.tree
        return
    }

    var gitTree *gitobject.Tree
    gitTree, err = c.gitCommit.Tree()
    if err != nil {
        err = errors.WithMessage(err, "unable to get tree")
        return
    }
    result = newTree(gitTree)

    c.memo.tree = result

    return
}

func (c *Commit) parents() gitobject.CommitIter {
    c.repository.mutex.Lock()
    defer c.repository.mutex.Unlock()

    return c.gitCommit.Parents()
}
