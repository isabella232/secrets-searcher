package git

import (
    "github.com/pantheon-systems/search-secrets/pkg/structures"
    "time"
)

type (
    CommitFilter struct {
        Hashes               structures.Set
        EarliestTime         time.Time
        LatestTime           time.Time
        EarliestCommit       string
        LatestCommit         string
        ExcludeNoDiffCommits bool
    }
    FileChangeFilter struct {
        ExcludeFileDeletions         bool
        IncludeMatchingPaths         structures.RegexpSet
        ExcludeMatchingPaths         structures.RegexpSet
        ExcludeBinaryOrEmpty         bool
        ExcludeOnesWithNoCodeChanges bool
    }
)

func (cf *CommitFilter) OldestCommitIsIncluded() (result bool) {
    return cf.EarliestCommit == "" && cf.EarliestTime.IsZero()
}

func (cf *CommitFilter) IsIncludedInLogResults(commit *Commit, hashSet *structures.Set) (result, more bool) {
    more = true

    // Filter by earliest commit
    if cf.EarliestCommit != "" && (hashSet == nil || hashSet.Contains(cf.EarliestCommit)) {
        more = false
        return
    }

    // Filter by time
    if !cf.LatestTime.IsZero() && commit.Time.After(cf.LatestTime) {
        return
    }
    if commit.Time.Before(cf.EarliestTime) {
        more = false
        return
    }

    // Filter out merge commits
    if cf.ExcludeNoDiffCommits && !commit.CanDiff() {
        return
    }

    result = true

    return
}

func (cf *CommitFilter) IsIncluded(commit *Commit) (result bool) {
    result, _ = cf.IsIncludedInLogResults(commit, nil)
    return
}
