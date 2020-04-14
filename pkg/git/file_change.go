package git

import (
    "bytes"
    diffpkg "github.com/pantheon-systems/search-secrets/pkg/diff"
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    gitdiff "gopkg.in/src-d/go-git.v4/plumbing/format/diff"
    gitobject "gopkg.in/src-d/go-git.v4/plumbing/object"
    "strings"
)

type (
    FileChange struct {
        Path          string
        gitFileChange *gitobject.Change
        memo          memo
    }
    Chunk struct {
        Type    DiffOperationEnum
        Content string
    }
    memo struct {
        gitPatch    *gitobject.Patch
        gitChunks   []gitdiff.Chunk
        chunks      []Chunk
        patchString string
        diff        *diffpkg.Diff
    }
)

func NewFileChange(gitFileChange *gitobject.Change) (result *FileChange) {
    return &FileChange{
        Path:          gitFileChange.To.Name,
        gitFileChange: gitFileChange,
        memo:          memo{},
    }
}

func (fcc *FileChange) IsBinaryOrEmpty() (result bool, err error) {
    var filePatch gitdiff.FilePatch
    filePatch, err = fcc.getGitFilePatch()
    if err != nil {
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
        return
    }

    for _, gitChunk := range filePatch.Chunks() {
        result = append(result, Chunk{
            Type:    NewDiffOperationFromGitOperation(gitChunk.Type()),
            Content: gitChunk.Content(),
        })
    }

    fcc.memo.chunks = result

    return
}

func (fcc *FileChange) Diff() (result *diffpkg.Diff, err error) {
    if fcc.memo.diff !=nil {
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
        return
    }

    var diff *diffpkg.Diff
    diff, err = diffpkg.New(lineStrings, lineMap)
    if err != nil {
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
        return
    }

    buf := bytes.NewBuffer(nil)
    encoder := gitdiff.NewUnifiedEncoder(buf, 3)
    if err = encoder.Encode(patch); err != nil {
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
    defer func() {
        if recovered := recover(); recovered != nil {
            err = errors.PanicWithMessage(recovered, "unable to retrieve patch")
        }
    }()

    if fcc.memo.gitPatch != nil {
        return fcc.memo.gitPatch, nil
    }

    fcc.memo.gitPatch, err = fcc.gitFileChange.Patch()
    result = fcc.memo.gitPatch

    return
}

func (fcc *FileChange) getGitFilePatch() (result gitdiff.FilePatch, err error) {
    var patch *gitobject.Patch
    patch, err = fcc.getGitPatch()
    if err != nil {
        return
    }

    if patch == nil {
        err =errors.New("Filepatches is nil?")
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
            if chunk.Type.Value() != DeleteEnum.Value() {
                fileLineNum += 1
            }
        }
    }

    return
}
