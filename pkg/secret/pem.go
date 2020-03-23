package secret

import (
	"fmt"
	"github.com/pantheon-systems/search-secrets/pkg/database"
	"github.com/pantheon-systems/search-secrets/pkg/database/enum/decision"
	"github.com/pantheon-systems/search-secrets/pkg/errors"
	diffpkg "github.com/pantheon-systems/search-secrets/pkg/secret/diff"
	"regexp"
	"strings"
)

var (
	pemTypePattern = regexp.MustCompile(`^-----BEGIN ([^-]+)-----$`)
)

type PEMParser struct {
	diff *diffpkg.Diff
}

func (p *PEMParser) Parse(finding *database.Finding, findingString *database.FindingString) (parsedSecrets []*parsedSecret, err error) {
	var keyLines []string

	matches := pemTypePattern.FindStringSubmatch(findingString.String)
	if len(matches) == 0 {
		return nil, errors.Errorv("unable to get PEM type from found string", findingString.String)
	}
	var pemType = matches[1]
	var header = fmt.Sprintf("-----BEGIN %s-----", pemType)
	var footer = fmt.Sprintf("-----END %s-----", pemType)

	p.diff = diffpkg.New(finding.Diff)
	p.diff.SetLine(findingString.StartLine)

	// -----BEGIN RSA PRIVATE KEY-----###########################################----END RSA PRIVATE KEY-----
	if p.diff.Line.CodeContains("-----BEGIN RSA PRIVATE KEY-----MIIJKAIBAAKCAgEAr8Aeb5G+VTuSQ/D1HXDDTYBf9/q----END RSA PRIVATE KEY-----") {
		p.diff.Increment()
		parsedSecrets = []*parsedSecret{{Value: "", Decision: decision.DoNotKnowYet{}.New()}}
		return
	}

	// Rotated key:
	//
	//  -----BEGIN RSA PRIVATE KEY-----
	// -[...]
	// -[...]
	// -[...]
	// +[...]
	// +[...]
	// +[...]
	//  -----END RSA PRIVATE KEY-----
	if !p.diff.Line.IsAddOrDel && p.diff.Line.CodeEndsWith(header) && p.diff.NextLine().IsDel {
		// Get removed key
		p.diff.Increment()
		keyLines = []string{}
		p.diff.CollectCodeWhile(func(line *diffpkg.Line) bool { return line.IsDel }, &keyLines)
		key1 := &parsedSecret{
			Value:    p.buildKey(strings.Join(keyLines, "\n"), pemType),
			Decision: decision.NeedsInvestigation{}.New(),
		}

		// Get added key
		p.diff.Increment()
		keyLines = []string{}
		p.diff.CollectCodeWhile(func(line *diffpkg.Line) bool { return line.IsAdd }, &keyLines)
		key2 := &parsedSecret{
			Value:    p.buildKeyFromLines(keyLines, pemType),
			Decision: decision.NeedsInvestigation{}.New(),
		}

		parsedSecrets = []*parsedSecret{key1, key2}
		return
	}

	// Deleted key:
	//
	// ------BEGIN RSA PRIVATE KEY-----
	// -[...]
	// -[...]
	// -[...]
	// ------END RSA PRIVATE KEY-----
	//
	// or added:
	//
	// +-----BEGIN RSA PRIVATE KEY-----
	// +[...]
	// +[...]
	// +[...]
	// +-----END RSA PRIVATE KEY-----
	//
	// or multiline comments (deleted or added):
	//
	// +        key = """-----BEGIN RSA PRIVATE KEY-----
	// +[...]
	// +[...]
	// +[...]
	// +-----END RSA PRIVATE KEY-----"""
	if p.diff.Line.IsAddOrDel && p.diff.Line.CodeEndsWith(header) {
		p.diff.Increment()
		keyLines = []string{}
		p.diff.CollectCodeUntil(func(line *diffpkg.Line) bool { return line.CodeStartsWith(footer) }, &keyLines)
		parsedSecrets = []*parsedSecret{{
			Value:    p.buildKeyFromLines(keyLines, pemType),
			Decision: decision.NeedsInvestigation{}.New(),
		}}
		return
	}

	// JSON object line:
	//
	// +    "key": "-----BEGIN RSA PRIVATE KEY-----\n[...]\n[...]\n[...]\n-----END RSA PRIVATE KEY-----\n",
	var oneLineJSONPattern = regexp.MustCompile(`\w*\"[a-zA-Z_]+\": \"-----BEGIN ` + pemType + `-----\\n(.*)\\n-----END ` + pemType + `-----\\n\",?$`)
	matches = oneLineJSONPattern.FindStringSubmatch(p.diff.Line.Code)
	if len(matches) > 0 {
		keyString := strings.ReplaceAll(matches[1], "\\n", "\n")
		parsedSecrets = []*parsedSecret{{
			Value:    p.buildKey(keyString, pemType),
			Decision: decision.NeedsInvestigation{}.New(),
		}}
		return
	}

	return nil, errors.Errorv("diff format not implemented", finding.Diff)
}

func (p *PEMParser) buildKeyFromLines(keyLines []string, pemType string) string {
	return p.buildKey(strings.Join(keyLines, "\n"), pemType)
}

func (p *PEMParser) buildKey(keyString, pemType string) string {
	return fmt.Sprintf("-----BEGIN %s-----\n%s\n-----END %s-----\n", pemType, keyString, pemType)
}
