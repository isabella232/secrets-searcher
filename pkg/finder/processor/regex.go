package processor

import (
    "github.com/pantheon-systems/search-secrets/pkg/finder"
    "github.com/pantheon-systems/search-secrets/pkg/git"
    "github.com/pantheon-systems/search-secrets/pkg/structures"
    "github.com/sirupsen/logrus"
    "regexp"
)

type RegexProcessor struct {
    name             string
    re               *regexp.Regexp
    whitelistCodeRes *structures.RegexpSet
}

func NewRegexProcessor(name, reString string, whitelistCodeRes *structures.RegexpSet) (result *RegexProcessor, err error) {
    var re *regexp.Regexp
    re, err = regexp.Compile(reString)
    if err != nil {
        return
    }

    result = &RegexProcessor{
        name:             name,
        re:               re,
        whitelistCodeRes: whitelistCodeRes,
    }

    return
}

func (p *RegexProcessor) Name() string {
    return p.name
}

func (p *RegexProcessor) FindInFileChange(*git.FileChange, *logrus.Entry) (result []*finder.Finding, ignore []*structures.FileRange, err error) {
    return
}

func (p *RegexProcessor) FindInLine(line string, _ *logrus.Entry) (result []*finder.FindingInLine, ignore []*structures.LineRange, err error) {
    indexPairs := p.re.FindAllStringIndex(line, -1)

    for _, pair := range indexPairs {
        lineRange := structures.NewLineRange(pair[0], pair[1])
        lineRangeValue := lineRange.ExtractValue(line)

        if p.isSecretWhitelisted(line, lineRangeValue) {
            ignore = append(ignore, lineRangeValue.LineRange)
            continue
        }

        result = append(result, &finder.FindingInLine{
            LineRange: lineRangeValue.LineRange,
            Secret:    &finder.Secret{Value: lineRangeValue.Value},
        })
    }

    return
}

func (p *RegexProcessor) isSecretWhitelisted(line string, secret *structures.LineRangeValue) bool {
    return p.whitelistCodeRes != nil && p.whitelistCodeRes.MatchAndTestSubmatchOrOverlap(line, secret.LineRange)
}
