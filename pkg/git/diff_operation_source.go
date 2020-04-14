package git

import gitdiff "gopkg.in/src-d/go-git.v4/plumbing/format/diff"

//go:generate sh -c "go run github.com/gdm85/go-genums DiffOperation value DiffOperationValue diff_operation_source.go > diff_operation.go"

type DiffOperationValue struct {
    Name   string
    Prefix string
}

var (
    DeleteEnum = Delete{}.New()

    valueEqual = DiffOperationValue{"equal", " "}
    valueDelete = DiffOperationValue{"delete", "-"}
    valueAdd = DiffOperationValue{"add", "+"}
)

func NewDiffOperationFromGitOperation(gdo gitdiff.Operation) DiffOperationEnum {
   switch gdo {
   case gitdiff.Equal:
       return Equal{}.New()
   case gitdiff.Delete:
       return Delete{}.New()
   case gitdiff.Add:
       return Add{}.New()
   default:
       panic("unable to create diff operation from git package operation")
   }
}
