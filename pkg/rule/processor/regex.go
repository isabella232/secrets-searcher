package processor

import (
    "github.com/pantheon-systems/search-secrets/pkg/rule"
    "github.com/pantheon-systems/search-secrets/pkg/structures"
    gitdiff "gopkg.in/src-d/go-git.v4/plumbing/format/diff"
    gitobject "gopkg.in/src-d/go-git.v4/plumbing/object"
    "regexp"
)

type RegexProcessor struct {
    re           *regexp.Regexp
    whitelistRes *structures.RegexpSet
}

func NewRegexProcessor(reString string, whitelistCodeMatch []string) (result *RegexProcessor, err error) {
    var re *regexp.Regexp
    re, err = regexp.Compile(reString)
    if err != nil {
        return
    }

    var whitelistRes structures.RegexpSet
    whitelistRes, err = structures.NewRegexpSetFromStrings(whitelistCodeMatch)
    if err != nil {
        return
    }

    result = &RegexProcessor{
        re:           re,
        whitelistRes: &whitelistRes,
    }

    return
}

func (p *RegexProcessor) FindInFileChange(*gitobject.Change, []gitdiff.Chunk, string) (result []*rule.FileChangeFinding, err error) {
    return
}

func (p *RegexProcessor) FindInLine(line string) (result []*rule.LineFinding, err error) {
    indexPairs := p.re.FindAllStringIndex(line, 1)

    for _, pair := range indexPairs {
        lineRange := &structures.LineRange{StartIndex: pair[0], EndIndex: pair[1]}
        secret := lineRange.GetStringFrom(line)

        if p.whitelistRes.MatchStringAny(secret) {
            continue
        }

        result = append(result, &rule.LineFinding{
            LineRange:        lineRange,
            SecretsProcessed: true,
            SecretValues:     []string{secret},
        })
    }

    return
}
