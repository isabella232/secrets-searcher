package pem

import (
    "bytes"
    "crypto"
    "crypto/rsa"
    "crypto/tls"
    "crypto/x509"
    "encoding/pem"
    "fmt"
    "github.com/grantae/certinfo"
    diffpkg "github.com/pantheon-systems/search-secrets/pkg/diff"
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    "github.com/pantheon-systems/search-secrets/pkg/finder"
    "github.com/pantheon-systems/search-secrets/pkg/git"
    "github.com/pantheon-systems/search-secrets/pkg/structures"
    "github.com/sirupsen/logrus"
    "golang.org/x/crypto/ssh"
    "path"
    "reflect"
    "regexp"
    "strings"
)

const (
    // Right now, only added or equal lines (rotation) are cared about. This can change if needed.
    skipDeletedHeaders = true
)

var (
    certificateRe = regexp.MustCompile(`(?s)-----BEGIN CERTIFICATE-----[^-]+-----END CERTIFICATE-----`)
)

type search struct {
    pemType            PEMTypeEnum
    header             string
    footer             string
    oneLineJSONPattern *regexp.Regexp
    whitelistCodeRes   structures.RegexpSet
    fileChange         *git.FileChange
    commit             *git.Commit
    diff               *diffpkg.Diff
    findings           *[]*finder.ProcFinding
    ignore             *[]*structures.FileRange
    log                logrus.FieldLogger
}

func (s *search) find() (err error) {
    var finding *finder.ProcFinding
    var dErr error

    for {
        // Advance to the next line that contains the header
        if dErr = s.diff.UntilTrueIncrement(func(line *diffpkg.Line) bool {

            // Right now, only added or equal lines (rotation) are cared about. This can change if needed.
            if skipDeletedHeaders && line.IsDel {
                return false
            }

            return line.CodeContains(s.header)
        }); dErr != nil {
            if !diffpkg.IsEOF(dErr) {
                err = errors.WithMessage(dErr, "unable to increment to header")
            }
            return
        }

        finding = s.findKey()
        if finding != nil {
            *s.findings = append(*s.findings, finding)
            continue
        }

        s.lg().Warn("unable to parse string in code")

        if err = s.diff.Increment(); err != nil {
            break
        }
    }

    return
}

// Find a potential key
func (s *search) findKey() (result *finder.ProcFinding) {
    var dErr error
    log := s.lg()
    originalLine := s.diff.Line.LineNum

    funcs := []func() (*finder.ProcFinding, error){
        s.findAddedKey,
        s.findRotatedKey,
        s.findJSONKey,
    }

    for _, fnc := range funcs {

        // Keep us on the same line for each try
        dErr = s.diff.SetLine(originalLine)
        if dErr != nil {
            if !diffpkg.IsEOF(dErr) {
                errors.ErrorLogger(log, dErr).Warn(dErr, "error setting line")
            }
            return
        }

        // Try to find
        result, dErr = fnc()
        if dErr != nil {
            if !diffpkg.IsEOF(dErr) {
                errors.ErrorLogger(log, dErr).Warn(dErr, "error while finding key")
                continue
            }
            return
        }

        // If we found something, return it
        // The functions are each responsible for incrementing the diff to the next line that needs to be searched
        if result != nil {
            return
        }
    }

    return
}

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
func (s *search) findAddedKey() (result *finder.ProcFinding, err error) {
    if !s.diff.Line.IsAdd || !strings.HasSuffix(s.diff.Line.Code, s.header) {
        return
    }

    var dErr error
    var keyLines []string

    // Get header line number
    headerLine := s.diff.Line.LineNumFile

    // Go to the first line of the block
    if err = s.diff.Increment(); err != nil {
        return
    }

    // Collect block lines
    if dErr = s.diff.UntilTrueCollectTrimmedCode(func(line *diffpkg.Line) bool {
        return strings.HasPrefix(strings.TrimLeft(line.Code, " "), s.footer)
    }, &keyLines, " "); dErr != nil && !diffpkg.IsEOF(dErr) {
        err = dErr
        return
    }
    keyString := s.buildKeyFromLines(keyLines)

    // Get footer line number
    footerLine := s.diff.Line.LineNumFile

    // Set the diff to the next line to be searched
    if dErr = s.diff.Increment(); dErr != nil && !diffpkg.IsEOF(dErr) {
        err = dErr
        return
    }

    fileRange := &structures.FileRange{
        StartLineNum: headerLine,
        StartIndex:   0,
        EndLineNum:   footerLine + 1,
        EndIndex:     0,
    }

    return s.buildKeyFinding(keyString, fileRange, true)
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
func (s *search) findRotatedKey() (result *finder.ProcFinding, err error) {
    if !s.diff.Line.IsEqu || !strings.HasSuffix(s.diff.Line.Code, s.header) {
        return
    }

    var dErr error
    var keyLines []string
    var nextLine *diffpkg.Line

    // The next line should be a delete
    nextLine, dErr = s.diff.PeekNextLine()
    if dErr != nil || !nextLine.IsDel {
        return
    }

    // Get header line number
    headerLine := s.diff.Line.LineNumFile

    // Increment until we hit the first add line
    if err = s.diff.UntilTrueIncrement(func(line *diffpkg.Line) bool { return line.IsAdd }); err != nil {
        return
    }

    // Collect all added lines
    if err = s.diff.WhileTrueCollectCode(func(line *diffpkg.Line) bool { return line.IsAdd }, &keyLines); err != nil {
        return
    }
    keyString := s.buildKeyFromLines(keyLines)

    // Get footer line number
    footerLine := s.diff.Line.LineNumFile

    // Set the diff to the next line to be searched
    if dErr = s.diff.Increment(); dErr != nil && !diffpkg.IsEOF(dErr) {
        err = dErr
        return
    }

    fileRange := &structures.FileRange{
        StartLineNum: headerLine,
        StartIndex:   0,
        EndLineNum:   footerLine + 1,
        EndIndex:     0,
    }

    return s.buildKeyFinding(keyString, fileRange, true)
}

// JSON object line:
//
// +    "key": "-----BEGIN RSA PRIVATE KEY-----\n[...]\n[...]\n[...]\n-----END RSA PRIVATE KEY-----\n",
func (s *search) findJSONKey() (result *finder.ProcFinding, err error) {
    matches := s.oneLineJSONPattern.FindStringSubmatch(s.diff.Line.Code)
    if len(matches) == 0 {
        return
    }

    var dErr error

    keyBlock := strings.ReplaceAll(matches[1], "\\n", "\n")
    keyString := s.buildKey(keyBlock)

    // Set the diff to the next line to be searched
    if dErr = s.diff.Increment(); dErr != nil && !diffpkg.IsEOF(dErr) {
        err = dErr
        return
    }

    fileRange := &structures.FileRange{
        StartLineNum: s.diff.Line.LineNumFile,
        StartIndex:   0,
        EndLineNum:   s.diff.Line.LineNumFile,
        EndIndex:     len(s.diff.Line.Code) - 1,
    }

    return s.buildKeyFinding(keyString, fileRange, false)
}

func (s *search) buildKeyFinding(keyString string, fileRange *structures.FileRange, isKeyFile bool) (result *finder.ProcFinding, err error) {
    switch s.pemType {
    case RSAPrivateKey{}.New():
        return s.buildRSAPrivateKeyFinding(keyString, fileRange, isKeyFile)
    default:
        return s.buildGeneralKeyFinding(keyString, fileRange)
    }
}

func (s *search) buildGeneralKeyFinding(keyString string, fileRange *structures.FileRange) (result *finder.ProcFinding, err error) {
    _, err = s.parseX509PEMString(keyString)
    if err != nil {
        err = errors.WithMessagev(err, "invalid x509 PEM key", keyString)
        return
    }

    result = &finder.ProcFinding{
        FileRange: fileRange,
        Secret:    &finder.ProcSecret{Value: keyString},
    }
    return
}

func (s *search) buildRSAPrivateKeyFinding(keyString string, fileRange *structures.FileRange, isKeyFile bool) (result *finder.ProcFinding, err error) {
    var secretExtras []*finder.ProcExtra
    var findingExtras []*finder.ProcExtra
    var key *rsa.PrivateKey

    key, err = s.parseRSAPrivateKeyX509PEMString(keyString)
    if err != nil {
        err = errors.WithMessagev(err, "invalid RSA private key", keyString)
        return
    }

    secretExtras = append(secretExtras, &finder.ProcExtra{
        Key:    "public-key-info",
        Header: "Public key info",
        Value:  s.publicKeyInfo(&key.PublicKey),
        Code:   true,
    })

    var keyPEMBlock *pem.Block
    keyPEMBlock, err = s.decodePEMString(keyString)
    if err != nil {
        err = errors.WithMessage(err, "unable to decode PEM string")
        return
    }

    if isKeyFile {
        var extras []*finder.ProcExtra
        extras, err = s.buildBundledCertExtras(keyPEMBlock)
        if err == nil && extras != nil {
            findingExtras = append(findingExtras, extras...)
        }

        if extras == nil {
            extras, err = s.buildPairedPublicKeyExtras()
            if err == nil && extras != nil {
                findingExtras = append(findingExtras, extras...)
            }
        }
    }

    result = &finder.ProcFinding{
        FileRange:     fileRange,
        Secret:        &finder.ProcSecret{Value: keyString},
        SecretExtras:  secretExtras,
        FindingExtras: findingExtras,
    }

    return
}

func (s *search) buildBundledCertExtras(keyPEMBlock *pem.Block) (result []*finder.ProcExtra, err error) {
    var cert *x509.Certificate
    var certPEMBlock *pem.Block
    var certPath string

    cert, certPEMBlock, certPath, err = s.lookForBundledCertificate(keyPEMBlock)
    if err != nil || cert == nil {
        return
    }

    buf := bytes.NewBuffer(nil)
    if err = pem.Encode(buf, certPEMBlock); err != nil {
        return
    }
    pemBlockString := buf.String()

    result = append(result, &finder.ProcExtra{
        Key:    "bundled-certificate-path",
        Header: "Bundled certificate path",
        Value:  certPath,
    })

    result = append(result, &finder.ProcExtra{
        Key:    "bundled-certificate",
        Header: "Bundled certificate",
        Value:  pemBlockString,
        Code:   true,
    })

    var certInfo string
    certInfo, err = certinfo.CertificateText(cert)
    if err != nil {
        err = errors.Wrap(err, "unable to get cert info")
        return
    }

    result = append(result, &finder.ProcExtra{
        Key:    "bundled-certificate-info",
        Header: "Bundled certificate info",
        Value:  certInfo,
        Code:   true,
    })

    return
}

func (s *search) buildPairedPublicKeyExtras() (result []*finder.ProcExtra, err error) {
    var pubKey *rsa.PublicKey
    var sshPubKey ssh.PublicKey
    var pubKeyPath string
    pubKey, sshPubKey, pubKeyPath, err = s.lookForPairedPublicKey()
    if err != nil {
        err = errors.Wrap(err, "unable to find paired public key")
        return
    }

    authorizedKey := ssh.MarshalAuthorizedKey(sshPubKey)

    result = append(result, &finder.ProcExtra{
        Key:    "paired-public-key-path",
        Header: "Public key file",
        Value:  pubKeyPath,
    })

    result = append(result, &finder.ProcExtra{
        Key:    "paired-public-key",
        Header: "Public key contents",
        Value:  bytes.NewBuffer(authorizedKey).String(),
        Code:   true,
    })

    result = append(result, &finder.ProcExtra{
        Key:    "paired-public-key-info",
        Header: "Public key info",
        Value:  s.publicKeyInfo(pubKey),
        Code:   true,
    })

    return
}

func (s *search) buildKeyFromLines(keyLines []string) string {
    return s.buildKey(strings.Join(keyLines, "\n"))
}

func (s *search) buildKey(keyBlock string) string {
    return fmt.Sprintf("-----BEGIN %s-----\n%s\n-----END %s-----\n", s.pemType.Value(), keyBlock, s.pemType.Value())
}

func (s *search) lookForPairedPublicKey() (result *rsa.PublicKey, sshPubKey ssh.PublicKey, pubKeyPath string, err error) {
    log := s.lg()

    pubPath := s.fileChange.Path + ".pub"
    fileContents, fcErr := s.commit.FileContents(pubPath)
    if fcErr != nil {
        err = fcErr
        return
    }

    fileBytes := []byte(fileContents)
    for {
        pubKey, comment, options, fileBytes, parseErr := ssh.ParseAuthorizedKey(fileBytes)
        if parseErr != nil {
            err = errors.Wrap(err, "unable to parse public key file")
            return
        }
        if pubKey == nil {
            log.Warn("no pub key found")
            break
        }

        if result != nil {
            log.Warn("we found another pubkey here, but it is being skipped")
            continue
        }

        log.WithField("path", pubPath).
            WithField("comment", comment).
            WithField("options", options).
            WithField("comment", comment).
            Debug("public key found")

        switch pubKey.Type() {
        case ssh.KeyAlgoRSA:
            rsaKey, ok := reflect.ValueOf(pubKey).Convert(reflect.TypeOf(&rsa.PublicKey{})).Interface().(*rsa.PublicKey)
            if !ok {
                errors.ErrorLogger(log, err).Warn("expecting an RSA public key in bundled certificate")
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

func (s *search) lookForBundledCertificate(keyPEMBlock *pem.Block) (result *x509.Certificate, pemBlock *pem.Block, pemPath string, err error) {
    if s.pemType != (RSAPrivateKey{}).New() {
        return
    }

    log := s.lg()

    var fileContentss []string
    var pemPaths []string
    fileContentss, pemPaths = s.findCertContents()
    if fileContentss == nil {
        return
    }

    for i, fileContents := range fileContentss {
        keyPEMBlockBytes := pem.EncodeToMemory(keyPEMBlock)

        fileBytes := []byte(fileContents)
        tlsCertificate, keyPairErr := tls.X509KeyPair(fileBytes, keyPEMBlockBytes)
        if keyPairErr != nil {
            log.Debug("invalid key pair")
            continue
        }

        x509Cert, parseErr := x509.ParseCertificate(tlsCertificate.Certificate[0])
        if parseErr != nil {
            err = errors.Wrap(parseErr, "unable to parse cert block")
            return
        }

        result = x509Cert // x509Cert.PublicKey could be *rsa.PublicKey, *ecdsa.PublicKey, or ed25519.PublicKey
        pemBlock = &pem.Block{
            Headers: make(map[string]string),
            Type:    Certificate{}.New().Value(),
            Bytes:   tlsCertificate.Certificate[0],
        }
        pemPath = pemPaths[i]

        return
    }

    return
}

func (s *search) findCertContents() (result, certPath []string) {
    var fcErr error
    var tryPath string
    var tryContents string
    ext := path.Ext(s.fileChange.Path)
    pathNoExt := strings.TrimSuffix(s.fileChange.Path, ext)

    if ext == ".key" {
        tryPath = pathNoExt + ".crt"
        tryContents, fcErr = s.commit.FileContents(tryPath)
        if fcErr == nil {
            result = append(result, tryContents)
            certPath = append(certPath, tryPath)
            return
        }
    }

    if ext == ".pem" {
        tryPath = s.fileChange.Path
        tryContents, fcErr = s.commit.FileContents(tryPath)
        if fcErr == nil {
            result = append(result, tryContents)
            certPath = append(certPath, tryPath)
            return
        }
    }

    if ext == ".py" {
        tryPath = s.fileChange.Path
        tryContents, fcErr = s.commit.FileContents(s.fileChange.Path)
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

func (s *search) parseX509PEMString(keyString string) (result crypto.PrivateKey, err error) {
    log := s.lg()

    var block *pem.Block
    block, err = s.decodePEMString(keyString)
    if err != nil {
        err = errors.Wrapv(err, "unable to parse PEM", keyString)
        return
    }

    if block.Type != s.pemType.Value() {
        err = errors.Errorf("PEM block should be \"%s\", not \"%s\"", s.pemType.Value(), block.Type)
        return
    }

    switch block.Type {
    case RSAPrivateKey{}.New().Value():
        var rsaKey *rsa.PrivateKey
        rsaKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
        if err != nil {
            err = errors.Wrapv(err, "unable to parse private key", keyString)
            return
        }
        if err = rsaKey.Validate(); err != nil {
            err = errors.Wrapv(err, "key is not valid", keyString)
            return
        }
        result = rsaKey
    default:
        log.Warnf("unsupported block type: %s", block.Type)
    }

    return
}

func (s *search) parseRSAPrivateKeyX509PEMString(keyString string) (result *rsa.PrivateKey, err error) {
    var privateKey crypto.PrivateKey
    privateKey, err = s.parseX509PEMString(keyString)
    if err != nil {
        err = errors.Wrapv(err, "unable to parse x509 PEM string", keyString)
        return
    }

    var ok bool
    result, ok = privateKey.(*rsa.PrivateKey)
    if !ok {
        err = errors.Wrap(err, "not an RSA private key")
    }

    return
}

func (s *search) decodePEMBytes(certBytes []byte) (result *pem.Block, err error) {
    var rest []byte
    result, rest = pem.Decode(certBytes)
    if result == nil {
        err = errors.Errorv("no blocks found in PEM string", string(certBytes))
        return
    }
    if len(rest) != 0 {
        err = errors.Errorv("extra input found in key string", string(rest))
        return
    }

    return
}

func (s *search) decodePEMString(certString string) (result *pem.Block, err error) {
    return s.decodePEMBytes([]byte(certString))
}

func (s *search) lg() logrus.FieldLogger {
    return s.log.WithFields(logrus.Fields{
        "code": s.diff.Line.Code,
        "line": s.diff.Line.LineNumFile,
    })
}

func (s *search) publicKeyInfo(rsaPublicKey *rsa.PublicKey) string {
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
