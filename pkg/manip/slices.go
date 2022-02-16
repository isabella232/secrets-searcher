package manip

import (
	"sort"
)

func SlicesAreEqual(a, b []interface{}) bool {
	if (a == nil) != (b == nil) {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func StringValuesEqualAfterSort(a, b []string) bool {
	if (a == nil) != (b == nil) {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	sort.Strings(a)
	sort.Strings(b)

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func FirstDuplicate(input []string) (result string, ok bool) {
	ss := make(map[string]interface{})
	for _, result = range input {
		if _, ok = ss[result]; ok {
			return
		}
	}
	return
}

func SliceContains(ss []string, findStr string) bool {
	for _, s := range ss {
		if s == findStr {
			return true
		}
	}
	return false
}

func DowncastSlice(ss []string) (result []interface{}) {
	result = make([]interface{}, len(ss))
	for i, s := range ss {
		result[i] = s
	}
	return
}
