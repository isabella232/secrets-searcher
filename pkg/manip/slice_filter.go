package manip

type SliceFilter struct {
	include Set
	exclude Set
}

func NewSliceFilter(include, exclude Set) *SliceFilter {
	if include == nil {
		include = NewEmptyBasicSet()
	}
	if exclude == nil {
		exclude = NewEmptyBasicSet()
	}
	return &SliceFilter{
		include: include,
		exclude: exclude,
	}
}

// Create from string slices
func StringFilter(include, exclude []string) *SliceFilter {
	return &SliceFilter{
		include: StringSet(include),
		exclude: StringSet(exclude),
	}
}

func IncludeFilter(include []string) *SliceFilter {
	return NewSliceFilter(StringSet(include), nil)
}

func ExcludeFilter(exclude []string) *SliceFilter {
	return NewSliceFilter(nil, StringSet(exclude))
}

// Returns true if any string is included
func (i *SliceFilter) IncludesAnything() bool {
	return i.include.IsEmpty() && i.exclude.IsEmpty()
}

// Returns true if an item is included
func (i *SliceFilter) Includes(value interface{}) bool {
	return !i.exclude.Contains(value) &&
		(i.include.IsEmpty() || i.include.Contains(value))
}

// Returns only included items in a BasicSet object
func (i *SliceFilter) FilterSet(items Set) {
	for _, item := range items.Values() {
		if !i.Includes(item) {
			items.Remove(item)
		}
	}
	return
}

// Returns only included items in a BasicSet object
func (i *SliceFilter) IncludesAllOf(items Set) bool {
	for _, item := range items.Values() {
		if !i.Includes(item) {
			return false
		}
	}
	return true
}

func (i *SliceFilter) IncludesAnyOf(items Set) bool {
	for _, item := range items.Values() {
		if i.Includes(item) {
			return true
		}
	}

	return false
}

// Returns true if an exact list of values that are included is available
func (i *SliceFilter) CanProvideExactValues() (result bool) {
	return !i.include.IsEmpty()
}

// Exact list of values that are included
func (i *SliceFilter) ExactValues() (result Set) {
	if !i.CanProvideExactValues() {
		panic("can't provide exact items, use CanProvideExactValues() to check first")
	}
	result = NewEmptyBasicSet()
	values := i.include.Values()
	for _, item := range values {
		if !i.exclude.Contains(item) {
			result.Add(item)
		}
	}
	return
}
