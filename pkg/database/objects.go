package database

import (
    "time"
)

const (
    CommitTable        = "commit"
    FindingTable       = "finding"
    RepoTable          = "repo"
    SecretTable        = "secret"
    SecretFindingTable = "secretfinding"
)

type (
    Commit struct {
        ID          string
        RepoID      string
        Commit      string
        CommitHash  string
        Date        time.Time
        AuthorFull  string
        AuthorEmail string
    }
    Finding struct {
        ID               string
        CommitID         string
        SecretID         string
        Processor        string
        Path             string
        StartLineNum     int
        StartIndex       int
        EndLineNum       int
        EndIndex         int
        StartDiffLineNum int
        EndDiffLineNum   int
        Code             string
        CodePadding      int
        Diff             string
        DiffPadding      int
    }
    Secret struct {
        ID           string
        Value        string
        ValueDecoded string
    }
    Repo struct {
        ID       string
        Name     string
        Owner    string
        FullName string
        SSHURL   string
        HTMLURL  string
        CloneDir string
    }
)
