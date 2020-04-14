package app

import (
    "fmt"
    "github.com/hako/durafmt"
    codepkg "github.com/pantheon-systems/search-secrets/pkg/code"
    "github.com/pantheon-systems/search-secrets/pkg/code/provider"
    "github.com/pantheon-systems/search-secrets/pkg/database"
    "github.com/pantheon-systems/search-secrets/pkg/database/enum/source_provider"
    "github.com/pantheon-systems/search-secrets/pkg/dev"
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    finderpkg "github.com/pantheon-systems/search-secrets/pkg/finder"
    interactpkg "github.com/pantheon-systems/search-secrets/pkg/interact"
    "github.com/pantheon-systems/search-secrets/pkg/logwriter"
    reporterpkg "github.com/pantheon-systems/search-secrets/pkg/reporter"
    "github.com/pantheon-systems/search-secrets/pkg/structures"
    "github.com/sirupsen/logrus"
    "os"
    "path/filepath"
    "strings"
    "time"
)

type (
    App struct {
        SecretCount      int
        code             *codepkg.Code
        finder           *finderpkg.Finder
        reporter         *reporterpkg.Reporter
        tester           *reporterpkg.Tester
        skipSourcePrep   bool
        reportDir        string
        codeDir          string
        startTime        time.Time
        db               *database.Database
        log              *logrus.Logger
        reportArchiveDir string
    }
    Config struct {
        SkipSourcePrep       bool
        Interactive          bool
        SourceDir            string
        OutputDir            string
        Refs         []string
        Processors   []finderpkg.Processor
        EarliestTime time.Time
        LatestTime           time.Time
        WhitelistPath        structures.RegexpSet
        WhitelistSecretIDSet structures.Set
        SkipReportSecrets    bool
        AppURL               string
        SourceConfig         *SourceConfig
        LogWriter            *logwriter.LogWriter
        Log                  *logrus.Logger
    }
    SourceConfig struct {
        SourceProvider string
        GithubToken    string
        Organization   string
        Repos          []string
        ExcludeRepos   []string
        ExcludeForks   bool
        LocalDir       string
    }
)

func New(cfg *Config) (result *App, err error) {
    startTime := time.Now()

    // Directories
    var outputDirAbs string
    outputDirAbs, err = filepath.Abs(cfg.OutputDir)
    if err != nil {
        err = errors.Wrapv(err, "unable to get absolute output dir", cfg.OutputDir)
        return
    }
    codeDir := filepath.Join(outputDirAbs, "code")
    dbDir := filepath.Join(outputDirAbs, "db")
    reportDir := filepath.Join(outputDirAbs, "report")
    reportArchiveDir := fmt.Sprintf("%s-%s", reportDir, startTime.Format("2006-01-02_15-04-05"))

    // Create database
    var db *database.Database
    db, err = database.New(dbDir)
    if err != nil {
        err = errors.WithMessagev(err, "unable to create database object for directory", dbDir)
        return
    }

    // Progress bars, etc
    interact := interactpkg.New(cfg.Interactive, cfg.LogWriter)

    // Create repo provider
    sourceProvider := buildSourceProvider(cfg.SourceConfig, cfg.Log)

    // Create code
    code := codepkg.New(sourceProvider, codeDir, interact, db, cfg.Log)

    // Create finder
    refFilter := buildRefFilter(cfg.Refs)
    finder := finderpkg.New(nil, refFilter, cfg.Processors, cfg.EarliestTime, cfg.LatestTime, cfg.WhitelistPath, cfg.WhitelistSecretIDSet, interact, db, cfg.Log)

    // Create reporter
    reporter := reporterpkg.New(reportDir, reportArchiveDir, cfg.SkipReportSecrets, cfg.AppURL, db, cfg.Log)

    // Create tester
    tester := reporterpkg.NewTester(cfg.Processors, cfg.WhitelistPath, cfg.WhitelistSecretIDSet, db, cfg.Log)

    result = &App{
        code:             code,
        finder:           finder,
        reporter:         reporter,
        tester:           tester,
        skipSourcePrep:   cfg.SkipSourcePrep,
        reportDir:        reportDir,
        reportArchiveDir: reportArchiveDir,
        codeDir:          codeDir,
        startTime:        startTime,
        db:               db,
        log:              cfg.Log,
    }

    return
}

func (a *App) Execute() (err error) {
    if dev.Enabled && dev.EnableTestMode {
        if err = a.executeCodePhase(); err != nil {
            return errors.WithMessage(err, "unable to execute code phase")
        }
        return
    }

    if !dev.Enabled || dev.EnableCodePhase {
        if err = a.executeCodePhase(); err != nil {
            return errors.WithMessage(err, "unable to execute code phase")
        }
    }
    if !dev.Enabled || dev.EnableSearchPhase {
        if err = a.executeSearchPhase(); err != nil {
            return errors.WithMessage(err, "unable to execute search phase")
        }
    }
    if !dev.Enabled || dev.EnableReportPhase {
        if err = a.executeReportPhase(); err != nil {
            return errors.WithMessage(err, "unable to execute reporting phase")
        }
    }

    return
}

func (a *App) executeCodePhase() (err error) {
    if a.skipSourcePrep {
        a.log.Debug("skipping source prep ... ")
        return
    }

    a.log.Debug("resetting filesystem to prepare for code phase ... ")
    if err = a.db.DeleteTableIfExists(database.RepoTable); err != nil {
        return errors.WithMessagev(err, "unable to delete table", database.RepoTable)
    }

    a.log.Info("preparing repos ... ")
    err = a.code.PrepareCode()
    if err != nil {
        return errors.WithMessage(err, "unable to prepare repos")
    }

    return
}

func (a *App) executeSearchPhase() (err error) {
    a.log.Debug("resetting filesystem to prepare for search phase ... ")
    for _, tableName := range []string{database.CommitTable, database.FindingTable, database.SecretTable, database.SecretFindingTable} {
        if err = a.db.DeleteTableIfExists(tableName); err != nil {
            return errors.WithMessagev(err, "unable to delete table", tableName)
        }
    }

    a.log.Info("finding secrets ... ")
    a.SecretCount, err = a.finder.Search()
    if err != nil {
        return errors.WithMessage(err, "unable to prepare findings")
    }

    return
}

func (a *App) executeReportPhase() (err error) {
    a.log.Debug("resetting filesystem to prepare for report phase ... ")
    if err = os.RemoveAll(a.reportDir); err != nil {
        return errors.Wrapv(err, "unable to delete directory", a.reportDir)
    }

    a.log.Info("creating report ... ")
    if err = a.reporter.PrepareReport(); err != nil {
        return errors.WithMessage(err, "unable to prepare report")
    }

    duration := durafmt.ParseShort(time.Now().Sub(a.startTime))
    a.log.Infof("command completed successfully (%s), view report at %s", duration, a.reportDir)

    return
}

func (a *App) executeTestPhase() (err error) {
    a.log.Info("testing existing secrets against processors ... ")
    if err = a.tester.Run(); err != nil {
        return errors.WithMessage(err, "unable to run tester")
    }

    return
}

func buildSourceProvider(sourceConfig *SourceConfig, log *logrus.Logger) (result codepkg.SourceProvider) {
    repoFilter := structures.NewFilter(sourceConfig.Repos, sourceConfig.ExcludeRepos)
    switch sourceConfig.SourceProvider {
    case source_provider.GitHub{}.New().Value():
        result = provider.NewGithubProvider(sourceConfig.GithubToken, sourceConfig.Organization, repoFilter, sourceConfig.ExcludeForks, log)
    case source_provider.Local{}.New().Value():
        result = provider.NewLocalProvider(sourceConfig.LocalDir, repoFilter, log)
    }
    return
}

func buildRefFilter(refs []string) (result *structures.Filter) {
    var values []string
    for _, ref := range refs {
        if !strings.Contains(ref, "/") {
            ref = "refs/heads/" + ref
        }
        values = append(values, ref)
    }
    return structures.NewFilter(values, nil)
}
