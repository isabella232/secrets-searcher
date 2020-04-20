package app

import (
    "github.com/hako/durafmt"
    codepkg "github.com/pantheon-systems/search-secrets/pkg/code"
    "github.com/pantheon-systems/search-secrets/pkg/database"
    "github.com/pantheon-systems/search-secrets/pkg/dbug"
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    finderpkg "github.com/pantheon-systems/search-secrets/pkg/finder"
    "github.com/pantheon-systems/search-secrets/pkg/logwriter"
    reporterpkg "github.com/pantheon-systems/search-secrets/pkg/reporter"
    "github.com/pantheon-systems/search-secrets/pkg/stats"
    "github.com/pantheon-systems/search-secrets/pkg/structures"
    "github.com/sirupsen/logrus"
    "os"
    "time"
)

type (
    App struct {
        code             *codepkg.Code
        finder           *finderpkg.Finder
        reporter         *reporterpkg.Reporter
        skipSourcePrep   bool
        reportDir        string
        reportArchiveDir string
        codeDir          string
        logFile          string
        nonZero          bool
        db               *database.Database
        logWriter        *logwriter.LogWriter
        log              *logrus.Entry
    }
    Config struct {
        *SourceConfig
        SkipSourcePrep          bool
        Interactive             bool
        SourceDir               string
        OutputDir               string
        Processors              []finderpkg.Processor
        EarliestTime            time.Time
        LatestTime              time.Time
        WhitelistPath           structures.RegexpSet
        WhitelistSecretIDSet    structures.Set
        AppURL                  string
        EnableReportDebugOutput bool
        NonZero                 bool
        ChunkSize               int
        WorkerCount             int
        ShowWorkersBars         bool
        CommitSearchTimeout     time.Duration
        Log                     *logrus.Entry
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
    return build(cfg)
}

func (a *App) Execute() (passed bool, err error) {
    stats.AppStartTime = time.Now()

    // Logging
    if err = a.prepareLogAndWriter(); err != nil {
        err = errors.WithMessage(err, "unable to prepare log")
        return
    }

    // Welcome message
    a.log.Info("=== Search Secrets is starting")
    if dbug.Cnf.Enabled {
        a.log.Info("DEV MODE ENABLED")
    }

    // Code phase
    if !dbug.Cnf.Enabled || dbug.Cnf.EnableCodePhase {
        if err = a.code.PrepareCode(); err != nil {
            err = errors.WithMessage(err, "unable to execute code phase")
            return
        }
        stats.CodePhaseCompleted = true
    }

    // Search phase
    if !dbug.Cnf.Enabled || dbug.Cnf.EnableSearchPhase {
        if err = a.finder.Search(); err != nil {
            err = errors.WithMessage(err, "unable to execute search phase")
            return
        }
        passed = !a.nonZero || stats.SecretsFoundCount == 0
        stats.SearchPhaseCompleted = true
    }

    // Report phase
    if !dbug.Cnf.Enabled || dbug.Cnf.EnableReportPhase {
        if err = a.reporter.PrepareReport(); err != nil {
            err = errors.WithMessage(err, "unable to prepare report")
            return
        }
        stats.ReportPhaseCompleted = true
    }

    stats.AppEndTime = time.Now()

    a.printDoneMessage()

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

    // Log
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

func (a *App) prepareLogAndWriter() (err error) {
    // Log file setup
    if _, statErr := os.Stat(a.logFile); os.IsNotExist(statErr) {
        // Log file does not exist so create one
        // FIXME This step shouldn't be necessary
        var empty *os.File
        empty, err = os.Create(a.logFile)
        if err != nil {
            err = errors.Wrapv(err, "unable to create log file", a.logFile)
            return
        }
        empty.Close()
    } else if statErr == nil {
        // Log file exists so truncate it
        // If you delete it, `tail -f` needs to be restarted
        if err = os.Truncate(a.logFile, 0); err != nil {
            err = errors.Wrapv(err, "unable to truncate log file", a.logFile)
            return
        }
    }

    // Set log writer into logger
    a.log.Logger.SetOutput(a.logWriter)

    return
}
