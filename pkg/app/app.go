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
        log              *logrus.Logger
        reportArchiveDir string
        Stats            *Stats
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
        SkipReportSecrets       bool
        AppURL                  string
        EnableReportDebugOutput bool
        SourceConfig            *SourceConfig
        LogWriter               *logwriter.LogWriter
        Log                     *logrus.Logger
        ChunkSize               int
        WorkerCount             int
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
    Stats struct {
        CommitsSearchedCount int64
        SecretsFoundCount    int64
        ExecutionStartTime   time.Time
        ExecutionEndTime     time.Time
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
    logEntry := logrus.NewEntry(cfg.Log)

    // Progress bars, etc
    interact := interactpkg.New(cfg.Interactive, cfg.LogWriter, logEntry)

    // Create repo provider
    sourceProvider := buildSourceProvider(cfg.SourceConfig, repoFilter, cfg.Log)

    // Create code
    code := codepkg.New(sourceProvider, repoFilter, codeDir, interact, db, logEntry)

    // Create finder
    finder := finderpkg.New(repoFilter, commitFilter, fileChangeFilter, cfg.ChunkSize, cfg.WorkerCount, cfg.Processors, cfg.WhitelistSecretIDSet, interact, db, logEntry)

    // Create reporter
    reporter := reporterpkg.New(reportDir, reportArchiveDir, cfg.SkipReportSecrets, cfg.AppURL, cfg.EnableReportDebugOutput, db, cfg.Log)

    result = &App{
        code:             code,
        finder:           finder,
        reporter:         reporter,
        skipSourcePrep:   cfg.SkipSourcePrep,
        reportDir:        reportDir,
        reportArchiveDir: reportArchiveDir,
        codeDir:          codeDir,
        db:               db,
        log:              cfg.Log,
        Stats:            &Stats{},
    }

    return
}

func (a *App) Execute() (err error) {
    a.Stats = &Stats{}
    a.Stats.ExecutionStartTime = time.Now()

    if !dbug.Cnf.Enabled || dbug.Cnf.EnableCodePhase {
        if err = a.executeCodePhase(); err != nil {
            return errors.WithMessage(err, "unable to execute code phase")
        }
    }
    if !dbug.Cnf.Enabled || dbug.Cnf.EnableSearchPhase {
        if err = a.executeSearchPhase(); err != nil {
            return errors.WithMessage(err, "unable to execute search phase")
        }
    }
    if !dbug.Cnf.Enabled || dbug.Cnf.EnableReportPhase {
        if err = a.executeReportPhase(); err != nil {
            return errors.WithMessage(err, "unable to execute reporting phase")
        }
    }

    a.Stats.ExecutionEndTime = time.Now()

    a.printDoneMessage()

    return
}

func (a *App) executeCodePhase() (err error) {
    if a.skipSourcePrep {
        a.log.Debug("skipping source prep ... ")
        return
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

    a.log.Info("finding secrets ... ")
    if err = a.finder.Search(); err != nil {
        return errors.WithMessage(err, "unable to prepare findings")
    }

    // Stats
    a.Stats.CommitsSearchedCount = a.finder.Stats.CommitsSearchedCount
    a.Stats.SecretsFoundCount = a.finder.Stats.SecretsFoundCount

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
    duration := a.Stats.ExecutionEndTime.Sub(a.Stats.ExecutionStartTime)
    durationHuman := durafmt.ParseShort(duration)

    // Duration per commit
    var commitDuration time.Duration
    commitDuration = time.Duration(int64(duration) / a.Stats.CommitsSearchedCount)

    a.log.Info("Command completed successfully")
    a.log.Infof("- Secrets found:       %d", a.Stats.SecretsFoundCount)
    a.log.Infof("- Report location:     %s", a.reportDir)
    if true || dbug.Cnf.Enabled {
        a.log.Infof("- Commits searched:    %d", a.Stats.CommitsSearchedCount)
        a.log.Infof("- Total duration:      %.2fs (%s)", duration.Seconds(), durationHuman)
        a.log.Infof("- Duration per commit: %dms (%dns)",
            commitDuration.Milliseconds(), commitDuration.Nanoseconds())
    }
}

func buildSourceProvider(sourceConfig *SourceConfig, repoFilter *structures.Filter, log *logrus.Logger) (result codepkg.SourceProvider) {
    switch sourceConfig.SourceProvider {
    case source_provider.GitHub{}.New().Value():
        result = provider.NewGithubProvider(source_provider.GitHub{}.New().Value(), sourceConfig.GithubToken, sourceConfig.Organization, repoFilter, sourceConfig.ExcludeForks, log)
    case source_provider.Local{}.New().Value():
        result = provider.NewLocalProvider(source_provider.Local{}.New().Value(), sourceConfig.LocalDir, repoFilter, log)
    }
    return
}
