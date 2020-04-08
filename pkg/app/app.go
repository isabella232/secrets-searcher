package app

import (
    codepkg "github.com/pantheon-systems/search-secrets/pkg/code"
    "github.com/pantheon-systems/search-secrets/pkg/database"
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    finderpkg "github.com/pantheon-systems/search-secrets/pkg/finder"
    "github.com/pantheon-systems/search-secrets/pkg/finder/comb"
    "github.com/pantheon-systems/search-secrets/pkg/finder/rule"
    "github.com/pantheon-systems/search-secrets/pkg/github"
    reporterpkg "github.com/pantheon-systems/search-secrets/pkg/reporter"
    "github.com/pantheon-systems/search-secrets/pkg/structures"
    "github.com/sirupsen/logrus"
    "os"
    "path/filepath"
    "strings"
    "time"
)

type App struct {
    code           *codepkg.Code
    finder         *finderpkg.Finder
    reporter       *reporterpkg.Reporter
    skipSourcePrep bool
    reportDir      string
    codeDir      string
    db             *database.Database
    log            *logrus.Logger
}

func New(skipSourcePrep bool, githubToken, organization, outputDir string, repos, refs []string, rules []*rule.Rule, earliestTime, latestTime time.Time, earliestCommit, latestCommit string, whitelistPath structures.RegexpSet, whitelistSecretIDSet structures.Set, log *logrus.Logger) (search *App, err error) {

    // Directories
    var outputDirAbs string
    outputDirAbs, err = filepath.Abs(outputDir)
    if err != nil {
        err = errors.Wrapv(err, "unable to get absolute output dir", outputDir)
        return
    }
    codeDir := filepath.Join(outputDirAbs, "code")
    dbDir := filepath.Join(outputDirAbs, "db")
    reportDir := filepath.Join(outputDirAbs, "report")

    // Create database
    var db *database.Database
    db, err = database.New(dbDir)
    if err != nil {
        err = errors.WithMessagev(err, "unable to create database object for directory", dbDir)
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
    reporter := reporterpkg.New(reportDir, db, log)

    search = &App{
        code:           code,
        finder:         finder,
        reporter:       reporter,
        skipSourcePrep: skipSourcePrep,
        reportDir: reportDir,
        codeDir: codeDir,
        db:             db,
        log:            log,
    }

    return
}

func (a *App) Execute() (err error) {
    a.log.Info("deleting existing output data ... ")
    if err = a.resetOutputDir(); err != nil {
        return errors.WithMessage(err, "unable to reset output dir")
    }

    if a.skipSourcePrep {
        a.log.Info("skipping source prep ... ")
    } else {
        a.log.Info("preparing repos ... ")
        err = a.code.PrepareCode()
        if err != nil {
            return errors.WithMessage(err, "unable to prepare repos")
        }
    }

    a.log.Info("preparing findings ... ")
    if err = a.finder.PrepareFindings(); err != nil {
        return errors.WithMessage(err, "unable to prepare findings")
    }

    a.log.Info("preparing report ... ")
    if err = a.reporter.PrepareReport(); err != nil {
        return errors.WithMessage(err, "unable to prepare report")
    }

    return
}

func (a *App) resetOutputDir() (err error) {
    for _, tableName := range []string{database.CommitTable, database.FindingTable, database.SecretTable, database.SecretFindingTable} {
        if err = a.db.DeleteTableIfExists(tableName); err != nil {
            return errors.WithMessagev(err, "unable to delete table", tableName)
        }
    }
    if err = os.RemoveAll(a.reportDir); err != nil {
        return errors.Wrapv(err, "unable to delete directory", a.reportDir)
    }
    if ! a.skipSourcePrep {
        if err = a.db.DeleteTableIfExists(database.RepoTable); err != nil {
            return errors.WithMessagev(err, "unable to delete table", database.RepoTable)
        }
        if err = os.RemoveAll(a.codeDir); err != nil {
            return errors.Wrapv(err, "unable to delete directory", a.codeDir)
        }
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
