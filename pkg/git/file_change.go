package git

import (
	"path"
	"strings"

	"github.com/pantheon-systems/search-secrets/pkg/errors"
	gitdiff "gopkg.in/src-d/go-git.v4/plumbing/format/diff"
	gitobject "gopkg.in/src-d/go-git.v4/plumbing/object"
)

type (
	FileChange struct {
		Commit          *Commit
		Path            string
		Chunks          []*Chunk
		IsBinaryOrEmpty bool
		fileChangeMemo
	}
	fileChangeMemo struct {
		diff *Diff
	}
	Chunk struct {
		Operation DiffOperation
		Content   string
	}
)

func NewFileChange(commit *Commit, gitFileChange *gitobject.Change) (result *FileChange, err error) {
	var chunks []*Chunk
	var isBinaryOrEmpty bool
	chunks, isBinaryOrEmpty, err = gatherPatchData(commit, gitFileChange)
	if err != nil {
		err = errors.WithMessage(err, "unable to get file patch")
		return
	}

	result = &FileChange{
		Commit:          commit,
		Path:            gitFileChange.To.Name,
		Chunks:          chunks,
		IsBinaryOrEmpty: isBinaryOrEmpty,
		fileChangeMemo:  fileChangeMemo{},
	}
	return
}

func (fcc *FileChange) Diff() (result *Diff, err error) {
	if fcc.diff != nil {
		result = fcc.diff
		return
	}

	if len(fcc.Chunks) == 0 {
		err = errors.New("no chunks")
		return
	}

	var lineMap map[int]int
	var lineStrings []string
	lineMap, lineStrings, err = buildDiffLineInfo(fcc.Chunks)
	if err != nil {
		err = errors.WithMessage(err, "unable to get build line info")
		return
	}

	var diff *Diff
	diff, err = NewDiff(lineStrings, lineMap)
	if err != nil {
		err = errors.WithMessage(err, "unable to build diff")
		return
	}

	result = diff

	return
}

func (fcc *FileChange) HasCodeChanges() (result bool) {
	return len(fcc.Chunks) > 0
}

func (fcc *FileChange) FileContents() (result string, err error) {
	return fcc.Commit.FileContents(fcc.Path)
}

func (fcc *FileChange) IsDeletion() bool {
	return fcc.Path == ""
}

func (fcc *FileChange) FileType() (result string) {
	result = path.Ext(fcc.Path)
	if result == "" {
		result = path.Base(fcc.Path)
	}
	return
}

func gatherPatchData(commit *Commit, gitFileChange *gitobject.Change) (chunks []*Chunk, isBinaryOrEmpty bool, err error) {
	var filePatch gitdiff.FilePatch
	filePatch, err = getFilePatch(commit, gitFileChange)
	if err != nil {
		err = errors.WithMessage(err, "unable to get file patch")
		return
	}

	// Get chunks
	gitChunks := filePatch.Chunks()
	chunks = make([]*Chunk, len(gitChunks))
	for i, gitChunk := range gitChunks {
		chunks[i] = &Chunk{
			Operation: NewDiffOperationFromGit(gitChunk.Type()),
			Content:   gitChunk.Content(),
		}
	}

	// Get binary flag
	isBinaryOrEmpty = filePatch.IsBinary()

	return
}

func getFilePatch(commit *Commit, gitFileChange *gitobject.Change) (result gitdiff.FilePatch, err error) {
	commit.repository.mutex.Lock()
	defer commit.repository.mutex.Unlock()

	defer errors.CatchPanicSetErr(&err, "panic getting file patch")

	// Get file patch
	var patch *gitobject.Patch
	patch, err = gitFileChange.Patch()
	if err != nil {
		err = errors.WithMessage(err, "unable to get file patch")
		return
	}
	if patch == nil {
		err = errors.New("patch is nil")
		return
	}
	filePatches := patch.FilePatches()
	if filePatches == nil {
		err = errors.New("file patches is nil")
		return
	}
	if filePatches[0] == nil {
		err = errors.New("file patch is nil")
		return
	}
	result = filePatches[0]

	return
}

func buildDiffLineInfo(chunks []*Chunk) (result map[int]int, diffLines []string, err error) {
	fileLineNum := 1
	diffLineNum := 1

	result = map[int]int{}
	for _, chunk := range chunks {
		chunkString := chunk.Content

		// Remove the trailing line break
		chunkLen := len(chunkString)
		if chunkLen > 0 && chunkString[chunkLen-1:] == "\n" {
			chunkString = chunkString[:chunkLen-1]
		}

		lines := strings.Split(chunkString, "\n")
		prefix := chunk.Operation.Prefix()

		for _, line := range lines {
			result[diffLineNum] = fileLineNum

			diffLines = append(diffLines, prefix+line)

			// Prepare for next
			diffLineNum += 1
			if chunk.Operation != Delete {
				fileLineNum += 1
			}
		}
	}

	return
}
