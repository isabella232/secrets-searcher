package diff

import (
    "strings"
)

type (
    Diff struct {
        LineStrings []string
        LineI       int
        Line        *Line
    }
    lineMatch func(line *Line) bool
)

func New(diff string) *Diff {
    d := &Diff{LineStrings: strings.Split(diff, "\n")}
    d.SetLine(0)
    return d
}

func (d *Diff) WhileTrueCollectCode(lineMatch lineMatch, collected *[]string) {
    for searching := lineMatch(d.Line); searching; searching = lineMatch(d.Line) {
        *collected = append(*collected, d.Line.Code)
        d.Increment()
    }
}

func (d *Diff) WhileTrueIncrement(lineMatch lineMatch) {
    for searching := lineMatch(d.Line); searching; searching = lineMatch(d.Line) {
        d.Increment()
    }
}

func (d *Diff) UntilTrueCollectCode(lineMatch lineMatch, collected *[]string) {
    d.WhileTrueCollectCode(func(line *Line) bool { return !lineMatch(line) }, collected)
}

func (d *Diff) UntilTrueIncrement(lineMatch lineMatch) {
    d.WhileTrueIncrement(func(line *Line) bool { return !lineMatch(line) })
}

func (d *Diff) Increment() {
    d.SetLine(d.LineI + 1)
}

func (d *Diff) SetLine(lineI int) {
    d.LineI = lineI
    d.Line = d.BuildLine(d.LineI)
}

func (d *Diff) NextLine() *Line {
    return d.BuildLine(d.LineI + 1)
}

func (d *Diff) BuildLine(lineI int) *Line {
    return NewLine(d.LineStrings[lineI], lineI)
}
