package processor

import (
    entropypkg "github.com/pantheon-systems/search-secrets/pkg/entropy"
    "github.com/pantheon-systems/search-secrets/pkg/rule"
    gitdiff "gopkg.in/src-d/go-git.v4/plumbing/format/diff"
    gitobject "gopkg.in/src-d/go-git.v4/plumbing/object"
)

type EntropyProcessor struct {
    Charset          string
    LengthThreshold  int
    EntropyThreshold float64
}

func NewEntropyProcessor(charset string, lengthThreshold int, entropyThreshold float64) (result *EntropyProcessor) {
    return &EntropyProcessor{
        Charset:          charset,
        LengthThreshold:  lengthThreshold,
        EntropyThreshold: entropyThreshold,
    }
}

func (p *EntropyProcessor) FindInFileChange(*gitobject.Change, []gitdiff.Chunk, string) (result []*rule.FileChangeFinding, err error) {
    return
}

func (p *EntropyProcessor) FindInLine(line string) (result []*rule.LineFinding, err error) {
    ranges := entropypkg.FindHighEntropyWords(line, p.Charset, p.LengthThreshold, p.EntropyThreshold)
    if ranges == nil {
        return
    }

    for _, rang := range ranges {
        result = append(result, &rule.LineFinding{LineRange: &rang})
    }

    return
}
