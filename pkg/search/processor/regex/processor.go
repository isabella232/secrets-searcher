package regex

import (
	"fmt"
	"regexp"

	"github.com/pantheon-systems/search-secrets/pkg/search"

	"github.com/pantheon-systems/search-secrets/pkg/logg"
	"github.com/pantheon-systems/search-secrets/pkg/manip"
	"github.com/pantheon-systems/search-secrets/pkg/search/contract"
)

type Processor struct {
	name          string
	re            *regexp.Regexp
	codeWhitelist *search.CodeWhitelist
	log           logg.Logg
}

func NewProcessor(name string, re *regexp.Regexp, codeWhitelist *search.CodeWhitelist, log logg.Logg) (result *Processor) {
	return &Processor{
		name:          name,
		re:            re,
		codeWhitelist: codeWhitelist,
		log:           log,
	}
}

func (p *Processor) GetName() string {
	return p.name
}

func (p *Processor) FindResultsInLine(job contract.LineProcessorJobI, line string) (err error) {
	matches := p.re.FindAllStringSubmatchIndex(line, -1)

	for _, match := range matches {
		var lineRange *manip.LineRange
		var contextRange *manip.LineRange

		switch len(match) {
		case 2: // No backref
			lineRange = manip.NewLineRange(match[0], match[1])
		case 4: // Backref
			lineRange = manip.NewLineRange(match[2], match[3])
			contextRange = manip.NewLineRange(match[0], match[1])
		default:
			panic(fmt.Sprintf("invalid match len: %d", len(match)))
		}

		if p.codeWhitelist.IsSecretWhitelisted(line, lineRange) {
			job.SubmitLineRangeIgnore(lineRange)
			continue
		}

		job.SubmitLineResult(&contract.LineResult{
			LineRange:        lineRange,
			SecretValue:      lineRange.ExtractValue(line).Value,
			ContextLineRange: contextRange,
		})
	}

	return
}
