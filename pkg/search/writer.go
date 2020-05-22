package search

import (
	"github.com/pantheon-systems/search-secrets/pkg/database"
	"github.com/pantheon-systems/search-secrets/pkg/errors"
	gitpkg "github.com/pantheon-systems/search-secrets/pkg/git"
	"github.com/pantheon-systems/search-secrets/pkg/logg"
	"github.com/pantheon-systems/search-secrets/pkg/manip"
	"github.com/pantheon-systems/search-secrets/pkg/search/contract"
)

const contextLenLimit = 50

type dbResultWriter struct {
	secretTracker manip.Set
	db            *database.Database
	log           logg.Logg
}

func NewDBResultWriter(db *database.Database, log logg.Logg) *dbResultWriter {
	return &dbResultWriter{
		secretTracker: manip.NewEmptyBasicSet(),
		db:            db,
		log:           log,
	}
}

func (w *dbResultWriter) WriteResult(result *contract.JobResult) (err error) {
	var dbCommit *database.Commit
	var dbSecret *database.Secret
	var dbFinding *database.Finding
	var dbSecretExtras database.SecretExtras
	var dbFindingExtras database.FindingExtras

	if dbCommit, dbFinding, dbSecret, dbSecretExtras, dbFindingExtras, err = w.buildDBObjects(result); err != nil {
		return
	}

	// Write commit
	if _, err = w.db.WriteCommitIfNotExists(dbCommit); err != nil {
		return
	}

	// Write secret
	var created bool
	if created, err = w.db.WriteSecretIfNotExists(dbSecret); err != nil {
		return
	}
	if created {
		w.secretTracker.Add(dbSecret.ID)
	}

	// Write finding
	if err = w.db.WriteFinding(dbFinding); err != nil {
		return
	}

	// Write secret extras
	for _, dbSecretExtra := range dbSecretExtras {
		if err = w.db.WriteSecretExtra(dbSecretExtra); err != nil {
			return
		}
	}

	// Write finding extras
	for _, dbFindingExtra := range dbFindingExtras {
		if err = w.db.WriteFindingExtra(dbFindingExtra); err != nil {
			return
		}
	}

	return
}

func (w *dbResultWriter) buildDBObjects(jobResult *contract.JobResult) (dbCommit *database.Commit, dbFinding *database.Finding, dbSecret *database.Secret, dbSecretExtras database.SecretExtras, dbFindingExtras database.FindingExtras, err error) {

	// Commit
	dbCommit = w.buildDBCommit(jobResult.FileChange.Commit, jobResult.RepoID)

	// Secret
	dbSecret = w.buildDBSecret(jobResult.SecretValue)

	// Finding
	dbFinding, err = w.buildDBFinding(jobResult, dbSecret.ID, dbCommit.ID)
	if err != nil {
		err = errors.WithMessage(err, "unable to build finding object for database")
		return
	}

	// Secret extras
	for i, secretExtra := range jobResult.SecretExtras {
		dbSecretExtras = append(dbSecretExtras, w.buildDBSecretExtra(secretExtra, dbSecret.ID, dbFinding.ID, i))
	}

	// Finding extras
	for i, findingExtra := range jobResult.FindingExtras {
		dbFindingExtras = append(dbFindingExtras, w.buildDBFindingExtra(findingExtra, dbFinding.ID, i))
	}

	return
}

func (w *dbResultWriter) buildDBCommit(commit *gitpkg.Commit, repoID string) *database.Commit {
	return &database.Commit{
		ID:          database.CreateHashID(repoID, commit.Hash),
		RepoID:      repoID,
		Commit:      commit.Message,
		CommitHash:  commit.Hash,
		Date:        commit.Date,
		AuthorName:  commit.AuthorName,
		AuthorEmail: commit.AuthorEmail,
	}
}

func (w *dbResultWriter) buildDBSecret(secretValue string) *database.Secret {
	return &database.Secret{
		ID:    database.CreateHashID(secretValue),
		Value: secretValue,
	}
}

func (w *dbResultWriter) buildDBFinding(jobResult *contract.JobResult, secretID, commitID string) (result *database.Finding, err error) {
	var fileContents string
	fileContents, err = jobResult.FileChange.FileContents()
	if err != nil {
		err = errors.WithMessagev(err, "unable to get file contents for path", jobResult.FileChange.Path)
		return
	}

	codeLineRange := manip.NewLineRangeFromFileRange(jobResult.FileRange, fileContents)

	// Code context
	var contextLineRange *manip.LineRange
	if jobResult.ContextFileRange != nil {
		contextLineRange = manip.NewLineRangeFromFileRange(jobResult.ContextFileRange, fileContents)
	}
	beforeCodeValue, afterCodeValue := manip.CodeContext(fileContents, codeLineRange, contextLineRange, contextLenLimit)
	beforeCode := beforeCodeValue.ExtractValue(fileContents).Value
	afterCode := afterCodeValue.ExtractValue(fileContents).Value
	code := codeLineRange.ExtractValue(fileContents).Value

	result = &database.Finding{
		ID: database.CreateHashID(
			commitID,
			jobResult.Processor.GetName(),
			jobResult.FileChange.Path,
			jobResult.FileRange.StartLineNum,
			jobResult.FileRange.StartIndex,
			jobResult.FileRange.EndLineNum,
			jobResult.FileRange.EndIndex,
		),
		CommitID:     commitID,
		SecretID:     secretID,
		Processor:    jobResult.Processor.GetName(),
		Path:         jobResult.FileChange.Path,
		StartLineNum: jobResult.FileRange.StartLineNum,
		StartIndex:   jobResult.FileRange.StartIndex,
		EndLineNum:   jobResult.FileRange.EndLineNum,
		EndIndex:     jobResult.FileRange.EndIndex,
		BeforeCode:   beforeCode,
		Code:         code,
		AfterCode:    afterCode,
		FileBasename: jobResult.FileBaseName,
	}

	return
}

func (w *dbResultWriter) buildDBSecretExtra(extra *contract.ResultExtra, secretID, findingID string, order int) *database.SecretExtra {
	return &database.SecretExtra{
		ID:        database.CreateHashID(secretID, extra.Key, order),
		SecretID:  secretID,
		FindingID: findingID,
		Order:     order,
		Key:       extra.Key,
		Header:    extra.Header,
		Value:     extra.Value,
		Code:      extra.Code,
		URL:       extra.URL,
		Debug:     extra.Debug,
	}
}

func (w *dbResultWriter) buildDBFindingExtra(extra *contract.ResultExtra, findingID string, order int) *database.FindingExtra {
	return &database.FindingExtra{
		ID:        database.CreateHashID(findingID, extra.Key, order),
		FindingID: findingID,
		Order:     order,
		Key:       extra.Key,
		Header:    extra.Header,
		Value:     extra.Value,
		Code:      extra.Code,
		URL:       extra.URL,
		Debug:     extra.Debug,
	}
}
