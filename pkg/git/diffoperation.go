package git

//go:generate stringer -type DiffOperation

import (
	"fmt"

	gitdiff "gopkg.in/src-d/go-git.v4/plumbing/format/diff"
)

type DiffOperation int

const (
	Equal DiffOperation = iota
	Delete
	Add
)

func NewDiffOperationFromGit(gdo gitdiff.Operation) DiffOperation {
	switch gdo {
	case gitdiff.Equal:
		return Equal
	case gitdiff.Delete:
		return Delete
	case gitdiff.Add:
		return Add
	default:
		panic("unknown git operation")
	}
}

func (i DiffOperation) Prefix() string {
	switch i {
	case Equal:
		return " "
	case Delete:
		return "-"
	case Add:
		return "+"
	default:
		return fmt.Sprintf("%d", int(i))
	}
}
