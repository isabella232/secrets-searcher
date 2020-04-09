package diff

import (
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    gitdiff "gopkg.in/src-d/go-git.v4/plumbing/format/diff"
    "strings"
)

const (
    EqualPrefix  = " "
    AddPrefix    = "+"
    DeletePrefix = "-"
)

type (
    Diff struct {
        Line        *Line
        lineStrings []string
        lineMap     map[int]int
    }
    lineMatch func(line *Line) bool
)

func NewFromChunks(chunks []gitdiff.Chunk) (result *Diff, err error) {
    if len(chunks) == 0 {
        err = errors.New("no chunks passed")
    }

    var lineMap map[int]int
    var lineStrings []string
    lineMap, lineStrings, err = buildDiffLineInfo(chunks)
    if err != nil {
        return
    }

    diff := &Diff{
        lineStrings: lineStrings,
        lineMap:     lineMap,
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

func (d *Diff) String() (result string) {
    return strings.Join(d.lineStrings, "\n")
}

func (d *Diff) lineExists(lineNum int) bool {
    return lineNum <= len(d.lineStrings)
}

func (d *Diff) buildLine(lineNumDiff int) (result *Line, ok bool) {
    if ! d.lineExists(lineNumDiff) {
        return
    }

    var lineNumFile = d.fileLineNum(lineNumDiff)

    result = NewLine(d.lineStrings[(lineNumDiff-1)], lineNumDiff, lineNumFile)

    ok = true
    return
}

func (d *Diff) fileLineNum(diffLineNum int) (result int) {
    result, _ = d.lineMap[diffLineNum]
    return
}

func buildDiffLineInfo(chunks []gitdiff.Chunk) (result map[int]int, diffLines []string, err error) {
    fileLineNum := 1
    diffLineNum := 1

    result = map[int]int{}
    for _, chunk := range chunks {
        chunkString := chunk.Content()

        // Remove the trailing line break
        chunkLen := len(chunkString)
        if chunkLen > 0 && chunkString[chunkLen-1:] == "\n" {
            chunkString = chunkString[:chunkLen-1]
        }

        lines := strings.Split(chunkString, "\n")
        prefix := GetPrefix(chunk.Type())

        for _, line := range lines {
            result[diffLineNum] = fileLineNum

            diffLines = append(diffLines, prefix+line)

            // Prepare for next
            diffLineNum += 1
            if chunk.Type() != gitdiff.Delete {
                fileLineNum += 1
            }
        }
    }

    return
}

func GetPrefix(chunkType gitdiff.Operation) (result string) {
    switch chunkType {
    case gitdiff.Equal:
        return EqualPrefix
    case gitdiff.Delete:
        return DeletePrefix
    case gitdiff.Add:
        return AddPrefix
    default:
        panic(errors.Errorv("unknown chunk type", chunkType))
    }
}
