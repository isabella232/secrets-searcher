package database

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/pantheon-systems/secrets-searcher/pkg/errors"
	"github.com/pantheon-systems/secrets-searcher/pkg/manip"
)

const (
	commitTable       = "commit"
	findingTable      = "finding"
	findingExtraTable = "finding-extra"
	repoTable         = "repo"
	secretTable       = "secret"
	secretExtraTable  = "secret-extra"
)

var searchTables = []string{
	commitTable,
	findingTable,
	findingExtraTable,
	secretTable,
	secretExtraTable,
}

func (d *Database) DeleteSearchTables() (err error) {
	return d.deleteTables(searchTables)
}

type ReportData struct {
	Secrets       Secrets
	Findings      Findings
	FindingExtras FindingExtras
	SecretExtras  SecretExtras
}

func (d *Database) GetBaseReportData() (result *ReportData, err error) {
	var (
		secrets       Secrets
		findings      Findings
		findingExtras FindingExtras
		secretExtras  SecretExtras
	)

	d.lockTables(searchTables)
	defer d.unlockTables(searchTables)

	// Secrets
	secrets, err = d.getSecrets(false)
	if err != nil {
		err = errors.WithMessage(err, "unable to get secrets")
		return
	}
	d.SortSecrets(secrets)

	// Findings
	findings, err = d.getFindings(false)
	if err != nil {
		err = errors.WithMessage(err, "unable to get findings")
		return
	}

	// Finding extras
	findingExtras, err = d.getFindingExtras(false)
	if err != nil {
		err = errors.WithMessage(err, "unable to get finding extras")
		return
	}
	d.SortFindingExtras(findingExtras)

	secretExtras, err = d.getSecretExtras(false)
	if err != nil {
		err = errors.WithMessage(err, "unable to get secret extras")
		return
	}
	d.SortSecretExtras(secretExtras)

	result = &ReportData{
		Secrets:       secrets,
		Findings:      findings,
		FindingExtras: findingExtras,
		SecretExtras:  secretExtras,
	}

	return
}

// Repo

func (d *Database) RepoTableExists() bool {
	return d.tableExists(repoTable)
}

func (d *Database) DeleteRepoTable() (err error) {
	return d.deleteTable(repoTable)
}

func (d *Database) GetRepo(id string) (result *Repo, err error) {
	err = d.read(repoTable, id, &result)
	return
}

func (d *Database) GetRepos() (result Repos, err error) {
	var lines []string
	lines, err = d.readAll(repoTable)
	if err != nil {
		err = errors.WithMessage(err, "unable to get repos")
		return
	}

	result = make(Repos, len(lines))
	for i, line := range lines {
		var obj *Repo
		if err = json.Unmarshal([]byte(line), &obj); err != nil {
			return
		}
		result[i] = obj
	}

	return
}

func (d *Database) GetReposFiltered(repoFilter manip.Filter) (result Repos, err error) {
	var repos Repos
	repos, err = d.GetRepos()
	if err != nil {
		err = errors.WithMessage(err, "unable to get filtered repos")
		return
	}

	for _, repo := range repos {
		if repoFilter.Includes(repo.Name) {
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

func (d *Database) GetReposFilteredSorted(repoFilter manip.Filter) (result Repos, err error) {
	result, err = d.GetReposFiltered(repoFilter)
	if err != nil {
		err = errors.WithMessage(err, "unable to get filtered sorted repos")
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
	err = d.write(repoTable, obj.ID, obj)
	return
}

func (d *Database) DeleteRepo(id string) (err error) {
	err = d.delete(repoTable, id)
	return
}

func (d *Database) SortRepos(objs Repos) {
	sort.Slice(objs, func(i, j int) bool { return strings.ToLower(objs[i].Name) < strings.ToLower(objs[j].Name) })
}

// Commit

func (d *Database) CommitTableExists() bool {
	return d.tableExists(commitTable)
}

func (d *Database) DeleteCommitTable() (err error) {
	return d.deleteTable(commitTable)
}

func (d *Database) GetCommit(id string) (result *Commit, err error) {
	err = d.read(commitTable, id, &result)
	return
}

func (d *Database) GetCommits() (result Commits, err error) {
	var lines []string
	lines, err = d.readAll(commitTable)
	if err != nil {
		err = errors.WithMessage(err, "unable to get commits")
		return
	}

	result = make(Commits, len(lines))
	for i, line := range lines {
		var obj *Commit
		if err = json.Unmarshal([]byte(line), &obj); err != nil {
			return
		}

		result[i] = obj
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
	created, err = d.writeIfNotExists(commitTable, obj.ID, obj)
	return
}

// Finding

func (d *Database) FindingTableExists() bool {
	return d.tableExists(findingTable)
}

func (d *Database) DeleteFindingTable() (err error) {
	return d.deleteTable(findingTable)
}

func (d *Database) GetFinding(id string) (result *Finding, err error) {
	err = d.read(findingTable, id, &result)
	return
}

func (d *Database) GetFindings() (result Findings, err error) {
	return d.getFindings(true)
}

func (d *Database) getFindings(lock bool) (result Findings, err error) {
	if lock {
		d.lockTable(findingTable)
		defer d.unlockTable(findingTable)
	}

	var lines []string
	lines, err = d.readAllUnsafe(findingTable)
	if err != nil {
		err = errors.WithMessage(err, "unable to read all findings")
		return
	}

	result = make(Findings, len(lines))
	for i, line := range lines {
		var obj *Finding
		if err = json.Unmarshal([]byte(line), &obj); err != nil {
			return
		}

		result[i] = obj
	}

	return
}

func (d *Database) WriteFinding(obj *Finding) (err error) {
	err = d.write(findingTable, obj.ID, obj)
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

func (d *Database) FindingExtraTableExists() bool {
	return d.tableExists(findingExtraTable)
}

func (d *Database) DeleteFindingExtraTable() (err error) {
	return d.deleteTable(findingExtraTable)
}

func (d *Database) GetFindingExtra(id string) (result *FindingExtra, err error) {
	err = d.read(findingExtraTable, id, &result)
	return
}

func (d *Database) GetFindingExtras() (result FindingExtras, err error) {
	return d.getFindingExtras(true)
}

func (d *Database) getFindingExtras(lock bool) (result FindingExtras, err error) {
	if lock {
		d.lockTable(findingExtraTable)
		defer d.unlockTable(findingExtraTable)
	}

	var lines []string
	lines, err = d.readAllUnsafe(findingExtraTable)
	if err != nil {
		err = errors.WithMessage(err, "unable to get finding extras")
		return
	}

	result = make(FindingExtras, len(lines))
	for i, line := range lines {
		var obj *FindingExtra
		if err = json.Unmarshal([]byte(line), &obj); err != nil {
			return
		}

		result[i] = obj
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
	err = d.write(findingExtraTable, obj.ID, obj)
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

func (d *Database) SecretTableExists() bool {
	return d.tableExists(secretTable)
}

func (d *Database) DeleteSecretTable() (err error) {
	return d.deleteTable(secretTable)
}

func (d *Database) GetSecret(id string) (result *Secret, err error) {
	err = d.read(secretTable, id, &result)
	return
}

func (d *Database) GetSecrets() (result Secrets, err error) {
	return d.getSecrets(true)
}

func (d *Database) getSecrets(lock bool) (result Secrets, err error) {
	if lock {
		d.lockTable(secretTable)
		defer d.unlockTable(secretTable)
	}

	var lines []string
	lines, err = d.readAllUnsafe(secretTable)
	if err != nil {
		err = errors.WithMessage(err, "unable to read all secrets")
		return
	}

	result = make(Secrets, len(lines))
	for i, line := range lines {
		var obj *Secret
		if err = json.Unmarshal([]byte(line), &obj); err != nil {
			return
		}

		result[i] = obj
	}

	d.SortSecrets(result)

	return
}

func (d *Database) SortSecrets(objs Secrets) {
	sort.Slice(objs, func(i, j int) bool { return objs[i].ID < objs[j].ID })
}

func (d *Database) WriteSecret(obj *Secret) (err error) {
	err = d.write(secretTable, obj.ID, obj)
	return
}

func (d *Database) WriteSecretIfNotExists(obj *Secret) (created bool, err error) {
	created, err = d.writeIfNotExists(secretTable, obj.ID, obj)
	return
}

// Secret extras

func (d *Database) SecretExtraTableExists() bool {
	return d.tableExists(secretExtraTable)
}

func (d *Database) DeleteSecretExtraTable() (err error) {
	return d.deleteTable(secretExtraTable)
}

func (d *Database) GetSecretExtra(id string) (result *SecretExtra, err error) {
	err = d.read(secretExtraTable, id, &result)
	return
}

func (d *Database) GetSecretExtras() (result SecretExtras, err error) {
	return d.getSecretExtras(true)
}

func (d *Database) getSecretExtras(lock bool) (result SecretExtras, err error) {
	if lock {
		d.lockTable(secretExtraTable)
		defer d.unlockTable(secretExtraTable)
	}

	var lines []string
	lines, err = d.readAllUnsafe(secretExtraTable)
	if err != nil {
		err = errors.WithMessage(err, "unable to get secret extras")
		return
	}

	result = make(SecretExtras, len(lines))
	for i, line := range lines {
		var obj *SecretExtra
		if err = json.Unmarshal([]byte(line), &obj); err != nil {
			return
		}

		result[i] = obj
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
	err = d.write(secretExtraTable, obj.ID, obj)
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
