package structures

type Filter struct {
    included Set
    excluded Set
}

func NewFilter(included, exclude []string) *Filter {
    return &Filter{
        included: NewSet(included),
        excluded: NewSet(exclude),
    }
}

func (i *Filter) IsEnabled() bool {
    return !i.included.IsEmpty()
}

func (i *Filter) IsIncluded(value string) bool {
    return !i.excluded.Contains(value) && (i.included.IsEmpty() || i.included.Contains(value))
}

func (i *Filter) Values() []string {
    return i.included.Values()
}
