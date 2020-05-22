package build

import (
	"regexp"
	"time"

	"github.com/pantheon-systems/search-secrets/pkg/app/config"
	"github.com/pantheon-systems/search-secrets/pkg/database"
	"github.com/pantheon-systems/search-secrets/pkg/dev"
	gitpkg "github.com/pantheon-systems/search-secrets/pkg/git"
	"github.com/pantheon-systems/search-secrets/pkg/manip"
)

func RepoFilter(sourceCfg *config.SourceConfig, rescanPrevious bool, db *database.Database) (result *manip.SliceFilter) {
	if dev.Params.Filter.Repo != "" {
		result = manip.StringFilter([]string{dev.Params.Filter.Repo}, nil)
		return
	}

	if rescanPrevious {
		var repos database.Repos
		var err error
		if repos, err = db.GetRepos(); err != nil {
			panic("unable to get repos")
		}

		repoNames := make([]string, len(repos))
		for i := range repos {
			repoNames[i] = repos[i].Name
		}

		result = manip.StringFilter(repoNames, nil)
		return
	}

	result = manip.StringFilter(sourceCfg.IncludeRepos, sourceCfg.ExcludeRepos)

	return
}

func CommitFilter(searchCfg *config.SearchConfig, rescanPrevious bool, db *database.Database) (result *gitpkg.CommitFilter) {
	if dev.Params.Filter.Commit != "" {
		return ExactCommitFilter(dev.Params.Filter.Commit)
	}

	if rescanPrevious {
		var commits database.Commits
		var err error
		if commits, err = db.GetCommits(); err != nil {
			panic("unable to get commits")
		}

		commitHashes := make([]string, len(commits))
		for i, commit := range commits {
			commitHashes[i] = commit.CommitHash
		}

		return ExactCommitFilter(commitHashes...)
	}

	earliestTime := searchCfg.EarliestTime
	latestTime := searchCfg.LatestTime

	const excludeNoDiffCommits = true

	result = gitpkg.NewCommitFilter(nil, earliestTime, latestTime, excludeNoDiffCommits)

	return
}

func ExactCommitFilter(commitHashes ...string) (result *gitpkg.CommitFilter) {
	commitHashSet := manip.StringSet(commitHashes)
	commitHashFilter := manip.NewSliceFilter(commitHashSet, nil)

	return gitpkg.NewCommitFilter(commitHashFilter, time.Time{}, time.Time{}, true)
}

func FileChangeFilter(searchCfg *config.SearchConfig) (result *gitpkg.FileChangeFilter) {
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

func SecretIDFilter(searchCfg *config.SearchConfig) (result *manip.SliceFilter) {
	result = manip.StringFilter(nil, searchCfg.WhitelistSecretIDs)
	return
}
