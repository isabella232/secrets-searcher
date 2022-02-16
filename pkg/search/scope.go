package search

import (
	"strings"
	"time"

	gitpkg "github.com/pantheon-systems/secrets-searcher/pkg/git"
	"github.com/pantheon-systems/secrets-searcher/pkg/search/contract"
	"github.com/pantheon-systems/secrets-searcher/pkg/stats"
)

type (
	Scope struct {
		*RepoScope
		enableProfiling bool
		stats           *stats.Stats
	}
	RepoScope struct {
		Repo  string
		start time.Time
		*RepoJobScope
	}
	RepoJobScope struct {
		RepoJob string
		start   time.Time
		*CommitScope
	}
	CommitScope struct {
		Commit *gitpkg.Commit
		start  time.Time
		*FileChangeScope
	}
	FileChangeScope struct {
		FileChange *gitpkg.FileChange
		start      time.Time
		*ProcScope
	}
	ProcScope struct {
		Proc  contract.NamedProcessorI
		start time.Time
		*LineScope
	}
	LineScope struct {
		Line  int
		start time.Time
	}
)

func NewScope(enableProfiling bool, stats *stats.Stats) *Scope {
	return &Scope{
		enableProfiling: enableProfiling,
		stats:           stats,
	}
}

func (s *Scope) StartRepo(repo string) {
	if !s.enableProfiling {
		s.RepoScope = &RepoScope{Repo: repo}
		return
	}

	now := time.Now()

	s.profileLine(now)
	s.profileProc(now)
	s.profileFileChange(now)
	s.profileCommit(now)
	s.profileRepoJob(now)
	s.profileRepo(now)

	s.RepoScope = &RepoScope{Repo: repo, start: now}
}

func (s *Scope) StartRepoJob(repoJob string) {
	if !s.hasRepo() {
		panic("can't enter repo job scope from here")
	}

	if !s.enableProfiling {
		s.RepoJobScope = &RepoJobScope{RepoJob: repoJob}
		return
	}

	now := time.Now()

	s.profileLine(now)
	s.profileProc(now)
	s.profileFileChange(now)
	s.profileCommit(now)
	s.profileRepoJob(now)

	s.RepoJobScope = &RepoJobScope{RepoJob: repoJob, start: now}
}

func (s *Scope) StartCommit(commit *gitpkg.Commit) {
	if !s.hasRepoJob() {
		panic("can't enter commit scope from here")
	}

	if !s.enableProfiling {
		s.CommitScope = &CommitScope{Commit: commit}
		return
	}

	now := time.Now()

	s.profileLine(now)
	s.profileProc(now)
	s.profileFileChange(now)
	s.profileCommit(now)

	s.CommitScope = &CommitScope{Commit: commit, start: now}
}

func (s *Scope) StartFileChange(fileChange *gitpkg.FileChange) {
	if !s.hasCommit() {
		panic("can't enter file change scope from here")
	}

	if !s.enableProfiling {
		s.FileChangeScope = &FileChangeScope{FileChange: fileChange}
		return
	}

	now := time.Now()

	s.profileLine(now)
	s.profileProc(now)
	s.profileFileChange(now)

	s.FileChangeScope = &FileChangeScope{FileChange: fileChange, start: now}
}

func (s *Scope) StartProc(proc contract.NamedProcessorI) {
	if !s.hasFileChange() {
		panic("can't enter processor scope from here")
	}

	if !s.enableProfiling {
		s.ProcScope = &ProcScope{Proc: proc}
		return
	}

	now := time.Now()

	s.profileLine(now)
	s.profileProc(now)

	s.ProcScope = &ProcScope{Proc: proc, start: now}
}

func (s *Scope) StartLine(line int) {
	if !s.hasProc() {
		panic("can't enter line scope from here")
	}

	if !s.enableProfiling {
		s.LineScope = &LineScope{Line: line}
		return
	}

	now := time.Now()

	s.profileLine(now)

	s.LineScope = &LineScope{Line: line, start: now}
}

func (s *Scope) FinishRepo() {
	now := time.Now()

	if !s.enableProfiling {
		s.RepoScope = nil
		return
	}

	s.profileLine(now)
	s.profileProc(now)
	s.profileFileChange(now)
	s.profileCommit(now)
	s.profileRepoJob(now)
	s.profileRepo(now)

	s.RepoScope = nil
}

func (s *Scope) profileRepo(now time.Time) {
	if s.hasRepo() {
		duration := now.Sub(s.RepoScope.start)
		s.stats.RepoDurations.SubmitAggregatedDuration(duration, s.Repo)
	}
}

func (s *Scope) profileRepoJob(_ time.Time) {
}

func (s *Scope) profileCommit(now time.Time) {
	if s.hasCommit() {
		duration := now.Sub(s.CommitScope.start)
		name := scopePath(s.Repo, s.Commit.Hash)
		s.stats.CommitDurations.SubmitUniqueDuration(duration, name)
	}
}

func (s *Scope) profileFileChange(now time.Time) {
	if s.hasFileChange() {
		duration := now.Sub(s.FileChangeScope.start)
		name := scopePath(s.Repo, s.Commit.Hash, s.FileChange.Path)
		s.stats.FileChangeDurations.SubmitUniqueDuration(duration, name)

		s.stats.FileTypeDurations.SubmitAggregatedDuration(duration, s.FileChange.FileType())
	}
}

func (s *Scope) profileProc(_ time.Time) {
}

func (s *Scope) profileLine(_ time.Time) {
}

func (s *Scope) hasRepo() bool {
	return s.RepoScope != nil
}

func (s *Scope) hasRepoJob() bool {
	return s.hasRepo() && s.RepoJobScope != nil
}

func (s *Scope) hasCommit() bool {
	return s.hasRepoJob() && s.CommitScope != nil
}

func (s *Scope) hasFileChange() bool {
	return s.hasCommit() && s.FileChangeScope != nil
}

func (s *Scope) hasProc() bool {
	return s.hasFileChange() && s.ProcScope != nil
}

func (s *Scope) hasLine() bool {
	return s.hasProc() && s.LineScope != nil
}

func (s *Scope) Fields() (result map[string]interface{}) {
	result = map[string]interface{}{}
	if s.hasRepo() {
		result["repo"] = s.Repo
	}
	if s.hasRepoJob() {
		result["job"] = s.RepoJob
	}
	if s.hasCommit() {
		result["commit"] = s.Commit.Hash
	}
	if s.hasFileChange() {
		result["path"] = s.FileChange.Path
	}
	if s.hasProc() {
		result["processor"] = s.Proc.GetName()
	}
	if s.hasLine() {
		result["line"] = s.Line
	}
	return
}

func scopePath(ss ...string) string {
	return strings.Join(ss, " / ")
}
