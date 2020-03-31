package structures

type Filter struct {
    set Set
}

func NewFilter(values []string) *Filter {
    return &Filter{
        set: NewSet(values),
    }
}

func (i *Filter) IsEnabled() bool {
    return ! i.set.IsEmpty()
}

func (i *Filter) IsIncluded(value string) bool {
    return i.set.IsEmpty() || i.set.Contains(value)
}

func (i *Filter) Values() []string {
    return i.set.Values()
}
