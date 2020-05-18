package build

import (
	"regexp"

	"github.com/pantheon-systems/search-secrets/pkg/builtin"

	"github.com/pantheon-systems/search-secrets/pkg/app/config"
	"github.com/pantheon-systems/search-secrets/pkg/database"
	"github.com/pantheon-systems/search-secrets/pkg/errors"
	gitpkg "github.com/pantheon-systems/search-secrets/pkg/git"
	interactpkg "github.com/pantheon-systems/search-secrets/pkg/interact"
	"github.com/pantheon-systems/search-secrets/pkg/logg"
	"github.com/pantheon-systems/search-secrets/pkg/manip"
	searchpkg "github.com/pantheon-systems/search-secrets/pkg/search"
	"github.com/pantheon-systems/search-secrets/pkg/search/contract"
)

var SecretFileMatch = regexp.MustCompile(`^secret-([0-9a-f]{5,40}).yaml$`)

func Search(
	searchCfg *config.SearchConfig,
	repoFilter *manip.SliceFilter,
	sourceDir string,
	commitFilter *gitpkg.CommitFilter,
	fileChangeFilter *gitpkg.FileChangeFilter,
	git *gitpkg.Git,
	interact *interactpkg.Interact,
	db *database.Database,
	searchLog logg.Logg,
) (result *searchpkg.Search, err error) {

	// Build search targets
	var targets *searchpkg.TargetSet
	targets, err = Targets(searchCfg)

	// Processors
	processorsLog := searchLog.AddPrefixPath("processor")
	var processors []contract.ProcessorI
	if processors, err = Procs(searchCfg, targets, processorsLog); err != nil {
		err = errors.WithMessage(err, "unable to build processors")
		return
	}

	// Execution parameters
	chunkSize := searchCfg.ChunkSize
	workerCount := searchCfg.WorkerCount
	showBarPerJob := searchCfg.ShowBarPerJob

	// Loggers
	searchJobBuilderLog := searchLog.AddPrefixPath("job-builder")
	workerLog := searchLog.AddPrefixPath("worker")
	writerLog := searchLog.AddPrefixPath("db-result-writer")

	// Search builder
	jobBuilder := searchpkg.NewJobBuilder(
		repoFilter,
		sourceDir,
		commitFilter,
		workerCount,
		chunkSize,
		showBarPerJob,
		git,
		interact,
		db,
		searchJobBuilderLog,
	)

	// Results writer
	dbResultWriter := searchpkg.NewDBResultWriter(db, writerLog)

	// Workers
	workers := make([]*searchpkg.Worker, workerCount)
	for i := 0; i < workerCount; i++ {
		workers[i] = searchpkg.NewWorker(processors, fileChangeFilter, workerLog)
	}

	// Search runner
	jobRunner := searchpkg.NewJobRunner(workers, dbResultWriter, searchLog)

	// Search service
	result = searchpkg.New(jobBuilder, jobRunner, interact, db, searchLog)

	return
}

func Targets(searchCfg *config.SearchConfig) (result *searchpkg.TargetSet, err error) {
	var targets []*searchpkg.Target
	targetFilter := manip.StringFilter(searchCfg.IncludeTargets, searchCfg.ExcludeTargets)

	// Custom targets names for filter
	customTargetNames := make([]string, len(searchCfg.CustomTargetConfigs))
	for i, targetConfig := range searchCfg.CustomTargetConfigs {
		targets = append(targets, Target(targetConfig))
		customTargetNames[i] = targetConfig.Name
	}

	// Core targets are run after custom targets
	coreTargetConfigs := builtin.TargetConfigs()
	for _, targetConfig := range coreTargetConfigs {
		if targetFilter.Includes(targetConfig.Name) && !manip.SliceContains(customTargetNames, targetConfig.Name) {
			targets = append(targets, Target(targetConfig))
		}
	}

	if len(targets) == 0 {
		err = errors.New("all targets were exluded by the filter")
		return
	}

	result = searchpkg.NewTargetSet(targets)

	return
}

func Target(targetConfig *config.TargetConfig) (result *searchpkg.Target) {
	return searchpkg.NewTarget(
		targetConfig.Name,
		targetConfig.KeyPatterns,
		targetConfig.ExcludeKeyPatterns,
		targetConfig.ValChars,
		targetConfig.ValLenMin,
		targetConfig.ValLenMax,
		targetConfig.ValEntropyMin,
		targetConfig.SkipFilePathLikeValues,
		targetConfig.SkipVariableLikeValues,
	)
}
