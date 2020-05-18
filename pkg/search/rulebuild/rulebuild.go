package rulebuild

import (
	"fmt"
	"strings"
)

const (

	// Quotes
	SingleQu                     = `'`
	DblQu                        = `"`
	Tick                         = "`"
	DblNoQuotePattern            = `"?`
	SingleDblQuotePattern        = `(?:'|")`
	SingleDblNoQuotePattern      = `(?:'|")?`
	SingleDblDollarQuotePattern  = `(?:'|"|\$)`
	SingleDblColonNoQuotePattern = `(?:'|"|\:)?`
	DblTickQuotePattern          = "(?:\"|`)"

	// Operators
	NoOper      = ``
	EqOper      = `=`
	ColonOper   = `:`
	EqArrowOper = `=>`
	GoOper      = `:?=`
	CommaOper   = `,`
	Slash       = `\/`

	SomeSpace = `\s{0,10}`

	//
	// Template names

	MainTmplName = "mainTmpl"
	KeyTmplName  = "keyTmpl"
	ValTmplName  = "valTmpl"

	//
	// Template variables

	Key        = `{{.Key}}`
	OpVar      = `{{.Op}}`
	Val        = `{{.Val}}`
	KeyTmplVar = `{{template "` + KeyTmplName + `" .}}`
	ValTmplVar = `{{template "` + ValTmplName + `" .}}`
	OpenBrace  = `{{"{"}}` // Otherwise Go gets confused
	CloseBrace = `{{"}"}}` // Otherwise Go gets confused

	//
	// Main templates

	StandardMainTmpl = KeyTmplVar + OpVar + ValTmplVar

	//
	// Key templates

	// Strings and symbols as keys
	VarKey                  = `\b` + Key                                                   // api_key
	SingleQuoteKey          = SingleQu + Key + SingleQu                                    // 'api_key'
	DblQuoteKey             = DblQu + Key + DblQu                                          // "api_key"
	DollarVarKey            = `\$` + Key                                                   // $api_key
	ColonSymbolKey          = `:` + Key                                                    // :api_key
	SingleDblQuoteKey       = SingleDblQuotePattern + Key + SingleDblQuotePattern          // "api_key", 'api_key'
	DblQuoteNoneKey         = DblNoQuotePattern + Key + DblNoQuotePattern                  // "api_key", api_key
	DblTickQuoteKey         = DblTickQuotePattern + Key + DblTickQuotePattern              // "api_key", `api_key`
	SingleDblQuoteNoneKey   = SingleDblNoQuotePattern + Key + SingleDblNoQuotePattern      // 'api_key', "api_key", api_key
	SingleDblColonKey       = SingleDblColonNoQuotePattern + Key + SingleDblNoQuotePattern // 'api_key', "api_key", :api_key
	SingleDblColonNoneKey   = SingleDblColonNoQuotePattern + Key + SingleDblNoQuotePattern // 'api_key', "api_key", :api_key, api_key
	SingleDblQuoteDollarKey = SingleDblDollarQuotePattern + Key + SingleDblNoQuotePattern  // 'api_key', "api_key", $api_key
	DeclaredStrVarKey       = `\b` + Key + `(?:\s{1,10}string)?`                           // api_key, api_key string
	//BracketVarKey                  = `\[` + VarKey + `]`                                          // [api_key]
	//BracketSingleQuoteKey          = `\[` + SingleQuoteKey + `]`                                  // ['api_key']
	//BracketDblQuoteKey             = `\[` + DblQuoteKey + `]`                                     // ["api_key"]
	//BracketDollarVarKey            = `\[` + DollarVarKey + `]`                                    // [$api_key]
	//BracketColonSymbolKey          = `\[` + ColonSymbolKey + `]`                                  // [:api_key]
	//BracketSingleDblQuoteKey       = `\[` + SingleDblQuoteKey + `]`                               // ["api_key"], ['api_key']
	BracketDblQuoteNoneKey         = `\[` + DblQuoteNoneKey + `]`         // ["api_key"], [api_key]
	BracketSingleDblQuoteNoneKey   = `\[` + SingleDblQuoteNoneKey + `]`   // ['api_key'], ["api_key"], [api_key]
	BracketSingleDblCommaNoneKey   = `\[` + SingleDblColonNoneKey + `]`   // ['api_key'], ["api_key"], [api_key], [:api_key]
	BracketSingleDblQuoteDollarKey = `\[` + SingleDblQuoteDollarKey + `]` // ['api_key'], ["api_key"], [$api_key]

	//
	// Value templates

	JustVal               = Val                                                     // shhh
	SingleQuoteVal        = SingleQu + Val + SingleQu                               // 'shhh'
	DblQuoteVal           = DblQu + Val + DblQu                                     // "shhh"
	TickVal               = Tick + Val + Tick                                       // `shhh`
	SingleDblQuoteVal     = SingleDblQuotePattern + Val + SingleDblQuotePattern     // 'shhh', "shhh"
	SingleDblQuoteJustVal = SingleDblNoQuotePattern + Val + SingleDblNoQuotePattern // 'shhh', "shhh", shhh
	DblTickQuoteVal       = DblTickQuotePattern + Val + DblTickQuotePattern         // "shhh", `shhh`
)

//
// Helpers

func Space(s string) string {
	return SomeSpace + s + SomeSpace
}

func Quote(s string, qq ...string) string {
	qu := MatchAnyCharOf(qq...)
	return qu + s + qu
}

func Maybe(s string) string {
	return s + `?`
}

func Bracket(s string) string {
	return `\[` + s + `]`
}

func Concat(ss ...string) string {
	return strings.Join(ss, "")
}

//
// File extensions

func AnyPath() []string {
	return []string{`.`}
}

func PHPExtPaths() []string {
	return []string{`\.php[s\d]?$`, `\.inc$`, `\.p?html$`, `\.module$`}
}

func PYExtPaths() []string {
	return []string{`\.py3?$`, `\.rst$`}
}

func YAMLExtPaths() []string {
	return []string{`\.ya?ml$`}
}

func JSExtPaths() []string {
	return []string{`(?:\.(?:build|bundle|min|slim)){0}\.jsx?$`, `\.ts$`, `\.es\d?$`, `\.jade$`}
}

func JSONExtPaths() []string {
	return []string{`\.json\d?$`, `\.(?:j|lib)sonnet$`}
}

func ShellScriptExtPaths() []string {
	return []string{`.`, `\.env$`, `\.sh$`, `\.bash$`, `\.zsh$`, `Makefile$`}
}

func SystemdConfExtPaths() []string {
	return []string{`\.service$`}
}

func ConfExtPaths() []string {
	return []string{`.`, `\.co?nf$`, `\.cfg$`, `\.cf$`, `\.ini$`}
}

func XMLExtPaths() []string {
	return []string{`\.xml$`}
}

func HTMLExtPaths() []string {
	return []string{`\.html?$`}
}

func RubyExtPaths() []string {
	return []string{`\.rb$`}
}

func GoExtPaths() []string {
	return []string{`\.go$`}
}

func TemplateExtPaths() []string {
	return TemplateExts()
}

// The patterns in TemplateExts() have a different role than the rest, which are just simple matchers.
// These are run through [re.ReplaceAllString()](https://golang.org/pkg/regexp/#Regexp.ReplaceAllString) in a loop
// to remove the template extension from a path.
// So each must exactly match the template extension and the period before it, and nothing more:
func TemplateExts() []string {
	return []string{`\.erb$`, `\.template$`}
}

//
// Character classes

func MatchAnyOf(patterns ...string) (result string) {
	return matchGroup(patterns, `(?:%s)`, `|`, false)
}

func MatchAnyCharOf(patterns ...string) (result string) {
	return matchGroup(patterns, `[%s]`, ``, false)
}

func NoMatchAnyCharOf(patterns ...string) (result string) {
	return matchGroup(patterns, `[^%s]`, ``, true)
}

func matchGroup(patterns []string, format, sep string, skipUnwrap bool) (result string) {
	if len(patterns) == 0 {
		panic("need at least one pattern")
	}
	if len(patterns) == 1 && !skipUnwrap {
		return patterns[0]
	}

	join := strings.Join(patterns, sep)
	return fmt.Sprintf(format, join)
}

// From app.vars

const AlphaChar = `a-z`

const HexAlphaChar = `a-f`

const DigitChar = `\d`

const WhitespaceChar = `\s`

const DashChar = `\-`

const UnderscoreChar = `\_`

const PeriodChar = `\.`

const EqChar = `\=`

const SlashChar = `\/`

const PlusChar = `\+`

const NoCase = `(?i)`

func AnyChars() []string {
	return nil
}

func Base64Chars() []string {
	return []string{AlphaChar, DigitChar, PlusChar, EqChar, SlashChar}
}

func HexChars() []string {
	return []string{HexAlphaChar, DigitChar}
}

func CommonURLPathChars() []string {
	return []string{AlphaChar, DigitChar, PeriodChar, DashChar, UnderscoreChar}
}

func Base64PeriodDashUnderscoreChars() []string {
	return []string{AlphaChar, DigitChar, PlusChar, EqChar, SlashChar, PeriodChar, DashChar, UnderscoreChar}
}
