package setter

import (
	"regexp"

	"github.com/pantheon-systems/search-secrets/pkg/search"
	"github.com/pantheon-systems/search-secrets/pkg/search/rulebuild"

	"github.com/pantheon-systems/search-secrets/pkg/manip"
)

var templateRes = manip.NewRegexpSetFromStringsMustCompile(rulebuild.TemplateExts())

type Rule struct {
	name          string
	targets       *search.TargetSet
	res           []*regexp.Regexp
	fileExtFilter manip.Filter
}

func NewRule(name string, targets *search.TargetSet, fileExtFilter manip.Filter, mainTmpl string, keyTmpls []string,
	keyChars []string, operator string, valTmpls []string, noWhitespace bool, notValChars []string) *Rule {

	reBuilder := &regexpBuilder{
		targets:      targets,
		mainTmpl:     mainTmpl,
		keyTmpls:     keyTmpls,
		keyChars:     keyChars,
		operator:     operator,
		valTmpls:     valTmpls,
		noWhitespace: noWhitespace,
		notValChars:  notValChars,
	}

	res := reBuilder.buildRes()

	return &Rule{
		name:          name,
		targets:       targets,
		res:           res,
		fileExtFilter: fileExtFilter,
	}
}

func (r *Rule) GetName() (result string) {
	return r.name
}

func (r *Rule) Res() (result []*regexp.Regexp) {
	return r.res
}

func (r *Rule) SupportsPath(path string) (result bool) {

	// Remove template extension from path
	for _, re := range templateRes.ReValues() {
		if newPath := re.ReplaceAllString(path, ""); newPath != path {
			path = newPath
			break
		}
	}

	return r.fileExtFilter.Includes(path)
}

func (r *Rule) FindNextSecret(line string) (contextValue, keyValue, secretValue *manip.LineRangeValue, matchingRe *regexp.Regexp, ok bool) {
	for _, re := range r.res {

		// In regexp_builder.go, we already validated that we have only the correct groups with the correct names,
		// which makes this logic simpler
		groupNames := re.SubexpNames()
		match := re.FindStringSubmatchIndex(line)
		if match == nil || match[2] == -1 || match[4] == -1 {
			continue
		}

		keyMatchI := 2
		valMatchI := 4
		if groupNames[1] == valBackrefName {
			keyMatchI = 4
			valMatchI = 2
		}

		contextValue = manip.NewLineRange(match[0], match[1]).ExtractValue(line)
		keyValue = manip.NewLineRange(match[keyMatchI], match[keyMatchI+1]).ExtractValue(line)
		secretValue = manip.NewLineRange(match[valMatchI], match[valMatchI+1]).ExtractValue(line)
		matchingRe = re
		ok = true
		break
	}

	return
}
