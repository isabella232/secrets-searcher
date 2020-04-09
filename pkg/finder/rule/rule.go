package rule

import (
    "github.com/pantheon-systems/search-secrets/pkg/structures"
    "github.com/sirupsen/logrus"
)

type (
    Rule struct {
        Name      string
        Processor Processor
    }
    Processor interface {
        FindInFileChange(context *FileChangeContext, log *logrus.Entry) (result []*FileChangeFinding, err error)
        FindInLine(line string, log *logrus.Entry) (result []*LineFinding, err error)
    }
    FileChangeFinding struct {
        FileRange    *structures.FileRange
        SecretValues []string
    }
    LineFinding struct {
        LineRange    *structures.LineRange
        SecretValues []string
    }
)

func New(name string, proc Processor) (result Rule) {
    return Rule{
        Name:      name,
        Processor: proc,
    }
}
