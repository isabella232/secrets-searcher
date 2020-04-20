package structures

import "sort"

func SlicesAreEqual(a, b []string) bool {
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

func SlicesAreEqualAfterSort(a, b []string) bool {
    sort.Strings(a)
    sort.Strings(b)

    return SlicesAreEqual(a, b)
}
