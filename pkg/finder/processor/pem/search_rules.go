package pem

import (
    diffpkg "github.com/pantheon-systems/search-secrets/pkg/diff"
    "github.com/pantheon-systems/search-secrets/pkg/finder"
    "github.com/pantheon-systems/search-secrets/pkg/structures"
    "regexp"
    "strings"
)

var onlyBase64Re = regexp.MustCompile(`^[a-zA-Z0-9+/]+={0,2}$`)

type searchRule func() (result *finder.ProcFinding, ignore *structures.FileRange, err error)

func (s *search) getRules() (result []searchRule) {
    return []searchRule{
        s.ignoreInvalidBlockData,
        s.findAddedKey,
        s.findRotatedKey,
        s.findJSONKey,
        s.ignoreMalformedKey,
    }
}

// Multiline blocks with invalid data
func (s *search) ignoreInvalidBlockData() (result *finder.ProcFinding, ignore *structures.FileRange, err error) {
    // Multiline
    if strings.Contains(s.diff.Line.Code, s.footer) {
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
    var i int
    if dErr = s.diff.UntilTrueCollectTrimmedCode(func(line *diffpkg.Line) bool {
        i += 1
        return strings.Contains(line.Code, s.footer)
    }, &keyLines, " "); dErr != nil && !diffpkg.IsEOF(dErr) {
        err = dErr
        return
    }

    // We didn't find it
    if i > 100 {
        if dErr = s.diff.SetLine(headerLine); dErr != nil && !diffpkg.IsEOF(dErr) {
            err = dErr
        }
        return
    }

    keyString := strings.Trim(strings.Join(keyLines, ""), " ")

    // Only base64 characters?
    if keyString == "" || onlyBase64Re.MatchString(keyString) {
        if dErr = s.diff.SetLine(headerLine); dErr != nil && !diffpkg.IsEOF(dErr) {
            err = dErr
        }
        return
    }

    footerLine := s.diff.Line.LineNumFile
    footerLineLen := len(s.diff.Line.Code)

    ignore = &structures.FileRange{
        StartLineNum: headerLine,
        StartIndex:   0,
        EndLineNum:   footerLine,
        EndIndex:     footerLineLen,
    }

    // Set the diff to the next line to be searched
    if dErr = s.diff.Increment(); dErr != nil && !diffpkg.IsEOF(dErr) {
        err = dErr
        return
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
func (s *search) findAddedKey() (result *finder.ProcFinding, _ *structures.FileRange, err error) {
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
    footerLine := s.diff.Line.LineNumFile
    footerLineLen := len(s.diff.Line.Code)

    fileRange := &structures.FileRange{
        StartLineNum: headerLine,
        StartIndex:   0,
        EndLineNum:   footerLine,
        EndIndex:     footerLineLen,
    }

    // Set the diff to the next line to be searched
    if dErr = s.diff.Increment(); dErr != nil && !diffpkg.IsEOF(dErr) {
        err = dErr
        return
    }

    return s.potentialFinding(keyString, fileRange, true)
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
func (s *search) findRotatedKey() (result *finder.ProcFinding, _ *structures.FileRange, err error) {
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
    footerLineLen := len(s.diff.Line.Code)
    fileRange := &structures.FileRange{
        StartLineNum: headerLine,
        StartIndex:   0,
        EndLineNum:   footerLine,
        EndIndex:     footerLineLen,
    }

    // Set the diff to the next line to be searched
    if dErr = s.diff.Increment(); dErr != nil && !diffpkg.IsEOF(dErr) {
        err = dErr
        return
    }

    return s.potentialFinding(keyString, fileRange, true)
}

// Escaped string object line:
// JSON:
// +    "key": "-----BEGIN RSA PRIVATE KEY-----\n[...]\n[...]\n[...]\n-----END RSA PRIVATE KEY-----\n",
// or Ruby:
// +    "key" =>"-----BEGIN RSA PRIVATE KEY-----[...]----END RSA PRIVATE KEY-----",
func (s *search) findJSONKey() (result *finder.ProcFinding, ignore *structures.FileRange, err error) {
    matches := s.oneLineEscapedStringKeyRe.FindStringSubmatch(s.diff.Line.Code)
    if len(matches) == 0 {
        return
    }

    matched := matches[1]
    keyBlock := strings.ReplaceAll(matched, "\\n", "\n")
    keyString := s.buildKey(keyBlock)

    fileRange := &structures.FileRange{
        StartLineNum: s.diff.Line.LineNumFile,
        StartIndex:   0,
        EndLineNum:   s.diff.Line.LineNumFile,
        EndIndex:     len(s.diff.Line.Code),
    }

    // If there weren't any line breaks it's invalid so ignore
    if keyBlock == matched {
        ignore = fileRange
        return
    }

    // Set the diff to the next line to be searched
    if dErr := s.diff.Increment(); dErr != nil && !diffpkg.IsEOF(dErr) {
        err = dErr
        return
    }

    return s.potentialFinding(keyString, fileRange, false)
}

// Ignore malformed keys
func (s *search) ignoreMalformedKey() (_ *finder.ProcFinding, ignore *structures.FileRange, err error) {
    match := s.oneLineKeyTooPermissiveRe.FindStringIndex(s.diff.Line.Code)
    if match != nil {
        lineRange := structures.NewLineRange(match[0], match[1])
        ignore = structures.NewFileRangeFromLineRange(lineRange, s.diff.Line.LineNumFile)
    }

    if dErr := s.diff.Increment(); dErr != nil && !diffpkg.IsEOF(dErr) {
        err = dErr
        return
    }

    return
}
