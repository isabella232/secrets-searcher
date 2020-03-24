package code

import (
    "github.com/pantheon-systems/search-secrets/pkg/code/github"
    "github.com/pantheon-systems/search-secrets/pkg/database"
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    filterpkg "github.com/pantheon-systems/search-secrets/pkg/filter"
    "github.com/sirupsen/logrus"
    "gopkg.in/src-d/go-git.v4"
    "os"
    "path/filepath"
    "strconv"
)

type Code struct {
    githubAPI    *github.API
    organization string
    codeDir      string
    filter       *filterpkg.Filter
    db           *database.Database
    log          *logrus.Logger
}

func New(githubToken, organization, codeDir string, filter *filterpkg.Filter, db *database.Database, log *logrus.Logger) (*Code, error) {
    githubAPI := github.New(githubToken)

    return &Code{
        githubAPI:    githubAPI,
        organization: organization,
        codeDir:      codeDir,
        filter:       filter,
        db:           db,
        log:          log,
    }, nil
}

func (c *Code) PrepareRepos() (err error) {
    err = c.pullRepoData()
    if err != nil {
        return err
    }

    return c.cloneRepos()
}

func (c *Code) CloneDir(repo *database.Repo) string {
    return filepath.Join(c.codeDir, repo.Name)
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
        id := database.CreateHashID(strconv.FormatInt(*ghRepo.ID, 10))
        repo := &database.Repo{
            ID:     id,
            Name:   *ghRepo.Name,
            SSHURL: *ghRepo.SSHURL,
        }

        err = c.db.Write(database.RepoTable, id, repo)
        if err != nil {
            return
        }
    }

    return
}

func (c *Code) cloneRepos() (err error) {
    repos, err := c.db.GetRepos()
    if err != nil {
        return err
    }

    for _, repo := range repos {
        if ! c.filter.Repos.Include(repo.Name) {
            continue
        }
        err = c.cloneRepo(repo)
        if err != nil {
            return err
        }
    }

    return
}

func (c *Code) cloneRepo(repo *database.Repo) (err error) {
    cloneDir := c.CloneDir(repo)

    // Check if the code dir exists for repo
    if _, err = os.Stat(cloneDir); err == nil {
        c.log.Warn("repo clone already exists, skipping clone")
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
