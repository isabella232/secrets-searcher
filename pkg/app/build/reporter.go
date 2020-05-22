package build

import (
	"path/filepath"

	"github.com/pantheon-systems/search-secrets/pkg/stats"

	"github.com/pantheon-systems/search-secrets/pkg/manip"

	"github.com/pantheon-systems/search-secrets/pkg/app/config"
	"github.com/pantheon-systems/search-secrets/pkg/database"
	"github.com/pantheon-systems/search-secrets/pkg/logg"
	reporterpkg "github.com/pantheon-systems/search-secrets/pkg/reporter"
	"github.com/pantheon-systems/search-secrets/pkg/source"
)

func Reporter(reporterCfg *config.ReportConfig, outputDir, url string, sourceProvider source.ProviderI, secretIDFilter *manip.SliceFilter, stats *stats.Stats, db *database.Database, log logg.Logg) *reporterpkg.Reporter {
	reportDir := reporterCfg.ReportDir
	if reportDir == "" {
		reportDir = filepath.Join(outputDir, "report")
	}
	reportArchivesDir := reporterCfg.ReportArchivesDir
	if reportArchivesDir == "" {
		reportArchivesDir = filepath.Join(outputDir, "report-archive")
	}

	return reporterpkg.New(
		reportDir,
		reportArchivesDir,
		url,
		reporterCfg.ShowDebugOutput,
		reporterCfg.EnablePreReports,
		reporterCfg.PreReportInterval,
		secretIDFilter,
		sourceProvider,
		stats,
		db,
		log,
	)
}
