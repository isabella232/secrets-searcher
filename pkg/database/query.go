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

func (d *Database) GetRepos() (result Repos, err error) {
    var lines []string
    lines, err = d.readAll(RepoTable)
    if err != nil {
        err = errors.WithMessage(err, "unable to get repos")
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

func (d *Database) GetReposFiltered(repoFilter *structures.Filter) (result Repos, err error) {
    var repos Repos
    repos, err = d.GetRepos()
    if err != nil {
        err = errors.WithMessage(err, "unable to get filtered repos")
        return
    }

    for _, repo := range repos {
        if repoFilter.IsIncluded(repo.Name) {
            result = append(result, repo)
        }
    }

    return
}

func (d *Database) GetReposSorted() (result Repos, err error) {
    result, err = d.GetRepos()
    if err != nil {
        err = errors.WithMessage(err, "unable to get repos")
        return
    }

    d.SortRepos(result)

    return
}

func (d *Database) GetReposFilteredSorted(repoFilter *structures.Filter) (result Repos, err error) {
    result, err = d.GetReposFiltered(repoFilter)
    if err != nil {
        err = errors.WithMessage(err, "unable to get filtered repos")
        return
    }

    d.SortRepos(result)

    return
}

func (d *Database) GetRepoByName(name string) (result *Repo, err error) {
    var repos Repos
    repos, err = d.GetRepos()
    if err != nil {
        err = errors.WithMessage(err, "unable to get repos by name")
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

func (d *Database) SortRepos(objs Repos) {
    sort.Slice(objs, func(i, j int) bool { return strings.ToLower(objs[i].Name) < strings.ToLower(objs[j].Name) })
}

// Commit

func (d *Database) GetCommit(id string) (result *Commit, err error) {
    err = d.read(CommitTable, id, &result)
    return
}

func (d *Database) GetCommits() (result Commits, err error) {
    var lines []string
    lines, err = d.readAll(CommitTable)
    if err != nil {
        err = errors.WithMessage(err, "unable to get commits")
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

func (d *Database) GetCommitsSortedByDate() (result Commits, err error) {
    result, err = d.GetCommits()
    if err != nil {
        err = errors.WithMessage(err, "unable to get commits sorted by date")
        return
    }

    d.SortCommitsByDate(result)

    return
}

func (d *Database) SortCommitsByDate(objs Commits) {
    sort.Slice(objs, func(i, j int) bool { return objs[i].Date.Before(objs[j].Date) })
}

func (d *Database) WriteCommitIfNotExists(obj *Commit) (created bool, err error) {
    created, err = d.writeIfNotExists(CommitTable, obj.ID, obj)
    return
}

// Finding

func (d *Database) GetFinding(id string) (result *Finding, err error) {
    err = d.read(FindingTable, id, &result)
    return
}

func (d *Database) GetFindings() (result Findings, err error) {
    var lines []string
    lines, err = d.readAll(FindingTable)
    if err != nil {
        err = errors.WithMessage(err, "unable to read all findings")
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

func (d *Database) GetFindingsWithIDIndex() (result map[string]*Finding, err error) {
    var objs Findings
    objs, err = d.GetFindings()
    if err != nil {
        err = errors.WithMessage(err, "unable to get findings")
        return
    }

    result = map[string]*Finding{}
    for _, obj := range objs {
        result[obj.ID] = obj
    }

    return
}

func (d *Database) GetFindingsSortedGroupedBySecretID() (result FindingGroups, err error) {
    var objs Findings
    objs, err = d.GetFindings()
    if err != nil {
        err = errors.WithMessage(err, "unable to get findings")
        return
    }

    result = make(FindingGroups)

    for _, obj := range objs {
        result[obj.SecretID] = append(result[obj.SecretID], obj)
    }

    return
}

// Finding extras

func (d *Database) GetFindingExtra(id string) (result *FindingExtra, err error) {
    err = d.read(FindingExtrasTable, id, &result)
    return
}

func (d *Database) GetFindingExtras() (result FindingExtras, err error) {
    var lines []string
    lines, err = d.readAll(FindingExtrasTable)
    if err != nil {
        err = errors.WithMessage(err, "unable to get finding extras")
        return
    }

    for _, line := range lines {
        var obj *FindingExtra
        if err = json.Unmarshal([]byte(line), &obj); err != nil {
            return
        }

        result = append(result, obj)
    }

    return
}

func (d *Database) GetFindingExtrasSorted() (result FindingExtras, err error) {
    result, err = d.GetFindingExtras()
    if err != nil {
        err = errors.WithMessage(err, "unable to get finding extras")
        return
    }

    d.SortFindingExtras(result)

    return
}

func (d *Database) SortFindingExtras(objs FindingExtras) {
    sort.Slice(objs, func(i, j int) bool { return objs[i].Order < objs[j].Order })
}

func (d *Database) WriteFindingExtra(obj *FindingExtra) (err error) {
    err = d.write(FindingExtrasTable, obj.ID, obj)
    return
}

func (d *Database) GetFindingExtrasSortedGroupedByFindingID() (result FindingExtraGroups, err error) {
    var objs FindingExtras
    objs, err = d.GetFindingExtrasSorted()
    if err != nil {
        err = errors.WithMessage(err, "unable to get sorted finding extras")
        return
    }

    result = make(FindingExtraGroups)

    for _, obj := range objs {
        result[obj.FindingID] = append(result[obj.FindingID], obj)
    }

    return
}

// Secret

func (d *Database) GetSecret(id string) (result *Secret, err error) {
    err = d.read(SecretTable, id, &result)
    return
}

func (d *Database) GetSecrets() (result Secrets, err error) {
    var lines []string
    lines, err = d.readAll(SecretTable)
    if err != nil {
        err = errors.WithMessage(err, "unable to read all secrets")
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

func (d *Database) GetSecretsSorted() (result Secrets, err error) {
    result, err = d.GetSecrets()
    if err != nil {
        err = errors.WithMessage(err, "unable to get secrets")
        return
    }

    d.SortSecrets(result)

    return
}

func (d *Database) SortSecrets(objs Secrets) {
    sort.Slice(objs, func(i, j int) bool { return objs[i].ID < objs[j].ID })
}

func (d *Database) GetSecretsSortedIndexed() (result SecretIndex, err error) {
    var secrets Secrets
    secrets, err = d.GetSecretsSorted()
    if err != nil {
        err = errors.WithMessage(err, "unable to get sorted secrets")
        return
    }

    result = SecretIndex{}
    for _, secret := range secrets {
        result[secret.ID] = secret
    }

    return
}

func (d *Database) WriteSecret(obj *Secret) (err error) {
    err = d.write(SecretTable, obj.ID, obj)
    return
}

func (d *Database) WriteSecretIfNotExists(obj *Secret) (created bool, err error) {
    created, err = d.writeIfNotExists(SecretTable, obj.ID, obj)
    return
}

// Secret extras

func (d *Database) GetSecretExtra(id string) (result *SecretExtra, err error) {
    err = d.read(SecretExtrasTable, id, &result)
    return
}

func (d *Database) GetSecretExtras() (result SecretExtras, err error) {
    var lines []string
    lines, err = d.readAll(SecretExtrasTable)
    if err != nil {
        err = errors.WithMessage(err, "unable to get secret extras")
        return
    }

    for _, line := range lines {
        var obj *SecretExtra
        if err = json.Unmarshal([]byte(line), &obj); err != nil {
            return
        }

        result = append(result, obj)
    }

    return
}

func (d *Database) GetSecretExtrasSorted() (result SecretExtras, err error) {
    result, err = d.GetSecretExtras()
    if err != nil {
        err = errors.WithMessage(err, "unable to get sorted secret extras")
        return
    }

    d.SortSecretExtras(result)

    return
}

func (d *Database) SortSecretExtras(objs SecretExtras) {
    sort.Slice(objs, func(i, j int) bool {
        return objs[i].Order < objs[j].Order
    })
}

func (d *Database) WriteSecretExtra(obj *SecretExtra) (err error) {
    err = d.write(SecretExtrasTable, obj.ID, obj)
    return
}

func (d *Database) GetSecretExtrasSortedGroupedBySecretID() (result SecretExtraGroups, err error) {
    var objs SecretExtras
    objs, err = d.GetSecretExtrasSorted()
    if err != nil {
        err = errors.WithMessage(err, "unable to get sorted, grouped secret extras")
        return
    }

    result = make(SecretExtraGroups)

    for _, obj := range objs {
        result[obj.SecretID] = append(result[obj.SecretID], obj)
    }

    return
}

// RepoCommitsCache

func (d *Database) GetRepoCommitsCache(id string) (result *RepoCommitsCache, err error) {
    err = d.read(RepoCommitsCacheTable, id, &result)
    return
}

func (d *Database) WriteRepoCommitsCache(obj *RepoCommitsCache) (err error) {
    err = d.write(RepoCommitsCacheTable, obj.RepoName, obj)
    return
}

func (d *Database) DeleteRepoCommitsCache(id string) (err error) {
    err = d.delete(RepoCommitsCacheTable, id)
    return
}
