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

func (r RegexpSet) MatchStringAny(input string) (result bool) {
    for _, re := range r {
        if re.MatchString(input) {
            return true
        }
    }
    return false
}
