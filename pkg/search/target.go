package search

import (
	"math/bits"
	"regexp"
	"sort"

	entropypkg "github.com/pantheon-systems/search-secrets/pkg/entropy"
	"github.com/pantheon-systems/search-secrets/pkg/logg"
	"github.com/pantheon-systems/search-secrets/pkg/manip"
	"github.com/pantheon-systems/search-secrets/pkg/search/rulebuild"
)

var (
	filePathLikeReSet = manip.NewRegexpSetFromStringsMustCompile([]string{
		`^[A-Za-z0-9-_/]+\.[A-Za-z0-9-_]{2,5}$`,
		`^/[A-Za-z0-9-_/]$`,
	})

	// TODO
	variableLikeReSet = manip.NewRegexpSetFromStringsMustCompile([]string{
		`zzzzzzzzzzzzzzzzzzzzzzz`,
	})
)

type Target struct {
	Name string

	// Patterns matching suspect key names
	KeyPatterns []string
	keyRe       *regexp.Regexp

	ExludeKeyPatterns []string
	exludeKeyRe       *regexp.Regexp

	// Pattern matching any single valid character in a value
	ValChars   []string
	valCharsRe *regexp.Regexp

	// Length restrictions
	ValLenMin int
	ValLenMax int

	ValEntropyMin float64

	SkipFilePathLikeValues bool
	SkipVariableLikeValues bool
}

func NewTarget(name string, keyPatterns, exludeKeyPatterns, valChars []string, valLenMin, valLenMax int, valEntropyMin float64, skipFilePathLikeValues, skipVariableLikeValues bool) (t *Target) {
	t = &Target{
		Name:                   name,
		KeyPatterns:            keyPatterns,
		ExludeKeyPatterns:      exludeKeyPatterns,
		ValChars:               valChars,
		ValLenMin:              valLenMin,
		ValLenMax:              valLenMax,
		ValEntropyMin:          valEntropyMin,
		SkipFilePathLikeValues: skipFilePathLikeValues,
		SkipVariableLikeValues: skipVariableLikeValues,
	}

	t.keyRe = regexp.MustCompile(rulebuild.NoCase + rulebuild.MatchAnyOf(t.KeyPatterns...))
	if t.ExludeKeyPatterns != nil {
		t.exludeKeyRe = regexp.MustCompile(rulebuild.NoCase + rulebuild.MatchAnyOf(t.ExludeKeyPatterns...))
	}
	if t.ValChars != nil {
		t.valCharsRe = regexp.MustCompile(rulebuild.NoCase + rulebuild.MatchAnyCharOf(t.ValChars...))
	}

	return t
}

func (t *Target) MatchValChars(val string) bool {
	return t.valCharsRe.MatchString(val)
}

func (t *Target) Matches(key, val string, log logg.Logg) (result bool, reason TargetMatchResult) {
	log = log.AddPrefixPath(t.Name)

	valLen := len(val)
	if valLen < t.ValLenMin {
		reason = ValTooShort
		log.WithField("reason", reason).
			Tracef("no match: value \"%s\" must have longer than %d characters", val, t.ValLenMin)
		return
	}
	if valLen > t.ValLenMax {
		reason = ValTooLong
		log.WithField("reason", reason).
			Tracef("no match: value \"%s\" must have fewer than %d characters", val, t.ValLenMax)
		return
	}
	if t.valCharsRe != nil && !t.valCharsRe.MatchString(val) {
		reason = ValNoMatch
		log.WithField("reason", reason).
			Tracef("no match: value \"%s\" doesn't match value regex: %s", val, t.valCharsRe.String())
		return
	}
	if !t.keyRe.MatchString(key) {
		reason = KeyNoMatch
		log.WithField("reason", reason).
			Tracef("no match: key \"%s\" doesn't match key regex: %s", key, t.keyRe.String())
		return
	}
	if t.exludeKeyRe != nil && t.exludeKeyRe.MatchString(key) {
		reason = KeyExcluded
		log.WithField("reason", reason).
			Tracef("no match: key \"%s\" matches key exclusion regex: %s", key, t.exludeKeyRe.String())
		return
	}
	if t.SkipFilePathLikeValues && filePathLikeReSet.MatchAny(val) {
		reason = ValFilePath
		log.WithField("reason", reason).
			Tracef("no match: value \"%s\" looks like a file path", val)
		return
	}
	if t.SkipVariableLikeValues && variableLikeReSet.MatchAny(val) {
		reason = ValVariable
		log.WithField("reason", reason).
			Tracef("no match: value \"%s\" looks like a variable", val)
		return
	}
	if t.ValEntropyMin > 0 {
		entropy := entropypkg.AgainstCharset(val, entropypkg.Base64CharsetName)
		if entropy < t.ValEntropyMin {
			reason = ValEntropy
			log.WithField("reason", reason).
				Tracef("no match: value \"%s\" has entropy %f which must be greater than %f", val, entropy, t.ValEntropyMin)
			return
		}
	}

	reason = Match
	result = true

	return
}

//
// TargetSet

type TargetSet struct {
	Targets      []*Target
	keyRe        *regexp.Regexp
	excludeKeyRe *regexp.Regexp
	valCharsRe   *regexp.Regexp
}

func NewTargetSet(targets []*Target) (tt *TargetSet) {
	tt = &TargetSet{Targets: targets}

	keyPatterns := tt.KeyPatterns()
	tt.keyRe = regexp.MustCompile(rulebuild.NoCase + rulebuild.MatchAnyOf(keyPatterns...))
	excludeKeyPatterns := tt.ExcludeKeyPatterns()
	if excludeKeyPatterns != nil {
		tt.excludeKeyRe = regexp.MustCompile(rulebuild.NoCase + rulebuild.MatchAnyOf(excludeKeyPatterns...))
	}
	valChars := tt.ValChars()
	if valChars != nil {
		tt.valCharsRe = regexp.MustCompile(rulebuild.NoCase + rulebuild.MatchAnyCharOf(valChars...))
	}

	return tt
}

func (tt *TargetSet) KeyPatterns() (result []string) {
	set := manip.NewEmptyBasicSet()
	for _, target := range tt.Targets {
		for _, value := range target.KeyPatterns {
			set.Add(value)
		}
	}
	result = set.StringValues()
	sort.Strings(result)
	return result
}

func (tt *TargetSet) ExcludeKeyPatterns() (result []string) {
	set := manip.NewEmptyBasicSet()
	for _, target := range tt.Targets {
		for _, value := range target.ExludeKeyPatterns {
			set.Add(value)
		}
	}
	result = set.StringValues()
	sort.Strings(result)
	return result
}

const MaxInt = 1<<(bits.UintSize-1) - 1

func (tt *TargetSet) ValLenMinMax() (min, max int) {
	max = MaxInt
	for _, target := range tt.Targets {
		if target.ValLenMin > min {
			min = target.ValLenMin
		}
		if target.ValLenMax < max {
			max = target.ValLenMax
		}
	}
	return
}

func (tt *TargetSet) ValChars() (result []string) {
	set := manip.NewEmptyBasicSet()
	for _, target := range tt.Targets {
		for _, value := range target.ValChars {
			set.Add(value)
		}
	}
	result = set.StringValues()
	sort.Strings(result)
	return result
}

func (tt *TargetSet) Matches(key, val string, log logg.Logg) (result bool) {
	return tt.FirstMatch(key, val, log) != nil
}

func (tt *TargetSet) FirstMatch(key, val string, log logg.Logg) (result *Target) {
	for _, t := range tt.Targets {
		matches, _ := t.Matches(key, val, log)
		if matches {
			result = t
			return
		}
	}
	return
}

func (tt *TargetSet) MatchValChars(val string) bool {
	return tt.valCharsRe == nil || tt.valCharsRe.MatchString(val)
}
