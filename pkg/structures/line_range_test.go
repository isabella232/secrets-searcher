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
