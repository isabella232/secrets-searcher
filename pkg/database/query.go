package database

import (
    "encoding/json"
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    "github.com/pantheon-systems/search-secrets/pkg/structures"
    "sort"
    "strings"
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
    var repos []*Repo
    repos, err = d.GetRepos()
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

func (d *Database) GetReposSorted() (result []*Repo, err error) {
    result, err = d.GetRepos()
    if err != nil {
        return
    }

    d.sortRepos(result)

    return
}

func (d *Database) GetReposFilteredSorted(repoFilter *structures.Filter) (result []*Repo, err error) {
    result, err = d.GetReposFiltered(repoFilter)
    if err != nil {
        return
    }

    d.sortRepos(result)

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

func (d *Database) sortRepos(objs []*Repo) {
    sort.Slice(objs, func(i, j int) bool {
        return strings.ToLower(objs[i].Name) < strings.ToLower(objs[j].Name)
    })
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

func (d *Database) WriteCommitIfNotExists(obj *Commit) (err error) {
    var exists bool
    exists, err = d.exists(CommitTable, obj.ID)
    if err != nil {
        return
    }

    if !exists {
        err = d.write(CommitTable, obj.ID, obj)
    }

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

func (d *Database) GetFindingsGroupedBySecret() (result map[*Secret][]*Finding, err error) {
    var secretIndex map[string]*Secret
    secretIndex, err = d.GetSecretsWithIDIndex()
    if err != nil {
        return
    }

    var findings []*Finding
    findings, err = d.GetFindings()
    if err != nil {
        return
    }

    result = make(map[*Secret][]*Finding)

    for _, finding := range findings {
        secret, ok := secretIndex[finding.SecretID]
        if !ok {
            err = errors.Errorv("no secret found for secret ID", finding.SecretID)
            return
        }
        result[secret] = append(result[secret], finding)
    }

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

func (d *Database) GetSecretsSorted() (result []*Secret, err error) {
    result, err = d.GetSecrets()
    if err != nil {
        return
    }

    d.SortSecrets(result)

    return
}

func (d *Database) SortSecrets(objs []*Secret) {
    sort.Slice(objs, func(i, j int) bool { return objs[i].ID < objs[j].ID })
}

func (d *Database) GetSecretsWithIDIndex() (result map[string]*Secret, err error) {
    var secrets []*Secret
    secrets, err = d.GetSecretsSorted()
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

func (d *Database) WriteSecretIfNotExists(obj *Secret) (err error) {
    var exists bool
    exists, err = d.exists(SecretTable, obj.ID)
    if err != nil {
        return
    }

    if !exists {
        err = d.write(SecretTable, obj.ID, obj)
    }

    return
}
