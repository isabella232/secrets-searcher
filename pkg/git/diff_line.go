package git

import (
	"fmt"
	"regexp"
	"strings"
)

type (
	Line struct {
		Num         int
		NumInFile   int
		Pre         string
		Code        string
		CodeTrimmed string
		IsEqu       bool
		IsAdd       bool
		IsDel       bool
	}
)

func NewLine(lineString string, num, numInFile int) *Line {
	if lineString == "" {
		return &Line{}
	}

	pre := lineString[:1]
	code := lineString[1:]
	isEqu := pre == Equal.Prefix()
	isAdd := pre == Add.Prefix()
	isDel := pre == Delete.Prefix()

	return &Line{
		Num:         num,
		NumInFile:   numInFile,
		Pre:         pre,
		Code:        code,
		CodeTrimmed: strings.Trim(code, " "),
		IsEqu:       isEqu,
		IsAdd:       isAdd,
		IsDel:       isDel,
	}
}

func (l *Line) String() string {
	return l.Pre + l.Code
}

func (l *Line) Trim() string {
	if l.CodeTrimmed == "" && l.Code != "" {
		l.CodeTrimmed = fmt.Sprintf("%q", l.Code)
	}
	return l.CodeTrimmed
}

func (l *Line) Contains(substr string) bool {
	return strings.Contains(l.Code, substr)
}

func (l *Line) StartsWith(substr string) bool {
	return strings.HasPrefix(l.Code, substr)
}

func (l *Line) EndsWith(substr string) bool {
	return strings.HasSuffix(l.Code, substr)
}

func (l *Line) Matches(regex *regexp.Regexp) bool {
	return regex.MatchString(l.Code)
}
