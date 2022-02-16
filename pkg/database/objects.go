package database

import (
	"time"
)

type (

	// Commit
	Commit struct {
		ID          string
		RepoID      string
		Commit      string
		CommitHash  string
		Date        time.Time
		AuthorName  string
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
		BeforeCode   string
		Code         string
		AfterCode    string
		FileBasename string
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
		Debug     bool
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
		ID        string
		SecretID  string
		FindingID string
		Order     int
		Key       string
		Header    string
		Value     string
		Code      bool
		URL       string
		Debug     bool
	}
	SecretExtras      []*SecretExtra
	SecretExtraGroups map[string]SecretExtras

	// Repo
	Repo struct {
		ID             string
		Name           string
		SourceProvider string
		RemoteURL      string
	}
	Repos      []*Repo
	RepoGroups map[string]Repos
)
