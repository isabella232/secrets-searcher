package diff

import (
    "fmt"
    "regexp"
    "strings"
)

type (
    Line struct {
        fmt.Stringer
        LineNumDiff int
        LineNumFile int
        Line        string
        Pre         string
        Code        string
        IsEqu       bool
        IsAdd       bool
        IsDel       bool
    }
)

func NewLine(lineString string, lineNumDiff, lineNumFile int) *Line {
    if lineString == "" {
        return &Line{}
    }

    pre := lineString[:1]
    code := lineString[1:]
    isEqu := pre == EqualPrefix
    isAdd := pre == AddPrefix
    isDel := pre == DeletePrefix

    return &Line{
        LineNumDiff: lineNumDiff,
        LineNumFile: lineNumFile,
        Line:        lineString,
        Pre:         pre,
        Code:        code,
        IsEqu:       isEqu,
        IsAdd:       isAdd,
        IsDel:       isDel,
    }
}

func (l *Line) String() string {
    return l.Line
}

func (l *Line) CodeContains(substr string) bool {
    return strings.Contains(l.Code, substr)
}

func (l *Line) CodeStartsWith(substr string) bool {
    return strings.HasPrefix(l.Code, substr)
}

func (l *Line) CodeEndsWith(substr string) bool {
    return strings.HasSuffix(l.Code, substr)
}

func (l *Line) CodeMatches(regex *regexp.Regexp) bool {
    return regex.MatchString(l.Code)
}
