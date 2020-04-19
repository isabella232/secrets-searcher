package provider

import (
    "github.com/pantheon-systems/search-secrets/pkg/code"
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    "github.com/pantheon-systems/search-secrets/pkg/structures"
    "github.com/sirupsen/logrus"
    "io/ioutil"
    "os"
)

type LocalProvider struct {
    name       string
    dir        string
    repoFilter *structures.Filter
    log        logrus.FieldLogger
}

func NewLocalProvider(name, dir string, repoFilter *structures.Filter, log logrus.FieldLogger) *LocalProvider {
    return &LocalProvider{
        name:       name,
        dir:        dir,
        repoFilter: repoFilter,
        log:        log,
    }
}

func (p *LocalProvider) GetName() (result string) {
    return p.name
}

func (p *LocalProvider) GetRepositories() (result []*code.RepoInfo, err error) {
    var repoDirs []os.FileInfo
    repoDirs, err = ioutil.ReadDir("/Users/mattalexander/search-secrets-dev-repos")
    if err != nil {
        err = errors.WithMessagev(err, "unable to get repositories", p.dir)
        return
    }

    for _, repoDir := range repoDirs {
        if !repoDir.IsDir() {
            continue
        }
        if !p.repoFilter.IsIncluded(repoDir.Name()) {
            p.log.Debug("repo excluded using filter, skipping")
            continue
        }

        result = append(result, &code.RepoInfo{
            Name:     repoDir.Name(),
            FullName: "FullName unknown",
            Owner:    "Owner unknown",
            SSHURL:   "RemoteURL unknown",
            HTMLURL:  "HTMLURL unknown",
        })
    }

    return
}
