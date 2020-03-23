package database

import (
	"github.com/pantheon-systems/search-secrets/pkg/database/enum/decision"
	"github.com/pantheon-systems/search-secrets/pkg/database/enum/reason"
)

const FindingTable = "finding"
const FindingStringTable = "finding-string"
const DecisionTable = "decision"
const FindingSecretTable = "finding-secret"
const SecretTable = "secret"
const RepoTable = "repo"

type (
	Finding struct {
		ID          string
		RepoID      string
		Branch      string
		Commit      string
		CommitHash  string
		Diff        string
		Path        string
		PrintDiff   string
		ReasonValue string
		Reason      reason.ReasonEnum `json:"-"`
	}
	FindingString struct {
		ID        string
		FindingID string
		String    string
		Index     int
		StartLine int
	}
	Decision struct {
		ID              string
		FindingStringID string
		SecretID        string
		DecisionValue   string
		Decision        decision.DecisionEnum `json:"-"`
	}
	Secret struct {
		ID    string
		Value string
	}
	Repo struct {
		ID     string
		Name   string
		SSHURL string
	}
)
