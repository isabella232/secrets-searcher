package manip

type Filter interface {
	IncludesAnything() bool
	Includes(interface{}) bool
	IncludesAllOf(items Set) bool
	IncludesAnyOf(items Set) bool
	FilterSet(Set)
	CanProvideExactValues() bool
	ExactValues() Set
}

type Set interface {
	Add(interface{})
	AddSliceValues([]interface{})
	Remove(item interface{})
	Contains(interface{}) bool
	IsEmpty() bool
	Values() []interface{}
	Len() int
	StringValues() []string
}
