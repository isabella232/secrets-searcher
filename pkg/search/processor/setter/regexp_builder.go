package setter

import (
	"bytes"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"text/template"

	"github.com/pantheon-systems/search-secrets/pkg/manip"
	"github.com/pantheon-systems/search-secrets/pkg/search"
	"github.com/pantheon-systems/search-secrets/pkg/search/rulebuild"
)

const (
	keyBackrefName = "key"
	valBackrefName = "val"
)

type (
	regexpBuilder struct {
		mainTmpl     string
		keyTmpls     []string
		keyChars     []string
		operator     string
		valTmpls     []string
		noWhitespace bool
		targets      *search.TargetSet
		notValChars  []string
	}
	tmplParams struct {
		Key string
		Op  string
		Val string
	}
)

func (r *regexpBuilder) buildRes() (result []*regexp.Regexp) {
	patterns := r.buildPatterns()
	result = make([]*regexp.Regexp, len(patterns))
	for i, pattern := range patterns {
		re := regexp.MustCompile(pattern)

		// Check backreferences
		groupNames := re.SubexpNames()[1:]
		sort.Strings(groupNames)
		if len(groupNames) != 2 || groupNames[0] != keyBackrefName || groupNames[1] != valBackrefName {
			panic("pattern must have two named backrefs, \"key\" and \"val\"")
		}

		result[i] = re
	}

	return
}

// TODO The generated negation felt frail when I was writing it
//      But it seems to work. Maybe it could be improved.
func (r *regexpBuilder) buildPatterns() (result []string) {

	// Create template references
	mainTmpl := template.New(rulebuild.MainTmplName)
	keyTmpl := mainTmpl.New(rulebuild.KeyTmplName)
	valTmpl := mainTmpl.New(rulebuild.ValTmplName)

	// Parse main template string
	// Case insensitivity is forced
	template.Must(mainTmpl.Parse(rulebuild.NoCase + r.mainTmpl))

	// Params
	key := r.buildKeyPattern()
	op := r.buildOperPattern()

	// We create an expression for each combination of
	// targets, key templates and value templates.
	for _, keyTmplString := range r.keyTmpls {
		template.Must(keyTmpl.Parse(keyTmplString))
		for _, valTmplString := range r.valTmpls {

			minLen, _ := r.targets.ValLenMinMax()
			valChars := r.targets.ValChars()

			// If no value characters is specified, it's essentially `.*`
			var valCharPatternUse string
			if valChars == nil {

				// ... so we need to selectively add some negated characters,
				// or the match will not know when to stop.
				var notChars []string
				if strings.Contains(valTmplString, `'`) {
					notChars = append(notChars, `'`)
				}
				if strings.Contains(valTmplString, `"`) {
					notChars = append(notChars, `"`)
				}
				if strings.HasSuffix(valTmplString, `?`) {
					notChars = append(notChars, ` `)
				}
				if len(notChars) == 0 {
					valCharPatternUse = `.`
				} else {
					valCharPatternUse = rulebuild.NoMatchAnyCharOf(notChars...)
				}
			} else {

				// Create list of val chars
				var chars []string
				for _, char := range valChars {
					if !manip.SliceContains(r.notValChars, char) {
						chars = append(chars, char)
					}
				}

				valCharPatternUse = rulebuild.MatchAnyCharOf(chars...)
			}

			var valTmplStringUse = valTmplString

			val := fmt.Sprintf(`(?P<val>%s{%d,})`, valCharPatternUse, minLen)
			template.Must(valTmpl.Parse(valTmplStringUse))

			params := tmplParams{Key: key, Op: op, Val: val}

			pattern := r.executeTmpl(mainTmpl, params)
			result = append(result, pattern)
		}
	}

	return
}

func (r *regexpBuilder) executeTmpl(mainTmpl *template.Template, params tmplParams) (result string) {
	buf := &bytes.Buffer{}

	// Execute template against params
	if err := mainTmpl.Execute(buf, params); err != nil {
		panic(err)
	}

	result = buf.String()

	if result == "" {
		panic("empty pattern")
	}
	if strings.Contains(result, "<no value>") {
		panic("template error")
	}

	return
}

func (r *regexpBuilder) buildKeyPattern() (result string) {
	keyPatterns := r.targets.KeyPatterns()
	keyPattern := rulebuild.MatchAnyOf(keyPatterns...)
	beforeAfter := rulebuild.MatchAnyCharOf(r.keyChars...) + `*`

	return fmt.Sprintf(`(?P<key>%s%s%s)`, beforeAfter, keyPattern, beforeAfter)
}

func (r *regexpBuilder) buildOperPattern() (result string) {
	if r.operator == "" {
		if !r.noWhitespace {
			return ""
		}
		return rulebuild.SomeSpace
	}

	result = r.operator

	if !r.noWhitespace {
		result = rulebuild.SomeSpace + result + rulebuild.SomeSpace
	}

	return
}

// We need at least one anchor at the end of the value, or the laziness we're putting on the value pattern
// will cause the matching to stop as soon as the minimum length is hit.
// If we have an end quote that doesn't match the secret's character regex, then we are all good because the quote will
// anchor it.
// If not, we need to include at least one anchor. The line end should work.
func (r *regexpBuilder) buildValEndAnchor(valTmplString string) (result string) {
	var anchorPieces []string
	_, closeQuote := r.findQuotes(valTmplString)
	// If the close quote isn't a match, there's no need for an extra anchor, so return ""
	if closeQuote != "" && !r.targets.MatchValChars(closeQuote) {
		return
	}

	// We need at least one anchor, or the laziness we're putting on the value pattern will cause the match
	// to stop as soon as the minimum length is hit.
	anchorPieces = append(anchorPieces, `$`)

	// Something tells me I'm going to be looking at this code again

	return rulebuild.MatchAnyOf(anchorPieces...)
}

func (r *regexpBuilder) findQuotes(valTmplString string) (open, close string) {
	quotes := strings.Split(valTmplString, rulebuild.Val)
	if len(quotes) != 2 {
		panic(fmt.Sprintf("You should have a single reference to \"%s\" in a value template string", rulebuild.Val))
	}
	if len(quotes[0]) > 0 {
		open = quotes[0][len(quotes[0])-1:]
	}
	if len(quotes[1]) > 0 {
		close = quotes[1][:1]
	}
	return
}
