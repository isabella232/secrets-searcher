package processor

import (
    "encoding/base64"
    "encoding/hex"
    diffpkg "github.com/pantheon-systems/search-secrets/pkg/diff"
    entropypkg "github.com/pantheon-systems/search-secrets/pkg/entropy"
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    "github.com/pantheon-systems/search-secrets/pkg/finder"
    "github.com/pantheon-systems/search-secrets/pkg/git"
    "github.com/pantheon-systems/search-secrets/pkg/structures"
    "github.com/sirupsen/logrus"
    "regexp"
    "strings"
)

var (
    pemBeginHeaderRegex = regexp.MustCompile("-----BEGIN [^-]+-----$")
    pemEndHeaderRegex   = regexp.MustCompile("-----END [^-]+-----")

    pemBeginPyMultilineRegex = regexp.MustCompile(`"""-----BEGIN [^-]+-----$`)
    pemEndPyMultilineRegex   = regexp.MustCompile(`^-----BEGIN [^-]+-----"""$`)

    pemJsonLineRegex = regexp.MustCompile(`"[a-zA-Z_]+": "-----BEGIN [^-]+----- ?(?:\\r)?\\n`)

    pemXMLStartRegex = regexp.MustCompile(`<ds:X509Certificate>(.+)$`)
    pemXMLEndRegex   = regexp.MustCompile(`(.+)</ds:X509Certificate>`)
)

type EntropyProcessor struct {
    name             string
    Charset          string
    LengthThreshold  int
    EntropyThreshold float64
    skipPEMs         bool
    whitelistCodeRes *structures.RegexpSet
}

func NewEntropyProcessor(name, charset string, lengthThreshold int, entropyThreshold float64, whitelistCodeRes *structures.RegexpSet, skipPEMs bool) (result *EntropyProcessor) {
    return &EntropyProcessor{
        name:             name,
        Charset:          charset,
        LengthThreshold:  lengthThreshold,
        EntropyThreshold: entropyThreshold,
        skipPEMs:         skipPEMs,
        whitelistCodeRes: whitelistCodeRes,
    }
}

func (p *EntropyProcessor) Name() string {
    return p.name
}

func (p *EntropyProcessor) FindInFileChange(fileChange *git.FileChange, log *logrus.Entry) (result []*finder.Finding, ignore []*structures.FileRange, err error) {
    var diff *diffpkg.Diff
    diff, err = fileChange.Diff()
    if err != nil {
        return
    }

    if p.skipPEMs && strings.HasSuffix(fileChange.Path, ".pem") {
        log.Debug("skipping PEM file because skipPEMs is true")
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
                continue
            }
            if diff.Line.CodeMatches(pemBeginPyMultilineRegex) {
                if ok := diff.UntilTrueIncrement(func(line *diffpkg.Line) bool {
                    return line.CodeMatches(pemEndPyMultilineRegex)
                }); !ok {
                    break
                }
                if ok := diff.Increment(); !ok {
                    break
                }
                continue
            }
            if diff.Line.CodeMatches(pemJsonLineRegex) {
                if ok := diff.Increment(); !ok {
                    break
                }
                continue
            }
            if diff.Line.CodeMatches(pemXMLStartRegex) {
                if ok := diff.UntilTrueIncrement(func(line *diffpkg.Line) bool {
                    return line.CodeMatches(pemXMLEndRegex)
                }); !ok {
                    break
                }
                if ok := diff.Increment(); !ok {
                    break
                }
                continue
            }
        }

        // Get to an add line
        if !diff.Line.IsAdd {
            if ok := diff.UntilTrueIncrement(func(line *diffpkg.Line) bool { return diff.Line.IsAdd }); !ok {
                break
            }
            continue
        }

        // Find entropy in line
        var findings []*finder.Finding
        findings, err := p.findEntropyInLine(diff.Line)
        if err != nil {
            err = errors.WithMessage(err, "unable to search for high entropy words, continuing to next line")
            if ok := diff.Increment(); !ok {
                break
            }
            continue
        }
        if findings != nil {
            result = append(result, findings...)
        }

        if ok := diff.Increment(); !ok {
            break
        }
    }

    return
}

func (p *EntropyProcessor) FindInLine(string, *logrus.Entry) (result []*finder.FindingInLine, ignore []*structures.LineRange, err error) {
    return
}

func (p *EntropyProcessor) findEntropyInLine(diffLine *diffpkg.Line) (result []*finder.Finding, err error) {
    var ranges []*structures.LineRangeValue
    ranges, err = entropypkg.FindHighEntropyWords(diffLine.Code, p.Charset, p.LengthThreshold, p.EntropyThreshold)
    if err != nil || ranges == nil {
        return
    }

    for _, rang := range ranges {
        secretValue := rang.Value

        // Try to decode base64
        var decoded []byte
        var decodedString string
        var decodeErr error
        switch p.Charset {
        case entropypkg.Base64CharsetName:
            decoded, decodeErr = base64.StdEncoding.DecodeString(secretValue)
            if decodeErr == nil {
                decodedString = string(decoded)
            }
        case entropypkg.HexCharsetName:
            decoded, decodeErr = hex.DecodeString(secretValue)
            if decodeErr == nil {
                decodedString = string(decoded)
            }
        }

        if p.isSecretWhitelisted(diffLine.Code, rang) {
            continue
        }

        result = append(result, &finder.Finding{
            FileRange: structures.NewFileRangeFromLineRange(rang.LineRange, diffLine.LineNumFile, diffLine.LineNumDiff),
            Secret:    &finder.Secret{Value: secretValue, Decoded: decodedString},
        })
    }

    return
}

func (p *EntropyProcessor) isSecretWhitelisted(line string, secret *structures.LineRangeValue) bool {
    return p.whitelistCodeRes != nil && p.whitelistCodeRes.MatchAndTestSubmatchOrOverlap(line, secret.LineRange)
}
