package config

import (
	"context"

	"github.com/pantheon-systems/search-secrets/pkg/manip"
	"github.com/pantheon-systems/search-secrets/pkg/source"

	va "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pantheon-systems/search-secrets/pkg/valid"
)

type SourceConfig struct {
	Provider             string   `param:"provider" env:"true"`
	IncludeRepos         []string `param:"include-repos" env:"true"`
	ExcludeRepos         []string `param:"exclude-repos" env:"true"`
	SkipFetch            bool     `param:"skip-fetch" env:"true"`
	LocalProviderConfig  `param:",squash"`
	GithubProviderConfig `param:",squash"`
}

func NewSourceConfig() (result *SourceConfig) {
	result = &SourceConfig{}
	result.SetDefaults()
	return
}

func (sourceCfg *SourceConfig) SetDefaults() {
	if sourceCfg.Provider == "" {
		sourceCfg.Provider = source.Local.Value()
	}
}

func (sourceCfg SourceConfig) ValidateWithContext(ctx context.Context) (err error) {
	err = va.ValidateStructWithContext(ctx, &sourceCfg,
		va.Field(&sourceCfg.Provider, va.Required, va.In(manip.DowncastSlice(source.ValidProviderValues())...)),
	)
	if err != nil {
		return
	}

	subCfg := sourceCfg.getSubCfg()
	return va.ValidateWithContext(ctx, subCfg)
}

func (sourceCfg SourceConfig) getSubCfg() (result va.ValidatableWithContext) {
	switch sourceCfg.Provider {
	case source.Local.Value():
		result = &sourceCfg.LocalProviderConfig
	case source.Github.Value():
		result = &sourceCfg.GithubProviderConfig
	default:
		panic("unknown provider: " + sourceCfg.Provider)
	}

	return
}

//
// Local provider

type LocalProviderConfig struct {
	LocalDir         string `param:"local-dir" env:"true"`
	MetadataProvider string `param:"metadata-provider" env:"true"`
}

func (localProviderCfg LocalProviderConfig) ValidateWithContext(ctx context.Context) (err error) {
	appCfg := getAppCfgToContext(ctx)
	return va.ValidateStructWithContext(ctx, &localProviderCfg,
		va.Field(&localProviderCfg.LocalDir, va.Required, valid.ExistingDir,
			valid.PathNotWithinParam(NewConfigParam(appCfg, &appCfg.OutputDir))),
		va.Field(&localProviderCfg.MetadataProvider, va.NotIn(source.Local.Value())),
	)
}

//
// Github provider

type GithubProviderConfig struct {
	APIToken     string `param:"api-token" env:"true"`
	User         string `param:"user" env:"true"`
	Organization string `param:"organization" env:"true"`
	SkipForks    bool   `param:"skip-forks" env:"true"`
}

func (githubProviderCfg GithubProviderConfig) ValidateWithContext(ctx context.Context) (err error) {
	return va.ValidateStruct(&githubProviderCfg,
		// TODO Implement user for Github provider, not just organization
		va.Field(&githubProviderCfg.User, va.NewStringRule(func(s string) bool { return s == "" },
			"currently, only \"source.organization\" is supported, \"%s\" is not")),
		va.Field(&githubProviderCfg.APIToken, va.Required),
		va.Field(&githubProviderCfg.Organization, va.Required),
	)
}
