package code

import (
    "github.com/pantheon-systems/search-secrets/pkg/database"
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    interactpkg "github.com/pantheon-systems/search-secrets/pkg/interact"
    "github.com/sirsean/go-pool"
    "github.com/sirupsen/logrus"
    "path/filepath"
    "sort"
    "strings"
)

type (
    Code struct {
        sourceProvider SourceProvider
        codeDir        string
        interact       interactpkg.Interactish
        db             *database.Database
        log            *logrus.Logger
    }
    SourceProvider interface {
        GetRepositories() (result []*RepoInfo, err error)
    }
    RepoInfo struct {
        Name     string
        FullName string
        Owner    string
        SSHURL   string
        HTMLURL  string
    }
)

func New(sourceProvider SourceProvider, codeDir string, interact interactpkg.Interactish, db *database.Database, log *logrus.Logger) *Code {
    return &Code{
        sourceProvider: sourceProvider,
        codeDir:        codeDir,
        interact:       interact,
        db:             db,
        log:            log,
    }
}

func (c *Code) PrepareCode() (err error) {
    if c.db.TableExists(database.RepoTable) {
        err = errors.New("one or more code tables already exist, cannot prepare code")
        return
    }

    var repoInfos []*RepoInfo
    repoInfos, err = c.sourceProvider.GetRepositories()
    if err != nil {
        err = errors.WithMessagev(err, "unable to get repositories")
        return
    }
    sort.Slice(repoInfos, func(i, j int) bool {
        return strings.ToLower(repoInfos[i].FullName) < strings.ToLower(repoInfos[j].FullName)
    })

    p := pool.NewPool(len(repoInfos), 5)
    p.Start()

    prog := c.interact.NewProgress()

    for _, repoInfo := range repoInfos {
        log := c.log.WithField("repo", repoInfo.Name)
        cloneDir := filepath.Join(c.codeDir, repoInfo.Name)

        p.Add(worker{
            repoInfo: repoInfo,
            cloneDir: cloneDir,
            prog:     prog,
            db:       c.db,
            log:      log,
        })
    }

    p.Close()

    return
}
