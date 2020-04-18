package pem

import (
    "fmt"
    diffpkg "github.com/pantheon-systems/search-secrets/pkg/diff"
    "github.com/pantheon-systems/search-secrets/pkg/finder"
    "github.com/pantheon-systems/search-secrets/pkg/git"
    "github.com/pantheon-systems/search-secrets/pkg/structures"
    "github.com/sirupsen/logrus"
    "regexp"
)

type Processor struct {
    name               string
    pemType            PEMTypeEnum
    header             string
    footer             string
    oneLineJSONPattern *regexp.Regexp
    whitelistCodeRes   structures.RegexpSet
}

func NewProcessor(name string, pemType PEMTypeEnum) (result *Processor) {
    header := fmt.Sprintf("-----BEGIN %s-----", pemType.Value())
    footer := fmt.Sprintf("-----END %s-----", pemType.Value())

    oneLineJSONPattern := regexp.MustCompile(`: *\"-----BEGIN ` + header + `-----\\n(.*)\\n` + footer + `\\n\",?$`)

    whitelistCodeRes := structures.NewRegexpSetFromStringsMustCompile([]string{
        // Incomplete/invalid/example keys
        // FIXME: These are too specific to Pantheon findings and should/can be generalized
        header + `.{43}` + footer + ``,
        `"` + header + `\n.{6}\.\.\."`,
        header + `,$`,
        `with ` + header,
    })

    return &Processor{
        name:               name,
        pemType:            pemType,
        header:             header,
        footer:             footer,
        oneLineJSONPattern: oneLineJSONPattern,
        whitelistCodeRes:   whitelistCodeRes,
    }
}

func (p *Processor) Name() string {
    return p.name
}

func (p *Processor) FindInLine(string, *logrus.Entry) (result []*finder.FindingInLine, ignore []*structures.LineRange, err error) {
    return
}

func (p *Processor) FindInFileChange(fileChange *git.FileChange, commit *git.Commit, log *logrus.Entry) (result []*finder.Finding, ignore []*structures.FileRange, err error) {
    var diff *diffpkg.Diff
    diff, err = fileChange.Diff()
    if err != nil {
        return
    }

    search := &search{
        pemType:            p.pemType,
        header:             p.header,
        footer:             p.footer,
        oneLineJSONPattern: p.oneLineJSONPattern,
        whitelistCodeRes:   p.whitelistCodeRes,
        fileChange:         fileChange,
        commit:             commit,
        diff:               diff,
        findings:           &result,
        ignore:             &ignore,
        log:                log,
    }
    err = search.find()

    return
}
