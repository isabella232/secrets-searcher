package git

import (
    "bytes"
    diffpkg "github.com/pantheon-systems/search-secrets/pkg/diff"
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    "github.com/pantheon-systems/search-secrets/pkg/git/diff_operation"
    gitdiff "gopkg.in/src-d/go-git.v4/plumbing/format/diff"
    gitobject "gopkg.in/src-d/go-git.v4/plumbing/object"
    "strings"
)

type (
    FileChange struct {
        commit        *Commit
        Path          string
        gitFileChange *gitobject.Change
        memo          fileChangeMemo
    }
    Chunk struct {
        Type    diff_operation.DiffOperationEnum
        Content string
    }
    fileChangeMemo struct {
        gitPatch    *gitobject.Patch
        gitChunks   []gitdiff.Chunk
        chunks      []Chunk
        patchString string
        diff        *diffpkg.Diff
    }
)

func NewFileChange(commit *Commit, gitFileChange *gitobject.Change) (result *FileChange) {
    return &FileChange{
        commit:        commit,
        Path:          gitFileChange.To.Name,
        gitFileChange: gitFileChange,
        memo:          fileChangeMemo{},
    }
}

func (fcc *FileChange) IsBinaryOrEmpty() (result bool, err error) {
    var filePatch gitdiff.FilePatch
    filePatch, err = fcc.getGitFilePatch()
    if err != nil {
        err = errors.WithMessage(err, "unable to get file patch")
        return
    }

    result = filePatch.IsBinary()
    return
}

func (fcc *FileChange) Chunks() (result []Chunk, err error) {
    if fcc.memo.chunks != nil {
        result = fcc.memo.chunks
        return
    }

    var filePatch gitdiff.FilePatch
    filePatch, err = fcc.getGitFilePatch()
    if err != nil {
        err = errors.WithMessage(err, "unable to get file patch")
        return
    }

    for _, gitChunk := range filePatch.Chunks() {
        result = append(result, Chunk{
            Type:    diff_operation.NewDiffOperationFromGitOperation(gitChunk.Type()),
            Content: gitChunk.Content(),
        })
    }

    fcc.memo.chunks = result

    return
}

func (fcc *FileChange) Diff() (result *diffpkg.Diff, err error) {
    if fcc.memo.diff != nil {
        result = fcc.memo.diff
        return
    }

    var chunks []Chunk
    chunks, err = fcc.Chunks()

    if len(chunks) == 0 {
        err = errors.New("no chunks passed")
        return
    }

    var lineMap map[int]int
    var lineStrings []string
    lineMap, lineStrings, err = buildDiffLineInfo(chunks)
    if err != nil {
        err = errors.WithMessage(err, "unable to get file patch")
        return
    }

    var diff *diffpkg.Diff
    diff, err = diffpkg.New(lineStrings, lineMap)
    if err != nil {
        err = errors.WithMessage(err, "unable to build diff")
        return
    }

    result = diff

    return
}

func (fcc *FileChange) PatchString() (result string, err error) {
    if fcc.memo.patchString != "" {
        result = fcc.memo.patchString
        return
    }

    var patch *gitobject.Patch
    patch, err = fcc.getGitPatch()
    if err != nil {
        err = errors.WithMessage(err, "unable to get file patch")
        return
    }

    buf := bytes.NewBuffer(nil)
    encoder := gitdiff.NewUnifiedEncoder(buf, 3)
    if err = encoder.Encode(patch); err != nil {
        err = errors.WithMessage(err, "unable to encode file patch")
        return
    }

    result = buf.String()

    fcc.memo.patchString = result
    return
}

func (fcc *FileChange) HasCodeChanges() (result bool, err error) {
    var chunks []Chunk
    chunks, err = fcc.Chunks()
    result = len(chunks) > 0
    return
}

func (fcc *FileChange) getGitPatch() (result *gitobject.Patch, err error) {
    defer errors.CatchPanicSetErr(&err, "unable to retrieve patch")

    if fcc.memo.gitPatch != nil {
        return fcc.memo.gitPatch, nil
    }

    fcc.memo.gitPatch, err = fcc.wrapPatch()
    result = fcc.memo.gitPatch

    return
}

func (fcc *FileChange) getGitFilePatch() (result gitdiff.FilePatch, err error) {
    var patch *gitobject.Patch
    patch, err = fcc.getGitPatch()
    if err != nil {
        err = errors.WithMessage(err, "unable to get file patch")
        return
    }

    if patch == nil {
        err = errors.New("Filepatches is nil?")
        return
    }
    result = patch.FilePatches()[0]
    return
}

func (fcc *FileChange) IsDeletion() bool {
    return fcc.Path == ""
}

func buildDiffLineInfo(chunks []Chunk) (result map[int]int, diffLines []string, err error) {
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
        prefix := chunk.Type.Value().Prefix

        for _, line := range lines {
            result[diffLineNum] = fileLineNum

            diffLines = append(diffLines, prefix+line)

            // Prepare for next
            diffLineNum += 1
            if chunk.Type.Value() != diff_operation.DeleteEnum.Value() {
                fileLineNum += 1
            }
        }
    }

    return
}

func (fcc *FileChange) wrapPatch() (result *gitobject.Patch, err error) {
    fcc.commit.repository.mutex.Lock()
    defer fcc.commit.repository.mutex.Unlock()

    return fcc.gitFileChange.Patch()
}
