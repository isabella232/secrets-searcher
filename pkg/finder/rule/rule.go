package rule

import (
    "github.com/pantheon-systems/search-secrets/pkg/structures"
)

type (
    Rule struct {
        Name      string
        Processor Processor
    }
    Processor interface {
        FindInFileChange(context *FileChangeContext) (result []*FileChangeFinding, err error)
        FindInLine(line string) (result []*LineFinding, err error)
    }
    FileChangeFinding struct {
        FileRange        *structures.FileRange
        SecretValues     []string
    }
    LineFinding struct {
        LineRange        *structures.LineRange
        SecretValues     []string
    }
)

func New(name string, proc Processor) (result *Rule) {
    return &Rule{
        Name:      name,
        Processor: proc,
    }
}
