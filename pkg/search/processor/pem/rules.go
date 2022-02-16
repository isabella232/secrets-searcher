package pem

import (
	"regexp"
	"strings"

	"github.com/pantheon-systems/secrets-searcher/pkg/git"

	"github.com/pantheon-systems/secrets-searcher/pkg/dev"
	"github.com/pantheon-systems/secrets-searcher/pkg/errors"
	"github.com/pantheon-systems/secrets-searcher/pkg/manip"
	"github.com/pantheon-systems/secrets-searcher/pkg/search/contract"
)

var onlyBase64Re = regexp.MustCompile(`^[a-zA-Z0-9+/]+={0,2}$`)

// Find a potential key
func (p *Processor) executeRules(job contract.ProcessorJobI) {
	var err error
	diff := job.Diff()

	// RULES, in order
	rules := []searchRule{
		{"findMultilineKey", p.findMultilineKey},
		{"findSingleLineKey", p.findSingleLineKey},
		{"ignoreMalformedKey", p.ignoreMalformedKey},
	}

	originalLine := diff.Line.Num

	for _, rule := range rules {
		dev.BreakBeforePemRule(rule.name)

		// Keep us on the same header line for each try
		diff.SetLine(originalLine)
		job.SearchingLine(originalLine)

		// Try to find
		job.Log(p.log).Debugf("running rule: %s", rule.name)

		// Execute rule
		var ok bool
		ok, err = rule.execute(job)
		if err != nil {
			errors.ErrLog(job.Log(p.log), err).Warn("error while finding key")
			continue
		}

		// If we parsed something, we're done
		if ok {
			return
		}
	}

	job.Log(p.log).Debug("unable to parse string in code")

	return
}

// Added keys:
//
// +-----BEGIN RSA PRIVATE KEY-----
// +[...]
// +[...]
// +[...]
// +-----END RSA PRIVATE KEY-----
//
// or stuff like this:
//
// +        key = """-----BEGIN RSA PRIVATE KEY-----
// +[...]
// +[...]
// +[...]
// +-----END RSA PRIVATE KEY-----"""
//
// or rotated key:
//
//  -----BEGIN RSA PRIVATE KEY-----
// -[...]
// -[...]
// -[...]
// +[...]
// +[...]
// +[...]
//  -----END RSA PRIVATE KEY-----
func (p *Processor) findMultilineKey(job contract.ProcessorJobI) (parsed bool, err error) {
	diff := job.Diff()

	// Must be multiline
	if strings.Contains(diff.Line.Code, p.footer) {
		return
	}

	// Get header line numbers
	headerLine := diff.Line

	// Go to the line after the header, first line of the block
	if !diff.Incr() {
		return
	}

	// Collect block lines, end then we'll be on the footer line
	var blockLines []string
	whileFun := func(line *git.Line) bool {
		return !strings.HasPrefix(line.Trim(), p.footer)
	}
	doFunc := func(line *git.Line) {
		if line.IsAdd {
			blockLines = append(blockLines, line.Trim())
		}
	}
	if ok := diff.WhileTrueDo(whileFun, doFunc); !ok {
		return
	}

	blockLinesLen := len(blockLines)
	footerLine := diff.Line
	fileRange := &manip.FileRange{
		StartLineNum: headerLine.NumInFile,
		StartIndex:   0,
		EndLineNum:   footerLine.NumInFile,
		EndIndex:     len(footerLine.Code),
	}
	keyString := p.buildKeyFromBlockLines(blockLines)

	// Validate length and charset
	blockString := strings.Trim(strings.Join(blockLines, ""), " ")
	if blockLinesLen < 5 || blockLinesLen > 100 || !onlyBase64Re.MatchString(blockString) {
		job.SubmitIgnore(fileRange)
		diff.Incr()
		return
	}

	// Collect result
	parsed, err = p.potentialFinding(job, keyString, fileRange, true)

	return
}

// Escaped string object line:
// JSON:
// +    "key": "-----BEGIN RSA PRIVATE KEY-----\n[...]\n[...]\n[...]\n-----END RSA PRIVATE KEY-----\n",
// or Ruby:
// +    "key" =>"-----BEGIN RSA PRIVATE KEY-----[...]----END RSA PRIVATE KEY-----",
func (p *Processor) findSingleLineKey(job contract.ProcessorJobI) (parsed bool, err error) {
	diff := job.Diff()

	matches := p.oneLineEscapedStringKeyRe.FindStringSubmatch(diff.Line.Code)
	if len(matches) == 0 {
		return
	}

	matched := matches[1]
	keyBlock := strings.ReplaceAll(matched, "\\n", "\n")
	keyString := p.buildKey(keyBlock)

	fileRange := &manip.FileRange{
		StartLineNum: diff.Line.NumInFile,
		StartIndex:   0,
		EndLineNum:   diff.Line.NumInFile,
		EndIndex:     len(diff.Line.Code),
	}

	// If there weren't any line breaks it's invalid so ignore
	if keyBlock == matched {
		job.SubmitIgnore(fileRange)
		parsed = true
		return
	}

	// Set the diff to the next line to be searched
	if !diff.Incr() {
		return
	}

	// Collect result
	parsed, err = p.potentialFinding(job, keyString, fileRange, true)

	return
}

// Ignore malformed keys
func (p *Processor) ignoreMalformedKey(job contract.ProcessorJobI) (parsed bool, err error) {
	diff := job.Diff()

	match := p.oneLineKeyTooPermissiveRe.FindStringIndex(diff.Line.Code)
	if match == nil {
		return
	}

	lineRange := manip.NewLineRange(match[0], match[1])
	fileRange := manip.NewFileRangeFromLineRange(lineRange, diff.Line.NumInFile)

	job.SubmitIgnore(fileRange)
	parsed = true

	diff.Incr()

	return
}
