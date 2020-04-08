package structures

import "regexp"

type RegexpSet []*regexp.Regexp

func NewRegexpSet(values []*regexp.Regexp) (result RegexpSet) {
    result = RegexpSet{}
    for _, value := range values {
        result = append(result, value)
    }
    return
}

func NewRegexpSetFromStrings(values []string) (result RegexpSet, err error) {
    result = RegexpSet{}
    for _, value := range values {
        var re *regexp.Regexp
        re, err = regexp.Compile(value)
        if err != nil {
            return
        }

        result = append(result, re)
    }

    return
}

func (r RegexpSet) MatchStringAny(input, substringToCompare string) (result bool) {
    for _, re := range r {
        matches := re.FindStringSubmatch(input)

        // Didn't match
        if matches == nil {
            continue
        }

        // Matches, but with no back reference, or nothing to compare, so return true
        if len(matches) == 1 || substringToCompare == "" {
            return true
        }

        // Compare the first backreference with the provided string
        return matches[1] == substringToCompare
    }

    return false
}
