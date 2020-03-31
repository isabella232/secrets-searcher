package structures

type FileRange struct {
    StartLineNum int
    StartIndex   int
    EndLineNum   int
    EndIndex     int
}

func (r FileRange) Overlaps(other *FileRange) bool {
    if ! r.DoLinesOverlap(other) {
        return false
    }
    if r.StartLineNum == other.EndLineNum {
        return r.StartIndex >= other.EndIndex
    }
    if other.StartLineNum == r.EndLineNum {
        return other.StartIndex >= r.EndIndex
    }

    return true
}

func (r FileRange) DoLinesOverlap(other *FileRange) bool {
    return r.EndLineNum >= other.StartLineNum && other.EndLineNum >= r.StartLineNum
}
