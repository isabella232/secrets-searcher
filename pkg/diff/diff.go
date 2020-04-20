package diff

import (
    "fmt"
    "github.com/pantheon-systems/search-secrets/pkg/dbug"
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    "github.com/sirupsen/logrus"
    "strings"
)

const (
    equalPrefix  = " "
    addPrefix    = "+"
    deletePrefix = "-"
)

var errEOF = eofError{}

type (
    Diff struct {
        Line                 *Line
        lineStrings          []string
        lineStringsLen       int
        diffToFileLineNumMap map[int]int // FIXME This shouldn't need to be passed in
    }
    lineMatch  func(line *Line) bool
    lineAction func(line *Line)
    eofError   struct{}
)

func New(lineStrings []string, lineMap map[int]int) (result *Diff, err error) {
    diff := &Diff{
        lineStrings:          lineStrings,
        lineStringsLen:       len(lineStrings),
        diffToFileLineNumMap: lineMap,
    }

    if err = diff.SetLine(1); err != nil {
        err = errors.New("unable to set line to 1")
        return
    }

    result = diff

    return
}

func (d *Diff) WhileTrueDo(lineMatch lineMatch, lineAction lineAction) (err error) {
    for searching := lineMatch(d.Line); searching; searching = lineMatch(d.Line) {
        lineAction(d.Line)
        if err = d.Increment(); err != nil {
            err = errors.WithMessage(err, "unable to incrememt once while searching line matches")
            return
        }
    }
    return
}

func (d *Diff) WhileTrueCollectCode(lineMatch lineMatch, collected *[]string) (err error) {
    err = d.WhileTrueDo(lineMatch, func(line *Line) {
        *collected = append(*collected, line.Code)
    })
    if err != nil {
        err = errors.WithMessage(err, "unable to collect code while line matches")
    }
    return
}

func (d *Diff) WhileTrueCollectTrimmedCode(lineMatch lineMatch, collected *[]string, cutset string) (err error) {
    err = d.WhileTrueDo(lineMatch, func(line *Line) {
        *collected = append(*collected, strings.Trim(line.Code, cutset))
    })
    if err != nil {
        err = errors.WithMessage(err, "unable to collect trimmed code while line matches")
    }
    return
}

func (d *Diff) WhileTrueIncrement(lineMatch lineMatch) (err error) {
    err = d.WhileTrueDo(lineMatch, func(line *Line) {})
    if err != nil {
        err = errors.WithMessage(err, "unable to increment line while line matches")
    }
    return
}

func (d *Diff) UntilTrueCollectCode(lineMatch lineMatch, collected *[]string) (err error) {
    err = d.WhileTrueCollectCode(func(line *Line) bool { return !lineMatch(line) }, collected)
    if err != nil {
        err = errors.WithMessage(err, "unable to collect code until line matches")
    }
    return
}

func (d *Diff) UntilTrueCollectTrimmedCode(lineMatch lineMatch, collected *[]string, cutset string) (err error) {
    err = d.WhileTrueCollectTrimmedCode(func(line *Line) bool { return !lineMatch(line) }, collected, cutset)
    if err != nil {
        err = errors.WithMessage(err, "unable to collect trimmed code until line matches")
    }
    return
}

func (d *Diff) UntilTrueIncrement(lineMatch lineMatch) (err error) {
    err = d.WhileTrueIncrement(func(line *Line) bool { return !lineMatch(line) })
    if err != nil {
        err = errors.WithMessage(err, "unable to increment line until line matches")
    }
    return
}

func (d *Diff) Increment() (err error) {
    num := d.Line.LineNum + 1
    if err = d.SetLine(num); err != nil {
        err = errors.WithMessagef(err, "unable to increment line to %d", num)
    }
    return
}

func (d *Diff) SetLine(lineNum int) (err error) {
    var line *Line
    line, err = d.buildLine(lineNum)
    if err != nil {
        err = errors.WithMessage(err, "unable to build line")
        return
    }

    d.Line = line

    if dbug.Cnf.Enabled {
        lineNumFile, _ := d.fileLineNum(lineNum)
        if dbug.Cnf.FilterConfig.Line > -1 && lineNumFile == dbug.Cnf.FilterConfig.Line {
            fmt.Print("") // For breakpoint
        }
    }

    return
}

func (d *Diff) PeekNextLine() (result *Line, err error) {
    result, err = d.buildLine(d.Line.LineNum + 1)
    if err != nil {
        err = errors.WithMessage(err, "unable to peek at next line")
    }
    return
}

func (d *Diff) Lines() (result []string) {
    return d.lineStrings
}

func (d *Diff) String() (result string) {
    return strings.Join(d.lineStrings, "\n")
}

func (e eofError) Error() (result string) {
    return "EOF"
}

func IsEOF(err error) bool {
    cause := errors.Cause(err)
    switch cause.(type) {
    case eofError:
        return true
    }
    return false
}

func (d *Diff) buildLine(lineNumDiff int) (result *Line, err error) {
    if lineNumDiff < 1 {
        err = errors.New("cannot build a line less than 1")
    }
    if lineNumDiff > len(d.lineStrings) {
        err = errors.WithMessage(errEOF, "end of file")
        return
    }

    lineNumFile, ok := d.fileLineNum(lineNumDiff)
    if !ok {
        err = errors.Errorv("unable to get mapped line", lineNumDiff)
        return
    }

    result = NewLine(d.lineStrings[(lineNumDiff-1)], lineNumDiff, lineNumFile)

    return
}

func (d *Diff) fileLineNum(diffLineNum int) (result int, ok bool) {
    result, ok = d.diffToFileLineNumMap[diffLineNum]
    return
}

func (d *Diff) IsLastLine() bool {
    return d.lineStringsLen == d.Line.LineNum
}

var eofWarning bool

// Stupid, I know. But it helps clear up these types of errors
func EOFErrFilter(err error, log logrus.FieldLogger) error {
    if err == nil || !IsEOF(err) || eofWarning {
        return err
    }

    errors.ErrLog(log, err).Warn("clean up your diff code, this is an uncaught EOF error!")
    eofWarning = true

    return nil
}
