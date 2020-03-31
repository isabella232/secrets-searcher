package structures

type LineRange struct {
    StartIndex int
    EndIndex   int
}

func (r LineRange) GetStringFrom(input string) (result string) {
    return input[r.StartIndex:r.EndIndex]
}
