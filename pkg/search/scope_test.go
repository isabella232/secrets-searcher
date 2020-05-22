package search_test

import (
	"testing"
	"time"

	"github.com/pantheon-systems/search-secrets/pkg/git"
	. "github.com/pantheon-systems/search-secrets/pkg/search"
	statspkg "github.com/pantheon-systems/search-secrets/pkg/stats"
	"github.com/stretchr/testify/assert"
)

type proc struct {
	Name string
}

func (p proc) GetName() string {
	return p.Name
}

func TestScope_Finish(t *testing.T) {
	subject := NewScope(false, nil)

	subject.StartRepo("repo")
	subject.StartRepoJob("repojob")
	subject.StartCommit(&git.Commit{Hash: "commitHash"})
	subject.StartFileChange(&git.FileChange{Path: "path"})
	subject.StartProc(&proc{})
	subject.StartLine(1)

	// Fire
	subject.FinishRepo()

	assert.Nil(t, subject.RepoScope)
}

func TestScope_Profiling_RepoDurations(t *testing.T) {
	stats := statspkg.New()
	subject := NewScope(true, stats)

	// Fire
	subject.StartRepo("repo0")
	time.Sleep(5 * time.Millisecond)
	subject.StartRepo("repo1")
	time.Sleep(10 * time.Millisecond)
	subject.FinishRepo()
	ss := stats.RepoDurations.Stats()

	assert.NotNil(t, ss)

	assert.Len(t, ss, 2)
	assert.Equal(t, "repo1", ss[0].Item)
	assert.Equal(t, "repo0", ss[1].Item)
}

func TestScope_WithScopeFields(t *testing.T) {
	subject := NewScope(false, nil)

	// Fire
	subject.StartRepo("repo")
	subject.StartRepoJob("repojob")
	subject.StartCommit(&git.Commit{Hash: "commitHash"})
	subject.StartFileChange(&git.FileChange{Path: "path"})
	subject.StartProc(&proc{"proc"})
	subject.StartLine(1)

	data := subject.Fields()
	assert.Equal(t, "repo", data["repo"])
	assert.Equal(t, "repojob", data["job"])
	assert.Equal(t, "commitHash", data["commit"])
	assert.Equal(t, "path", data["path"])
	assert.Equal(t, "proc", data["processor"])
	assert.Equal(t, 1, data["line"])

	subject.FinishRepo()

	data = subject.Fields()
	assert.Nil(t, data["repo"])
	assert.Nil(t, data["job"])
	assert.Nil(t, data["commit"])
	assert.Nil(t, data["path"])
	assert.Nil(t, data["processor"])
	assert.Nil(t, data["line"])
}
