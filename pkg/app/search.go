package app

import (
    codepkg "github.com/pantheon-systems/search-secrets/pkg/code"
    "github.com/pantheon-systems/search-secrets/pkg/database"
    finderpkg "github.com/pantheon-systems/search-secrets/pkg/finder"
    "github.com/pantheon-systems/search-secrets/pkg/finder/comb"
    "github.com/pantheon-systems/search-secrets/pkg/github"
    reporterpkg "github.com/pantheon-systems/search-secrets/pkg/reporter"
    "github.com/pantheon-systems/search-secrets/pkg/rule"
    "github.com/pantheon-systems/search-secrets/pkg/structures"
    "github.com/sirupsen/logrus"
    "path/filepath"
    "strings"
    "time"
)

type Search struct {
    code     *codepkg.Code
    finder   *finderpkg.Finder
    log      *logrus.Logger
    reporter *reporterpkg.Reporter
}

func NewSearch(githubToken, organization, outputDir string, repos, refs []string, rules []*rule.Rule, earliestTime, latestTime time.Time, earliestCommit, latestCommit string, whitelistPath structures.RegexpSet, whitelistSecretIDSet structures.Set, log *logrus.Logger) (search *Search, err error) {

    // Directories
    var outputDirAbs string
    outputDirAbs, err = filepath.Abs(outputDir)
    if err != nil {
        return
    }
    codeDir := filepath.Join(outputDirAbs, "code")
    dbDir := filepath.Join(outputDirAbs, "db")
    reportDir := filepath.Join(outputDirAbs, "report")

    // Create database
    var db *database.Database
    db, err = database.New(dbDir)
    if err != nil {
        return
    }

    // Create filters
    repoFilter := structures.NewFilter(repos)
    refFilter := buildRefFilter(refs)

    // Create Github API
    githubAPI := github.NewAPI(githubToken)

    // Create code
    code := codepkg.New(githubAPI, organization, codeDir, repoFilter, db, log)

    // Create driver
    driver := comb.New(log)

    // Create finder
    finder := finderpkg.New(driver, code, repoFilter, refFilter, rules, earliestTime, latestTime, earliestCommit, latestCommit, whitelistPath, whitelistSecretIDSet, db, log)

    // Create reporter
    reporter := reporterpkg.New(githubAPI, reportDir, db, log)

    search = &Search{
        code:     code,
        finder:   finder,
        reporter: reporter,
        log:      log,
    }
    return
}

func (s *Search) Execute() (err error) {
    s.log.Infof("Preparing repos ... ")
    err = s.code.PrepareRepos()
    if err != nil {
        return
    }

    s.log.Infof("Preparing findings ... ")
    err = s.finder.PrepareFindings()
    if err != nil {
        return
    }

    s.log.Infof("Preparing report ... ")
    err = s.reporter.PrepareReport()
    if err != nil {
        return
    }

    return
}

func buildRefFilter(refs []string) (result *structures.Filter) {
    var values []string
    for _, ref := range refs {
        if ! strings.Contains(ref, "/") {
            ref = "refs/heads/" + ref
        }
        values = append(values, ref)
    }
    return structures.NewFilter(values)
}
