package config

import (
	"context"
	"time"

	va "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pantheon-systems/search-secrets/pkg/errors"
	"github.com/pantheon-systems/search-secrets/pkg/manip"
	"github.com/pantheon-systems/search-secrets/pkg/valid"
)

type SearchConfig struct {
	CustomTargetConfigs []*TargetConfig `param:"targets"`
	IncludeTargets      []string        `param:"include-targets"`
	ExcludeTargets      []string        `param:"exclude-targets"`

	ProcessorConfigs  []*ProcessorConfig `param:"processors"`
	IncludeProcessors []string           `param:"include-processors"`
	ExcludeProcessors []string           `param:"exclude-processors"`

	EarliestTime              time.Time `param:"earliest-date" env:"true"`
	LatestTime                time.Time `param:"latest-date" env:"true"`
	WhitelistPathMatchStrings []string  `param:"whitelist-path-match"`
	WhitelistSecretIDs        []string  `param:"whitelist-secret-ids"`
	WhitelistSecretDir        string    `param:"whitelist-secret-dir" env:"true"`
	ChunkSize                 int       `param:"chunk-size" env:"true"`
	WorkerCount               int       `param:"worker-count" env:"true"`
	ShowBarPerJob             bool      `param:"show-bar-per-job" env:"true"`
	DetailedStats             bool      `param:"detailed-stats" env:"true"`
}

func NewSearchConfig() (result *SearchConfig) {
	result = &SearchConfig{}
	result.SetDefaults()
	return
}

func (searchCfg *SearchConfig) SetDefaults() {
	if searchCfg.ChunkSize == 0 {
		searchCfg.ChunkSize = 500
	}
	if searchCfg.WorkerCount == 0 {
		searchCfg.WorkerCount = 8
	}
}

func (searchCfg SearchConfig) ValidateWithContext(ctx context.Context) (err error) {
	return va.ValidateStructWithContext(ctx, &searchCfg,
		va.Field(&searchCfg.CustomTargetConfigs, va.By(noDupeTargetNames)),
		va.Field(&searchCfg.IncludeTargets, va.Each(va.Required)),
		va.Field(&searchCfg.ExcludeTargets, va.Each(va.Required)),
		va.Field(&searchCfg.ProcessorConfigs, validProcessorNames()),
		va.Field(&searchCfg.EarliestTime, valid.WhenBothNotZero(
			valid.BeforeTime(manip.NewBasicParam(&searchCfg, &searchCfg.LatestTime)))),
		va.Field(&searchCfg.WhitelistPathMatchStrings, va.Each(va.Required, valid.RegexpPattern)),
		va.Field(&searchCfg.WhitelistSecretIDs, va.Each(va.Required)),
		va.Field(&searchCfg.WhitelistSecretDir, va.When(searchCfg.WhitelistSecretDir != "", valid.ExistingDir)),
		va.Field(&searchCfg.ChunkSize, va.Required),
		va.Field(&searchCfg.WorkerCount, va.Required),
	)
}

func validProcessorNames() va.Rule {
	i := 0
	procNameRegistry := manip.NewEmptyBasicSet()
	return va.By(func(value interface{}) (err error) {
		procCfgs := value.([]*ProcessorConfig)
		for _, subValue := range procCfgs {
			name := subValue.Name
			if name == "" {
				err = errors.Errorv("processor %d name empty", i)
			}
			if procNameRegistry.Contains(name) {
				err = errors.Errorv("duplicate processor name", name)
			}
			procNameRegistry.Add(name)
			i += 1
		}
		return
	})
}

func noDupeTargetNames(value interface{}) (err error) {
	names := targetNames(value.([]*TargetConfig))
	if dupeCustom, ok := manip.FirstDuplicate(names); ok {
		return va.NewError("valid_no_dupe_rule_names", "duplicate rule name: "+dupeCustom)
	}
	return
}

func targetNames(targetConfigs []*TargetConfig) (result []string) {
	result = make([]string, len(targetConfigs))
	for i, config := range targetConfigs {
		result[i] = config.Name
	}
	return
}

//
// TargetConfig

type TargetConfig struct {

	// Unique name
	Name string `param:"name"`

	// Patterns matching suspect key names
	KeyPatterns []string `param:"key-patterns"`

	// Patterns that will exlude false positives
	ExcludeKeyPatterns []string `param:"exlude-key-patterns"`

	// Pattern matching any single valid character in a value
	ValChars []string `param:"value-char-patterns"`

	// Length restrictions
	ValLenMin     int     `param:"value-length-min"`
	ValLenMax     int     `param:"value-length-max"`
	ValEntropyMin float64 `param:"value-entropy-threshold"`

	// Exclude values that look like file paths
	SkipFilePathLikeValues bool `param:"skip-file-path-like-values"`

	// Exclude values that look like variables
	SkipVariableLikeValues bool `param:"skip-variable-like-values"`
}

func (targetCfg *TargetConfig) Validate() (err error) {
	return va.ValidateStruct(targetCfg,
		va.Field(&targetCfg.Name, va.Required),
		va.Field(&targetCfg.KeyPatterns, va.Required, va.Each(va.Required, valid.RegexpPattern)),
		va.Field(&targetCfg.KeyPatterns, va.Each(va.Required, valid.RegexpPattern)),
		va.Field(&targetCfg.ValChars, va.NilOrNotEmpty, va.Each(valid.RegexpPattern)),
		va.Field(&targetCfg.ValLenMin, va.Required),
		va.Field(&targetCfg.ValLenMax, va.Min(targetCfg.ValLenMin+1)),
	)
}
