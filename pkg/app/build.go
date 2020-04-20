package app

import (
    "context"
    "github.com/google/go-github/v29/github"
    scribble "github.com/nanobox-io/golang-scribble"
    "github.com/pantheon-systems/search-secrets/pkg/app/source_provider"
    codepkg "github.com/pantheon-systems/search-secrets/pkg/code"
    "github.com/pantheon-systems/search-secrets/pkg/code/provider"
    "github.com/pantheon-systems/search-secrets/pkg/database"
    "github.com/pantheon-systems/search-secrets/pkg/dbug"
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    finderpkg "github.com/pantheon-systems/search-secrets/pkg/finder"
    gitpkg "github.com/pantheon-systems/search-secrets/pkg/git"
    interactpkg "github.com/pantheon-systems/search-secrets/pkg/interact"
    "github.com/pantheon-systems/search-secrets/pkg/logwriter"
    reporterpkg "github.com/pantheon-systems/search-secrets/pkg/reporter"
    "github.com/pantheon-systems/search-secrets/pkg/structures"
    "github.com/sirupsen/logrus"
    "golang.org/x/oauth2"
    "path/filepath"
)

func build(cfg *Config) (result *App, err error) {
    log := cfg.Log

    // Files/Dirs
    outputDirAbs, _ := filepath.Abs(cfg.OutputDir)
    codeDir := filepath.Join(outputDirAbs, "code")
    dbDir := filepath.Join(outputDirAbs, "db")
    reportDir := filepath.Join(outputDirAbs, "report")
    reportArchivesDir := filepath.Join(outputDirAbs, "report-archive")
    logFile := filepath.Join(outputDirAbs, "run.log")

    // Filters
    var repoFilter *structures.Filter
    var commitFilter *gitpkg.CommitFilter
    var fileChangeFilter *gitpkg.FileChangeFilter
    repoFilter, commitFilter, fileChangeFilter = buildFilters(cfg)

    // Create database
    var db *database.Database
    db, err = buildDatabase(dbDir, log)
    if err != nil {
        err = errors.Wrapv(err, "unable to build database for directory", dbDir)
        return
    }

    // Services
    logWriter := logwriter.New(logFile)
    interact := interactpkg.New(cfg.Interactive, logWriter, log.WithField("prefix", "interact"))
    code := buildCode(cfg, codeDir, repoFilter, interact, db, log)
    finder := buildFinder(cfg, repoFilter, commitFilter, fileChangeFilter, interact, db, log)
    reporter := buildReporter(cfg, reportDir, reportArchivesDir, db, log)

    // Build app
    result = &App{
        code:             code,
        finder:           finder,
        reporter:         reporter,
        skipSourcePrep:   cfg.SkipSourcePrep,
        reportDir:        reportDir,
        reportArchiveDir: reportArchivesDir,
        logFile:          logFile,
        codeDir:          codeDir,
        nonZero:          cfg.NonZero,
        db:               db,
        logWriter:        logWriter,
        log:              log.WithField("prefix", "app"),
    }

    return
}

//func NewCommitFilter(
//    hashes structures.Set,
//    earliestTime time.Time,
//    latestTime time.Time,
//    earliestCommit string,
//    latestCommit string,
//    excludeNoDiffCommits bool,
//) (result *CommitFilter) {
//    return &CommitFilter{
//        Hashes:               structures.Set{},
//        EarliestTime:         time.Time{},
//        LatestTime:           time.Time{},
//        LatestTimeSet:        false,
//        EarliestCommit:       "",
//        LatestCommit:         "",
//        ExcludeNoDiffCommits: false,
//    }
//}

func buildFilters(cfg *Config) (repoFilter *structures.Filter, commitFilter *gitpkg.CommitFilter, fileChangeFilter *gitpkg.FileChangeFilter) {
    repoFilter = structures.NewFilter(
        cfg.SourceConfig.Repos,
        cfg.SourceConfig.ExcludeRepos,
    )
    commitFilter = gitpkg.NewCommitFilter(
        structures.NewSet(nil),
        cfg.EarliestTime,
        cfg.LatestTime,
        "",
        "",
        true,
    )
    fileChangeFilter = &gitpkg.FileChangeFilter{
        IncludeMatchingPaths:         nil,
        ExcludeFileDeletions:         true,
        ExcludeMatchingPaths:         cfg.WhitelistPath,
        ExcludeBinaryOrEmpty:         true,
        ExcludeOnesWithNoCodeChanges: true,
    }

    // Debug
    if dbug.Cnf.Enabled {
        if dbug.Cnf.FilterConfig.Repo != "" {
            repoFilter = structures.NewFilter([]string{dbug.Cnf.FilterConfig.Repo}, nil)
        }
        if dbug.Cnf.FilterConfig.Commit != "" {
            commitFilter.Hashes = structures.NewSet([]string{dbug.Cnf.FilterConfig.Commit})
        }
        if dbug.Cnf.FilterConfig.Path != "" {
            fileChangeFilter.IncludeMatchingPaths = structures.NewRegexpSetFromStringsMustCompile([]string{dbug.Cnf.FilterConfig.Path})
        }
        cfg.Interactive = dbug.Cnf.EnableInteract
        cfg.EnableReportDebugOutput = true
    }

    return
}

func buildDatabase(dbDir string, log logrus.FieldLogger) (result *database.Database, err error) {

    // Scribble
    var dbDriver *scribble.Driver
    dbDriver, err = scribble.New(dbDir, nil)
    if err != nil {
        err = errors.Wrapv(err, "unable to scribble driver for directory", dbDir)
        return
    }

    // Create database
    result = database.New(dbDir, dbDriver, log)

    return
}

func buildCode(cfg *Config, codeDir string, repoFilter *structures.Filter, interact interactpkg.Interactish, db *database.Database, log logrus.FieldLogger) (result *codepkg.Code) {

    // Create repo provider
    sourceProvider := buildSourceProvider(
        cfg.SourceConfig,
        repoFilter,
        log.WithField("prefix", "source"),
    )

    // Create code
    result = codepkg.New(
        sourceProvider,
        repoFilter,
        codeDir,
        interact,
        cfg.SkipSourcePrep,
        db,
        log.WithField("prefix", "code"),
    )

    return
}

func buildFinder(cfg *Config, repoFilter *structures.Filter, commitFilter *gitpkg.CommitFilter, fileChangeFilter *gitpkg.FileChangeFilter, interact interactpkg.Interactish, db *database.Database, log logrus.FieldLogger) (result *finderpkg.Finder) {

    git := gitpkg.New(log.WithField("prefix", "git"))

    // Search builder
    searchBuilder := finderpkg.NewSearchBuilder(
        git,
        repoFilter,
        commitFilter,
        fileChangeFilter,
        cfg.CommitSearchTimeout,
        cfg.Processors,
        cfg.ChunkSize,
        cfg.WorkerCount,
        cfg.ShowWorkersBars,
        interact,
        db,
        log.WithField("prefix", "search"),
    )

    // Search writer
    writer := finderpkg.NewWriter(
        cfg.WhitelistSecretIDSet,
        db,
        log.WithField("prefix", "search"),
    )

    // Create finder
    return finderpkg.New(
        writer,
        searchBuilder,
        cfg.WorkerCount,
        interact,
        log.WithField("prefix", "search"),
    )
}

func buildReporter(cfg *Config, reportDir, reportArchivesDir string, db *database.Database, log logrus.FieldLogger) *reporterpkg.Reporter {
    return reporterpkg.New(
        reportDir,
        reportArchivesDir,
        cfg.AppURL,
        cfg.EnableReportDebugOutput,
        db,
        log.WithField("prefix", "report"),
    )
}

func buildSourceProvider(sourceConfig *SourceConfig, repoFilter *structures.Filter, log logrus.FieldLogger) (result codepkg.SourceProvider) {
    switch sourceConfig.SourceProvider {

    // GitHub source provider
    case source_provider.GitHub{}.New().Value():
        // Client
        ctx := context.Background()
        tc := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: sourceConfig.GithubToken}))
        client := github.NewClient(tc)

        // Provider
        result = provider.NewGithubProvider(
            sourceConfig.SourceProvider,
            sourceConfig.Organization,
            client,
            repoFilter,
            sourceConfig.ExcludeForks,
            log,
        )

    // Local source provider
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
