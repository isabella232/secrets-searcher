package build

import (
	"context"

	"github.com/google/go-github/v29/github"
	"github.com/pantheon-systems/search-secrets/pkg/app/config"
	"github.com/pantheon-systems/search-secrets/pkg/database"
	gitpkg "github.com/pantheon-systems/search-secrets/pkg/git"
	interactpkg "github.com/pantheon-systems/search-secrets/pkg/interact"
	"github.com/pantheon-systems/search-secrets/pkg/logg"
	"github.com/pantheon-systems/search-secrets/pkg/manip"
	"github.com/pantheon-systems/search-secrets/pkg/source"
	providerpkg "github.com/pantheon-systems/search-secrets/pkg/source/providers"
	"golang.org/x/oauth2"
)

func Source(sourceDir string, skipFetch bool, repoFilter *manip.SliceFilter, git *gitpkg.Git, sourceProvider source.ProviderI, interact *interactpkg.Interact, db *database.Database, sourceLog logg.Logg) (result *source.Source) {
	result = source.New(sourceDir, skipFetch, repoFilter, git, sourceProvider, interact, db, sourceLog)
	return
}

func buildSourceProvider(sourceCfg *config.SourceConfig, git *gitpkg.Git, sourceLog logg.Logg) (result source.ProviderI) {
	providerLog := sourceLog.AddPrefixPath("provider")
	switch sourceCfg.Provider {

	// Github source provider
	case source.Local.Value():
		result = buildLocalProvider(sourceCfg, git, providerLog)
	case source.Github.Value():
		githubClient := buildGithubClient(sourceCfg)
		result = buildGithubProvider(sourceCfg, githubClient, providerLog)
	}

	return
}

func buildLocalProvider(sourceCfg *config.SourceConfig, git *gitpkg.Git, log logg.Logg) *providerpkg.LocalProvider {

	// Metadata provider
	var metadataProvider source.ProviderI
	switch sourceCfg.MetadataProvider {
	case source.Github.Value():
		metadataProvider = buildGithubProvider(sourceCfg, nil, log)
	}

	providerLog := log.AddPrefixPath("local-provider")
	return providerpkg.NewLocalProvider(sourceCfg.Provider, sourceCfg.LocalDir, git, metadataProvider, providerLog)
}

func buildGithubProvider(sourceCfg *config.SourceConfig, gitHubClient *github.Client, log logg.Logg) *providerpkg.GithubProvider {
	providerLog := log.AddPrefixPath("github-provider")
	return providerpkg.NewGithubProvider(sourceCfg.Provider, sourceCfg.Organization, gitHubClient, sourceCfg.SkipForks, providerLog)
}

func buildGithubClient(sourceCfg *config.SourceConfig) *github.Client {
	ctx := context.Background()
	tc := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: sourceCfg.APIToken}))
	return github.NewClient(tc)
}
