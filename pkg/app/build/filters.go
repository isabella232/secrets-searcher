package build

import (
	"regexp"

	"github.com/pantheon-systems/search-secrets/pkg/app/config"
	"github.com/pantheon-systems/search-secrets/pkg/dev"
	gitpkg "github.com/pantheon-systems/search-secrets/pkg/git"
	"github.com/pantheon-systems/search-secrets/pkg/manip"
)

func buildFileChangeFilter(searchCfg *config.SearchConfig) (result *gitpkg.FileChangeFilter) {
	var pathFilter *manip.RegexpFilter

	var include []string
	var exclude []string
	if dev.Params.Filter.Path != "" {
		include = []string{regexp.QuoteMeta(dev.Params.Filter.Path)}
	} else {
		exclude = searchCfg.WhitelistPathMatchStrings
	}
	pathFilter = manip.NewStringRegexpFilter(include, exclude)

	const (
		excludeFileDeletions         = true
		excludeBinaryOrEmpty         = true
		excludeOnesWithNoCodeChanges = true
	)

	result = gitpkg.NewFileChangeFilter(pathFilter, excludeFileDeletions, excludeBinaryOrEmpty, excludeOnesWithNoCodeChanges)

	return
}

func buildCommitFilter(searchCfg *config.SearchConfig) (result *gitpkg.CommitFilter) {
	var commitHashes []string
	if dev.Params.Filter.Commit != "" {
		commitHashes = []string{dev.Params.Filter.Commit}
	}
	commitHashFilter := manip.StringFilter(commitHashes, nil)

	earliestTime := searchCfg.EarliestTime
	latestTime := searchCfg.LatestTime

	const excludeNoDiffCommits = true

	result = gitpkg.NewCommitFilter(commitHashFilter, earliestTime, latestTime, excludeNoDiffCommits)

	return
}

func buildRepoFilter(sourceCfg *config.SourceConfig) (result *manip.SliceFilter) {
	if dev.Params.Filter.Repo != "" {
		result = manip.StringFilter([]string{dev.Params.Filter.Repo}, nil)
		return
	}

	result = manip.StringFilter(sourceCfg.IncludeRepos, sourceCfg.ExcludeRepos)

	return
}

func buildSecretIDFilter(searchCfg *config.SearchConfig) (result *manip.SliceFilter) {
	result = manip.StringFilter(nil, searchCfg.WhitelistSecretIDs)
	return
}
