package manip

import (
	"fmt"
	"strings"
)

type (
	LineRange struct {
		StartIndex int
		EndIndex   int
	}
	LineRangeValue struct {
		LineRange *LineRange
		Value     string
	}
)

func NewLineRange(startIndex, endIndex int) (result *LineRange) {
	if endIndex < startIndex {
		panic(fmt.Sprintf("end index must be equal or greater than start index (%d >= %d)", startIndex, endIndex))
	}
	return &LineRange{StartIndex: startIndex, EndIndex: endIndex}
}

func FindLineRange(val, sub string) (result *LineRange) {
	index := strings.Index(val, sub)
	if index == -1 {
		return
	}
	return NewLineRange(index, index+len(sub))
}

func NewLineRangeFromFileRange(fileRange *FileRange, content string) (result *LineRange) {

	// FIXME: This part only makes sense if the line number is 1. Reassess this, it's wrong
	if fileRange.StartLineNum == 1 && fileRange.EndLineNum == 1 {
		return NewLineRange(fileRange.StartIndex, fileRange.EndIndex)
	}

	var lineRangeStartIndex int
	var lineRangeEndIndex int

	lines := strings.Split(content, "\n")

	beforeLines := lines[:fileRange.StartLineNum-1]
	beforeLinesLen := len(strings.Join(beforeLines, "")) + len(beforeLines)

	startLine := lines[fileRange.StartLineNum-1]
	startLineLen := len(startLine) + 1

	if fileRange.StartLineNum == fileRange.EndLineNum {
		lineRangeStartIndex = beforeLinesLen + fileRange.StartIndex
		lineRangeEndIndex = beforeLinesLen + fileRange.EndIndex

		return NewLineRange(lineRangeStartIndex, lineRangeEndIndex)
	}

	middleLines := lines[fileRange.StartLineNum : fileRange.EndLineNum-1]
	middleLinesLen := len(strings.Join(middleLines, "")) + len(middleLines)

	lineRangeStartIndex = beforeLinesLen + fileRange.StartIndex
	lineRangeEndIndex = beforeLinesLen + startLineLen + middleLinesLen + fileRange.EndIndex

	return NewLineRange(lineRangeStartIndex, lineRangeEndIndex)
}

func (r *LineRange) Shifted(by int) (result *LineRange) {
	return NewLineRange(r.StartIndex+by, r.EndIndex+by)
}

func (r *LineRange) NewValue(valueString string) (result *LineRangeValue) {
	return &LineRangeValue{LineRange: r, Value: valueString}
}

func (r *LineRange) ExtractValue(input string) (result *LineRangeValue) {
	return r.NewValue(r.extractString(input))
}

func (r *LineRange) Equals(other *LineRange) bool {
	return r.StartIndex == other.StartIndex && other.EndIndex == r.EndIndex
}

func (r *LineRange) Len() int {
	return r.EndIndex - r.StartIndex
}

func (r *LineRange) HasContent() bool {
	return r.StartIndex < r.EndIndex
}

func (r *LineRange) Overlaps(other *LineRange) bool {
	return r.HasContent() && other.HasContent() && r.EndIndex >= other.StartIndex && other.EndIndex >= r.StartIndex
}

func (r *LineRange) extractString(input string) (result string) {
	return input[r.StartIndex:r.EndIndex]
}
