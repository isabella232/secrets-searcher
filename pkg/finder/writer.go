package finder

import (
    "github.com/pantheon-systems/search-secrets/pkg/database"
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    gitpkg "github.com/pantheon-systems/search-secrets/pkg/git"
    "github.com/pantheon-systems/search-secrets/pkg/structures"
    "github.com/sirupsen/logrus"
    "strings"
)

type (
    Writer struct {
        whitelistSecretIDSet structures.Set
        secretTracker        structures.Set
        db                   *database.Database
        log                  *logrus.Entry
    }
)

func newWriter(whitelistSecretIDSet structures.Set, db *database.Database, log *logrus.Entry) *Writer {
    return &Writer{
        whitelistSecretIDSet: whitelistSecretIDSet,
        secretTracker:        structures.NewSet(nil),
        db:                   db,
        log:                  log,
    }
}

func (f *Writer) persistResult(result *workerResult) (err error) {
    dbCommit, dbFindings, dbSecrets, dbSecretExtras, dbFindingExtras, ok := f.buildDBObjects(result)
    if !ok {
        return
    }

    if _, err = f.db.WriteCommitIfNotExists(dbCommit); err != nil {
        return
    }
    for _, dbSecret := range dbSecrets {
        var created bool
        if created, err = f.db.WriteSecretIfNotExists(dbSecret); err != nil {
            return
        }
        if created {
            f.secretTracker.Add(dbSecret.ID)
        }
    }
    for _, dbFinding := range dbFindings {
        if err = f.db.WriteFinding(dbFinding); err != nil {
            return
        }
    }
    for _, dbSecretExtra := range dbSecretExtras {
        if err = f.db.WriteSecretExtra(dbSecretExtra); err != nil {
            return
        }
    }
    for _, dbFindingExtra := range dbFindingExtras {
        if err = f.db.WriteFindingExtra(dbFindingExtra); err != nil {
            return
        }
    }

    return
}

func (f *Writer) buildDBObjects(result *workerResult) (dbCommit *database.Commit, dbFindings []*database.Finding, dbSecrets []*database.Secret, dbSecretExtras database.SecretExtras, dbFindingExtras database.FindingExtras, ok bool) {
    var commit *database.Commit
    var secrets []*database.Secret
    var secretExtras database.SecretExtras
    var findings []*database.Finding
    var findingExtras database.FindingExtras

    commit = f.buildDBCommit(result.Commit, result.RepoID)

    log := f.log.WithFields(logrus.Fields{
        "repo":       commit.RepoID,
        "commitHash": commit.CommitHash,
    })

    for _, findingResult := range result.FindingResults {
        for _, finding := range findingResult.Findings {
            dbSecret := f.buildDBSecret(finding.Secret)

            // Check whitelist
            if f.whitelistSecretIDSet.Contains(dbSecret.ID) {
                log.WithField("secret", dbSecret.ID).Debug("secret whitelisted by ID, skipping finding")
                continue
            }

            var dbSecretExtras database.SecretExtras
            for i, secretExtra := range finding.SecretExtras {
                dbSecretExtras = append(dbSecretExtras, f.buildDBSecretExtra(secretExtra, dbSecret.ID, i))
            }

            dbFinding, findingErr := f.buildDBFinding(finding, result.Commit, findingResult.FileChange, dbSecret.ID, commit.ID)
            if findingErr != nil {
                errors.ErrorLogForEntry(log, findingErr).Error("unable to build finding object for database")
                continue
            }

            var dbFindingExtras database.FindingExtras
            for i, findingExtra := range finding.FindingExtras {
                dbFindingExtras = append(dbFindingExtras, f.buildDBFindingExtra(findingExtra, dbFinding.ID, i))
            }

            secrets = append(secrets, dbSecret)
            findings = append(findings, dbFinding)
            secretExtras = append(secretExtras, dbSecretExtras...)
            findingExtras = append(findingExtras, dbFindingExtras...)
        }
    }

    if findings != nil {
        dbCommit = commit
        dbFindings = findings
        dbSecrets = secrets
        dbSecretExtras = secretExtras
        dbFindingExtras = findingExtras
        ok = true
    }

    return
}

func (f *Writer) buildDBCommit(commit *gitpkg.Commit, repoID string) *database.Commit {
    return &database.Commit{
        ID:          database.CreateHashID(repoID, commit.Hash),
        RepoID:      repoID,
        Commit:      commit.Message,
        CommitHash:  commit.Hash,
        Date:        commit.Time,
        AuthorFull:  commit.AuthorFull,
        AuthorEmail: commit.AuthorEmail,
    }
}

func (f *Writer) buildDBSecret(secret *Secret) *database.Secret {
    return &database.Secret{
        ID:    database.CreateHashID(secret.Value),
        Value: secret.Value,
    }
}

func (f *Writer) buildDBSecretExtra(extra *Extra, secretID string, order int) *database.SecretExtra {
    return &database.SecretExtra{
        ID:       database.CreateHashID(secretID, extra.Key, order),
        SecretID: secretID,
        Order:    order,
        Key:      extra.Key,
        Header:   extra.Header,
        Value:    extra.Value,
        Code:     extra.Code,
        URL:      extra.URL,
    }
}

func (f *Writer) buildDBFindingExtra(extra *Extra, findingID string, order int) *database.FindingExtra {
    return &database.FindingExtra{
        ID:        database.CreateHashID(findingID, extra.Key, order),
        FindingID: findingID,
        Order:     order,
        Key:       extra.Key,
        Header:    extra.Header,
        Value:     extra.Value,
        Code:      extra.Code,
        URL:       extra.URL,
    }
}

func (f *Writer) buildDBFinding(finding *Finding, commit *gitpkg.Commit, fileChange *gitpkg.FileChange, secretID, commitID string) (result *database.Finding, err error) {
    var fileContents string
    fileContents, err = commit.FileContents(fileChange.Path)
    if err != nil {
        return
    }

    // Get code and diff
    code := getExcerpt(fileContents, finding.FileRange.StartLineNum, finding.FileRange.EndLineNum)
    const maxLength = 2000
    if len(code) > maxLength {
        code = code[:maxLength] + " [...]"
    }

    result = &database.Finding{
        ID: database.CreateHashID(
            commitID,
            finding.ProcessorName,
            fileChange.Path,
            finding.FileRange.StartLineNum,
            finding.FileRange.StartIndex,
            finding.FileRange.EndLineNum,
            finding.FileRange.EndIndex,
        ),
        CommitID:     commitID,
        SecretID:     secretID,
        Processor:    finding.ProcessorName,
        Path:         fileChange.Path,
        StartLineNum: finding.FileRange.StartLineNum,
        StartIndex:   finding.FileRange.StartIndex,
        EndLineNum:   finding.FileRange.EndLineNum,
        EndIndex:     finding.FileRange.EndIndex,
        Code:         code,
    }

    return
}

func getExcerpt(contents string, fromLineNum int, toLineNum int) (result string) {
    lineNum := 1
    theRest := contents
    for {
        index := strings.Index(theRest, "\n")
        if index == -1 {
            result += theRest
            return
        }
        if lineNum >= fromLineNum {
            result += theRest[:index+1]
        }
        theRest = theRest[index+1:]
        lineNum += 1
        if lineNum == toLineNum+1 {
            return
        }
    }
}
