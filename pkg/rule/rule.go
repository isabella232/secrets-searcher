package rule

import (
    "github.com/pantheon-systems/search-secrets/pkg/structures"
    gitdiff "gopkg.in/src-d/go-git.v4/plumbing/format/diff"
    gitobject "gopkg.in/src-d/go-git.v4/plumbing/object"
)

type (
    Rule struct {
        Name      string
        Processor Processor
    }
    Processor interface {
        FindInFileChange(fileChange *gitobject.Change, chunks []gitdiff.Chunk, diffString string) (result []*FileChangeFinding, err error)
        FindInLine(line string) (result []*LineFinding, err error)
    }
    FileChangeFinding struct {
        FileRange        *structures.FileRange
        SecretsProcessed bool
        SecretValues     []string
    }
    LineFinding struct {
        LineRange        *structures.LineRange
        SecretsProcessed bool
        SecretValues     []string
    }
)

func New(name string, proc Processor) (result *Rule) {
    return &Rule{
        Name:      name,
        Processor: proc,
    }
}
