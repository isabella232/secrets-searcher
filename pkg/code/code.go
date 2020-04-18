package code

import (
    "github.com/pantheon-systems/search-secrets/pkg/database"
    gitpkg "github.com/pantheon-systems/search-secrets/pkg/git"
    interactpkg "github.com/pantheon-systems/search-secrets/pkg/interact"
    "github.com/pantheon-systems/search-secrets/pkg/structures"
    "github.com/sirsean/go-pool"
    "github.com/sirupsen/logrus"
    "path/filepath"
    "sort"
    "strings"
)

type (
    Code struct {
        sourceProvider SourceProvider
        repoFilter     *structures.Filter
        codeDir        string
        interact       interactpkg.Interactish
        db             *database.Database
        log            *logrus.Entry
    }
    SourceProvider interface {
        GetName() string
        GetRepositories() (result []*RepoInfo, err error)
    }
    RepoInfo struct {
        Name           string
        SourceProvider string
        FullName       string
        Owner          string
        SSHURL         string
        HTMLURL        string
    }
)

func New(sourceProvider SourceProvider, repoFilter *structures.Filter, codeDir string, interact interactpkg.Interactish, db *database.Database, log *logrus.Entry) *Code {
    return &Code{
        sourceProvider: sourceProvider,
        repoFilter:     repoFilter,
        codeDir:        codeDir,
        interact:       interact,
        db:             db,
        log:            log,
    }
}

func (c *Code) PrepareCode() (err error) {
    var repoInfos []*RepoInfo
    repoInfos, err = c.getRepoInfos()
    if err != nil {
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
            git:      gitpkg.New(c.log),
            prog:     prog,
            db:       c.db,
            log:      log,
        })
    }

    p.Close()
    if prog != nil {
        prog.Wait()
    }

    return
}

func (c *Code) getRepoInfos() (result []*RepoInfo, err error) {
    // If the filter matches the existing list of repos in the database, we'll just proceed
    if c.db.TableExists(database.RepoTable) && c.repoFilter.CanProvideExactValues() {

        // Build repo infos from database repos
        var repoInfos []*RepoInfo
        var repoNames structures.Set
        repoInfos, repoNames, err = c.getFilteredRepoInfosFromDatabase()
        if err != nil {
            return
        }

        // See if the filtered list equals the filter provided exactly
        if c.repoFilter.ExactValues().EqualsAfterBothSorted(repoNames) {
            c.log.Debug("using repo information from database since it contains all repos from the filter, skipping source provider")
            result = repoInfos
        }

        // Not we have repo data in memory so delete the repo table. the worker class will save the correct list again
        c.log.Debug("deleting repo table")
        if err = c.db.DeleteTableIfExists(database.RepoTable); err != nil {
            return
        }
    }

    // Get repo infos from source provider if we haven't already gotten the repo list from the database
    if result == nil {
        c.log.Debug("querying source provider for new repo info")
        result, err = c.sourceProvider.GetRepositories()
    }

    return
}

func (c *Code) getFilteredRepoInfosFromDatabase() (result []*RepoInfo, repoNames structures.Set, err error) {
    var dbRepos []*database.Repo
    dbRepos, err = c.db.GetRepos()
    if err != nil {
        return
    }

    repoNames = structures.NewSet(nil)
    for _, dbRepo := range dbRepos {
        if !c.repoFilter.IsIncluded(dbRepo.Name) {
            continue
        }

        result = append(result, &RepoInfo{
            Name:           dbRepo.Name,
            SourceProvider: c.sourceProvider.GetName(),
            FullName:       dbRepo.FullName,
            Owner:          dbRepo.Owner,
            SSHURL:         dbRepo.SSHURL,
            HTMLURL:        dbRepo.HTMLURL,
        })

        repoNames.Add(dbRepo.Name)
    }
    return
}
