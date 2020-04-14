package processor

import (
    "fmt"
    diffpkg "github.com/pantheon-systems/search-secrets/pkg/diff"
    "github.com/pantheon-systems/search-secrets/pkg/finder"
    "github.com/pantheon-systems/search-secrets/pkg/git"
    "github.com/pantheon-systems/search-secrets/pkg/structures"
    "github.com/sirupsen/logrus"
    "regexp"
    "strings"
)

type (
    PEMProcessor struct {
        name             string
        pemType            string
        header             string
        footer             string
        oneLineJSONPattern *regexp.Regexp
        whitelistCodeRes   structures.RegexpSet
    }
)

func NewPEMProcessor(name, pemType string) (result *PEMProcessor) {
    header := fmt.Sprintf("-----BEGIN %s-----", pemType)
    footer := fmt.Sprintf("-----END %s-----", pemType)
    oneLineJSONPattern := regexp.MustCompile(`: *\"-----BEGIN ` + header + `-----\\n(.*)\\n` + footer + `\\n\",?$`)
    whitelistCodeRes := structures.NewRegexpSetFromStringsMustCompile([]string{
        // Incomplete/invalid/example keys
        // FIXME: These are too specific to Pantheon findings and should/can be generalized
        header + `.{43}` + footer + ``,
        `"` + header + `\n.{6}\.\.\."`,
        header + `,$`,
        `with ` + header,
    })

    return &PEMProcessor{
        name:             name,
        pemType:            pemType,
        header:             header,
        footer:             footer,
        oneLineJSONPattern: oneLineJSONPattern,
        whitelistCodeRes:   whitelistCodeRes,
    }
}

func (p *PEMProcessor) Name() string {
    return p.name
}

func (p *PEMProcessor) FindInFileChange(fileChange *git.FileChange, log *logrus.Entry) (result []*finder.Finding, ignore []*structures.FileRange, err error) {

    // Quick out
    var patchString string
    patchString, err = fileChange.PatchString()
    if !strings.Contains(patchString, p.header) {
        return
    }

    var keyLines []string
    var startLineNum int
    var startDiffNum int

    var diff *diffpkg.Diff
    diff, err = fileChange.Diff()
    if err != nil {
        return
    }

    for {
        // Advance to the next line that contains the header
        if ok := diff.UntilTrueIncrement(func(line *diffpkg.Line) bool {
            return strings.Contains(line.Code, p.header)
        }); !ok {
            break
        }

        // Any occurences of the PEM header with a line break after it
        if diff.Line.CodeEndsWith(p.header) {

            // Added keys:
            //
            // +-----BEGIN RSA PRIVATE KEY-----
            // +[...]
            // +[...]
            // +[...]
            // +-----END RSA PRIVATE KEY-----
            //
            // or stuff like this:
            //
            // +        key = """-----BEGIN RSA PRIVATE KEY-----
            // +[...]
            // +[...]
            // +[...]
            // +-----END RSA PRIVATE KEY-----"""
            if diff.Line.IsAdd {
                startLineNum = diff.Line.LineNumFile
                startDiffNum = diff.Line.LineNumDiff

                if ok := diff.Increment(); !ok {
                    break
                }

                keyLines = []string{}
                areMoreLines := diff.UntilTrueCollectCode(func(line *diffpkg.Line) bool {
                    return line.CodeStartsWith(p.footer)
                }, &keyLines)

                secret := p.buildKeyFromLines(keyLines)

                result = append(result, &finder.Finding{
                    FileRange: &structures.FileRange{
                        StartLineNum:     startLineNum,
                        StartIndex:       0,
                        EndLineNum:       diff.Line.LineNumFile + 1,
                        EndIndex:         0,
                        StartDiffLineNum: startDiffNum,
                        EndDiffLineNum:   diff.Line.LineNumDiff + 1,
                    },
                    Secret: &finder.Secret{Value: secret},
                })

                if !areMoreLines {
                    break
                }

                continue
            }

            // Rotated key:
            //
            //  -----BEGIN RSA PRIVATE KEY-----
            // -[...]
            // -[...]
            // -[...]
            // +[...]
            // +[...]
            // +[...]
            //  -----END RSA PRIVATE KEY-----
            nextLine, nextLineExists := diff.PeekNextLine()
            if diff.Line.IsEqu && nextLineExists && nextLine.IsDel {

                // Start of entire block
                startLineNum = diff.Line.LineNumFile
                startDiffNum = diff.Line.LineNumDiff
                if ok := diff.Increment(); !ok {
                    break
                }

                // Pass removed key lines
                if ok := diff.WhileTrueIncrement(func(line *diffpkg.Line) bool { return line.IsDel }); !ok {
                    break
                }

                // Get added key
                keyLines = []string{}
                areMoreLines := diff.WhileTrueCollectCode(func(line *diffpkg.Line) bool { return line.IsAdd }, &keyLines)
                secret := p.buildKeyFromLines(keyLines)

                result = append(result, &finder.Finding{
                    FileRange: &structures.FileRange{
                        StartLineNum:     startLineNum,
                        StartIndex:       0,
                        EndLineNum:       diff.Line.LineNumFile + 1,
                        EndIndex:         0,
                        StartDiffLineNum: startDiffNum,
                        EndDiffLineNum:   diff.Line.LineNumDiff + 1,
                    },
                    Secret: &finder.Secret{Value: secret},
                })

                if !areMoreLines {
                    break
                }
                continue
            }

            // If we're here, the key is unchanged in this commit
            if ok := diff.Increment(); !ok {
                break
            }
            continue
        }

        // JSON object line:
        //
        // +    "key": "-----BEGIN RSA PRIVATE KEY-----\n[...]\n[...]\n[...]\n-----END RSA PRIVATE KEY-----\n",
        matches := p.oneLineJSONPattern.FindStringSubmatch(diff.Line.Code)
        if len(matches) > 0 {
            keyString := strings.ReplaceAll(matches[1], "\\n", "\n")
            secret := p.buildKey(keyString)

            result = append(result, &finder.Finding{
                FileRange: &structures.FileRange{
                    StartLineNum:     diff.Line.LineNumFile,
                    StartIndex:       0,
                    EndLineNum:       diff.Line.LineNumFile,
                    EndIndex:         len(diff.Line.Code) - 1,
                    StartDiffLineNum: diff.Line.LineNumDiff,
                    EndDiffLineNum:   diff.Line.LineNumDiff,
                },
                Secret: &finder.Secret{Value: secret},
            })

            if ok := diff.Increment(); !ok {
                break
            }
            continue
        }

        log.WithField("line", diff.Line.Code).Warn("unable to parse string in code")

        if ok := diff.Increment(); !ok {
            break
        }
    }

    return
}

func (p *PEMProcessor) FindInLine(string, *logrus.Entry) (result []*finder.FindingInLine, ignore []*structures.LineRange, err error) {
    return
}

func (p *PEMProcessor) buildKeyFromLines(keyLines []string) string {
    return p.buildKey(strings.Join(keyLines, "\n"))
}

func (p *PEMProcessor) buildKey(keyString string) string {
    return fmt.Sprintf("-----BEGIN %s-----\n%s\n-----END %s-----\n", p.pemType, keyString, p.pemType)
}
