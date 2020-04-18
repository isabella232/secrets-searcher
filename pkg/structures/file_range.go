package structures

type FileRange struct {
    StartLineNum int
    StartIndex   int
    EndLineNum   int
    EndIndex     int
}

func NewFileRangeFromLineRange(lineRange *LineRange, lineNum int) (result *FileRange) {
    return &FileRange{
        StartLineNum: lineNum,
        StartIndex:   lineRange.StartIndex,
        EndLineNum:   lineNum,
        EndIndex:     lineRange.EndIndex,
    }
}

func (r FileRange) Overlaps(other *FileRange) bool {
    if !r.DoLinesOverlap(other) {
        return false
    }
    // Same line, single line
    if r.StartLineNum == r.EndLineNum && other.StartLineNum == other.EndLineNum &&
        r.StartLineNum == other.StartLineNum {
        lineRange := NewLineRange(r.StartIndex, r.EndIndex)
        otherLineRange := NewLineRange(other.StartIndex, other.EndIndex)
        return lineRange.Overlaps(otherLineRange)
    }
    // Other starts at end
    if other.StartLineNum == r.EndLineNum {
        return other.StartIndex <= r.EndIndex
    }
    // Other ends at start
    if other.EndLineNum == r.StartLineNum {
        return other.EndIndex >= r.StartIndex
    }

    return true
}

func (r FileRange) DoLinesOverlap(other *FileRange) bool {
    return r.EndLineNum >= other.StartLineNum && other.EndLineNum >= r.StartLineNum
}
