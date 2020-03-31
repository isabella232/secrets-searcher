package structures

type Set map[string]struct{}

func NewSet(values []string) (result Set) {
    result = Set{}
    for _, value := range values {
        result[value] = struct{}{}
    }

    return result
}

func (s Set) Add(value string) {
    s[value] = struct{}{}
}

func (s Set) Contains(value string) (result bool) {
    _, result = s[value]
    return
}

func (s Set) IsEmpty() (result bool) {
    return len(s) == 0
}

func (s Set) Values() (result []string) {
    for key := range s {
        result = append(result, key)
    }
    return
}
