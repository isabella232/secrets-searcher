package processor

import (
    "github.com/pantheon-systems/search-secrets/pkg/finder/rule"
    "github.com/pantheon-systems/search-secrets/pkg/structures"
    "github.com/sirupsen/logrus"
    "regexp"
)

type RegexProcessor struct {
    re           *regexp.Regexp
    whitelistRes *structures.RegexpSet
    log          *logrus.Logger
}

func NewRegexProcessor(reString string, whitelistRes *structures.RegexpSet, log *logrus.Logger) (result *RegexProcessor, err error) {
    var re *regexp.Regexp
    re, err = regexp.Compile(reString)
    if err != nil {
        return
    }

    result = &RegexProcessor{
        re:           re,
        whitelistRes: whitelistRes,
        log: log,
    }

    return
}

func (p *RegexProcessor) FindInFileChange(*rule.FileChangeContext) (result []*rule.FileChangeFinding, err error) {
    return
}

func (p *RegexProcessor) FindInLine(line string) (result []*rule.LineFinding, err error) {
    indexPairs := p.re.FindAllStringIndex(line, 1)

    for _, pair := range indexPairs {
        lineRange := &structures.LineRange{StartIndex: pair[0], EndIndex: pair[1]}
        secret := lineRange.GetStringFrom(line)

        if p.whitelistRes.MatchStringAny(secret, secret) {
            continue
        }

        result = append(result, &rule.LineFinding{
            LineRange:        lineRange,
            SecretValues:     []string{secret},
        })
    }

    return
}
