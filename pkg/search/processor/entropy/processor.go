package entropy

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/pantheon-systems/search-secrets/pkg/search"

	entropypkg "github.com/pantheon-systems/search-secrets/pkg/entropy"
	"github.com/pantheon-systems/search-secrets/pkg/errors"
	"github.com/pantheon-systems/search-secrets/pkg/git"
	"github.com/pantheon-systems/search-secrets/pkg/logg"
	"github.com/pantheon-systems/search-secrets/pkg/manip"
	"github.com/pantheon-systems/search-secrets/pkg/search/contract"
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
	charset          string
	lengthThreshold  int
	entropyThreshold float64
	skipPEMs         bool
	codeWhitelist    *search.CodeWhitelist
	log              logg.Logg
}

func NewProcessor(name, charset string, lengthThreshold int, entropyThreshold float64, codeWhitelist *search.CodeWhitelist, skipPEMs bool, log logg.Logg) (result *Processor) {
	return &Processor{
		name:             name,
		charset:          charset,
		lengthThreshold:  lengthThreshold,
		entropyThreshold: entropyThreshold,
		skipPEMs:         skipPEMs,
		codeWhitelist:    codeWhitelist,
		log:              log,
	}
}

func (p *Processor) GetName() string {
	return p.name
}

func (p *Processor) FindResultsInFileChange(job contract.ProcessorJobI) (err error) {
	diff := job.Diff()
	if p.skipPEMs && strings.HasSuffix(job.FileChange().Path, ".pem") {
		job.Log(p.log).Debug("skipping PEM file because skipPEMs is true")
		return
	}

	defer func() {
		if err != nil {
			err = errors.WithMessagev(err, "????", job.Log(p.log))
		}
	}()

	for {
		// Skip PEM files of all types
		if p.skipPEMs {
			if ok := p.skipPEMsInDiff(diff); !ok {
				break
			}
		}

		// Get to an add line
		if !diff.UntilTrueIncrement(func(line *git.Line) bool { return diff.Line.IsAdd }) {
			break
		}

		// Find entropy in line
		job.SearchingLine(diff.Line.NumInFile)
		p.findEntropyInLine(job, diff.Line)

		if !diff.Incr() {
			break
		}
	}

	return
}

func (p *Processor) findEntropyInLine(job contract.ProcessorJobI, diffLine *git.Line) {
	entResults := entropypkg.FindHighEntropyWords(diffLine.Code, p.charset, p.lengthThreshold, p.entropyThreshold)
	if entResults == nil {
		return
	}

	for _, entResult := range entResults {
		if p.codeWhitelist.IsSecretWhitelisted(diffLine.Code, entResult.LineRange.LineRange) {
			continue
		}
		secretValue := entResult.LineRange.Value

		var secretExtras []*contract.ResultExtra

		// Entropy value extra
		secretExtras = append(secretExtras, &contract.ResultExtra{
			Key:    fmt.Sprintf("entropy-%s", p.charset),
			Header: fmt.Sprintf("Entropy (%s)", p.charset),
			Value:  fmt.Sprintf("%f", entResult.Entropy),
		})

		// Try to decode base64
		//var decoded []byte
		//var decodedString string
		//var encoding string
		//var decodeErr error
		//switch p.charset {
		//case entropypkg.Base64CharsetName:
		//	decoded, decodeErr = base64.StdEncoding.DecodeString(secretValue)
		//	if decodeErr == nil {
		//		decodedString = string(decoded)
		//		encoding = "base64"
		//	}
		//case entropypkg.HexCharsetName:
		//	decoded, decodeErr = hex.DecodeString(secretValue)
		//	if decodeErr == nil {
		//		decodedString = string(decoded)
		//		encoding = "hex"
		//	}
		//}
		//
		//// Secret extras
		//
		//// Decoded value extra
		//if decodedString != "" {
		//	secretExtras = append(secretExtras, &contract.ResultExtra{
		//		Key:    fmt.Sprintf("decoded-%s", encoding),
		//		Header: fmt.Sprintf("Decoded (%s)", encoding),
		//		Value:  decodedString,
		//		Code:   true,
		//	})
		//}

		fileRange := manip.NewFileRangeFromLineRange(entResult.LineRange.LineRange, diffLine.NumInFile)
		job.SubmitResult(&contract.Result{
			FileRange:    fileRange,
			SecretValue:  secretValue,
			SecretExtras: secretExtras,
		})
	}

	return
}

func (p *Processor) skipPEMsInDiff(diff *git.Diff) (ok bool) {
	for {

		// If we see a PEM header, skip until the footer, then increment once more
		if diff.Line.Matches(pemBeginHeaderRegex) {
			if ok = !diff.UntilMatchesIncrement(pemEndHeaderRegex); !ok {
				return
			}
			if ok = !diff.Incr(); !ok {
				return
			}
			continue
		}

		// If we see a PEM header in python code, skip until the footer, then increment once more
		if diff.Line.Matches(pemBeginPyMultilineRegex) {
			if ok = !diff.UntilMatchesIncrement(pemEndPyMultilineRegex); !ok {
				return
			}
			if ok = !diff.UntilMatchesIncrement(pemEndPyMultilineRegex); !ok {
				return
			}
			if ok = !diff.Incr(); !ok {
				return
			}
			continue
		}

		// If we see a PEM header in specific XML code, skip until the footer
		if diff.Line.Matches(pemXMLStartRegex) {
			if !diff.UntilMatchesIncrement(pemXMLEndRegex) {
				return
			}
			continue
		}

		// If we see a PEM header in a single line JSON line, increment to skip it
		if diff.Line.Matches(pemJsonLineRegex) {
			if ok = !diff.Incr(); !ok {
				return
			}
			continue
		}

		// If nothing is matching, we're done for now
		break
	}

	return
}
