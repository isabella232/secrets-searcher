package git

import (
	"io"
	"sync"

	"gopkg.in/src-d/go-git.v4/config"

	"github.com/pantheon-systems/secrets-searcher/pkg/errors"
	"github.com/pantheon-systems/secrets-searcher/pkg/logg"
	gitvendor "gopkg.in/src-d/go-git.v4"
	gitplumbing "gopkg.in/src-d/go-git.v4/plumbing"
	gitobject "gopkg.in/src-d/go-git.v4/plumbing/object"
)

type Repository struct {
	git      *Git
	gitRepo  *gitvendor.Repository
	cloneDir string
	mutex    *sync.Mutex
	log      logg.Logg
}

func newRepository(git *Git, gitRepo *gitvendor.Repository, cloneDir string, log logg.Logg) (result *Repository) {
	return &Repository{
		git:      git,
		gitRepo:  gitRepo,
		cloneDir: cloneDir,
		mutex:    &sync.Mutex{},
		log:      log,
	}
}

func (r *Repository) FetchAll(url string) (err error) {
	remoteConfig := &config.RemoteConfig{Name: "origin", URLs: []string{url}}

	var remoteExists = true
	if _, remErr := r.gitRepo.Remote(remoteConfig.Name); remErr != nil {
		if remErr == gitvendor.ErrRemoteNotFound {
			remoteExists = false
		} else {
			err = errors.Wrap(remErr, "unable to get remote")
			return
		}
	}

	if remoteExists {
		if err = r.gitRepo.DeleteRemote(remoteConfig.Name); err != nil {
			err = errors.Wrap(err, "unable to delete remote")
			return
		}
	}

	if _, err = r.gitRepo.CreateRemote(remoteConfig); err != nil {
		err = errors.Wrap(err, "unable to create remote")
		return
	}

	if fetchErr := r.gitRepo.Fetch(&gitvendor.FetchOptions{}); fetchErr != nil && fetchErr != gitvendor.NoErrAlreadyUpToDate {
		err = errors.Wrap(fetchErr, "unable to fetch all")
		return
	}

	return
}

func (r *Repository) Log(commitFilter *CommitFilter) (result []*Commit, err error) {
	if commitFilter == nil {
		commitFilter = NewEmptyCommitFilter()
	}
	logOptions := &gitvendor.LogOptions{Order: gitvendor.LogOrderCommitterTime}

	// If we have a list, let's just grab them directly
	if commitFilter.CanProvideExactCommitHashValues() {
		commitHashValues := commitFilter.ExactCommitHashValues()

		for _, hashString := range commitHashValues.StringValues() {
			var commit *Commit
			commit, err = r.Commit(hashString)
			if err != nil {
				err = errors.Wrapv(err, "unable to retrieve commit", hashString)
				return
			}

			if !commitFilter.Includes(commit) {
				continue
			}

			result = append(result, commit)
		}
		return
	}

	logOptions.All = true

	var fromCommits gitobject.CommitIter
	fromCommits, err = r.gitRepo.Log(logOptions)
	if err != nil {
		err = errors.Wrap(err, "unable to get git log")
		return
	}

	var gitCommit *gitobject.Commit
	var commit *Commit
	var lastCommit *Commit
	var iterErr error

	for {
		gitCommit, iterErr = fromCommits.Next()
		if iterErr == io.EOF {

			// In certain situations, we can tell which is the oldest commit
			if lastCommit != nil && (commitFilter == nil || commitFilter.OldestCommitIsIncluded()) {
				lastCommit.Oldest = true

				// Try again with the new knighthood
				included := true
				if commitFilter != nil {
					included, _ = commitFilter.IsIncludedInLogResults(commit)
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
			err = errors.WithMessage(err, "unable to create new commit")
			return
		}

		included := true
		more := true
		if commitFilter != nil {
			included, more = commitFilter.IsIncludedInLogResults(commit)
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

func (r *Repository) GetOldestCommit() (result *Commit, err error) {
	var commits []*Commit
	commits, err = r.Log(nil)
	if err != nil {
		err = errors.WithMessage(err, "unable to run git log")
		return
	}

	for _, commit := range commits {
		if commit.Oldest {
			result = commit
			return
		}
	}

	return
}

func (r *Repository) Commit(hashString string) (result *Commit, err error) {
	var gitCommit *gitobject.Commit
	gitCommit, err = r.wrapCommitObject(gitplumbing.NewHash(hashString))
	if err != nil {
		err = errors.WithMessage(err, "unable to get commit")
		return
	}

	result, err = newCommit(r, gitCommit)

	return
}

func (r *Repository) Spawn() (result *Repository, err error) {
	return r.git.OpenRepository(r.cloneDir)
}

func (r *Repository) newCommitsFromIter(iter gitobject.CommitIter) (result []*Commit, err error) {
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
