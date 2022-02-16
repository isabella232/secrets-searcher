package manip

import (
	"fmt"
	"sort"
	"sync"
)

type BasicSet struct {
	data map[interface{}]struct{}
	lock *sync.Mutex
}

func NewBasicSet(values []interface{}) (result *BasicSet) {
	result = &BasicSet{
		data: map[interface{}]struct{}{},
		lock: &sync.Mutex{},
	}
	result.AddSliceValues(values)

	return
}

func StringSet(values []string) (result *BasicSet) {
	result = NewBasicSet(nil)
	for _, value := range values {
		result.Add(value)
	}

	return
}

func NewEmptyBasicSet() (result *BasicSet) {
	return NewBasicSet(nil)
}

func (s *BasicSet) Add(value interface{}) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.data[value] = struct{}{}
}

func (s *BasicSet) AddSliceValues(values []interface{}) {
	for _, value := range values {
		s.Add(value)
	}
}

func (s *BasicSet) Remove(value interface{}) {
	s.lock.Lock()
	defer s.lock.Unlock()

	delete(s.data, value)
}

func (s *BasicSet) Contains(value interface{}) (result bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	_, result = s.data[value]
	return
}

func (s *BasicSet) IsEmpty() (result bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	return len(s.data) == 0
}

func (s *BasicSet) Values() (result []interface{}) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if len(s.data) == 0 {
		return nil
	}

	result = make([]interface{}, len(s.data))
	i := 0
	for key := range s.data {
		result[i] = key
		i += 1
	}

	return
}

func (s *BasicSet) StringValues() (result []string) {
	if len(s.data) == 0 {
		return nil
	}

	result = make([]string, len(s.data))
	i := 0
	for key := range s.data {
		result[i] = fmt.Sprintf("%v", key)
		i += 1
	}

	sort.Strings(result)

	return
}

func (s *BasicSet) Len() (result int) {
	s.lock.Lock()
	defer s.lock.Unlock()

	result = len(s.data)
	return
}
