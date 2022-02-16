package app

import (
	"os"
	"time"

	va "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/hako/durafmt"
	"github.com/pantheon-systems/secrets-searcher/pkg/app/build"
	"github.com/pantheon-systems/secrets-searcher/pkg/app/config"
	"github.com/pantheon-systems/secrets-searcher/pkg/database"
	"github.com/pantheon-systems/secrets-searcher/pkg/errors"
	"github.com/pantheon-systems/secrets-searcher/pkg/logg"
	reporterpkg "github.com/pantheon-systems/secrets-searcher/pkg/reporter"
	searchpkg "github.com/pantheon-systems/secrets-searcher/pkg/search"
	sourcepkg "github.com/pantheon-systems/secrets-searcher/pkg/source"
	"github.com/pantheon-systems/secrets-searcher/pkg/stats"
)

type App struct {
	outputDir       string
	logFile         string
	nonZero         bool
	enableProfiling bool

	enableSourcePhase bool
	enableSearchPhase bool
	enableReportPhase bool

	sourcePhaseCompleted bool
	searchPhaseCompleted bool
	reportPhaseCompleted bool

	source   *sourcepkg.Source
	search   *searchpkg.Search
	reporter *reporterpkg.Reporter
	stats    *stats.Stats
	db       *database.Database
	log      logg.Logg
}

func New(appCfg *config.AppConfig) (a *App, err error) {

	// Validate config
	if err = va.Validate(appCfg); err != nil {
		err = errors.WithMessage(err, "invalid configuration")
		return
	}

	var params *build.AppParams
	params, err = build.App(appCfg)
	if err != nil {
		err = errors.WithMessage(err, "unable to build app")
		return
	}

	a = &App{
		outputDir:         params.OutputDir,
		logFile:           params.LogFile,
		nonZero:           params.NonZero,
		enableSourcePhase: params.EnableSourcePhase,
		enableSearchPhase: params.EnableSearchPhase,
		enableReportPhase: params.EnableReportPhase,
		enableProfiling:   params.EnableProfiling,
		source:            params.Source,
		search:            params.Search,
		reporter:          params.Reporter,
		stats:             params.Stats,
		db:                params.DB,
		log:               params.AppLog,
	}

	return
}

func (a *App) Execute() (passed bool, err error) {
	passed = true
	a.stats.AppStartTime = time.Now()

	// Create output directory
	if err = os.MkdirAll(a.outputDir, 0700); err != nil {
		err = errors.Wrapv(err, "unable to create output directory", a.outputDir)
		return
	}

	// Truncate log file if it exists (if you delete it, `tail -f` needs to be restarted)
	if _, statErr := os.Stat(a.logFile); statErr == nil {
		if err = os.Truncate(a.logFile, 0); err != nil {
			err = errors.Wrapv(err, "unable to truncate log file", a.logFile)
			return
		}
	}

	// Welcome message
	a.log.Info("=== Search Secrets is starting")

	// Create database
	if err = a.db.PrepareFilesystemForWriting(); err != nil {
		err = errors.WithMessage(err, "unable to prepare filesystem for database")
		return
	}

	// Source phase
	if a.enableSourcePhase {
		if err = a.source.PrepareSource(); err != nil {
			err = errors.WithMessage(err, "unable to execute source phase")
			return
		}
		a.sourcePhaseCompleted = true
	}

	// Search phase
	if a.enableSearchPhase {

		// Reset search tables
		if err = a.db.DeleteSearchTables(); err != nil {
			err = errors.WithMessage(err, "unable to delete search tables")
		}

		// Pre reporting
		if a.enableReportPhase {

			// Prepare fs (reset ./output/report, create ./output/report-archive)
			if err = a.reporter.PrepareFilesystem(); err != nil {
				err = errors.WithMessage(err, "unable to prepare filesystem for pre reporting")
				return
			}

			a.reporter.RunPreReporting()
		}

		// Searching
		if err = a.search.Search(); err != nil {
			err = errors.WithMessage(err, "unable to execute search phase")
			return
		}
		passed = !a.nonZero || a.stats.SecretsFoundCount == 0
		a.searchPhaseCompleted = true
	}

	// Report phase
	if a.enableReportPhase {

		// Prepare fs (reset ./output/report, create ./output/report-archive)
		if err = a.reporter.PrepareFilesystem(); err != nil {
			err = errors.WithMessage(err, "unable to prepare filesystem for pre reporting")
			return
		}

		if err = a.reporter.PrepareFinalReport(); err != nil {
			err = errors.WithMessage(err, "unable to prepare report")
			return
		}
		a.reportPhaseCompleted = true
	}

	a.stats.AppEndTime = time.Now()

	a.printDoneMessage()

	return
}

func (a *App) printDoneMessage() {
	// Execution duration
	duration := a.stats.SearchEndTime.Sub(a.stats.SearchStartTime)
	durationHuman := durafmt.ParseShort(duration)

	// Duration per commit
	var commitDuration time.Duration
	if a.stats.CommitsSearchedCount != 0 {
		commitDuration = time.Duration(int64(duration) / a.stats.CommitsSearchedCount)
	}

	if a.enableProfiling {
		a.log.Info("Long running repos:")
		a.logDurationStats(a.stats.RepoDurations.Stats())
		a.log.Info("Long running commits:")
		a.logDurationStats(a.stats.CommitDurations.Stats())
		a.log.Info("Long running file changes:")
		a.logDurationStats(a.stats.FileChangeDurations.Stats())
		a.log.Info("Long running file types:")
		a.logDurationStats(a.stats.FileTypeDurations.Stats())
	}

	// Log
	a.log.Info("Command completed successfully")
	if a.searchPhaseCompleted {
		a.log.Infof("- Secrets found:       %d", a.stats.SecretsFoundCount)
	}
	if a.reportPhaseCompleted {
		a.log.Infof("- Report location:     %s", a.reporter.ReportDir)
		a.log.Infof("- Report archive:      %s", a.reporter.ReportArchivesDir)
	}
	if a.searchPhaseCompleted {
		a.log.Infof("- Search duration:     %.2fs (%s)", duration.Seconds(), durationHuman)
		a.log.Infof("- Commits searched:    %d", a.stats.CommitsSearchedCount)
		a.log.Infof("- Duration per commit: %dms (%dns)",
			commitDuration.Milliseconds(), commitDuration.Nanoseconds())
	}
}

func (a *App) logDurationStats(stats []*stats.DurationStat) {
	if stats == nil {
		a.log.Info("- [none]")
	}
	for _, stat := range stats {
		a.log.Infof("- %-20s %s", stat.Dur, stat.Item)
	}
}
