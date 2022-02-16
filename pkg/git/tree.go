package git

import (
	"sync"

	gitobject "gopkg.in/src-d/go-git.v4/plumbing/object"
)

type (
	Tree struct {
		gitTree *gitobject.Tree
		mutex   *sync.Mutex
	}
)

func newTree(gitTree *gitobject.Tree) (result *Tree) {
	result = &Tree{
		gitTree: gitTree,
		mutex:   &sync.Mutex{},
	}

	return
}

func (c *Tree) wrapDiff(to *Tree) (gitobject.Changes, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.gitTree.Diff(to.gitTree)
}
