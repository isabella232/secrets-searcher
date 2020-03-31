package diff

import (
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    "strings"
)

const (
    EqualPrefix  = " "
    AddPrefix    = "+"
    DeletePrefix = "-"
)

var (
    ErrEOL = errors.New("EOL")
)

type (
    Diff struct {
        LineStrings []string
        LineNum     int
        Line        *Line
    }
    lineMatch func(line *Line) bool
)

func New(lineStrings []string) (result *Diff, err error) {
    result = &Diff{LineStrings: lineStrings}
    err = result.SetLine(1)
    return
}

func NewFromString(diff string) (result *Diff, err error) {
    return New(strings.Split(diff, "\n"))
}

func (d *Diff) WhileTrueCollectCode(lineMatch lineMatch, collected *[]string) (err error) {
    for searching := lineMatch(d.Line); searching; searching = lineMatch(d.Line) {
        *collected = append(*collected, d.Line.Code)
        if err = d.Increment(); err != nil {
            return
        }
    }
    return
}

func (d *Diff) WhileTrueIncrement(lineMatch lineMatch) (err error) {
    for searching := lineMatch(d.Line); searching; searching = lineMatch(d.Line) {
        if err = d.Increment(); err != nil {
            return
        }
    }
    return
}

func (d *Diff) UntilTrueCollectCode(lineMatch lineMatch, collected *[]string) (err error) {
    return d.WhileTrueCollectCode(func(line *Line) bool { return !lineMatch(line) }, collected)
}

func (d *Diff) UntilTrueIncrement(lineMatch lineMatch) (err error) {
    return d.WhileTrueIncrement(func(line *Line) bool { return !lineMatch(line) })
}

func (d *Diff) Increment() (err error) {
    return d.SetLine(d.LineNum + 1)
}

func (d *Diff) SetLine(lineNum int) (err error) {
    var line *Line
    line, err = d.BuildLine(lineNum)
    if err != nil {
        return
    }

    d.LineNum = lineNum
    d.Line = line

    return
}

func (d *Diff) NextLine() (result *Line, err error) {
    return d.BuildLine(d.LineNum + 1)
}

func (d *Diff) BuildLine(lineNum int) (result *Line, err error) {
    lineI := lineNum - 1
    if len(d.LineStrings) < lineNum {
        err = ErrEOL
        return
    }

    result = NewLine(d.LineStrings[lineI], lineNum)

    return
}

func (d *Diff) RequireNextLine() (result *Line) {
    var err error
    result, err = d.NextLine()
    if err != nil {
        panic(err)
    }

    return
}

func (d *Diff) RequireIncrement() {
    if err := d.Increment(); err != nil {
        panic(err)
    }
}
