package pem

import (
	"bytes"
	"crypto"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"path"
	"reflect"
	"regexp"
	"strings"

	"github.com/grantae/certinfo"
	"github.com/pantheon-systems/secrets-searcher/pkg/search"
	"golang.org/x/crypto/ssh"

	"github.com/pantheon-systems/secrets-searcher/pkg/errors"
	"github.com/pantheon-systems/secrets-searcher/pkg/git"
	"github.com/pantheon-systems/secrets-searcher/pkg/logg"
	"github.com/pantheon-systems/secrets-searcher/pkg/manip"
	"github.com/pantheon-systems/secrets-searcher/pkg/search/contract"
)

const (
	RSAPrivateKey = "RSA PRIVATE KEY"
	Certificate   = "CERTIFICATE"

	// Right now, only added or equal lines (rotation) are cared about. This can change if needed.
	skipDeletedHeaders = true
)

var (
	certificateRe = regexp.MustCompile(`(?s)-----BEGIN CERTIFICATE-----[^-]+-----END CERTIFICATE-----`)
)

type (
	Processor struct {
		name                      string
		pemType                   string
		header                    string
		footer                    string
		oneLineKeyRe              *regexp.Regexp
		oneLineKeyTooPermissiveRe *regexp.Regexp
		oneLineEscapedStringKeyRe *regexp.Regexp
		codeWhitelist             *search.CodeWhitelist
		log                       logg.Logg
	}
	searchRule struct {
		name    string
		execute func(job contract.ProcessorJobI) (parsed bool, err error)
	}
)

func NewProcessor(name string, pemType string, log logg.Logg) (result *Processor) {
	header := fmt.Sprintf("-----BEGIN %s-----", pemType)
	footer := fmt.Sprintf("-----END %s-----", pemType)

	oneLineKeyRe := regexp.MustCompile(header + `\\n(.*)\\n` + footer)
	oneLineKeyTooPermissiveRe := regexp.MustCompile(`-BEGIN ` + pemType + `-(.*)-END ` + pemType + `-`)
	oneLineEscapedStringKeyRe := regexp.MustCompile(`"` + header + `\\n(.*)\\n` + footer + `\\?n?"`)

	// Incomplete/invalid/example keys
	// FIXME: These are too specific to Pantheon findings and should/can be generalized
	// FIXME: Code whitelist doesn't make sense in this processor, remove it somehow
	codeWhitelist := search.NewCodeWhitelist(nil, log.WithPrefix("code-whitelist"))
	codeWhitelist.Res.Add(regexp.MustCompile(header + `.{43}` + footer))
	codeWhitelist.Res.Add(regexp.MustCompile(`"` + header + `\n.{6}\.\.\."`))
	codeWhitelist.Res.Add(regexp.MustCompile(header + `,$`))
	codeWhitelist.Res.Add(regexp.MustCompile(`with ` + header))

	return &Processor{
		name:                      name,
		pemType:                   pemType,
		header:                    header,
		footer:                    footer,
		oneLineKeyRe:              oneLineKeyRe,
		oneLineKeyTooPermissiveRe: oneLineKeyTooPermissiveRe,
		oneLineEscapedStringKeyRe: oneLineEscapedStringKeyRe,
		codeWhitelist:             codeWhitelist,
		log:                       log,
	}
}

func (p *Processor) GetName() string {
	return p.name
}

func (p *Processor) FindResultsInFileChange(job contract.ProcessorJobI) (err error) {
	diff := job.Diff()

	for {

		// Advance to the next line that contains the header
		// Right now, only added or equal lines (cert rotation) are cared about.
		// This can change if needed.
		if !diff.UntilTrueIncrement(func(line *git.Line) bool {
			if skipDeletedHeaders && line.IsDel {
				return false
			}
			return line.Contains(p.header)
		}) {
			return
		}

		// Execute rules to parse the key we might have found.
		job.SearchingLine(diff.Line.NumInFile)
		p.executeRules(job)

		// If the key was parsed, we should be on the last line of the key lines.
		// Whether the key was parsed or not, we should continue.
		if !diff.Incr() {
			break
		}
	}

	return
}

func (p *Processor) potentialFinding(job contract.ProcessorJobI, keyString string, fileRange *manip.FileRange, isKeyFile bool) (parsed bool, err error) {
	var fileContents string
	fileContents, err = job.FileChange().FileContents()
	if err != nil {
		return
	}

	lineRange := manip.NewLineRangeFromFileRange(fileRange, fileContents)

	if p.codeWhitelist.IsSecretWhitelisted(keyString, lineRange) {
		return
	}

	parsed, err = p.buildKeyFinding(job, keyString, fileRange, isKeyFile)

	return
}

func (p *Processor) buildKeyFinding(job contract.ProcessorJobI, keyString string, fileRange *manip.FileRange, isKeyFile bool) (parsed bool, err error) {
	switch p.pemType {
	case RSAPrivateKey:
		return p.buildRSAPrivateKeyFinding(job, keyString, fileRange, isKeyFile)
	default:
		return p.buildGeneralKeyFinding(job, keyString, fileRange)
	}
}

func (p *Processor) buildGeneralKeyFinding(job contract.ProcessorJobI, keyString string, fileRange *manip.FileRange) (parsed bool, err error) {
	_, err = p.parseX509PEMString(job, keyString)
	if err != nil {
		err = errors.WithMessagev(err, "invalid x509 PEM key", keyString, job.Log(p.log))
		return
	}

	fileBasename := path.Base(job.FileChange().Path)

	job.SubmitResult(&contract.Result{
		FileRange:    fileRange,
		SecretValue:  keyString,
		FileBasename: fileBasename,
	})
	parsed = true

	return
}

func (p *Processor) buildKeyFromBlockLines(keyLines []string) string {
	return p.buildKey(strings.Join(keyLines, "\n"))
}

func (p *Processor) buildKey(keyBlock string) string {
	return fmt.Sprintf("-----BEGIN %s-----\n%s\n-----END %s-----\n", p.pemType, keyBlock, p.pemType)
}
func (p *Processor) buildRSAPrivateKeyFinding(job contract.ProcessorJobI, keyString string, fileRange *manip.FileRange, isKeyFile bool) (parsed bool, err error) {
	var secretExtras []*contract.ResultExtra
	var findingExtras []*contract.ResultExtra
	var key *rsa.PrivateKey

	key, err = p.parseRSAPrivateKeyX509PEMString(job, keyString)
	if err != nil {
		err = errors.WithMessagev(err, "invalid RSA private key", keyString, job.Log(p.log))
		return
	}

	secretExtras = append(secretExtras, &contract.ResultExtra{
		Key:    "public-key-info",
		Header: "Public key info",
		Value:  p.publicKeyInfo(&key.PublicKey),
		Code:   true,
	})

	var keyPEMBlock *pem.Block
	keyPEMBlock, err = p.decodePEMString(job, keyString)
	if err != nil {
		err = errors.WithMessagev(err, "unable to decode PEM string", job.Log(p.log))
		return
	}

	if isKeyFile {
		var extras []*contract.ResultExtra
		extras, err = p.buildBundledCertExtras(job, keyPEMBlock)
		if err == nil && extras != nil {
			findingExtras = append(findingExtras, extras...)
		}

		if extras == nil {
			extras, err = p.buildPairedPublicKeyExtras(job)
			if err == nil && extras != nil {
				findingExtras = append(findingExtras, extras...)
			}
		}
	}

	fileBasename := path.Base(job.FileChange().Path)
	fileExt := path.Ext(job.FileChange().Path)
	if fileExt != ".pem" {
		fileBasename += ".pem"
	}

	job.SubmitResult(&contract.Result{
		FileRange:     fileRange,
		SecretValue:   keyString,
		SecretExtras:  secretExtras,
		FindingExtras: findingExtras,
		FileBasename:  fileBasename,
	})
	parsed = true

	return
}

func (p *Processor) buildBundledCertExtras(job contract.ProcessorJobI, keyPEMBlock *pem.Block) (result []*contract.ResultExtra, err error) {
	var cert *x509.Certificate
	var certPEMBlock *pem.Block
	var certPath string

	cert, certPEMBlock, certPath, err = p.lookForBundledCertificate(job, keyPEMBlock)
	if err != nil || cert == nil {
		return
	}

	buf := bytes.NewBuffer(nil)
	if err = pem.Encode(buf, certPEMBlock); err != nil {
		return
	}
	pemBlockString := buf.String()

	result = append(result, &contract.ResultExtra{
		Key:    "bundled-certificate-path",
		Header: "Bundled certificate path",
		Value:  certPath,
	})

	result = append(result, &contract.ResultExtra{
		Key:    "bundled-certificate",
		Header: "Bundled certificate",
		Value:  pemBlockString,
		Code:   true,
	})

	var certInfo string
	certInfo, err = certinfo.CertificateText(cert)
	if err != nil {
		err = errors.Wrapv(err, "unable to get cert info", job.Log(p.log))
		return
	}

	result = append(result, &contract.ResultExtra{
		Key:    "bundled-certificate-info",
		Header: "Bundled certificate info",
		Value:  certInfo,
		Code:   true,
	})

	return
}

func (p *Processor) buildPairedPublicKeyExtras(job contract.ProcessorJobI) (result []*contract.ResultExtra, err error) {
	var pubKey *rsa.PublicKey
	var sshPubKey ssh.PublicKey
	var pubKeyPath string
	pubKey, sshPubKey, pubKeyPath, err = p.lookForPairedPublicKey(job)
	if err != nil || sshPubKey == nil {
		return
	}

	authorizedKey := ssh.MarshalAuthorizedKey(sshPubKey)

	result = append(result, &contract.ResultExtra{
		Key:    "paired-public-key-path",
		Header: "Public key file",
		Value:  pubKeyPath,
	})

	result = append(result, &contract.ResultExtra{
		Key:    "paired-public-key",
		Header: "Public key contents",
		Value:  bytes.NewBuffer(authorizedKey).String(),
		Code:   true,
	})

	result = append(result, &contract.ResultExtra{
		Key:    "paired-public-key-info",
		Header: "Public key info",
		Value:  p.publicKeyInfo(pubKey),
		Code:   true,
	})

	return
}

func (p *Processor) lookForPairedPublicKey(job contract.ProcessorJobI) (result *rsa.PublicKey, sshPubKey ssh.PublicKey, pubKeyPath string, err error) {
	fileChange := job.FileChange()
	pubPath := fileChange.Path + ".pub"
	fileContents, fcErr := fileChange.FileContents()
	if fcErr != nil {
		return
	}

	fileBytes := []byte(fileContents)
	for {
		pubKey, comment, options, fileBytes, parseErr := ssh.ParseAuthorizedKey(fileBytes)
		if parseErr != nil || pubKey == nil {
			return
		}

		job.Log(p.log).WithField("path", pubPath).
			WithField("comment", comment).
			WithField("options", options).
			WithField("comment", comment).
			Debug("public key found")

		switch pubKey.Type() {
		case ssh.KeyAlgoRSA:
			rsaKey, ok := reflect.ValueOf(pubKey).Convert(reflect.TypeOf(&rsa.PublicKey{})).Interface().(*rsa.PublicKey)
			if !ok {
				errors.ErrLog(job.Log(p.log), err).Warn("expecting an RSA public key in bundled certificate")
				continue
			}
			result = rsaKey
			sshPubKey = pubKey
			pubKeyPath = pubPath
		}

		if len(fileBytes) == 0 {
			break
		}
	}

	return
}

func (p *Processor) lookForBundledCertificate(job contract.ProcessorJobI, keyPEMBlock *pem.Block) (result *x509.Certificate, pemBlock *pem.Block, pemPath string, err error) {
	if p.pemType != RSAPrivateKey {
		return
	}

	var fileContentss []string
	var pemPaths []string
	fileContentss, pemPaths = p.findCertContents(job)
	if fileContentss == nil {
		return
	}

	for i, fileContents := range fileContentss {
		keyPEMBlockBytes := pem.EncodeToMemory(keyPEMBlock)

		fileBytes := []byte(fileContents)
		tlsCertificate, keyPairErr := tls.X509KeyPair(fileBytes, keyPEMBlockBytes)
		if keyPairErr != nil {
			continue
		}

		x509Cert, parseErr := x509.ParseCertificate(tlsCertificate.Certificate[0])
		if parseErr != nil {
			err = errors.Wrapv(parseErr, "unable to parse cert block", tlsCertificate.Certificate[0], job.Log(p.log))
			return
		}

		result = x509Cert // x509Cert.PublicKey could be *rsa.PublicKey, *ecdsa.PublicKey, or ed25519.PublicKey
		pemBlock = &pem.Block{
			Headers: make(map[string]string),
			Type:    Certificate,
			Bytes:   tlsCertificate.Certificate[0],
		}
		pemPath = pemPaths[i]

		return
	}

	return
}

func (p *Processor) findCertContents(job contract.ProcessorJobI) (result, certPath []string) {
	commit := job.Commit()
	fileChange := job.FileChange()
	var fcErr error
	var tryPath string
	var tryContents string
	ext := path.Ext(fileChange.Path)
	pathNoExt := strings.TrimSuffix(fileChange.Path, ext)

	if ext == ".key" {
		tryPath = pathNoExt + ".crt"
		tryContents, fcErr = commit.FileContents(tryPath)
		if fcErr == nil {
			result = append(result, tryContents)
			certPath = append(certPath, tryPath)
			return
		}
	}

	if ext == ".pem" {
		tryPath = fileChange.Path
		tryContents, fcErr = commit.FileContents(tryPath)
		if fcErr == nil {
			result = append(result, tryContents)
			certPath = append(certPath, tryPath)
			return
		}
	}

	if ext == ".py" {
		tryPath = fileChange.Path
		tryContents, fcErr = commit.FileContents(fileChange.Path)
		if fcErr == nil {
			matches := certificateRe.FindAllString(tryContents, -1)
			if matches != nil {
				for _, cert := range matches {
					result = append(result, cert)
					certPath = append(certPath, tryPath)
				}
				return
			}
		}
	}

	return
}

func (p *Processor) parseX509PEMString(job contract.ProcessorJobI, keyString string) (result crypto.PrivateKey, err error) {
	var block *pem.Block
	block, err = p.decodePEMString(job, keyString)
	if err != nil {
		err = errors.Wrapv(err, "unable to parse PEM", keyString, job.Log(p.log))
		return
	}

	if block.Type != p.pemType {
		err = errors.Errorv(fmt.Sprintf("PEM block should be \"%s\", not \"%s\"", p.pemType, block.Type), job.Log(p.log))
		return
	}

	switch block.Type {
	case RSAPrivateKey:
		var rsaKey *rsa.PrivateKey
		rsaKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			err = errors.Wrapv(err, "unable to parse private key", keyString, job.Log(p.log))
			return
		}
		if err = rsaKey.Validate(); err != nil {
			err = errors.Wrapv(err, "key is not valid", keyString, job.Log(p.log))
			return
		}
		result = rsaKey
	default:
		job.Log(p.log).Warnf("unsupported block type: %s", block.Type)
	}

	return
}

func (p *Processor) parseRSAPrivateKeyX509PEMString(job contract.ProcessorJobI, keyString string) (result *rsa.PrivateKey, err error) {
	var privateKey crypto.PrivateKey
	privateKey, err = p.parseX509PEMString(job, keyString)
	if err != nil {
		err = errors.Wrapv(err, "unable to parse x509 PEM string", keyString, job.Log(p.log))
		return
	}

	var ok bool
	result, ok = privateKey.(*rsa.PrivateKey)
	if !ok {
		err = errors.Wrapv(err, "not an RSA private key", job.Log(p.log))
	}

	return
}

func (p *Processor) decodePEMBytes(job contract.ProcessorJobI, certBytes []byte) (result *pem.Block, err error) {
	var rest []byte
	result, rest = pem.Decode(certBytes)
	if result == nil {
		err = errors.Errorv("no blocks found in PEM string", string(certBytes), job.Log(p.log))
		return
	}
	if len(rest) != 0 {
		err = errors.Errorv("extra input found in key string", string(rest), job.Log(p.log))
		return
	}

	return
}

func (p *Processor) decodePEMString(job contract.ProcessorJobI, certString string) (result *pem.Block, err error) {
	return p.decodePEMBytes(job, []byte(certString))
}

func (p *Processor) publicKeyInfo(rsaPublicKey *rsa.PublicKey) string {
	var buf = bytes.NewBuffer(nil)
	buf.WriteString(fmt.Sprintf("Public Key Algorithm: RSA\n"))
	buf.WriteString(fmt.Sprintf("%4sPublic-Key: (%d bit)\n", "", rsaPublicKey.N.BitLen()))
	buf.WriteString(fmt.Sprintf("%4sModulus:", ""))
	for i, val := range rsaPublicKey.N.Bytes() {
		if (i % 15) == 0 {
			buf.WriteString(fmt.Sprintf("\n%20s", ""))
		}
		buf.WriteString(fmt.Sprintf("%02x", val))
		if i != len(rsaPublicKey.N.Bytes())-1 {
			buf.WriteString(":")
		}
	}
	buf.WriteString(fmt.Sprintf("\n%4sExponent: %d (%#x)\n", "", rsaPublicKey.E, rsaPublicKey.E))

	return buf.String()
}
