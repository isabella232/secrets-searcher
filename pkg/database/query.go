package database

import (
    "encoding/json"
    "github.com/pantheon-systems/search-secrets/pkg/database/enum/decision"
    "github.com/pantheon-systems/search-secrets/pkg/database/enum/reason"
    "github.com/pantheon-systems/search-secrets/pkg/errors"
)

func (d *Database) GetRepos() (repos []*Repo, err error) {
    lines, err := d.ReadAll(RepoTable)
    if err != nil {
        return
    }

    for _, line := range lines {
        var r *Repo
        if err = json.Unmarshal([]byte(line), &r); err != nil {
            return nil, errors.Wrapv(err, "unable to unmarshal json", line)
        }
        repos = append(repos, r)
    }

    return
}

func (d *Database) GetFindings() (findings []*Finding, err error) {
    lines, err := d.ReadAll(FindingTable)
    if err != nil {
        return
    }

    for _, line := range lines {
        var finding *Finding
        if err = json.Unmarshal([]byte(line), &finding); err != nil {
            return nil, errors.Wrapv(err, "unable to unmarshal json", line)
        }

        finding.Reason = reason.NewReasonFromValue(finding.ReasonValue)
        if finding.Reason == nil {
            err = errors.Errorv("unknown finding reason", finding.ReasonValue)
            return
        }

        findings = append(findings, finding)
    }

    return
}

func (d *Database) GetSecrets() (secrets []*Secret, err error) {
    lines, err := d.ReadAll(SecretTable)
    if err != nil {
        return
    }

    for _, line := range lines {
        var r *Secret
        if err = json.Unmarshal([]byte(line), &r); err != nil {
            return nil, errors.Wrapv(err, "unable to unmarshal json", line)
        }
    }

    return
}

func (d *Database) GetDecisions() (secrets []*Decision, err error) {
    lines, err := d.ReadAll(DecisionTable)
    if err != nil {
        return
    }

    for _, line := range lines {
        var r *Decision
        if err = json.Unmarshal([]byte(line), &r); err != nil {
            return nil, errors.Wrapv(err, "unable to unmarshal json", line)
        }

        r.Decision = decision.NewDecisionFromValue(r.DecisionValue)
        if r.Decision == nil {
            err = errors.Errorv("unknown finding reason", r.DecisionValue)
            return
        }
    }

    return
}

func (d *Database) GetFindingStringsForFinding(finding *Finding) (findingStrings []*FindingString, err error) {
    strings, err := d.ReadAll(FindingStringTable)
    if err != nil {
        return
    }

    for _, line := range strings {
        var findingString *FindingString
        if err = json.Unmarshal([]byte(line), &findingString); err != nil {
            return nil, errors.Wrapv(err, "unable to unmarshal json", line)
        }
        if findingString.FindingID == finding.ID {
            findingStrings = append(findingStrings, findingString)
        }
    }

    return
}
