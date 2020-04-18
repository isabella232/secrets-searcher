package structures

type Filter struct {
    Included Set
    Excluded Set
}

func NewFilter(included, exclude []string) *Filter {
    return &Filter{
        Included: NewSet(included),
        Excluded: NewSet(exclude),
    }
}

func (i *Filter) IsEnabled() bool {
    return !i.Included.IsEmpty() || !i.Excluded.IsEmpty()
}

func (i *Filter) IsIncluded(value string) bool {
    return !i.Excluded.Contains(value) && (i.Included.IsEmpty() || i.Included.Contains(value))
}

func (i *Filter) AnyExcluded(items []string) bool {
    for _, item := range items {
        if i.Excluded.Contains(item) {
            return false
        }
    }
    return true
}

func (i *Filter) AnyExcludedSet(items Set) bool {
    return i.AnyExcluded(items.Values())
}

func (i *Filter) Values() []string {
    return i.Included.Values()
}

func (i *Filter) FilteredSet(items Set) (result Set) {
    result = NewSet(nil)
    for _, item := range items.Values() {
        if i.IsIncluded(item) {
            result.Add(item)
        }
    }
    return
}

func (i *Filter) CanProvideExactValues() (result bool) {
    return !i.Included.IsEmpty()
}

func (i *Filter) ExactValues() (result Set) {
    if !i.CanProvideExactValues() {
        panic("can't provide exact items, use CanProvideExactValues() to check first")
    }
    result = NewSet(nil)
    for _, item := range i.Included.Values() {
        if !i.Excluded.Contains(item) {
            result.Add(item)
        }
    }
    return
}
