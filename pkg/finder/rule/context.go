package rule

import (
    "bytes"
    diffpkg "github.com/pantheon-systems/search-secrets/pkg/diff"
    gitdiff "gopkg.in/src-d/go-git.v4/plumbing/format/diff"
    gitobject "gopkg.in/src-d/go-git.v4/plumbing/object"
)

type FileChangeContext struct {
    FileChange  *gitobject.Change
    patch       *gitobject.Patch
    chunks      []gitdiff.Chunk
    patchString string
    diff        *diffpkg.Diff
}

func NewFileChangeContext(fileChange *gitobject.Change) (result *FileChangeContext) {
    return &FileChangeContext{
        FileChange: fileChange,
    }
}

func (fcc *FileChangeContext) Patch() (result *gitobject.Patch, err error) {
    if fcc.patch == nil {
        fcc.patch, err = fcc.FileChange.Patch()
        if err != nil {
            return
        }
    }
    result = fcc.patch
    return
}

func (fcc *FileChangeContext) FilePatch() (result gitdiff.FilePatch, err error) {
    var patch *gitobject.Patch
    patch, err = fcc.Patch()
    if err != nil {
        return
    }

    result = patch.FilePatches()[0]
    return
}

func (fcc *FileChangeContext) Chunks() (result []gitdiff.Chunk, err error) {
    if fcc.chunks == nil {
        var filePatch gitdiff.FilePatch
        filePatch, err = fcc.FilePatch()
        if err != nil {
            return
        }

        fcc.chunks = filePatch.Chunks()
    }
    result = fcc.chunks
    return
}

func (fcc *FileChangeContext) PatchString() (result string, err error) {
    if fcc.patchString == "" {
        var patch *gitobject.Patch
        patch, err = fcc.Patch()
        if err != nil {
            return
        }

        buf := bytes.NewBuffer(nil)
        encoder := gitdiff.NewUnifiedEncoder(buf, 3)
        if err = encoder.Encode(patch); err != nil {
            return
        }

        fcc.patchString = buf.String()
    }

    result = fcc.patchString
    return
}

func (fcc *FileChangeContext) Diff() (result *diffpkg.Diff, err error) {
    if fcc.diff == nil {
        var chunks []gitdiff.Chunk
        chunks, err = fcc.Chunks()
        if err != nil {
            return
        }

        fcc.diff, err = diffpkg.NewFromChunks(chunks)
        if err != nil {
            return
        }
    }

    result = fcc.diff
    return
}



func (fcc *FileChangeContext) HasCodeChanges() (result bool, err error) {
    var chunks []gitdiff.Chunk
    chunks, err = fcc.Chunks()
    if err != nil {
        return
    }

    result = len(chunks) > 0
    return
}
