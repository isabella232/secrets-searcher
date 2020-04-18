package git

import (
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    "github.com/pantheon-systems/search-secrets/pkg/structures"
    "github.com/sirupsen/logrus"
    gitvendor "gopkg.in/src-d/go-git.v4"
    gitplumbing "gopkg.in/src-d/go-git.v4/plumbing"
    gitobject "gopkg.in/src-d/go-git.v4/plumbing/object"
    "io"
    "sync"
)

type Repository struct {
    git      *Git
    gitRepo  *gitvendor.Repository
    cloneDir string
    mutex    *sync.Mutex
    log      *logrus.Entry
}

func newRepository(git *Git, gitRepo *gitvendor.Repository, cloneDir string, log *logrus.Entry) (result *Repository, err error) {
    result = &Repository{
        git:      git,
        gitRepo:  gitRepo,
        cloneDir: cloneDir,
        mutex:    &sync.Mutex{},
        log:      log,
    }

    return
}

func (r *Repository) Log(filter *CommitFilter) (result []*Commit, err error) {
    logOptions := &gitvendor.LogOptions{Order: gitvendor.LogOrderCommitterTime}

    // If we have a list, let's just grab them directly
    if !filter.Hashes.IsEmpty() {

        // Filter by earliest commit
        if filter.EarliestCommit != "" || filter.LatestCommit != "" {
            r.log.Warn("passing EarliestCommit or LatestCommit won't work if you pass Hashes in commit filter because we're not using log")
        }

        for _, hashString := range filter.Hashes.Values() {
            var commit *Commit
            var gitCommit *gitobject.Commit

            gitCommit, err = r.wrapCommitObject(gitplumbing.NewHash(hashString))
            if err != nil {
                err = errors.Wrapv(err, "unable to retrieve commit", hashString)
                return
            }

            commit, err = newCommit(r, gitCommit)
            if err != nil {
                return
            }

            if !filter.IsIncluded(commit) {
                continue
            }

            result = append(result, commit)
        }
        return
    }

    // Filter by latest commit
    if filter != nil && filter.LatestCommit != "" {
        logOptions.From = gitplumbing.NewHash(filter.LatestCommit)
    } else {
        logOptions.All = true
    }

    var fromCommits gitobject.CommitIter
    fromCommits, err = r.gitRepo.Log(logOptions)
    if err != nil {
        err = errors.Wrap(err, "unable to get git log")
        return
    }

    hashSet := structures.NewSet(nil)
    var gitCommit *gitobject.Commit
    var commit *Commit
    var lastCommit *Commit
    var iterErr error

    for {
        gitCommit, iterErr = fromCommits.Next()
        if iterErr == io.EOF {

            // In certain situations, we can tell which is the oldest commit
            if lastCommit != nil && (filter == nil || filter.OldestCommitIsIncluded()) {
                lastCommit.Oldest = true

                // Try again with the new knighthood
                included := true
                if filter != nil {
                    included, _ = filter.IsIncludedInLogResults(commit, &hashSet)
                }
                if included {
                    result = append(result, commit)
                }
            }

            break
        }
        if iterErr != nil {
            err = iterErr
            return
        }

        commit, err = newCommit(r, gitCommit)
        if err != nil {
            return
        }

        included := true
        more := true
        if filter != nil {
            included, more = filter.IsIncludedInLogResults(commit, &hashSet)
        }
        if included {
            result = append(result, commit)
        }
        if !more {
            break
        }

        lastCommit = commit
    }

    return
}

func (r *Repository) Commit(hashString string) (result *Commit, err error) {
    var gitCommit *gitobject.Commit
    gitCommit, err = r.wrapCommitObject(gitplumbing.NewHash(hashString))
    if err != nil {
        return
    }
    result, err = newCommit(r, gitCommit)
    return
}

func (r *Repository) Spawn() (result *Repository, err error) {
    return r.git.NewRepository(r.cloneDir)
}

func (r *Repository) newCommitsFromItem(iter gitobject.CommitIter) (result []*Commit, err error) {
    r.mutex.Lock()
    defer r.mutex.Unlock()

    err = iter.ForEach(func(gitCommit *gitobject.Commit) (err error) {
        var commit *Commit
        commit, err = newCommit(r, gitCommit)
        result = append(result, commit)
        return
    })

    return
}

func (r *Repository) wrapCommitObject(h gitplumbing.Hash) (*gitobject.Commit, error) {
    r.mutex.Lock()
    defer r.mutex.Unlock()

    return r.gitRepo.CommitObject(h)
}

func (r *Repository) EmptyTree() *Tree {
    return newTree(&gitobject.Tree{})
}
