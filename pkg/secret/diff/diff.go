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

func (d *Diff) CollectCodeWhile(lineMatch lineMatch, collected *[]string) {
	for searching := true; searching; searching = lineMatch(d.Line) {
		*collected = append(*collected, d.Line.Code)
		d.Increment()
	}
}

func (d *Diff) IncrementWhile(lineMatch lineMatch) {
	for searching := true; searching; searching = lineMatch(d.Line) {
		d.Increment()
	}
}

func (d *Diff) CollectCodeUntil(lineMatch lineMatch, collected *[]string) {
	d.CollectCodeWhile(func(line *Line) bool { return !lineMatch(line) }, collected)
}

func (d *Diff) IncrementUntil(lineMatch lineMatch) {
	d.IncrementWhile(func(line *Line) bool { return !lineMatch(line) })
}

func (d *Diff) Increment() {
	d.SetLine(d.LineI + 1)
}

func (d *Diff) SetLine(lineI int) {
	d.LineI = lineI
	d.Line = NewLine(d.LineStrings[d.LineI])
}

func (d *Diff) NextLine() *Line {
	return NewLine(d.LineStrings[d.LineI+1])
}
