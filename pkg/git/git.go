package git

import (
	"github.com/pantheon-systems/secrets-searcher/pkg/errors"
	"github.com/pantheon-systems/secrets-searcher/pkg/logg"
	gitvendor "gopkg.in/src-d/go-git.v4"
	gittransport "gopkg.in/src-d/go-git.v4/plumbing/transport"
)

type Git struct {
	log logg.Logg
}

func New(log logg.Logg) *Git {
	return &Git{
		log: log,
	}
}

func (g *Git) OpenRepository(cloneDir string) (result *Repository, err error) {
	var gitRepo *gitvendor.Repository
	gitRepo, err = gitvendor.PlainOpen(cloneDir)
	if err != nil {
		err = errors.Wrapv(err, "unable to open directory", cloneDir)
		return
	}

	return newRepository(g, gitRepo, cloneDir, g.log), nil
}

func (g *Git) Clone(url, cloneDir string) (result *Repository, err error) {
	var gitRepo *gitvendor.Repository
	co := &gitvendor.CloneOptions{URL: url}
	if gitRepo, err = gitvendor.PlainClone(cloneDir, false, co); err != nil {
		err = errors.Wrapf(err, "unable to clone from %s to %s", url, cloneDir)
		return
	}

	return newRepository(g, gitRepo, cloneDir, g.log), nil
}

func (g *Git) ValidateClone(cloneDir string) (err error) {
	var repo *Repository
	repo, err = g.OpenRepository(cloneDir)
	if err != nil {
		return errors.WithMessagev(err, "invalid clone", cloneDir)
	}

	_, err = repo.gitRepo.Head()

	return err
}

func (g *Git) IsCloneValid(cloneDir string) (result bool) {
	return g.ValidateClone(cloneDir) == nil
}

func IsErrEmptyRemoteRepository(err error) bool {
	return errors.WasCausedBy(err, gittransport.ErrEmptyRemoteRepository)
}
