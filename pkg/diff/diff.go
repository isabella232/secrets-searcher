package diff

import (
    "fmt"
    "github.com/pantheon-systems/search-secrets/pkg/dev"
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    "strings"
)

const (
    equalPrefix  = " "
    addPrefix    = "+"
    deletePrefix = "-"
)

type (
    Diff struct {
        Line                 *Line
        lineStrings          []string
        diffToFileLineNumMap map[int]int // FIXME This shouldn't need to be passed in
    }
    lineMatch func(line *Line) bool
)

func New(lineStrings []string, lineMap map[int]int) (result *Diff, err error) {
    diff := &Diff{
        lineStrings:          lineStrings,
        diffToFileLineNumMap: lineMap,
    }

    if ok := diff.SetLine(1); !ok {
        err = errors.Errorv("unable to set line to 1")
        return
    }

    result = diff

    return
}

func (d *Diff) WhileTrueCollectCode(lineMatch lineMatch, collected *[]string) (ok bool) {
    for searching := lineMatch(d.Line); searching; searching = lineMatch(d.Line) {
        *collected = append(*collected, d.Line.Code)
        if ok = d.Increment(); !ok {
            return
        }
    }
    return true
}

func (d *Diff) WhileTrueIncrement(lineMatch lineMatch) (ok bool) {
    for searching := lineMatch(d.Line); searching; searching = lineMatch(d.Line) {
        if ok = d.Increment(); !ok {
            return
        }
    }
    return true
}

func (d *Diff) UntilTrueCollectCode(lineMatch lineMatch, collected *[]string) (ok bool) {
    return d.WhileTrueCollectCode(func(line *Line) bool { return !lineMatch(line) }, collected)
}

func (d *Diff) UntilTrueIncrement(lineMatch lineMatch) (ok bool) {
    return d.WhileTrueIncrement(func(line *Line) bool { return !lineMatch(line) })
}

func (d *Diff) Increment() (ok bool) {
    return d.SetLine(d.Line.LineNumDiff + 1)
}

func (d *Diff) SetLine(lineNumDiff int) (ok bool) {
    if dev.Enabled && dev.DiffLine > 0 && lineNumDiff == dev.DiffLine {
        fmt.Print("")
    }

    var line *Line
    line, ok = d.buildLine(lineNumDiff)
    if !ok {
        return
    }

    d.Line = line

    return
}

func (d *Diff) PeekNextLine() (result *Line, ok bool) {
    return d.buildLine(d.Line.LineNumDiff + 1)
}

func (d *Diff) Lines() (result []string) {
    return d.lineStrings
}

func (d *Diff) String() (result string) {
    return strings.Join(d.lineStrings, "\n")
}

func (d *Diff) lineExists(lineNum int) bool {
    return lineNum <= len(d.lineStrings)
}

func (d *Diff) buildLine(lineNumDiff int) (result *Line, ok bool) {
    if !d.lineExists(lineNumDiff) {
        return
    }

    var lineNumFile = d.fileLineNum(lineNumDiff)

    result = NewLine(d.lineStrings[(lineNumDiff-1)], lineNumDiff, lineNumFile)

    ok = true
    return
}

func (d *Diff) fileLineNum(diffLineNum int) (result int) {
    result, _ = d.diffToFileLineNumMap[diffLineNum]
    return
}
