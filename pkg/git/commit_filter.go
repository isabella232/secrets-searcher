package git

import (
	"time"

	"github.com/pantheon-systems/search-secrets/pkg/manip"
)

type CommitFilter struct {
	HashFilter           manip.Filter
	EarliestTime         time.Time
	LatestTime           time.Time
	LatestTimeSet        bool
	ExcludeNoDiffCommits bool
}

func NewCommitFilter(
	hashFilter manip.Filter,
	earliestTime time.Time,
	latestTime time.Time,
	excludeNoDiffCommits bool,
) (result *CommitFilter) {
	if hashFilter == nil {
		hashFilter = manip.StringFilter(nil, nil)
	}
	return &CommitFilter{
		HashFilter:           hashFilter,
		EarliestTime:         earliestTime,
		LatestTime:           latestTime,
		LatestTimeSet:        false,
		ExcludeNoDiffCommits: excludeNoDiffCommits,
	}
}

func NewEmptyCommitFilter() *CommitFilter {
	return NewCommitFilter(
		nil,
		time.Time{},
		time.Time{},
		false,
	)
}

func (cf *CommitFilter) OldestCommitIsIncluded() (result bool) {
	return cf.EarliestTime.IsZero()
}

func (cf *CommitFilter) IsIncludedInLogResults(commit *Commit) (result, more bool) {
	more = true

	// SliceFilter by time
	if !cf.LatestTime.IsZero() && commit.Date.After(cf.LatestTime) {
		return
	}
	if commit.Date.Before(cf.EarliestTime) {
		more = false
		return
	}

	// SliceFilter out merge commits
	if cf.ExcludeNoDiffCommits && !commit.CanDiff() {
		return
	}

	if !cf.HashFilter.Includes(commit.Hash) {
		return
	}

	result = true

	return
}

func (cf *CommitFilter) Includes(commit interface{}) (result bool) {
	result, _ = cf.IsIncludedInLogResults(commit.(*Commit))
	return
}

func (cf *CommitFilter) IncludesAnything() bool {
	return cf.EarliestTime.IsZero() && cf.LatestTime.IsZero() &&
		(cf.HashFilter == nil || cf.HashFilter.IncludesAnything())
}

func (cf *CommitFilter) IncludesAllOf(items manip.Set) bool {
	for _, item := range items.Values() {
		if !cf.Includes(item) {
			return false
		}
	}
	return true
}

func (cf *CommitFilter) IncludesAnyOf(items manip.Set) bool {
	for _, item := range items.Values() {
		if cf.Includes(item) {
			return true
		}
	}
	return false
}

func (cf *CommitFilter) FilterSet(items manip.Set) {
	for _, item := range items.Values() {
		if !cf.Includes(item) {
			items.Remove(item)
		}
	}
}

func (cf *CommitFilter) CanProvideExactValues() bool {
	return false
}

func (cf *CommitFilter) ExactValues() (result manip.Set) {
	if !cf.CanProvideExactValues() {
		panic("use CanProvideExactValues to avoid this")
	}
	return
}

func (cf *CommitFilter) CanProvideExactCommitHashValues() bool {
	// FIXME This doesn't need to ignore everything but HashFilter, we could filter
	return cf.EarliestTime.IsZero() && cf.LatestTime.IsZero() &&
		cf.HashFilter.CanProvideExactValues()
}

func (cf *CommitFilter) ExactCommitHashValues() (result manip.Set) {
	if !cf.CanProvideExactCommitHashValues() {
		panic("use CanProvideExactValues to avoid this")
	}
	return cf.HashFilter.ExactValues()
}
