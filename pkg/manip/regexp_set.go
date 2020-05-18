package manip

import (
	"regexp"
)

type RegexpSet struct {
	resSet Set
}

func (r *RegexpSet) Res() Set {
	return r.resSet
}

func NewRegexpSet(values []*regexp.Regexp) (result *RegexpSet) {
	res := make([]interface{}, len(values))
	for i := range values {
		res[i] = values[i]
	}
	return &RegexpSet{NewBasicSet(res)}
}

func NewRegexpSetFromStringsMustCompile(values []string) (result *RegexpSet) {
	res := make([]interface{}, len(values))
	for i, value := range values {
		res[i] = regexp.MustCompile(value)
	}

	return &RegexpSet{NewBasicSet(res)}
}

func (r *RegexpSet) FindStringSubmatchAny(input string) (result []string) {
	for _, re := range r.resSet.Values() {
		matches := re.(*regexp.Regexp).FindStringSubmatch(input)
		if matches == nil {
			continue
		}
		return matches
	}

	return
}

func (r *RegexpSet) FirstMatchingRe(input string) (result *regexp.Regexp) {
	for _, re := range r.resSet.Values() {
		reCast := re.(*regexp.Regexp)
		if reCast.MatchString(input) {
			return reCast
		}
	}

	return
}

func (r *RegexpSet) FindStringSubmatchIndex(line string) (result []int, re *regexp.Regexp) {
	for _, reInt := range r.resSet.Values() {
		reCast := reInt.(*regexp.Regexp)
		result = reCast.FindStringSubmatchIndex(line)
		if result != nil {
			re = reCast
			return
		}
	}

	return
}

func (r *RegexpSet) MatchAny(input string) (result bool) {
	for _, re := range r.resSet.Values() {
		if re.(*regexp.Regexp).MatchString(input) {
			return true
		}
	}

	return false
}

func (r *RegexpSet) Add(i interface{}) {
	r.resSet.Add(i.(*regexp.Regexp))
}

func (r *RegexpSet) AddSliceValues(values []interface{}) {
	for _, value := range values {
		r.resSet.Add(value)
	}
}

func (r *RegexpSet) Remove(item interface{}) {
	r.resSet.Remove(item)
}

func (r *RegexpSet) Contains(i interface{}) bool {
	return r.resSet.Contains(i)
}

func (r *RegexpSet) Values() []interface{} {
	return r.resSet.Values()
}

func (r *RegexpSet) ReValues() (result []*regexp.Regexp) {
	values := r.resSet.Values()
	result = make([]*regexp.Regexp, len(values))
	for i := range values {
		result[i] = values[i].(*regexp.Regexp)
	}
	return
}

func (r *RegexpSet) Len() int {
	return r.resSet.Len()
}

func (r *RegexpSet) StringValues() []string {
	return r.resSet.StringValues()
}

func (r *RegexpSet) IsEmpty() (result bool) {
	return r.resSet.IsEmpty()
}
