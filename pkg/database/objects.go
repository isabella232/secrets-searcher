package database

import (
    "time"
)

const (
    CommitTable        = "commit"
    FindingTable       = "finding"
    FindingExtrasTable = "finding-extras"
    RepoTable          = "repo"
    SecretTable        = "secret"
    SecretExtrasTable  = "secret-extras"
)

type (

    // Commit
    Commit struct {
        ID          string
        RepoID      string
        Commit      string
        CommitHash  string
        Date        time.Time
        AuthorFull  string
        AuthorEmail string
    }
    Commits      []*Commit
    CommitGroups map[string]Commits

    // Finding
    Finding struct {
        ID           string
        CommitID     string
        SecretID     string
        Processor    string
        Path         string
        StartLineNum int
        StartIndex   int
        EndLineNum   int
        EndIndex     int
        Code         string
        CodePadding  int
    }
    Findings      []*Finding
    FindingGroups map[string]Findings

    // FindingExtra
    FindingExtra struct {
        ID        string
        FindingID string
        Order     int
        Key       string
        Header    string
        Value     string
        Code      bool
        URL       string
    }
    FindingExtras      []*FindingExtra
    FindingExtraGroups map[string]FindingExtras

    // Secret
    Secret struct {
        ID    string
        Value string
    }
    Secrets     []*Secret
    SecretIndex map[string]*Secret

    // SecretExtra
    SecretExtra struct {
        ID       string
        SecretID string
        Order    int
        Key      string
        Header   string
        Value    string
        Code     bool
        URL      string
    }
    SecretExtras      []*SecretExtra
    SecretExtraGroups map[string]SecretExtras

    // Repo
    Repo struct {
        ID             string
        Name           string
        Owner          string
        SourceProvider string
        FullName       string
        SSHURL         string
        HTMLURL        string
        CloneDir       string
    }
    Repos      []*Repo
    RepoGroups map[string]Repos
)
