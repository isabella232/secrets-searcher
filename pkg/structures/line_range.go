package structures

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
    return &LineRange{StartIndex: startIndex, EndIndex: endIndex}
}

func (r LineRange) NewValue(valueString string) (result *LineRangeValue) {
    return &LineRangeValue{LineRange: &r, Value: valueString}
}

func (r LineRange) ExtractValue(input string) (result *LineRangeValue) {
    return r.NewValue(r.extractString(input))
}

func (r LineRange) Equals(other *LineRange) bool {
    return r.StartIndex == other.StartIndex && other.EndIndex == r.EndIndex
}

func (r LineRange) HasContent() bool {
    return r.StartIndex < r.EndIndex
}

func (r LineRange) Overlaps(other *LineRange) bool {
    return r.HasContent() && other.HasContent() && r.EndIndex >= other.StartIndex && other.EndIndex >= r.StartIndex
}

func (r LineRange) extractString(input string) (result string) {
    return input[r.StartIndex:r.EndIndex]
}
