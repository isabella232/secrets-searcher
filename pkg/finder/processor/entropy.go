package processor

import (
    diffpkg "github.com/pantheon-systems/search-secrets/pkg/diff"
    entropypkg "github.com/pantheon-systems/search-secrets/pkg/entropy"
    "github.com/pantheon-systems/search-secrets/pkg/finder/rule"
    "github.com/pantheon-systems/search-secrets/pkg/structures"
    "github.com/sirupsen/logrus"
    "regexp"
)

var (
    pemBeginHeaderRegex = regexp.MustCompile("-----BEGIN [^-]+-----$")
    pemEndHeaderRegex   = regexp.MustCompile("-----END [^-]+-----")
    pemJsonLineRegex    = regexp.MustCompile(`\w*\"[a-zA-Z_]+\": \"-----BEGIN [^-]+-----\\n.*\\n-----END [^-]+-----\\n\",?$`)
)

type EntropyProcessor struct {
    Charset          string
    LengthThreshold  int
    EntropyThreshold float64
    skipPEMs         bool
    whitelistRes     *structures.RegexpSet
    log              *logrus.Logger
}

func NewEntropyProcessor(charset string, lengthThreshold int, entropyThreshold float64, whitelistRes *structures.RegexpSet, skipPEMs bool, log *logrus.Logger) (result *EntropyProcessor) {
    return &EntropyProcessor{
        Charset:          charset,
        LengthThreshold:  lengthThreshold,
        EntropyThreshold: entropyThreshold,
        skipPEMs:         skipPEMs,
        whitelistRes:     whitelistRes,
        log:              log,
    }
}

func (p *EntropyProcessor) FindInFileChange(context *rule.FileChangeContext) (result []*rule.FileChangeFinding, err error) {
    var diff *diffpkg.Diff
    diff, err = context.Diff()
    if err != nil {
        return
    }

    for {
        // Skip PEM files of all types
        if p.skipPEMs {
            if diff.Line.CodeMatches(pemBeginHeaderRegex) {
                if ok := diff.UntilTrueIncrement(func(line *diffpkg.Line) bool {
                    return line.CodeMatches(pemEndHeaderRegex)
                }); !ok {
                    break
                }
                if ok := diff.Increment(); !ok {
                    break
                }
            }
            if diff.Line.CodeMatches(pemJsonLineRegex) {
                if ok := diff.Increment(); !ok {
                    break
                }
            }
        }

        if p.whitelistRes.MatchStringAny(diff.Line.Code, "") {
            if ok := diff.Increment(); !ok {
                break
            }
            continue
        }

        if ok := diff.UntilTrueIncrement(func(line *diffpkg.Line) bool { return diff.Line.IsAdd }); !ok {
            break
        }

        // Find entropy in line
        ranges := entropypkg.FindHighEntropyWords(diff.Line.Code, p.Charset, p.LengthThreshold, p.EntropyThreshold)
        if ranges == nil {
            if ok := diff.Increment(); !ok {
                break
            }
            continue
        }

        for _, rang := range ranges {
            secret := rang.GetStringFrom(diff.Line.Code)

            if p.whitelistRes.MatchStringAny(diff.Line.Code, secret) {
                continue
            }

            result = append(result, &rule.FileChangeFinding{
                FileRange: &structures.FileRange{
                    StartLineNum:     diff.Line.LineNumFile,
                    StartIndex:       rang.StartIndex,
                    EndLineNum:       diff.Line.LineNumFile,
                    EndIndex:         rang.EndIndex,
                    StartDiffLineNum: diff.Line.LineNumDiff,
                    EndDiffLineNum:   diff.Line.LineNumDiff,
                },
                SecretValues: []string{secret},
            })
        }

        if ok := diff.Increment(); !ok {
            break
        }
    }

    return
}

func (p *EntropyProcessor) FindInLine(string) (result []*rule.LineFinding, err error) {
    return
}
