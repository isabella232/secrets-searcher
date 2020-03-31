package code

import (
    "github.com/pantheon-systems/search-secrets/pkg/database"
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    "github.com/pantheon-systems/search-secrets/pkg/github"
    "github.com/pantheon-systems/search-secrets/pkg/structures"
    "github.com/sirupsen/logrus"
    "gopkg.in/src-d/go-git.v4"
    "os"
    "path/filepath"
)

type Code struct {
    githubAPI    *github.API
    organization string
    codeDir      string
    repoFilter   *structures.Filter
    db           *database.Database
    log          *logrus.Logger
}

func New(githubAPI *github.API, organization, codeDir string, repoFilter *structures.Filter, db *database.Database, log *logrus.Logger) *Code {
    return &Code{
        githubAPI:    githubAPI,
        organization: organization,
        codeDir:      codeDir,
        repoFilter:   repoFilter,
        db:           db,
        log:          log,
    }
}

func (c *Code) PrepareRepos() (err error) {
    err = c.pullRepoData()
    if err != nil {
        return
    }

    return c.cloneRepos()
}

func (c *Code) CloneDir(repoName string) string {
    return filepath.Join(c.codeDir, repoName)
}

func (c *Code) pullRepoData() (err error) {
    if c.db.TableExists(database.RepoTable) {
        c.log.Warn("repo table already exists, skipping")
        return
    }

    ghRepos, err := c.githubAPI.GetRepositoriesByOrganization(c.organization)
    if err != nil {
        return
    }

    for _, ghRepo := range ghRepos {
        if ! c.repoFilter.IsIncluded(ghRepo.GetName()) {
            continue
        }
        // TODO This should be in config
        if ghRepo.GetFork() {
            continue
        }
        err = c.db.WriteRepo(&database.Repo{
            ID:       database.CreateHashID(ghRepo.GetFullName()),
            Name:     ghRepo.GetName(),
            Owner:    ghRepo.Owner.GetLogin(),
            FullName: ghRepo.GetFullName(),
            SSHURL:   ghRepo.GetSSHURL(),
            HTMLURL:  ghRepo.GetHTMLURL(),
        })
        if err != nil {
            return
        }
    }

    return
}

func (c *Code) cloneRepos() (err error) {
    repos, err := c.db.GetReposFiltered(c.repoFilter)
    if err != nil {
        return
    }

    for _, repo := range repos {
        err = c.cloneRepo(repo)
        if err != nil {
            return
        }
    }

    return
}

func (c *Code) cloneRepo(repo *database.Repo) (err error) {
    cloneDir := c.CloneDir(repo.Name)

    // Check if the code dir exists for repo
    if _, err = os.Stat(cloneDir); err == nil {
        c.log.WithField("repo", repo.Name).Warn("repo clone already exists, skipping clone")
        return
    }

    // Clone
    co := &git.CloneOptions{URL: repo.SSHURL, Progress: os.Stdout}
    if _, err = git.PlainClone(cloneDir, false, co); err != nil {
        err = errors.Wrapv(err, "unable to clone repo", repo.SSHURL, cloneDir)
        return
    }

    return
}
