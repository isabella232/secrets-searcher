package git

import (
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    "github.com/pantheon-systems/search-secrets/pkg/structures"
    gitvendor "gopkg.in/src-d/go-git.v4"
    gitplumbing "gopkg.in/src-d/go-git.v4/plumbing"
    gitobject "gopkg.in/src-d/go-git.v4/plumbing/object"
    gitstorer "gopkg.in/src-d/go-git.v4/plumbing/storer"
    "time"
)

type (
    Repository struct {
        git     *Git
        gitRepo *gitvendor.Repository
    }
    CommitFilter struct {
        EarliestTime                time.Time
        LatestTime                  time.Time
        EarliestCommit              string
        LatestCommit                string
        ExcludeMergeCommits         bool
        ExcludeCommitsWithNoParents bool
    }
)

func newRepository(git *Git, gitRepo *gitvendor.Repository) (result *Repository, err error) {
    result = &Repository{
        git:     git,
        gitRepo: gitRepo,
    }

    return
}

func (r *Repository) Log(filter *CommitFilter) (result []*Commit, err error) {
    logOptions := &gitvendor.LogOptions{Order: gitvendor.LogOrderCommitterTime}

    // Filter by latest commit
    if filter.LatestCommit != "" {
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
    err = fromCommits.ForEach(func(gitCommit *gitobject.Commit) (err error) {

        // Filter by time
        commitTime := gitCommit.Committer.When
        if !filter.LatestTime.IsZero() && commitTime.After(filter.LatestTime) {
            return
        }
        if commitTime.Before(filter.EarliestTime) {
            return gitstorer.ErrStop
        }

        // Filter by earliest commit
        if filter.EarliestCommit != "" && hashSet.Contains(filter.EarliestCommit) {
            return gitstorer.ErrStop
        }

        commit := newCommit(gitCommit)

        // Filter out merge commits
        if filter.ExcludeMergeCommits && commit.IsMergeCommit() {
            return
        }

        // Filter out commits with no parent
        if filter.ExcludeCommitsWithNoParents && !commit.HasParents() {
            return
        }

        result = append(result, commit)

        return
    })
    if err != nil {
        err = errors.Wrap(err, "unable to find in ancestor commits of commit")
        return
    }

    return
}
