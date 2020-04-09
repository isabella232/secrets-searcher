package database

import (
    "encoding/json"
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

func (d *Database) DeleteRepo(id string) (err error) {
    err = d.delete(RepoTable, id)
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

func (d *Database) GetSecretsWithIDIndex() (result map[string]*Secret, err error) {
    var secrets []*Secret
    secrets, err = d.GetSecrets()
    if err != nil {
        return
    }

    result = map[string]*Secret{}
    for _, secret := range secrets {
        result[secret.ID] = secret
    }

    return
}

func (d *Database) WriteSecret(obj *Secret) (err error) {
    err = d.write(SecretTable, obj.ID, obj)
    return
}

// SecretFinding

func (d *Database) GetSecretFinding(id string) (result *SecretFinding, err error) {
    if err = d.read(SecretFindingTable, id, &result); err != nil {
        return
    }

    return
}

func (d *Database) GetSecretFindings() (result []*SecretFinding, err error) {
    lines, err := d.readAll(SecretFindingTable)
    if err != nil {
        return
    }

    for _, line := range lines {
        var obj *SecretFinding
        if err = json.Unmarshal([]byte(line), &obj); err != nil {
            return
        }
        result = append(result, obj)
    }

    return
}

func (d *Database) GetSecretFindingsGroupedBySecret() (result map[*Secret][]*SecretFinding, err error) {
    var secretIndex map[string]*Secret
    secretIndex, err = d.GetSecretsWithIDIndex()
    if err != nil {
        return
    }

    var sfs []*SecretFinding
    sfs, err = d.GetSecretFindings()
    if err != nil {
        return
    }

    result = make(map[*Secret][]*SecretFinding)

    for _, sf := range sfs {
        secret, ok := secretIndex[sf.SecretID]
        if !ok {
            err = errors.Errorv("no secret found for secret ID", sf.SecretID)
            return
        }
        result[secret] = append(result[secret], sf)
    }

    return
}

func (d *Database) WriteSecretFinding(obj *SecretFinding) (err error) {
    err = d.write(SecretFindingTable, obj.ID, obj)
    return
}
