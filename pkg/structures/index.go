package structures

type Index struct {
	Values  []string
	Enabled bool
	set     map[string]struct{}
}

func NewIndex(values []string) *Index {
	return &Index{
		Values:  values,
		Enabled: len(values) > 0,
		set:     sliceToIndex(values),
	}
}

func (i *Index) Include(value string) bool {
	if ! i.Enabled {
		return true
	}

	_, ok := i.set[value]

	return ok
}

func sliceToIndex(values []string) map[string]struct{} {
	result := map[string]struct{}{}
	for _, value := range values {
		result[value] = struct{}{}
	}

	return result
}
