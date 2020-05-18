package git

import (
	"regexp"
	"strings"

	"github.com/pantheon-systems/search-secrets/pkg/dev"
	"github.com/pantheon-systems/search-secrets/pkg/errors"
)

var (
	emptyLineAction = func(line *Line) {}
)

type (
	Diff struct {
		Line                 *Line
		lineStrings          []string
		lines                []*Line
		lineStringsLen       int
		diffToFileLineNumMap map[int]int // FIXME This shouldn't need to be passed in
		SetLineHook          lineAction
	}
	lineMatch  func(line *Line) bool
	lineAction func(line *Line)
)

func NewDiff(lineStrings []string, lineMap map[int]int) (result *Diff, err error) {
	lineStringsLen := len(lineStrings)

	diff := &Diff{
		lineStrings:          lineStrings,
		lines:                make([]*Line, lineStringsLen),
		lineStringsLen:       lineStringsLen,
		diffToFileLineNumMap: lineMap,
	}

	if ok := diff.SetLine(1); !ok {
		err = errors.New("file is empty")
		return
	}

	result = diff

	return
}

func (d *Diff) SetLine(lineNum int) (ok bool) {

	// Set lines
	d.Line = d.getLineObject(lineNum)

	// Call breakpoint
	if d.Line != nil {
		lineNumFile, _ := d.fileLineNum(d.Line.Num)
		dev.BreakOnDiffLine(lineNumFile)
		if d.SetLineHook != nil {
			d.SetLineHook(d.Line)
		}
	}

	return d.Line != nil
}

func (d *Diff) String() (result string) {
	return strings.Join(d.lineStrings, "\n")
}

// Navigation

func (d *Diff) Incr() (ok bool) {
	return d.IncrBy(1)
}

func (d *Diff) IncrBy(by int) (ok bool) {
	return d.Line != nil && d.SetLine(d.Line.Num+by)
}

func (d *Diff) WhileTrueDo(lineMatch lineMatch, lineAction lineAction) (ok bool) {
	ok = true
	for searching := lineMatch(d.Line); searching; searching = lineMatch(d.Line) {
		lineAction(d.Line)
		if !d.Incr() {
			ok = false
			break
		}
	}
	return
}

func (d *Diff) WhileTrueIncrement(lineMatch lineMatch) (ok bool) {
	return d.WhileTrueDo(lineMatch, emptyLineAction)
}
func (d *Diff) UntilTrueIncrement(lineMatch lineMatch) (ok bool) {
	return d.WhileTrueIncrement(reverseMatch(lineMatch))
}

func (d *Diff) WhileTrueCollectCode(lineMatch lineMatch, collected *[]string) (ok bool) {
	return d.WhileTrueDo(lineMatch, collectCode(collected))
}

func (d *Diff) UntilTrueCollectCode(lineMatch lineMatch, collected *[]string) (ok bool) {
	return d.WhileTrueCollectCode(reverseMatch(lineMatch), collected)
}

func (d *Diff) WhileContainsCollectCode(substring string, collected *[]string) (ok bool) {
	return d.WhileTrueDo(lineContains(substring), collectCode(collected))
}

func (d *Diff) UntilContainsCollectCode(substring string, collected *[]string) (ok bool) {
	return d.WhileTrueDo(reverseMatch(lineContains(substring)), collectCode(collected))
}

func (d *Diff) WhileTrueCollectTrimmedCode(lineMatch lineMatch, collected *[]string) (ok bool) {
	return d.WhileTrueDo(lineMatch, collectTrimmedCode(collected))
}

func (d *Diff) UntilTrueCollectTrimmedCode(lineMatch lineMatch, collected *[]string) (ok bool) {
	return d.WhileTrueCollectTrimmedCode(reverseMatch(lineMatch), collected)
}

func (d *Diff) WhileContainsCollectTrimmedCode(substring string, collected *[]string) (ok bool) {
	return d.WhileTrueDo(lineContains(substring), collectTrimmedCode(collected))
}

func (d *Diff) UntilContainsCollectTrimmedCode(substring string, collected *[]string) (ok bool) {
	return d.WhileTrueDo(reverseMatch(lineContains(substring)), collectTrimmedCode(collected))
}

func (d *Diff) WhileMatchesIncrement(re *regexp.Regexp) (ok bool) {
	return d.WhileTrueDo(lineMatchRe(re), emptyLineAction)
}

func (d *Diff) UntilMatchesIncrement(re *regexp.Regexp) (ok bool) {
	return d.WhileTrueDo(reverseMatch(lineMatchRe(re)), emptyLineAction)
}

// Query

func (d *Diff) OnLastLine() bool {
	return d.IsLastLine(d.Line.Num)
}

func (d *Diff) IsLastLine(lineNum int) bool {
	return d.lineStringsLen == lineNum
}

// Internal

func (d *Diff) getLineObject(lineNum int) (result *Line) {
	var ok = true
	var lineIndex = lineNum - 1

	if lineNum < 1 || lineNum > d.lineStringsLen {
		return
	}

	// Return cache
	result = d.lines[lineIndex]
	if result != nil {
		return
	}

	// Get line string
	lineString := d.lineStrings[lineIndex]

	// Get mapped line number
	var lineNumFile int
	if lineNumFile, ok = d.fileLineNum(lineNum); !ok {
		lineNumFile = -1
	}

	// Build line
	result = NewLine(lineString, lineNum, lineNumFile)

	// Write cache
	d.lines[lineIndex] = result

	return
}

func (d *Diff) fileLineNum(lineNum int) (result int, ok bool) {
	result, ok = d.diffToFileLineNumMap[lineNum]
	return
}

// Callbacks

func reverseMatch(lineMatch lineMatch) (result lineMatch) {
	return func(line *Line) bool { return !lineMatch(line) }
}

func lineContains(substring string) (result lineMatch) {
	return func(line *Line) bool { return line.Contains(substring) }
}

func lineMatchRe(re *regexp.Regexp) (result lineMatch) {
	return func(line *Line) bool { return line.Matches(re) }
}

func collectCode(collected *[]string) (result lineAction) {
	return func(line *Line) { *collected = append(*collected, line.Code) }
}

func collectTrimmedCode(collected *[]string) (result lineAction) {
	return func(line *Line) { *collected = append(*collected, line.Trim()) }
}
