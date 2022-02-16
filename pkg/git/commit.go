package git

import (
	"sync"
	"time"

	"github.com/pantheon-systems/secrets-searcher/pkg/errors"
	gitobject "gopkg.in/src-d/go-git.v4/plumbing/object"
)

type (
	Commit struct {
		repository  *Repository
		Hash        string
		Message     string
		Date        time.Time
		AuthorName  string
		AuthorEmail string
		Oldest      bool
		Tree        *Tree
		gitCommit   *gitobject.Commit
		commitState
	}
	commitState struct {
		mutex            *sync.Mutex
		fileContentIndex map[string]string
		parentsCache     []*Commit
	}
)

func newCommit(repository *Repository, gitCommit *gitobject.Commit) (result *Commit, err error) {
	var tree *Tree
	tree, err = getTree(repository, gitCommit)
	if err != nil {
		err = errors.WithMessage(err, "unable to get tree")
		return
	}

	result = &Commit{
		repository:  repository,
		Hash:        gitCommit.Hash.String(),
		Message:     gitCommit.Message,
		Date:        gitCommit.Committer.When,
		AuthorName:  gitCommit.Author.Name,
		AuthorEmail: gitCommit.Author.Email,
		Tree:        tree,
		gitCommit:   gitCommit,
		commitState: commitState{
			fileContentIndex: map[string]string{},
			mutex:            &sync.Mutex{},
		},
	}

	return
}

func (c *Commit) Parents() (result []*Commit, err error) {
	defer errors.CatchPanicSetErr(&err, "unable to retrieve parent commits")

	if c.parentsCache != nil {
		result = c.parentsCache
		return
	}

	// Get parents
	result, err = c.repository.newCommitsFromIter(c.parents())
	if err != nil {
		err = errors.WithMessage(err, "unable to get parent commits of commit")
		return
	}

	c.parentsCache = result

	return
}

func (c *Commit) HasParents() (result bool, err error) {
	var parents []*Commit
	parents, err = c.Parents()
	result = err == nil && len(parents) > 0
	return
}

func (c *Commit) IsMergeCommit() (result bool, err error) {
	var parents []*Commit
	parents, err = c.Parents()
	result = err == nil && len(parents) > 1
	return
}

func (c *Commit) CanDiff() (result bool) {
	if c.Oldest {
		return true
	}

	var parents []*Commit
	parents, _ = c.Parents()

	c.recheckIsOldest(true)
	if c.Oldest {
		return true
	}

	result = len(parents) == 1

	return
}

func (c *Commit) FileChanges(filter *FileChangeFilter) (result []*FileChange, err error) {

	// Are we going to get
	if !c.CanDiff() {
		return
	}

	// Get parent
	var parents []*Commit
	parents, err = c.Parents()
	if err != nil {
		err = errors.WithMessage(err, "unable to get parents")
		return
	}
	parentsLen := len(parents)

	// Get parent tree
	var parentCommitTree *Tree
	if parentsLen == 1 {
		parentCommitTree = parents[0].Tree
	} else if c.Oldest {
		parentCommitTree = c.repository.EmptyTree()
	} else {
		err = errors.New("unable to find parent (we shouldn't be here since we're calling CanDiff above)")
		return
	}

	// Diff them together
	var gitFileChanges gitobject.Changes
	gitFileChanges, err = parentCommitTree.wrapDiff(c.Tree)
	if err != nil {
		err = errors.WithMessage(err, "unable to diff")
		return
	}

	// For each file change, build a FileChange object
	for _, gitFileChange := range gitFileChanges {
		var fileChange *FileChange
		if fileChange, err = NewFileChange(c, gitFileChange); err != nil {
			err = errors.WithMessage(err, "unable build file change")
			return
		}
		if filter != nil && !filter.Includes(fileChange) {
			continue
		}
		result = append(result, fileChange)
	}

	return
}

func (c *Commit) FileContents(path string) (result string, err error) {
	var ok bool
	result, ok = c.fileContentIndex[path]
	if ok {
		return
	}

	var file *gitobject.File
	file, err = c.gitCommit.File(path)
	if err != nil {
		err = errors.WithMessage(err, "unable to get file at commit")
		return
	}

	result, err = file.Contents()

	c.fileContentIndex[path] = result

	return
}

func (c *Commit) parents() gitobject.CommitIter {
	c.repository.mutex.Lock()
	defer c.repository.mutex.Unlock()

	return c.gitCommit.Parents()
}

// Expensive
func (c *Commit) recheckIsOldest(force bool) (result bool) {
	return // FIXME This is causing a hang for some reason
	if !force {
		result = c.Oldest
		return
	}

	var oldest *Commit
	oldest, _ = c.repository.GetOldestCommit()

	result = oldest != nil && oldest.Hash == c.Hash
	c.Oldest = result

	return
}

func getTree(repository *Repository, gitCommit *gitobject.Commit) (result *Tree, err error) {
	//repository.mutex.Lock()
	//defer repository.mutex.Unlock()

	var gitTree *gitobject.Tree
	gitTree, err = gitCommit.Tree()
	if err != nil {
		err = errors.WithMessage(err, "unable to get tree")
		return
	}

	result = newTree(gitTree)

	return
}
