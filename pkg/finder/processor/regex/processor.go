package regex

import (
    "github.com/pantheon-systems/search-secrets/pkg/finder"
    "github.com/pantheon-systems/search-secrets/pkg/git"
    "github.com/pantheon-systems/search-secrets/pkg/structures"
    "github.com/sirupsen/logrus"
    "regexp"
)

type Processor struct {
    name             string
    re               *regexp.Regexp
    whitelistCodeRes *structures.RegexpSet
}

func NewProcessor(name string, re *regexp.Regexp, whitelistCodeRes *structures.RegexpSet) (result *Processor) {
    return &Processor{
        name:             name,
        re:               re,
        whitelistCodeRes: whitelistCodeRes,
    }
}

func (p *Processor) Name() string {
    return p.name
}

func (p *Processor) FindInFileChange(*git.FileChange, *git.Commit, logrus.FieldLogger) (result []*finder.ProcFinding, ignore []*structures.FileRange, err error) {
    return
}

func (p *Processor) FindInLine(line string, _ logrus.FieldLogger) (result []*finder.ProcFindingInLine, ignore []*structures.LineRange, err error) {
    indexPairs := p.re.FindAllStringIndex(line, -1)

    for _, pair := range indexPairs {
        lineRange := structures.NewLineRange(pair[0], pair[1])
        lineRangeValue := lineRange.ExtractValue(line)

        if p.isSecretWhitelisted(line, lineRangeValue) {
            ignore = append(ignore, lineRangeValue.LineRange)
            continue
        }

        result = append(result, &finder.ProcFindingInLine{
            LineRange: lineRangeValue.LineRange,
            Secret:    &finder.ProcSecret{Value: lineRangeValue.Value},
        })
    }

    return
}

func (p *Processor) isSecretWhitelisted(line string, secret *structures.LineRangeValue) bool {
    return p.whitelistCodeRes != nil && p.whitelistCodeRes.MatchAndTestSubmatchOrOverlap(line, secret.LineRange)
}
