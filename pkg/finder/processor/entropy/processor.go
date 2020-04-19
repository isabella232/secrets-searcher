package entropy

import (
    "encoding/base64"
    "encoding/hex"
    "fmt"
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

type Processor struct {
    name             string
    Charset          string
    LengthThreshold  int
    EntropyThreshold float64
    skipPEMs         bool
    whitelistCodeRes *structures.RegexpSet
}

func NewProcessor(name, charset string, lengthThreshold int, entropyThreshold float64, whitelistCodeRes *structures.RegexpSet, skipPEMs bool) (result *Processor) {
    return &Processor{
        name:             name,
        Charset:          charset,
        LengthThreshold:  lengthThreshold,
        EntropyThreshold: entropyThreshold,
        skipPEMs:         skipPEMs,
        whitelistCodeRes: whitelistCodeRes,
    }
}

func (p *Processor) Name() string {
    return p.name
}

func (p *Processor) FindInFileChange(fileChange *git.FileChange, _ *git.Commit, log logrus.FieldLogger) (result []*finder.ProcFinding, ignore []*structures.FileRange, err error) {
    var diff *diffpkg.Diff
    var dErr error
    diff, err = fileChange.Diff()
    if err != nil {
        err = errors.WithMessagev(err, "unable to get diff for file change", fileChange.Path)
        return
    }

    if p.skipPEMs && strings.HasSuffix(fileChange.Path, ".pem") {
        log.Debug("skipping PEM file because skipPEMs is true")
        return
    }

    for {
        // Skip PEM files of all types
        if p.skipPEMs {
            dErr = p.skipPEMsInDiff(diff)
            if diffpkg.IsEOF(dErr) {
                break
            } else if dErr != nil {
                err = errors.WithMessage(dErr, "unable to skip PEMs")
                return
            }
        }

        // Get to an add line
        dErr = diff.UntilTrueIncrement(func(line *diffpkg.Line) bool { return diff.Line.IsAdd })
        if diffpkg.IsEOF(dErr) {
            break
        } else if dErr != nil {
            err = errors.WithMessage(dErr, "unable to skip PEMs")
            return
        }

        // Find entropy in line
        var findings []*finder.ProcFinding
        findings, err = p.findEntropyInLine(diff.Line)
        if err != nil {
            err = errors.WithMessage(err, "unable to search for high entropy words, continuing to next line")
            if err = diff.Increment(); err != nil {
                return
            }
            continue
        }
        if findings != nil {
            result = append(result, findings...)
        }

        if err = diff.Increment(); err != nil {
            return
        }
    }

    return
}

func (p *Processor) skipPEMsInDiff(diff *diffpkg.Diff) (err error) {
    for {
        if diff.Line.CodeMatches(pemBeginHeaderRegex) {
            if err = diff.UntilTrueIncrement(func(line *diffpkg.Line) bool {
                return line.CodeMatches(pemEndHeaderRegex)
            }); err != nil {
                return
            }
            if err = diff.Increment(); err != nil {
                return
            }
            continue
        }
        if diff.Line.CodeMatches(pemBeginPyMultilineRegex) {
            if err = diff.UntilTrueIncrement(func(line *diffpkg.Line) bool {
                return line.CodeMatches(pemEndPyMultilineRegex)
            }); err != nil {
                return
            }
            if err = diff.Increment(); err != nil {
                return
            }
            continue
        }
        if diff.Line.CodeMatches(pemXMLStartRegex) {
            if err = diff.UntilTrueIncrement(func(line *diffpkg.Line) bool {
                return line.CodeMatches(pemXMLEndRegex)
            }); err != nil {
                return
            }
            continue
        }
        if diff.Line.CodeMatches(pemJsonLineRegex) {
            if err = diff.Increment(); err != nil {
                return
            }
            continue
        }
        break
    }
    return
}

func (p *Processor) FindInLine(string, logrus.FieldLogger) (result []*finder.ProcFindingInLine, ignore []*structures.LineRange, err error) {
    return
}

func (p *Processor) findEntropyInLine(diffLine *diffpkg.Line) (result []*finder.ProcFinding, err error) {
    var ranges []*structures.LineRangeValue
    ranges, err = entropypkg.FindHighEntropyWords(diffLine.Code, p.Charset, p.LengthThreshold, p.EntropyThreshold)
    if err != nil || ranges == nil {
        return
    }

    for _, rang := range ranges {
        if p.isSecretWhitelisted(diffLine.Code, rang) {
            continue
        }

        var extras []*finder.ProcExtra

        secretValue := rang.Value

        // Try to decode base64
        var decoded []byte
        var decodedString string
        var encoding string
        var decodeErr error
        switch p.Charset {
        case entropypkg.Base64CharsetName:
            decoded, decodeErr = base64.StdEncoding.DecodeString(secretValue)
            if decodeErr == nil {
                decodedString = string(decoded)
                encoding = "base64"
            }
        case entropypkg.HexCharsetName:
            decoded, decodeErr = hex.DecodeString(secretValue)
            if decodeErr == nil {
                decodedString = string(decoded)
                encoding = "hex"
            }
        }

        if decodedString != "" {
            extras = append(extras, &finder.ProcExtra{
                Key:    fmt.Sprintf("decoded-%s", encoding),
                Header: fmt.Sprintf("Decoded (%s)", encoding),
                Value:  decodedString,
                Code:   true,
            })
        }

        result = append(result, &finder.ProcFinding{
            FileRange:    structures.NewFileRangeFromLineRange(rang.LineRange, diffLine.LineNumFile),
            Secret:       &finder.ProcSecret{Value: secretValue},
            SecretExtras: extras,
        })
    }

    return
}

func (p *Processor) isSecretWhitelisted(line string, secret *structures.LineRangeValue) bool {
    return p.whitelistCodeRes != nil && p.whitelistCodeRes.MatchAndTestSubmatchOrOverlap(line, secret.LineRange)
}
