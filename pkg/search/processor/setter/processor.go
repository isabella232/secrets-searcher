package setter

import (
	"strings"

	"github.com/pantheon-systems/secrets-searcher/pkg/dev"
	"github.com/pantheon-systems/secrets-searcher/pkg/search"

	cmap "github.com/orcaman/concurrent-map"
	"github.com/pantheon-systems/secrets-searcher/pkg/logg"
	"github.com/pantheon-systems/secrets-searcher/pkg/manip"
	"github.com/pantheon-systems/secrets-searcher/pkg/search/contract"
)

const KeyValueExtraName = "setter-key-value"

type (
	Processor struct {
		name    string
		Rules   []*Rule
		targets *search.TargetSet
		log     logg.Logg
		processorState
	}
	processorState struct {
		rulesByExt cmap.ConcurrentMap
	}
)

func NewProcessor(name string, rules []*Rule, targets *search.TargetSet, log logg.Logg) (result *Processor) {
	result = &Processor{
		name:    name,
		Rules:   rules,
		targets: targets,
		log:     log,
		processorState: processorState{
			rulesByExt: cmap.New(),
		},
	}

	return
}

func (p *Processor) GetName() string {
	return p.name
}

func (p *Processor) supportedRules(path string) (result []*Rule) {
	if uncast, ok := p.rulesByExt.Get(path); ok {
		return uncast.([]*Rule)
	}

	for _, rule := range p.Rules {
		if rule.SupportsPath(path) {
			p.log.Tracef("rule %s supports path %s", rule.GetName(), path)
			result = append(result, rule)
		}
	}

	p.rulesByExt.Set(path, result)

	return
}

func (p *Processor) FindResultsInLine(job contract.LineProcessorJobI, line string) (err error) {
	path := job.FileChange().Path
	rules := p.supportedRules(path)
	//targetLog := job.Log(p.log).WithPrefix("target")
	targetLog := p.log

	var startIndex int
	for {
		if strings.TrimSpace(line[startIndex:]) == "" {
			break
		}

		// Execute rules
		// If one returnes found=true, we continue with the rest of the line
		if found, endIndex := p.findResultsWithRules(job, rules, line, startIndex, targetLog); found {
			startIndex = endIndex
			continue
		}

		// If rules find nothing in the line we're done
		break
	}

	return
}

func (p *Processor) findResultsWithRules(job contract.LineProcessorJobI, rules []*Rule, line string, startIndex int, targetLog logg.Logg) (found bool, endIndex int) {
	for _, rule := range rules {
		if found, endIndex = p.findFirstSecretInStringUsingRule(job, rule, line, startIndex, targetLog); found {
			return
		}
	}
	return
}

func (p *Processor) findFirstSecretInStringUsingRule(job contract.LineProcessorJobI, rule *Rule, line string, startIndex int, targetLog logg.Logg) (found bool, endIndex int) {
	dev.BreakBeforeSetterRule(rule.GetName())

	// Execute rule on the sub string
	rest := line[startIndex:]
	contextValue, keyValue, secretValue, matchingRe, ok := rule.FindNextSecret(rest)

	// If we found nothing, we're done
	if !ok {
		return
	}

	// Check targets
	if !p.targets.Matches(keyValue.Value, secretValue.Value, targetLog) {
		return
	}

	// We need to shift the line range to account for the beginning of the line string.
	lineRange := secretValue.LineRange.Shifted(startIndex)
	contextRange := contextValue.LineRange.Shifted(startIndex)

	// There's just entropy check and code match left, so we should submit an ignore,
	// since we're having a good enough look. Always do this after SubmitLineResult() or we'll
	// ignore outselves.
	defer job.SubmitLineRangeIgnore(lineRange)

	// Extras
	var secretExtras []*contract.ResultExtra

	var findingExtras []*contract.ResultExtra

	findingExtras = append(findingExtras, &contract.ResultExtra{
		Key:    "setter-rule",
		Header: "Proc. rule",
		Value:  rule.GetName(),
	})

	findingExtras = append(findingExtras, &contract.ResultExtra{
		Key:    "setter-regex",
		Header: "Regex",
		Value:  matchingRe.String(),
		Code:   true,
		Debug:  true,
	})

	testLink := dev.RegexpTestLink(matchingRe, rest)
	findingExtras = append(findingExtras, &contract.ResultExtra{
		Key:    "setter-regex",
		Header: "Debug regex",
		Value:  manip.Truncate(testLink, 50) + "...",
		URL:    testLink,
		Debug:  true,
	})

	findingExtras = append(findingExtras, &contract.ResultExtra{
		Key:    KeyValueExtraName,
		Header: "Processor key",
		Value:  keyValue.Value,
	})

	job.SubmitLineResult(&contract.LineResult{
		LineRange:        lineRange,
		ContextLineRange: contextRange,
		SecretValue:      secretValue.Value,
		SecretExtras:     secretExtras,
		FindingExtras:    findingExtras,
	})

	found = true
	endIndex = lineRange.EndIndex

	return
}
