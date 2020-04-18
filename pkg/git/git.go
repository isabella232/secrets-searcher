package git

import (
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    "github.com/sirupsen/logrus"
    gitvendor "gopkg.in/src-d/go-git.v4"
    gittransport "gopkg.in/src-d/go-git.v4/plumbing/transport"
)

type Git struct {
    log *logrus.Entry
}

func New(log *logrus.Entry) *Git {
    return &Git{
        log: log,
    }
}

func (g *Git) NewRepository(cloneDir string) (result *Repository, err error) {
    var gitRepo *gitvendor.Repository
    gitRepo, err = gitvendor.PlainOpen(cloneDir)
    if err != nil {
        err = errors.Wrapv(err, "unable to open directory", cloneDir)
        return
    }

    return newRepository(g, gitRepo, cloneDir, g.log)
}

func (g *Git) Clone(url, cloneDir string) (result *Repository, err error) {
    var gitRepo *gitvendor.Repository
    co := &gitvendor.CloneOptions{URL: url}
    if gitRepo, err = gitvendor.PlainClone(cloneDir, false, co); err != nil {
        return
    }

    return newRepository(g, gitRepo, cloneDir, g.log)
}

func (g *Git) IsCloneValid(cloneDir string) (result bool) {
    var err error

    var repo *Repository
    repo, err = g.NewRepository(cloneDir)
    if err != nil {
        return false
    }

    _, err = repo.gitRepo.Head()

    return err == nil
}

func IsErrEmptyRemoteRepository(err error) bool {
    return err == gittransport.ErrEmptyRemoteRepository
}
