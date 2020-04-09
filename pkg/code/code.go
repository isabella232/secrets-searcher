package code

import (
    "github.com/pantheon-systems/search-secrets/pkg/database"
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    "github.com/pantheon-systems/search-secrets/pkg/github"
    "github.com/pantheon-systems/search-secrets/pkg/structures"
    "github.com/sirupsen/logrus"
    "gopkg.in/src-d/go-git.v4"
    gittransport "gopkg.in/src-d/go-git.v4/plumbing/transport"
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

func (c *Code) PrepareCode() (err error) {
    if c.db.TableExists(database.RepoTable) {
        return errors.New("one or more code tables already exist, cannot prepare code")
    }

    if err = c.pullRepoData(); err != nil {
        return errors.WithMessage(err, "unable to pull repo data")
    }

    return c.cloneRepos()
}

func (c *Code) CloneDir(repoName string) string {
    return filepath.Join(c.codeDir, repoName)
}

func (c *Code) pullRepoData() (err error) {
    ghRepos, err := c.githubAPI.GetRepositoriesByOrganization(c.organization)
    if err != nil {
        return errors.WithMessagev(err, "unable to get repositories", c.organization)
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
            return errors.WithMessage(err, "unable to write repo")
        }
    }

    return
}

func (c *Code) cloneRepos() (err error) {
    repos, err := c.db.GetReposFiltered(c.repoFilter)
    if err != nil {
        return errors.WithMessage(err, "unable to get filtered list of repos")
    }

    for _, repo := range repos {
        newLog := c.log.WithField("repo", repo.Name)
        err = c.cloneRepo(repo, newLog)
        if err != nil {

            // Empty repos should not crash the execution
            if err == gittransport.ErrEmptyRemoteRepository {
                if err = c.db.DeleteRepo(repo.ID); err != nil {
                    newLog.Warn("repo is empty, skipped search")
                }
                continue
            }

            return errors.WithMessage(err, "unable to clone repo")
        }
    }

    return
}

func (c *Code) cloneRepo(repo *database.Repo, log *logrus.Entry) (err error) {
    cloneDir := c.CloneDir(repo.Name)

    // Check if the code dir exists for repo
    if _, err = os.Stat(cloneDir); err == nil {
        c.log.WithField("repo", repo.Name).Warn("repo clone already exists, skipping clone")
        return
    }

    // Clone
    log.Debug("cloning repo")
    co := &git.CloneOptions{URL: repo.SSHURL}
    if _, err = git.PlainClone(cloneDir, false, co); err != nil {

        // Empty repos should not crash the execution
        if err == gittransport.ErrEmptyRemoteRepository {
            return
        }

        err = errors.Wrapv(err, "unable to clone repo", repo.SSHURL, cloneDir)
        return
    }

    return
}
