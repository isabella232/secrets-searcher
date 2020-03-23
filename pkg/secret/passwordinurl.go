package secret

import (
	"github.com/pantheon-systems/search-secrets/pkg/database"
	"github.com/pantheon-systems/search-secrets/pkg/database/enum/decision"
	"github.com/pantheon-systems/search-secrets/pkg/errors"
	diffpkg "github.com/pantheon-systems/search-secrets/pkg/secret/diff"
	"regexp"
)

var (
	ignoreExamplePairs = []*pair{
		{"Me", "Secret"},
		{"user", "pass"},
		{"username", "password"},
		{"USERNAME", "PASSWORD"},
		{"myuser", "EXTRA_S3CUR3_PASS"},
	}
	needsInvestigationPairs = []*pair{
		{"bob", "loblaw"},
	}
	ignoreTemplatePasswords = []string{"{password}", "{0}", "$db_pass"}

	re = regexp.MustCompile(`[a-zA-Z]{3,10}://([^/\s:@]{3,20}):([^/\s:@]{3,20})@`)
)

type (
	PasswordInURLParser struct {
		diff *diffpkg.Diff
	}
	pair struct {
		username, password string
	}
)

func (p *PasswordInURLParser) Parse(finding *database.Finding, findingString *database.FindingString) (parsedSecrets []*parsedSecret, err error) {
	matches := re.FindStringSubmatch(findingString.String)
	if len(matches) == 0 {
		err = errors.New("unable to parse finding secret string")
		return
	}

	pair := &pair{username: matches[1], password: matches[2]}

	if stringInSlice(ignoreTemplatePasswords, pair.password) {
		parsedSecrets = []*parsedSecret{{Value: findingString.String, Decision: decision.IgnoreTemplateVars{}.New()}}
		return
	}
	if p.Contains(ignoreExamplePairs, pair) {
		parsedSecrets = []*parsedSecret{{Value: findingString.String, Decision: decision.IgnoreExampleCode{}.New()}}
		return
	}

	if p.Contains(needsInvestigationPairs, pair) {
		parsedSecrets = []*parsedSecret{{Value: findingString.String, Decision: decision.NeedsInvestigation{}.New()}}
		return
	}

	p.diff = diffpkg.New(finding.Diff)
	p.diff.SetLine(findingString.StartLine)

	return nil, errors.Errorv("dsiff format not implemented", finding.Diff)
}

func stringInSlice(slice []string, str string) bool {
	for _, item := range slice {
		if str == item {
			return true
		}
	}
	return false
}

func (p *PasswordInURLParser) Contains(pairs []*pair, pair *pair) bool {
	for _, pa := range pairs {
		if p.PairsEqual(pa, pair) {
			return true
		}
	}
	return false
}

func (p *PasswordInURLParser) PairsEqual(pair1 *pair, pair2 *pair) bool {
	return pair1.username == pair2.username && pair1.password == pair2.password
}
