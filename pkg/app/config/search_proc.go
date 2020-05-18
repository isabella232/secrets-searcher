package config

import (
	va "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pantheon-systems/search-secrets/pkg/entropy"
	"github.com/pantheon-systems/search-secrets/pkg/manip"
	"github.com/pantheon-systems/search-secrets/pkg/search"
	"github.com/pantheon-systems/search-secrets/pkg/valid"
)

type ProcessorConfig struct {
	Name                   string `param:"name"`
	Processor              string `param:"processor"`
	RegexProcessorConfig   `param:",squash"`
	PEMProcessorConfig     `param:",squash"`
	SetterProcessorConfig  `param:",squash"`
	EntropyProcessorConfig `param:",squash"`
}

func (procCfg *ProcessorConfig) GetName() string {
	return procCfg.Name
}

func (procCfg *ProcessorConfig) Validate() (err error) {
	err = va.ValidateStruct(procCfg,
		va.Field(&procCfg.Name, va.Required),
		va.Field(&procCfg.Processor, va.Required, va.In(manip.DowncastSlice(search.ValidProcessorTypeValues())...)),
	)
	if err != nil {
		return
	}

	return va.Validate(procCfg.getSubCfg())
}

func (procCfg *ProcessorConfig) getSubCfg() (result va.Validatable) {
	switch procCfg.Processor {
	case search.Regex.String():
		result = &procCfg.RegexProcessorConfig
	case search.PEM.String():
		result = &procCfg.PEMProcessorConfig
	case search.Setter.String():
		result = &procCfg.SetterProcessorConfig
	case search.Entropy.String():
		result = &procCfg.EntropyProcessorConfig
	default:
		panic("unknown processor: " + procCfg.Processor)
	}

	return
}

//
// Regex processor

type RegexProcessorConfig struct {
	RegexString        string   `param:"regex"`
	WhitelistCodeMatch []string `param:"whitelist-code-match"`
}

func (regexProcCfg *RegexProcessorConfig) Validate() (err error) {
	return va.ValidateStruct(regexProcCfg,
		va.Field(&regexProcCfg.RegexString, va.Required, valid.RegexpPattern),
		va.Field(&regexProcCfg.WhitelistCodeMatch, va.Each(valid.RegexpPattern)),
	)
}

//
// PEM processor

type PEMProcessorConfig struct {
	PEMType string `param:"pem-type"`
}

func (pemProcCfg *PEMProcessorConfig) Validate() (err error) {
	return va.ValidateStruct(pemProcCfg,
		va.Field(&pemProcCfg.PEMType, va.Required),
	)
}

//
// Setter processor

// In search_proc_setter.go

//
// Entropy processor

type EntropyProcessorConfig struct {
	Charset             string   `param:"charset"`
	WordLengthThreshold int      `param:"word-length-threshold"`
	Threshold           float64  `param:"threshold"`
	SkipPEMs            bool     `param:"skip-pems"`
	WhitelistCodeMatch  []string `param:"whitelist-code-match"`
}

// FIXME This is being skipped during viper/mapstructure creation of appConfig, so the defaults won't work
func NewEntropyProcessorConfig() *EntropyProcessorConfig {
	return &EntropyProcessorConfig{
		SkipPEMs: true,
	}
}

func (entropyProcCfg *EntropyProcessorConfig) Validate() (err error) {
	return va.ValidateStruct(entropyProcCfg,
		va.Field(&entropyProcCfg.Charset, va.Required, va.In(manip.DowncastSlice(entropy.ValidCharsets())...)),
		va.Field(&entropyProcCfg.WordLengthThreshold, va.Required),
		va.Field(&entropyProcCfg.Threshold),
		va.Field(&entropyProcCfg.WhitelistCodeMatch, va.Each(valid.RegexpPattern)),
	)
}
