package build

import (
	"path/filepath"

	"github.com/pantheon-systems/search-secrets/pkg/app/config"
	"github.com/pantheon-systems/search-secrets/pkg/app/vars"
	"github.com/pantheon-systems/search-secrets/pkg/database"
	"github.com/pantheon-systems/search-secrets/pkg/dev"
	"github.com/pantheon-systems/search-secrets/pkg/errors"
	gitpkg "github.com/pantheon-systems/search-secrets/pkg/git"
	interactpkg "github.com/pantheon-systems/search-secrets/pkg/interact"
	"github.com/pantheon-systems/search-secrets/pkg/logg"
	reporterpkg "github.com/pantheon-systems/search-secrets/pkg/reporter"
	searchpkg "github.com/pantheon-systems/search-secrets/pkg/search"
	sourcepkg "github.com/pantheon-systems/search-secrets/pkg/source"
)

type AppParams struct {
	OutputDir         string
	LogFile           string
	NonZero           bool
	EnableSourcePhase bool
	EnableSearchPhase bool
	EnableReportPhase bool
	Source            *sourcepkg.Source
	Search            *searchpkg.Search
	Reporter          *reporterpkg.Reporter
	DB                *database.Database
	AppLog            logg.Logg
}

func App(appCfg *config.AppConfig) (result *AppParams, err error) {
	// Set dev params singleton
	dev.Params = Dev(&appCfg.DevConfig)

	// Files/Dirs
	outputDir, _ := filepath.Abs(appCfg.OutputDir)
	sourceDir := filepath.Join(outputDir, "source")
	dbDir := filepath.Join(outputDir, "db")
	logFile := filepath.Join(outputDir, "run.log")

	// Filters
	repoFilter := buildRepoFilter(&appCfg.SourceConfig)
	commitFilter := buildCommitFilter(&appCfg.SearchConfig)
	fileChangeFilter := buildFileChangeFilter(&appCfg.SearchConfig)
	secretIDFilter := buildSecretIDFilter(&appCfg.SearchConfig)

	// Init logger
	var initLog *logg.LogrusLogg
	if initLog, err = buildInitLog(appCfg.LogLevel); err != nil {
		err = errors.WithMessage(err, "unable to build logger")
		return
	}
	initLog = initLog.WithPrefix("init").(*logg.LogrusLogg)

	// App loggers
	var appLog logg.Logg
	if appLog, err = buildAppLog(initLog, logFile); err != nil {
		err = errors.WithMessage(err, "unable to build logger")
		return
	}
	dbLog := appLog.WithPrefix("db")
	gitLog := appLog.WithPrefix("git")
	interactLog := appLog.WithPrefix("interact")
	sourceLog := appLog.WithPrefix("source")
	searchLog := appLog.WithPrefix("search")
	reporterLog := appLog.WithPrefix("report")

	// Database
	var db *database.Database
	db, err = database.New(dbDir, dbLog)
	if err != nil {
		err = errors.Wrapv(err, "unable to build database for directory", dbDir)
		return
	}

	// Git service
	git := gitpkg.New(gitLog)

	// Interact service
	interact := interactpkg.New(appCfg.Interactive, interactLog)

	// Source provider
	sourceProvider := buildSourceProvider(&appCfg.SourceConfig, git, sourceLog)

	// Source service
	source := Source(sourceDir, appCfg.SourceConfig.SkipFetch, repoFilter, git, sourceProvider, interact, db, sourceLog)

	// Search service
	var search *searchpkg.Search
	if search, err = Search(
		&appCfg.SearchConfig,
		repoFilter,
		sourceDir,
		commitFilter,
		fileChangeFilter,
		git,
		interact,
		db,
		searchLog,
	); err != nil {
		err = errors.WithMessage(err, "unable to build search")
	}

	// Reporter service
	reporter := Reporter(&appCfg.ReporterConfig, outputDir, vars.URL, sourceProvider, secretIDFilter, db, reporterLog)

	// Build app params
	result = &AppParams{
		OutputDir:         outputDir,
		LogFile:           logFile,
		NonZero:           appCfg.NonZero,
		EnableSourcePhase: appCfg.EnableSourcePhase,
		EnableSearchPhase: appCfg.EnableSearchPhase,
		EnableReportPhase: appCfg.EnableReportPhase,
		Source:            source,
		Search:            search,
		Reporter:          reporter,
		DB:                db,
		AppLog:            appLog,
	}

	return
}
