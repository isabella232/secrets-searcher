package config

import (
	va "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pantheon-systems/search-secrets/pkg/search/rulebuild"
	"github.com/pantheon-systems/search-secrets/pkg/valid"
)

type SetterProcessorConfig struct {

	// File extensions to search
	FileExts []string `param:"file-exts"`

	// Patterns each match a possible complete matching pattern
	// Use "{{.KeyTmpl}}", "{{.Op}}", and "{{.ValTmpl}}" to signify where things should go.
	MainTmpl string `param:"main-tmpl"`

	// Patterns each match a possible key segment syntax (quoted, starts with a dollar sign, etc)
	// Use "{{.Key}}" to signify where the pattern matching the key itself should go.
	KeyTmpls []string `param:"key-tmpls"`

	// Patterns each match a possible value segment syntax (double quote, single quote, etc)
	// Use "{{.Val}}" to signify where the pattern matching the value should go.
	ValTmpls []string `param:"val-tmpls"`

	// Pattern matches the operator
	Operator string `param:"operator"`

	// If true, no whitespace will be matched around operator
	NoWhitespace bool `param:"no-whitespace"`

	// Pattern matching any single valid character in a key
	// Used to match the characters surrounding the more specific patterns defined in target.KeyPatterns
	KeyChars []string `param:"key-chars"`

	// Disable certain value characters set in the target if the context requires it
	// Like, in a URL path, you wouldn't want to match slashes in the value if you were looking at its segments
	NotValChars []string
}

func (setterProcCfg *SetterProcessorConfig) SetDefaults() {
	if setterProcCfg.MainTmpl == "" {
		setterProcCfg.MainTmpl = rulebuild.StandardMainTmpl
	}
	if setterProcCfg.KeyTmpls == nil {
		setterProcCfg.KeyTmpls = []string{rulebuild.VarKey}
	}
	if setterProcCfg.KeyChars == nil {
		setterProcCfg.KeyChars = rulebuild.Base64PeriodDashUnderscoreChars()
	}
}

func (setterProcCfg *SetterProcessorConfig) Validate() (err error) {
	return va.ValidateStruct(setterProcCfg,
		va.Field(&setterProcCfg.FileExts, va.Required, va.Each(valid.RegexpPattern)),
		va.Field(&setterProcCfg.MainTmpl, va.Required, valid.RegexpTmpl),
		va.Field(&setterProcCfg.KeyTmpls, va.Each(va.Required, valid.RegexpTmpl)),
		va.Field(&setterProcCfg.ValTmpls, va.Each(va.Required, valid.RegexpTmpl)),
		va.Field(&setterProcCfg.Operator, valid.RegexpPattern),
		va.Field(&setterProcCfg.KeyChars, va.Required, va.Each(va.Required, valid.RegexpPattern)),
		va.Field(&setterProcCfg.NotValChars, va.Each(va.Required, valid.RegexpPattern)),
	)
}
