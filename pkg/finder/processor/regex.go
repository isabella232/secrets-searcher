package processor

import (
    "github.com/pantheon-systems/search-secrets/pkg/finder/rule"
    "github.com/pantheon-systems/search-secrets/pkg/structures"
    "github.com/sirupsen/logrus"
    "regexp"
)

type RegexProcessor struct {
    re               *regexp.Regexp
    whitelistCodeRes *structures.RegexpSet
}

func NewRegexProcessor(reString string, whitelistCodeRes *structures.RegexpSet) (result *RegexProcessor, err error) {
    var re *regexp.Regexp
    re, err = regexp.Compile(reString)
    if err != nil {
        return
    }

    result = &RegexProcessor{
        re:               re,
        whitelistCodeRes: whitelistCodeRes,
    }

    return
}

func (p *RegexProcessor) FindInFileChange(*rule.FileChangeContext, *logrus.Entry) (result []*rule.FileChangeFinding, ignore []*structures.FileRange, err error) {
    return
}

func (p *RegexProcessor) FindInLine(line string, _ *logrus.Entry) (result []*rule.LineFinding, ignore []*structures.LineRange, err error) {
    indexPairs := p.re.FindAllStringIndex(line, -1)

    for _, pair := range indexPairs {
        lineRange := structures.NewLineRange(pair[0], pair[1])
        lineRangeValue := lineRange.ExtractValue(line)

        if p.isSecretWhitelisted(line, lineRangeValue) {
            ignore = append(ignore, lineRangeValue.LineRange)
            continue
        }

        result = append(result, &rule.LineFinding{
            LineRange: lineRangeValue.LineRange,
            Secrets:   []*rule.Secret{{Value: lineRangeValue.Value}},
        })
    }

    return
}

func (p *RegexProcessor) isSecretWhitelisted(line string, secret *structures.LineRangeValue) bool {
    return p.whitelistCodeRes != nil && p.whitelistCodeRes.MatchAndTestSubmatchOrOverlap(line, secret.LineRange)
}
