package manip_test

import (
	"testing"

	. "github.com/pantheon-systems/secrets-searcher/pkg/manip"
	"github.com/stretchr/testify/require"
)

var (
	subject0          = &FileRange{StartLineNum: 2, EndLineNum: 10, StartIndex: 2, EndIndex: 10}
	subjectSingleLine = &FileRange{StartLineNum: 2, EndLineNum: 2, StartIndex: 2, EndIndex: 10}
)

func TestFileRangeOverlap_NoOverlap(t *testing.T) {
	other := &FileRange{StartLineNum: 11, EndLineNum: 12}

	// Fire
	response := subject0.Overlaps(other)

	require.False(t, response)
}

func TestFileRangeOverlap_OtherWithinLines(t *testing.T) {
	other := &FileRange{StartLineNum: 3, EndLineNum: 9}

	// Fire
	response := subject0.Overlaps(other)

	require.True(t, response)
}

func TestFileRangeOverlap_OtherOverlapsBothLines(t *testing.T) {
	other := &FileRange{StartLineNum: 1, EndLineNum: 11}

	// Fire
	response := subject0.Overlaps(other)

	require.True(t, response)
}

func TestFileRangeOverlap_OtherOverlapsStartLine(t *testing.T) {
	other := &FileRange{StartLineNum: 1, EndLineNum: 3}

	// Fire
	response := subject0.Overlaps(other)

	require.True(t, response)
}

func TestFileRangeOverlap_OtherOverlapsEndLine(t *testing.T) {
	other := &FileRange{StartLineNum: 9, EndLineNum: 11}

	// Fire
	response := subject0.Overlaps(other)

	require.True(t, response)
}

func TestFileRangeOverlap_OtherLinesStartAtEnd_OtherIndexBefore(t *testing.T) {
	other := &FileRange{StartLineNum: 10, EndLineNum: 11, StartIndex: 9}

	// Fire
	response := subject0.Overlaps(other)

	require.True(t, response)
}

func TestFileRangeOverlap_OtherLinesStartAtEnd_OtherIndexSame(t *testing.T) {
	other := &FileRange{StartLineNum: 10, EndLineNum: 11, StartIndex: 10}

	// Fire
	response := subject0.Overlaps(other)

	require.True(t, response)
}

func TestFileRangeOverlap_OtherLinesStartAtEnd_OtherIndexAfter(t *testing.T) {
	other := &FileRange{StartLineNum: 10, EndLineNum: 11, StartIndex: 11}

	// Fire
	response := subject0.Overlaps(other)

	require.False(t, response)
}

func TestFileRangeOverlap_OtherLinesEndAtStart_OtherIndexBefore(t *testing.T) {
	other := &FileRange{StartLineNum: 1, EndLineNum: 2, EndIndex: 1}

	// Fire
	response := subject0.Overlaps(other)

	require.False(t, response)
}

func TestFileRangeOverlap_OtherLinesEndAtStart_OtherIndexSame(t *testing.T) {
	other := &FileRange{StartLineNum: 1, EndLineNum: 2, EndIndex: 2}

	// Fire
	response := subject0.Overlaps(other)

	require.True(t, response)
}

func TestFileRangeOverlap_OtherLinesEndAtStart_OtherIndexAfter(t *testing.T) {
	other := &FileRange{StartLineNum: 1, EndLineNum: 2, EndIndex: 3}

	// Fire
	response := subject0.Overlaps(other)

	require.True(t, response)
}

func TestFileRangeOverlap_OtherLinesSame(t *testing.T) {
	other := &FileRange{StartLineNum: 2, EndLineNum: 10, StartIndex: 2, EndIndex: 10}

	// Fire
	response := subject0.Overlaps(other)

	require.True(t, response)
}

func TestFileRangeOverlap_SameSingleLine_NoOverlap(t *testing.T) {
	other := &FileRange{StartLineNum: 2, EndLineNum: 2, StartIndex: 11, EndIndex: 12}

	// Fire
	response := subjectSingleLine.Overlaps(other)

	require.False(t, response)
}

func TestFileRangeOverlap_SameSingleLine_OtherWithinIndexes(t *testing.T) {
	other := &FileRange{StartLineNum: 2, EndLineNum: 2, StartIndex: 3, EndIndex: 9}

	// Fire
	response := subjectSingleLine.Overlaps(other)

	require.True(t, response)
}

func TestFileRangeOverlap_SameSingleLine_OtherOverlapsBothIndexes(t *testing.T) {
	other := &FileRange{StartLineNum: 2, EndLineNum: 2, StartIndex: 1, EndIndex: 11}

	// Fire
	response := subjectSingleLine.Overlaps(other)

	require.True(t, response)
}

func TestFileRangeOverlap_SameSingleLine_OtherOverlapStartIndex(t *testing.T) {
	other := &FileRange{StartLineNum: 2, EndLineNum: 2, StartIndex: 1, EndIndex: 3}

	// Fire
	response := subjectSingleLine.Overlaps(other)

	require.True(t, response)
}

func TestFileRangeOverlap_SameSingleLine_OtherOverlapEndIndex(t *testing.T) {
	other := &FileRange{StartLineNum: 2, EndLineNum: 2, StartIndex: 9, EndIndex: 11}

	// Fire
	response := subjectSingleLine.Overlaps(other)

	require.True(t, response)
}

func TestFileRangeOverlap_SameSingleLine_OtherSameIndex(t *testing.T) {
	other := &FileRange{StartLineNum: 2, EndLineNum: 2, StartIndex: 2, EndIndex: 10}

	// Fire
	response := subjectSingleLine.Overlaps(other)

	require.True(t, response)
}

func TestFileRangeOverlap_SameSingleLine_OtherWithinContextEqualIndex(t *testing.T) {
	other := &FileRange{StartLineNum: 2, EndLineNum: 2, StartIndex: 3, EndIndex: 3}

	// Fire
	response := subjectSingleLine.Overlaps(other)

	require.False(t, response)
}
