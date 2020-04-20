package pem

import (
    "fmt"
    diffpkg "github.com/pantheon-systems/search-secrets/pkg/diff"
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    "github.com/pantheon-systems/search-secrets/pkg/finder"
    "github.com/pantheon-systems/search-secrets/pkg/git"
    "github.com/pantheon-systems/search-secrets/pkg/structures"
    "github.com/sirupsen/logrus"
    "regexp"
)

type Processor struct {
    name                      string
    pemType                   PEMTypeEnum
    header                    string
    footer                    string
    oneLineKeyRe              *regexp.Regexp
    oneLineKeyTooPermissiveRe *regexp.Regexp
    oneLineEscapedStringKeyRe *regexp.Regexp
    whitelistCodeRes          *structures.RegexpSet
}

func NewProcessor(name string, pemType PEMTypeEnum, whitelistCodeRes *structures.RegexpSet) (result *Processor) {
    header := fmt.Sprintf("-----BEGIN %s-----", pemType.Value())
    footer := fmt.Sprintf("-----END %s-----", pemType.Value())

    oneLineKeyRe := regexp.MustCompile(header + `\\n(.*)\\n` + footer)
    oneLineKeyTooPermissiveRe := regexp.MustCompile(`-BEGIN ` + pemType.Value() + `-(.*)-END ` + pemType.Value() + `-`)

    oneLineEscapedStringKeyRe := regexp.MustCompile(`"` + header + `\\n(.*)\\n` + footer + `\\?n?"`)

    *whitelistCodeRes = append(*whitelistCodeRes, structures.NewRegexpSetFromStringsMustCompile([]string{
        // Incomplete/invalid/example keys
        // FIXME: These are too specific to Pantheon findings and should/can be generalized
        header + `.{43}` + footer + ``,
        `"` + header + `\n.{6}\.\.\."`,
        header + `,$`,
        `with ` + header,
    })...)

    return &Processor{
        name:                      name,
        pemType:                   pemType,
        header:                    header,
        footer:                    footer,
        oneLineKeyRe:              oneLineKeyRe,
        oneLineKeyTooPermissiveRe: oneLineKeyTooPermissiveRe,
        oneLineEscapedStringKeyRe: oneLineEscapedStringKeyRe,
        whitelistCodeRes:          whitelistCodeRes,
    }
}

func (p *Processor) Name() string {
    return p.name
}

func (p *Processor) FindInLine(string, logrus.FieldLogger) (result []*finder.ProcFindingInLine, ignore []*structures.LineRange, err error) {
    return
}

func (p *Processor) FindInFileChange(fileChange *git.FileChange, commit *git.Commit, log logrus.FieldLogger) (result []*finder.ProcFinding, ignore []*structures.FileRange, err error) {
    var diff *diffpkg.Diff
    diff, err = fileChange.Diff()
    if err != nil {
        err = errors.WithMessagev(err, "unable to get diff for file change", fileChange.Path)
        return
    }

    search := &search{
        pemType:                   p.pemType,
        header:                    p.header,
        footer:                    p.footer,
        oneLineKeyRe:              p.oneLineKeyRe,
        oneLineKeyTooPermissiveRe: p.oneLineKeyTooPermissiveRe,
        oneLineEscapedStringKeyRe: p.oneLineEscapedStringKeyRe,
        whitelistCodeRes:          p.whitelistCodeRes,
        fileChange:                fileChange,
        commit:                    commit,
        diff:                      diff,
        findings:                  &result,
        ignores:                   &ignore,
        log:                       log,
    }
    err = search.find()

    return
}
