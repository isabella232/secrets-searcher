package structures

import "sync"

type Set struct {
    data map[string]struct{}
    lock *sync.Mutex
}

func NewSet(values []string) (result Set) {
    var data = map[string]struct{}{}
    for _, value := range values {
        data[value] = struct{}{}
    }

    result = Set{
        data: data,
        lock: &sync.Mutex{},
    }

    return
}

func (s Set) Add(value string) {
    s.lock.Lock()
    defer s.lock.Unlock()

    s.data[value] = struct{}{}
}

func (s Set) Contains(value string) (result bool) {
    s.lock.Lock()
    defer s.lock.Unlock()

    _, result = s.data[value]
    return
}

func (s Set) IsEmpty() (result bool) {
    s.lock.Lock()
    defer s.lock.Unlock()

    return len(s.data) == 0
}

func (s Set) Values() (result []string) {
    s.lock.Lock()
    defer s.lock.Unlock()

    for key := range s.data {
        result = append(result, key)
    }
    return
}

func (s Set) Len() (result int) {
    s.lock.Lock()
    defer s.lock.Unlock()

    result = len(s.data)
    return
}

func (s Set) Equals(other Set) (result bool) {
    return SlicesAreEqual(s.Values(), other.Values())
}

func (s Set) EqualsAfterBothSorted(other Set) (result bool) {
    return s.Len() == other.Len() && SlicesAreEqualAfterSort(s.Values(), other.Values())
}
