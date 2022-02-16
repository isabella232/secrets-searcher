package git

import (
	"github.com/pantheon-systems/secrets-searcher/pkg/manip"
)

type FileChangeFilter struct {
	PathFilter                   *manip.RegexpFilter
	ExcludeFileDeletions         bool
	ExcludeBinaryOrEmpty         bool
	ExcludeOnesWithNoCodeChanges bool
}

func NewFileChangeFilter(
	pathFilter *manip.RegexpFilter,
	excludeFileDeletions bool,
	excludeBinaryOrEmpty bool,
	excludeOnesWithNoCodeChanges bool,
) (result *FileChangeFilter) {
	return &FileChangeFilter{
		PathFilter:                   pathFilter,
		ExcludeFileDeletions:         excludeFileDeletions,
		ExcludeBinaryOrEmpty:         excludeBinaryOrEmpty,
		ExcludeOnesWithNoCodeChanges: excludeOnesWithNoCodeChanges,
	}
}

func (cf *FileChangeFilter) Includes(input interface{}) (result bool) {
	fileChange := input.(*FileChange)

	// Filter out deletions
	if cf.ExcludeFileDeletions && fileChange.IsDeletion() {
		return false
	}

	if !cf.PathFilter.Includes(fileChange.Path) {
		return false
	}

	// Filter out ones with no code changes
	if cf.ExcludeOnesWithNoCodeChanges && !fileChange.HasCodeChanges() {
		return false
	}

	// Filter out empty or binary files
	if cf.ExcludeBinaryOrEmpty && fileChange.IsBinaryOrEmpty {
		return false
	}

	return true
}

func (cf *FileChangeFilter) IncludesAnything() (result bool) {
	return cf.PathFilter.IncludesAnything()
}

func (cf *FileChangeFilter) IncludesAllOf(items manip.Set) bool {
	for _, item := range items.Values() {
		if !cf.Includes(item) {
			return false
		}
	}
	return true
}

func (cf *FileChangeFilter) IncludesAnyOf(items manip.Set) bool {
	for _, item := range items.Values() {
		if !cf.Includes(item) {
			return true
		}
	}
	return false
}

func (cf *FileChangeFilter) FilterSet(fileChanges manip.Set) {
	for _, fileChange := range fileChanges.Values() {
		if !cf.Includes(fileChange) {
			fileChanges.Remove(fileChange)
		}
	}

	return
}

func (cf *FileChangeFilter) CanProvideExactValues() bool {
	return false
}

func (cf *FileChangeFilter) ExactValues() manip.Set {
	panic("never")
}
