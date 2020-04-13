package rule

import (
    "bytes"
    diffpkg "github.com/pantheon-systems/search-secrets/pkg/diff"
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    "github.com/sirupsen/logrus"
    gitdiff "gopkg.in/src-d/go-git.v4/plumbing/format/diff"
    gitobject "gopkg.in/src-d/go-git.v4/plumbing/object"
)

type FileChangeContext struct {
    FileChange  *gitobject.Change
    patch       *gitobject.Patch
    chunks      []gitdiff.Chunk
    patchString string
    diff        *diffpkg.Diff
    commit      *gitobject.Commit
    repoName    string
    log         *logrus.Entry
}

func NewFileChangeContext(repoName string, commit *gitobject.Commit, fileChange *gitobject.Change, log *logrus.Entry) (result *FileChangeContext) {
    return &FileChangeContext{
        FileChange: fileChange,
        commit:     commit,
        repoName:   repoName,
        log:        log,
    }
}

func (fcc *FileChangeContext) Patch() (result *gitobject.Patch, err error) {
    defer func() {
        if recovered := recover(); recovered != nil {
            err = errors.PanicWithMessage(recovered, "unable to retrieve patch")
        }
    }()

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

    if patch == nil {
        fcc.log.Error("Filepatches is nil?")
        return nil, nil
    }
    result = patch.FilePatches()[0]
    return
}

func (fcc *FileChangeContext) IsBinaryOrEmpty() (result bool, err error) {
    var filePatch gitdiff.FilePatch
    filePatch, err = fcc.FilePatch()
    if err != nil {
        return
    }

    result = filePatch.IsBinary()
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
