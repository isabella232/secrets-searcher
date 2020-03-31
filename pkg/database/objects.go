package database

import (
    "github.com/pantheon-systems/search-secrets/pkg/database/enum/decision"
    "time"
)

const (
    CommitTable   = "commit"
    DecisionTable = "decision"
    FindingTable  = "finding"
    RepoTable     = "repo"
    SecretTable   = "secret"
)

type (
    Commit struct {
        ID          string
        RepoID      string
        Commit      string
        CommitHash  string
        Date        time.Time
        AuthorEmail string
    }
    Decision struct {
        ID            string
        FindingID     string
        SecretID      string
        Decision      decision.DecisionEnum `json:"-"`
        DecisionValue string
    }
    Finding struct {
        ID               string
        CommitID         string
        Rule             string
        Path             string
        StartLineNum     int
        StartIndex       int
        EndLineNum       int
        EndIndex         int
        Code             string
        Diff             string
        SecretsProcessed bool
    }
    Secret struct {
        ID    string
        Value string
    }
    Repo struct {
        ID       string
        Name     string
        Owner    string
        FullName string
        SSHURL   string
        HTMLURL  string
    }
)
