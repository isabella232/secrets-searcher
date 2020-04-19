package app

import (
    "fmt"
    "github.com/hako/durafmt"
    "github.com/pantheon-systems/search-secrets/pkg/app/source_provider"
    codepkg "github.com/pantheon-systems/search-secrets/pkg/code"
    "github.com/pantheon-systems/search-secrets/pkg/code/provider"
    "github.com/pantheon-systems/search-secrets/pkg/database"
    "github.com/pantheon-systems/search-secrets/pkg/dbug"
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    finderpkg "github.com/pantheon-systems/search-secrets/pkg/finder"
    "github.com/pantheon-systems/search-secrets/pkg/git"
    interactpkg "github.com/pantheon-systems/search-secrets/pkg/interact"
    "github.com/pantheon-systems/search-secrets/pkg/logwriter"
    reporterpkg "github.com/pantheon-systems/search-secrets/pkg/reporter"
    "github.com/pantheon-systems/search-secrets/pkg/stats"
    "github.com/pantheon-systems/search-secrets/pkg/structures"
    "github.com/sirupsen/logrus"
    "os"
    "path/filepath"
    "time"
)

type (
    App struct {
        code             *codepkg.Code
        finder           *finderpkg.Finder
        reporter         *reporterpkg.Reporter
        skipSourcePrep   bool
        reportDir        string
        codeDir          string
        db               *database.Database
        log              logrus.FieldLogger
        reportArchiveDir string
    }
    Config struct {
        SkipSourcePrep          bool
        Interactive             bool
        SourceDir               string
        OutputDir               string
        Refs                    []string
        Processors              []finderpkg.Processor
        EarliestTime            time.Time
        LatestTime              time.Time
        WhitelistPath           structures.RegexpSet
        WhitelistSecretIDSet    structures.Set
        AppURL                  string
        EnableReportDebugOutput bool
        SourceConfig            *SourceConfig
        LogWriter               *logwriter.LogWriter
        Log                     logrus.FieldLogger
        ChunkSize               int
        WorkerCount             int
        CommitSearchTimeout     time.Duration
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
    log := cfg.Log

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

    // Filters
    repoFilter := structures.NewFilter(cfg.SourceConfig.Repos, cfg.SourceConfig.ExcludeRepos)
    commitFilter := &git.CommitFilter{
        Hashes:               structures.NewSet(nil),
        EarliestTime:         cfg.EarliestTime,
        LatestTime:           cfg.LatestTime,
        ExcludeNoDiffCommits: true,
    }
    fileChangeFilter := &git.FileChangeFilter{
        IncludeMatchingPaths:         nil,
        ExcludeFileDeletions:         true,
        ExcludeMatchingPaths:         cfg.WhitelistPath,
        ExcludeBinaryOrEmpty:         true,
        ExcludeOnesWithNoCodeChanges: true,
    }

    // Debug
    if dbug.Cnf.Enabled {
        if dbug.Cnf.Filter.Repo != "" {
            repoFilter = structures.NewFilter([]string{dbug.Cnf.Filter.Repo}, nil)
        }
        if dbug.Cnf.Filter.Commit != "" {
            commitFilter.Hashes = structures.NewSet([]string{dbug.Cnf.Filter.Commit})
        }
        if dbug.Cnf.Filter.Path != "" {
            fileChangeFilter.IncludeMatchingPaths = structures.NewRegexpSetFromStringsMustCompile([]string{dbug.Cnf.Filter.Path})
        }
        cfg.Interactive = dbug.Cnf.EnableInteract
        cfg.EnableReportDebugOutput = true
    }

    // Progress bars, etc
    interact := interactpkg.New(cfg.Interactive, cfg.LogWriter, log.WithField("prefix", "interact"))

    // Create repo provider
    sourceProvider := buildSourceProvider(cfg.SourceConfig,
        repoFilter,
        log.WithField("prefix", "source"),
    )

    // Create code
    code := codepkg.New(
        sourceProvider,
        repoFilter,
        codeDir,
        interact,
        db,
        log.WithField("prefix", "code"),
    )

    // Create finder
    finder := finderpkg.New(
        repoFilter,
        commitFilter,
        fileChangeFilter,
        cfg.ChunkSize,
        cfg.WorkerCount,
        cfg.CommitSearchTimeout,
        cfg.Processors,
        cfg.WhitelistSecretIDSet,
        interact,
        db,
        log.WithField("prefix", "search"),
    )

    // Create reporter
    reporter := reporterpkg.New(reportDir, reportArchiveDir, cfg.AppURL, cfg.EnableReportDebugOutput, db, log.WithField("prefix", "report"), )

    result = &App{
        code:             code,
        finder:           finder,
        reporter:         reporter,
        skipSourcePrep:   cfg.SkipSourcePrep,
        reportDir:        reportDir,
        reportArchiveDir: reportArchiveDir,
        codeDir:          codeDir,
        db:               db,
        log:              log.WithField("prefix", "app"),
    }

    return
}

func (a *App) Execute() (err error) {
    stats.AppStartTime = time.Now()

    if !dbug.Cnf.Enabled || dbug.Cnf.EnableCodePhase {
        if err = a.executeCodePhase(); err != nil {
            return errors.WithMessage(err, "unable to execute code phase")
        }
        stats.CodePhaseCompleted = true
    }
    if !dbug.Cnf.Enabled || dbug.Cnf.EnableSearchPhase {
        if err = a.executeSearchPhase(); err != nil {
            return errors.WithMessage(err, "unable to execute search phase")
        }
        stats.SearchPhaseCompleted = true
    }
    if !dbug.Cnf.Enabled || dbug.Cnf.EnableReportPhase {
        if err = a.executeReportPhase(); err != nil {
            return errors.WithMessage(err, "unable to execute reporting phase")
        }
        stats.ReportPhaseCompleted = true
    }

    stats.AppEndTime = time.Now()

    a.printDoneMessage()

    return
}

func (a *App) executeCodePhase() (err error) {
    if a.skipSourcePrep {
        a.log.Debug("skipping source prep ... ")
        return
    }

    err = a.code.PrepareCode()
    if err != nil {
        return errors.WithMessage(err, "unable to prepare repos")
    }

    return
}

func (a *App) executeSearchPhase() (err error) {
    a.log.Debug("resetting filesystem to prepare for search phase ... ")
    searchTables := []string{
        database.CommitTable,
        database.FindingTable,
        database.FindingExtrasTable,
        database.SecretTable,
        database.SecretExtrasTable,
    }
    for _, tableName := range searchTables {
        if err = a.db.DeleteTableIfExists(tableName); err != nil {
            return errors.WithMessagev(err, "unable to delete table", tableName)
        }
    }

    if err = a.finder.Search(); err != nil {
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

    return
}

func (a *App) printDoneMessage() {
    // Execution duration
    duration := stats.SearchEndTime.Sub(stats.SearchStartTime)
    durationHuman := durafmt.ParseShort(duration)

    // Duration per commit
    var commitDuration time.Duration
    if stats.CommitsSearchedCount != 0 {
        commitDuration = time.Duration(int64(duration) / stats.CommitsSearchedCount)
    }

    a.log.Info("Command completed successfully")
    if stats.SearchPhaseCompleted {
        a.log.Infof("- Secrets found:       %d", stats.SecretsFoundCount)
    }
    if stats.ReportPhaseCompleted {
        a.log.Infof("- Report location:     %s", a.reportDir)
    }
    if true || dbug.Cnf.Enabled {
        if stats.SearchPhaseCompleted {
            a.log.Infof("- Search duration:     %.2fs (%s)", duration.Seconds(), durationHuman)
            a.log.Infof("- Commits searched:    %d", stats.CommitsSearchedCount)
            a.log.Infof("- Duration per commit: %dms (%dns)",
                commitDuration.Milliseconds(), commitDuration.Nanoseconds())
        }
    }
}

func buildSourceProvider(sourceConfig *SourceConfig, repoFilter *structures.Filter, log logrus.FieldLogger) (result codepkg.SourceProvider) {
    switch sourceConfig.SourceProvider {
    case source_provider.GitHub{}.New().Value():
        result = provider.NewGithubProvider(
            sourceConfig.SourceProvider,
            sourceConfig.GithubToken,
            sourceConfig.Organization,
            repoFilter,
            sourceConfig.ExcludeForks,
            log,
        )
    case source_provider.Local{}.New().Value():
        result = provider.NewLocalProvider(
            sourceConfig.SourceProvider,
            sourceConfig.LocalDir,
            repoFilter,
            log,
        )
    }
    return
}
