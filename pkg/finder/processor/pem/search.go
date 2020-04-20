package pem

import (
    "fmt"
    diffpkg "github.com/pantheon-systems/search-secrets/pkg/diff"
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    "github.com/pantheon-systems/search-secrets/pkg/finder"
    "github.com/pantheon-systems/search-secrets/pkg/git"
    "github.com/pantheon-systems/search-secrets/pkg/structures"
    "github.com/sirupsen/logrus"
    "regexp"
    "strings"
)

const (
    // Right now, only added or equal lines (rotation) are cared about. This can change if needed.
    skipDeletedHeaders = true
)

type search struct {
    pemType                   PEMTypeEnum
    header                    string
    footer                    string
    oneLineKeyRe              *regexp.Regexp
    oneLineKeyTooPermissiveRe *regexp.Regexp
    oneLineEscapedStringKeyRe *regexp.Regexp
    whitelistCodeRes          *structures.RegexpSet
    fileChange                *git.FileChange
    commit                    *git.Commit
    diff                      *diffpkg.Diff
    findings                  *[]*finder.ProcFinding
    ignores                   *[]*structures.FileRange
    log                       logrus.FieldLogger
}

func (s *search) find() (err error) {
    var finding *finder.ProcFinding
    var ignore *structures.FileRange
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

        finding, ignore = s.findKey()
        if finding != nil || ignore != nil {
            if finding != nil {
                *s.findings = append(*s.findings, finding)
            }
            if ignore != nil {
                *s.ignores = append(*s.ignores, ignore)
            }
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
func (s *search) findKey() (result *finder.ProcFinding, ignore *structures.FileRange) {
    var dErr error

    log := s.lg()
    originalLine := s.diff.Line.LineNum

    for _, rule := range s.getRules() {

        // Keep us on the same line for each try
        dErr = s.diff.SetLine(originalLine)
        if dErr != nil {
            if !diffpkg.IsEOF(dErr) {
                errors.ErrLog(log, dErr).Warn(dErr, "error setting line")
            }
            return
        }

        // Try to find
        result, ignore, dErr = rule()
        if dErr != nil {
            if !diffpkg.IsEOF(dErr) {
                errors.ErrLog(log, dErr).Warn(dErr, "error while finding key")
                continue
            }
            return
        }

        // If we found something, return it
        // The functions are each responsible for incrementing the diff to the next line that needs to be searched
        if result != nil || ignore != nil {
            return
        }
    }

    return
}

func (s *search) potentialFinding(keyString string, fileRange *structures.FileRange, isKeyFile bool) (result *finder.ProcFinding, _ *structures.FileRange, err error) {
    var fileContents string
    fileContents, err = s.fileChange.FileContents()
    if err != nil {
        return
    }

    lineRange := structures.NewLineRangeFromFileRange(fileRange, fileContents)
    if s.isSecretWhitelisted(keyString, lineRange.NewValue(keyString)) {
        return
    }

    found := found{pemType: s.pemType, fileChange: s.fileChange, commit: s.commit, log: s.lg()}
    return func() (result *finder.ProcFinding, _ *structures.FileRange, err error) {
        result, err = found.buildKeyFinding(keyString, fileRange, isKeyFile)
        return
    }()
}

func (s *search) buildKeyFromLines(keyLines []string) string {
    return s.buildKey(strings.Join(keyLines, "\n"))
}

func (s *search) buildKey(keyBlock string) string {
    return fmt.Sprintf("-----BEGIN %s-----\n%s\n-----END %s-----\n", s.pemType.Value(), keyBlock, s.pemType.Value())
}

func (s *search) isSecretWhitelisted(input string, secret *structures.LineRangeValue) bool {
    return s.whitelistCodeRes != nil && s.whitelistCodeRes.MatchAndTestSubmatchOrOverlap(input, secret.LineRange)
}

func (s *search) lg() logrus.FieldLogger {
    return s.log.WithFields(logrus.Fields{
        "code": s.diff.Line.Code,
        "line": s.diff.Line.LineNumFile,
    })
}
