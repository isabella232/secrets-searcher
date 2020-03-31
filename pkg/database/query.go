package database

import (
    "encoding/json"
    "github.com/pantheon-systems/search-secrets/pkg/database/enum/decision"
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    "github.com/pantheon-systems/search-secrets/pkg/structures"
)

// Repo

func (d *Database) GetRepo(id string) (result *Repo, err error) {
    err = d.read(RepoTable, id, &result)
    return
}

func (d *Database) GetRepos() (result []*Repo, err error) {
    lines, err := d.readAll(RepoTable)
    if err != nil {
        return
    }

    for _, line := range lines {
        var obj *Repo
        if err = json.Unmarshal([]byte(line), &obj); err != nil {
            return
        }
        result = append(result, obj)
    }

    return
}

func (d *Database) GetReposFiltered(repoFilter *structures.Filter) (result []*Repo, err error) {
    repos, err := d.GetRepos()
    if err != nil {
        return
    }

    for _, repo := range repos {
        if repoFilter.IsIncluded(repo.Name) {
            result = append(result, repo)
        }
    }

    return
}

func (d *Database) GetRepoByName(name string) (result *Repo, err error) {
    var repos []*Repo
    repos, err = d.GetRepos()
    if err != nil {
        return
    }

    for _, repo := range repos {
        if repo.Name == name {
            result = repo
            return
        }
    }

    return
}

func (d *Database) WriteRepo(obj *Repo) (err error) {
    err = d.write(RepoTable, obj.ID, obj)
    return
}

// Commit

func (d *Database) GetCommit(id string) (result *Commit, err error) {
    err = d.read(CommitTable, id, &result)
    return
}

func (d *Database) GetCommits() (result []*Commit, err error) {
    lines, err := d.readAll(CommitTable)
    if err != nil {
        return
    }

    for _, line := range lines {
        var commit *Commit
        if err = json.Unmarshal([]byte(line), &commit); err != nil {
            return
        }

        result = append(result, commit)
    }

    return
}

func (d *Database) WriteCommit(obj *Commit) (err error) {
    err = d.write(CommitTable, obj.ID, obj)
    return
}

// Finding

func (d *Database) GetFinding(id string) (result *Finding, err error) {
    err = d.read(FindingTable, id, &result)
    return
}

func (d *Database) GetFindings() (result []*Finding, err error) {
    lines, err := d.readAll(FindingTable)
    if err != nil {
        return
    }

    for _, line := range lines {
        var obj *Finding
        if err = json.Unmarshal([]byte(line), &obj); err != nil {
            return
        }

        result = append(result, obj)
    }

    return
}

func (d *Database) WriteFinding(obj *Finding) (err error) {
    err = d.write(FindingTable, obj.ID, obj)
    return
}

// Secret

func (d *Database) GetSecret(id string) (result *Secret, err error) {
    err = d.read(SecretTable, id, &result)
    return
}

func (d *Database) GetSecrets() (result []*Secret, err error) {
    lines, err := d.readAll(SecretTable)
    if err != nil {
        return
    }

    for _, line := range lines {
        var obj *Secret
        if err = json.Unmarshal([]byte(line), &obj); err != nil {
            return
        }

        result = append(result, obj)
    }

    return
}

func (d *Database) WriteSecret(obj *Secret) (err error) {
    err = d.write(SecretTable, obj.ID, obj)
    return
}

// Decision

func (d *Database) GetDecision(id string) (result *Decision, err error) {
    if err = d.read(DecisionTable, id, &result); err != nil {
        return
    }

    result.Decision = decision.NewDecisionFromValue(result.DecisionValue)
    if result.Decision == nil {
        err = errors.Errorv("unknown decision", result.DecisionValue)
        return
    }

    return
}

func (d *Database) GetDecisions() (result []*Decision, err error) {
    lines, err := d.readAll(DecisionTable)
    if err != nil {
        return
    }

    for _, line := range lines {
        var obj *Decision
        if err = json.Unmarshal([]byte(line), &obj); err != nil {
            return
        }

        obj.Decision = decision.NewDecisionFromValue(obj.DecisionValue)
        if obj.Decision == nil {
            err = errors.Errorv("unknown decision", obj.DecisionValue)
            return
        }

        result = append(result, obj)
    }

    return
}

func (d *Database) GetDecisionsForSecret(secret *Secret) (result []*Decision, err error) {
    decs, err := d.GetDecisions()
    if err != nil {
        return
    }

    for _, obj := range decs {
        if obj.SecretID == secret.ID {
            result = append(result, obj)
        }
    }

    return
}

func (d *Database) WriteDecision(obj *Decision) (err error) {
    obj.DecisionValue = obj.Decision.Value()
    err = d.write(DecisionTable, obj.ID, obj)
    return
}
