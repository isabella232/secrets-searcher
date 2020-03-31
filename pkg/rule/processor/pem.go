package processor

import (
    "fmt"
    diffpkg "github.com/pantheon-systems/search-secrets/pkg/diff"
    "github.com/pantheon-systems/search-secrets/pkg/rule"
    "github.com/pantheon-systems/search-secrets/pkg/structures"
    "github.com/sirupsen/logrus"
    gitdiff "gopkg.in/src-d/go-git.v4/plumbing/format/diff"
    gitobject "gopkg.in/src-d/go-git.v4/plumbing/object"
    "regexp"
    "strings"
)

var pemTypePattern = regexp.MustCompile(`^-----BEGIN ([^-]+)-----$`)

const (
    weirdKey = "-----BEGIN RSA PRIVATE KEY-----MIIJKAIBAAKCAgEAr8Aeb5G+VTuSQ/D1HXDDTYBf9/q----END RSA PRIVATE KEY-----"
)

type (
    PEMProcessor struct {
        pemType            string
        oneLineJSONPattern *regexp.Regexp
        log                *logrus.Logger
    }
)

func NewPEMProcessor(pemType string, log *logrus.Logger) (result *PEMProcessor) {
    oneLineJSONPattern := regexp.MustCompile(`\w*\"[a-zA-Z_]+\": \"-----BEGIN ` + pemType + `-----\\n(.*)\\n-----END ` + pemType + `-----\\n\",?$`)

    return &PEMProcessor{
        pemType:            pemType,
        oneLineJSONPattern: oneLineJSONPattern,
        log:                log,
    }
}

func (p *PEMProcessor) FindInFileChange(fileChange *gitobject.Change, chunks []gitdiff.Chunk, diffString string) (result []*rule.FileChangeFinding, err error) {
    var header = fmt.Sprintf("-----BEGIN %s-----", p.pemType)
    var footer = fmt.Sprintf("-----END %s-----", p.pemType)

    // Quick out
    if ! strings.Contains(diffString, header) {
        return
    }

    var keyLines []string
    var startLineNum int
    var endLineNum int

    var chunksDiff *diffpkg.ChunksDiff
    chunksDiff, err = diffpkg.NewChunksDiff(chunks)
    if err != nil {
        return
    }
    diff := chunksDiff.Diff

    for {
        // Advance to the next line that contains the header
        differr := diff.UntilTrueIncrement(func(line *diffpkg.Line) bool {
            return strings.Contains(line.Code, header)
        })
        if differr == diffpkg.ErrEOL {
            break
        }
        if differr != nil {
            err = differr
            return
        }

        if diff.Line.CodeEndsWith(header) {

            // Added or deleted keys:
            //
            // ------BEGIN RSA PRIVATE KEY-----
            // -[...]
            // -[...]
            // -[...]
            // ------END RSA PRIVATE KEY-----
            //
            // or:
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
            if diff.Line.IsAddOrDel {
                startLineNum = chunksDiff.RequireFileLineNum(diff.LineNum)
                diff.RequireIncrement()
                keyLines = []string{}
                err = diff.UntilTrueCollectCode(func(line *diffpkg.Line) bool {
                    return line.CodeStartsWith(footer)
                }, &keyLines)
                if err != nil {
                    return
                }
                endLineNum = chunksDiff.RequireFileLineNum(diff.LineNum) + 1
                result = append(result, &rule.FileChangeFinding{
                    FileRange: &structures.FileRange{
                        StartLineNum: startLineNum,
                        StartIndex:   0,
                        EndLineNum:   endLineNum,
                        EndIndex:     0,
                    },
                    SecretValues:     []string{p.buildKeyFromLines(keyLines)},
                    SecretsProcessed: true,
                    //Decision:    decision.NeedsInvestigation{}.New(),
                })
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
            if diff.Line.IsEqu && diff.RequireNextLine().IsDel {
                startLineNum = chunksDiff.RequireFileLineNum(diff.LineNum)

                // Get removed key
                diff.RequireIncrement()
                keyLines = []string{}
                err = diff.WhileTrueCollectCode(func(line *diffpkg.Line) bool { return line.IsDel }, &keyLines)
                if err != nil {
                    return
                }
                endLineNum = chunksDiff.RequireFileLineNum(diff.LineNum) + 1
                key1 := p.buildKey(strings.Join(keyLines, "\n"))

                // Get added key
                diff.RequireIncrement()
                startLineNum = chunksDiff.RequireFileLineNum(diff.LineNum)
                keyLines = []string{}
                err = diff.WhileTrueCollectCode(func(line *diffpkg.Line) bool { return line.IsAdd }, &keyLines)
                if err != nil {
                    return
                }
                endLineNum = chunksDiff.RequireFileLineNum(diff.LineNum) + 1
                key2 := p.buildKeyFromLines(keyLines)

                result = append(result, &rule.FileChangeFinding{
                    FileRange: &structures.FileRange{
                        StartLineNum: startLineNum,
                        StartIndex:   0,
                        EndLineNum:   endLineNum,
                        EndIndex:     0,
                    },
                    SecretValues:     []string{key1, key2},
                    SecretsProcessed: true,
                    //Decision:    decision.NeedsInvestigation{}.New(),
                })
                continue
            }

            // If we're here, the key is unchanged in this commit
            continue
        }

        // JSON object line:
        //
        // +    "key": "-----BEGIN RSA PRIVATE KEY-----\n[...]\n[...]\n[...]\n-----END RSA PRIVATE KEY-----\n",
        matches := p.oneLineJSONPattern.FindStringSubmatch(diff.Line.Code)
        if len(matches) > 0 {
            startLineNum = chunksDiff.RequireFileLineNum(diff.LineNum)
            keyString := strings.ReplaceAll(matches[1], "\\n", "\n")
            result = []*rule.FileChangeFinding{{
                FileRange: &structures.FileRange{
                    StartLineNum: startLineNum,
                    StartIndex:   0,
                    EndLineNum:   startLineNum,
                    EndIndex:     len(diff.Line.Code) - 1,
                },
                SecretValues:     []string{p.buildKey(keyString)},
                SecretsProcessed: true,
                //Decision:    decision.NeedsInvestigation{}.New(),
            }}
            diff.RequireIncrement()
            continue
        }

        // -----BEGIN RSA PRIVATE KEY-----###########################################----END RSA PRIVATE KEY-----
        if p.pemType == "RSA PRIVATE KEY" && diff.Line.CodeContains(weirdKey) {
            startLineNum = chunksDiff.RequireFileLineNum(diff.LineNum)
            result = append(result, &rule.FileChangeFinding{
                FileRange: &structures.FileRange{
                    StartLineNum: startLineNum,
                    StartIndex:   0,
                    EndLineNum:   startLineNum,
                    EndIndex:     len(diff.Line.Code) - 1,
                },
                SecretValues:     []string{weirdKey},
                SecretsProcessed: true,
                //Decision:        decision.DoNotKnowYet{}.New(),
            })
            diff.RequireIncrement()
            continue
        }

        p.log.WithField("line", diff.Line.Code).Warn("unable to parse string in code")
        diff.RequireIncrement()
    }

    return
}

func (p *PEMProcessor) FindInLine(string) (result []*rule.LineFinding, err error) {
    return
}

func (p *PEMProcessor) countHeaderInChunks(chunks []gitdiff.Chunk) (result int) {
    var header = fmt.Sprintf("-----BEGIN %s-----", p.pemType)

    for _, chunk := range chunks {
        if chunk.Type() == gitdiff.Delete {
            continue
        }
        result += strings.Count(chunk.Content(), header)
    }
    return
}

func (p *PEMProcessor) buildKeyFromLines(keyLines []string) string {
    return p.buildKey(strings.Join(keyLines, "\n"))
}

func (p *PEMProcessor) buildKey(keyString string) string {
    return fmt.Sprintf("-----BEGIN %s-----\n%s\n-----END %s-----\n", p.pemType, keyString, p.pemType)
}
