package structures_test

import (
    . "github.com/pantheon-systems/search-secrets/pkg/structures"
    "github.com/stretchr/testify/require"
    "testing"
)

var (
    subject = &LineRange{StartIndex: 2, EndIndex: 10}
)

func TestLineRangeOverlap_NoOverlap(t *testing.T) {
    other := &LineRange{StartIndex: 11, EndIndex: 12}

    // Fire
    response := subject.Overlaps(other)

    require.False(t, response)
}

func TestLineRangeOverlap_OtherWithinIndexes(t *testing.T) {
    other := &LineRange{StartIndex: 3, EndIndex: 9}

    // Fire
    response := subject.Overlaps(other)

    require.True(t, response)
}

func TestLineRangeOverlap_OtherOverlapsBothIndexes(t *testing.T) {
    other := &LineRange{StartIndex: 1, EndIndex: 11}

    // Fire
    response := subject.Overlaps(other)

    require.True(t, response)
}

func TestLineRangeOverlap_OtherOverlapStartIndex(t *testing.T) {
    other := &LineRange{StartIndex: 1, EndIndex: 3}

    // Fire
    response := subject.Overlaps(other)

    require.True(t, response)
}

func TestLineRangeOverlap_OtherOverlapEndIndex(t *testing.T) {
    other := &LineRange{StartIndex: 9, EndIndex: 11}

    // Fire
    response := subject.Overlaps(other)

    require.True(t, response)
}

func TestLineRangeOverlap_OtherSameIndex(t *testing.T) {
    other := &LineRange{StartIndex: 2, EndIndex: 10}

    // Fire
    response := subject.Overlaps(other)

    require.True(t, response)
}

func TestLineRangeOverlap_OtherWithinContextEqualIndex(t *testing.T) {
    other := &LineRange{StartIndex: 3, EndIndex: 3}

    // Fire
    response := subject.Overlaps(other)

    require.False(t, response)
}

func TestNewLineRangeFromFileRange_Happy(t *testing.T) {
    content := "123\n456\n789\n"
    expected := "123\n45"
    fileRange := NewFileRange(1, 0, 2, 2)

    // Fire
    response := NewLineRangeFromFileRange(fileRange, content)

    require.Equal(t, expected, response.ExtractValue(content).Value)
}

func TestNewLineRangeFromFileRange_SecondLine(t *testing.T) {
    content := "123\n456\n789\n"
    expected := "456\n78"
    fileRange := NewFileRange(2, 0, 3, 2)

    // Fire
    response := NewLineRangeFromFileRange(fileRange, content)

    require.Equal(t, expected, response.ExtractValue(content).Value)
}

func TestNewLineRangeFromFileRange_LastCharIsZeroIndex(t *testing.T) {
    content := "123\n456\n789\n"
    expected := "123\n456\n"
    fileRange := NewFileRange(1, 0, 3, 0)

    // Fire
    response := NewLineRangeFromFileRange(fileRange, content)

    require.Equal(t, expected, response.ExtractValue(content).Value)
}

func TestNewLineRangeFromFileRange_FirstLineIsEmpty(t *testing.T) {
    content := "123\n456\n789\n"
    expected := "\n456\n"
    fileRange := NewFileRange(1, 3, 3, 0)

    // Fire
    response := NewLineRangeFromFileRange(fileRange, content)

    require.Equal(t, expected, response.ExtractValue(content).Value)
}

func TestNewLineRangeFromFileRange_Entire(t *testing.T) {
    content := "123\n456\n789\n"
    expected := "2"
    fileRange := NewFileRange(1, 1, 1, 2)

    // Fire
    response := NewLineRangeFromFileRange(fileRange, content)

    require.Equal(t, expected, response.ExtractValue(content).Value)
}

func TestNewLineRangeFromFileRange_OneWholeLine(t *testing.T) {
    content := "123\n456\n789\n"
    expected := "123"
    fileRange := NewFileRange(1, 0, 1, 3)

    // Fire
    response := NewLineRangeFromFileRange(fileRange, content)

    require.Equal(t, expected, response.ExtractValue(content).Value)
}

func TestNewLineRangeFromFileRange_OnePartialLine(t *testing.T) {
    content := "123\n456\n789\n"
    expected := "\n456\n"
    fileRange := NewFileRange(1, 3, 3, 0)

    // Fire
    response := NewLineRangeFromFileRange(fileRange, content)

    require.Equal(t, expected, response.ExtractValue(content).Value)
}
