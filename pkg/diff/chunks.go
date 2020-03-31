package diff

import (
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    gitdiff "gopkg.in/src-d/go-git.v4/plumbing/format/diff"
    "strings"
)

type ChunksDiff struct {
    Diff *Diff
    Map  map[int]int
}

func NewChunksDiff(chunks []gitdiff.Chunk) (result *ChunksDiff, err error) {
    var diffLines []string
    var mp map[int]int
    var diff *Diff

    diffLines, mp, err = buildDiffStringAndMap(chunks)
    if err != nil {
        return
    }

    diff, err = New(diffLines)
    if err != nil {
        return
    }

    result = &ChunksDiff{
        Diff: diff,
        Map:  mp,
    }

    return
}

func (d ChunksDiff) FileLineNum(diffLineNum int) (result int, err error) {
    var ok bool
    result, ok = d.Map[diffLineNum]
    if !ok {
        err = errors.Errorv("key does not exist in map", diffLineNum)
    }

    return
}

func (d ChunksDiff) RequireFileLineNum(diffLineNum int) (result int) {
    var err error
    result, err = d.FileLineNum(diffLineNum)
    if err != nil {
        panic(err)
    }

    return
}

func buildDiffStringAndMap(chunks []gitdiff.Chunk) (diffLines []string, mp map[int]int, err error) {
    mp = map[int]int{}
    fileLineNum := 1
    diffLineNum := 1

    for _, chunk := range chunks {
        chunkString := chunk.Content()

        // Remove the trailing line break
        chunkLen := len(chunkString)
        if chunkLen > 0 && chunkString[chunkLen-1:] == "\n" {
            chunkString = chunkString[:chunkLen-1]
        }

        lines := strings.Split(chunkString, "\n")
        prefix := getPrefix(chunk.Type())

        for _, line := range lines {
            mp[diffLineNum] = fileLineNum

            diffLines = append(diffLines, prefix+line)

            // Prepare for next
            diffLineNum += 1
            if chunk.Type() != gitdiff.Delete {
                fileLineNum += 1
            }
        }
    }

    return
}

func getPrefix(chunkType gitdiff.Operation) (result string) {
    switch chunkType {
    case gitdiff.Equal:
        return EqualPrefix
    case gitdiff.Delete:
        return DeletePrefix
    case gitdiff.Add:
        return AddPrefix
    default:
        panic(errors.Errorv("unknown chunk type", chunkType))
    }
}
