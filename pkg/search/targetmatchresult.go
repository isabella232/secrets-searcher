package search

//go:generate stringer -type TargetMatchResult

import "strings"

type TargetMatchResult int

const (
	Match TargetMatchResult = iota
	KeyNoMatch
	KeyExcluded
	ValTooShort
	ValTooLong
	ValNoMatch
	ValFilePath
	ValVariable
	ValEntropy
)

func TargetNoMatchReasons() []TargetMatchResult {
	return []TargetMatchResult{
		Match,
		KeyNoMatch,
		KeyExcluded,
		ValTooShort,
		ValTooLong,
		ValNoMatch,
		ValFilePath,
		ValVariable,
		ValEntropy,
	}
}

func (i TargetMatchResult) Value() string {
	return strings.ToLower(i.String())
}

func NewTargetNoMatchReasonFromValue(val string) TargetMatchResult {
	for _, e := range TargetNoMatchReasons() {
		if e.Value() == val {
			return e
		}
	}
	panic("unknown target match reason: " + val)
}
