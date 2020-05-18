package manip

import "fmt"

type RegexpFilter struct {
	include *RegexpSet
	exclude *RegexpSet
}

func NewRegexpFilter(included, exclude *RegexpSet) *RegexpFilter {
	return &RegexpFilter{
		include: included,
		exclude: exclude,
	}
}

func NewRegexpIncludeFilter(include []string) *RegexpFilter {
	return &RegexpFilter{
		include: NewRegexpSetFromStringsMustCompile(include),
		exclude: NewRegexpSet(nil),
	}
}

func NewRegexpExcludeFilter(exclude []string) *RegexpFilter {
	return &RegexpFilter{
		include: NewRegexpSet(nil),
		exclude: NewRegexpSetFromStringsMustCompile(exclude),
	}
}

// Create from string slices
func NewStringRegexpFilter(include, exclude []string) *RegexpFilter {
	return &RegexpFilter{
		include: NewRegexpSetFromStringsMustCompile(include),
		exclude: NewRegexpSetFromStringsMustCompile(exclude),
	}
}

// Returns true if any string is included
func (f *RegexpFilter) IncludesAnything() bool {
	return !f.include.IsEmpty() && !f.exclude.IsEmpty()
}

func (f *RegexpFilter) IncludesAllOf(items Set) bool {
	for _, item := range items.Values() {
		if !f.Includes(item) {
			return false
		}
	}
	return true
}

func (f *RegexpFilter) IncludesAnyOf(items Set) bool {
	for _, item := range items.Values() {
		if f.Includes(item) {
			return true
		}
	}
	return false
}

// Returns true if an item is included
func (f *RegexpFilter) Includes(value interface{}) bool {
	stringValue := fmt.Sprintf("%v", value)
	return !f.exclude.MatchAny(stringValue) && (f.include.IsEmpty() || f.include.MatchAny(stringValue))
}

// Returns only included items in a BasicSet object
func (f *RegexpFilter) FilterSet(items Set) {
	for _, item := range items.Values() {
		if !f.Includes(item) {
			items.Remove(item)
		}
	}
	return
}

func (f *RegexpFilter) CanProvideExactValues() bool {
	return false
}

func (f *RegexpFilter) ExactValues() Set {
	return nil
}
