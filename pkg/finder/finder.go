package finder

import (
    "github.com/pantheon-systems/search-secrets/pkg/database"
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    "github.com/pantheon-systems/search-secrets/pkg/git"
    "github.com/pantheon-systems/search-secrets/pkg/interact"
    "github.com/pantheon-systems/search-secrets/pkg/structures"
    "github.com/sirsean/go-pool"
    "github.com/sirupsen/logrus"
    "runtime"
    "strings"
    "time"
)

type (
    Finder struct {
        repoFilter           *structures.Filter
        refFilter            *structures.Filter
        processors           []Processor
        earliestTime         time.Time
        latestTime           time.Time
        whitelistPath        structures.RegexpSet
        whitelistSecretIDSet structures.Set
        secretTracker        structures.Set
        interact             interact.Interactish
        db                   *database.Database
        log                  *logrus.Logger
    }
    Processor interface {
        FindInFileChange(fileChange *git.FileChange, log *logrus.Entry) (result []*Finding, ignore []*structures.FileRange, err error)
        FindInLine(line string, log *logrus.Entry) (result []*FindingInLine, ignore []*structures.LineRange, err error)
        Name() string
    }
    Finding struct {
        ProcessorName string
        FileRange     *structures.FileRange
        Secret        *Secret
    }
    FindingInLine struct {
        ProcessorName string
        LineRange     *structures.LineRange
        Secret        *Secret
    }
    Secret struct {
        Value   string
        Decoded string
    }
    result struct {
        RepoID         string
        Commit         *git.Commit
        FindingResults []*findingResult
    }
    findingResult struct {
        FileChange *git.FileChange
        Findings   []*Finding
    }
)

func New(repoFilter *structures.Filter, refFilter *structures.Filter, processors []Processor, earliestTime, latestTime time.Time, whitelistPath structures.RegexpSet, whitelistSecretIDSet structures.Set, interact interact.Interactish, db *database.Database, log *logrus.Logger) *Finder {
    return &Finder{
        repoFilter:           repoFilter,
        refFilter:            refFilter,
        processors:           processors,
        earliestTime:         earliestTime,
        latestTime:           latestTime,
        whitelistPath:        whitelistPath,
        whitelistSecretIDSet: whitelistSecretIDSet,
        secretTracker:        structures.NewSet(nil),
        interact:             interact,
        db:                   db,
        log:                  log,
    }
}

func (f *Finder) Search() (secretCount int, err error) {
    for _, tableName := range []string{database.CommitTable, database.FindingTable, database.SecretTable, database.SecretFindingTable} {
        if f.db.TableExists(tableName) {
            err = errors.Errorv("one or more finder-specific tables already exist, cannot prepare findings", tableName)
            return
        }
    }

    var repos []*database.Repo
    if f.repoFilter != nil {
        repos, err = f.db.GetReposFilteredSorted(f.repoFilter)
    } else {
        repos, err = f.db.GetReposSorted()
    }
    if err != nil {
        return
    }

    numCPU := runtime.NumCPU()
    runtime.GOMAXPROCS(numCPU)

    out := make(chan *result)

    pl := pool.NewPool(len(repos), numCPU*2)
    pl.Start()

    prog := f.interact.NewProgress()

    for _, repo := range repos {
        log := f.log.WithField("repo", repo.Name)

        log.Debug("adding find worker for repo")
        pl.Add(NewWorker(
            repo,
            f.processors,
            &git.CommitFilter{
                EarliestTime:                f.earliestTime,
                LatestTime:                  f.latestTime,
                ExcludeMergeCommits:         true,
                ExcludeCommitsWithNoParents: true,
            },
            &git.FileChangeFilter{
                ExcludeFileDeletions:         true,
                ExcludeMatchingPaths:         f.whitelistPath,
                ExcludeBinaryOrEmpty:         true,
                ExcludeOnesWithNoCodeChanges: true,
            },
            prog,
            out,
            f.db,
            f.log.WithField("repo", repo.Name,
            ),
        ))
        log.Debug("worker added")
    }

    go func() {
        pl.Close()
        if prog != nil {
            prog.Wait()
        }
        close(out)
    }()

    // Process findings from channel
    for dr := range out {
        log := f.log.WithField("repo", dr.RepoID)
        log.Debug("received finding from channel")

        if err = f.persistResult(dr); err != nil {
            errors.ErrorLog(f.log, errors.WithMessage(err, "error processing commit"))
            continue
        }
    }

    secretCount = f.secretTracker.Len()
    f.log.Infof("completed, found %d secrets", secretCount)

    return
}

func (f *Finder) persistResult(result *result) (err error) {
    dbCommit, dbFindings, dbSecrets, ok := f.buildDBObjects(result)
    if !ok {
        return
    }

    if err = f.db.WriteCommitIfNotExists(dbCommit); err != nil {
        return
    }
    for _, dbSecret := range dbSecrets {
        if err = f.db.WriteSecretIfNotExists(dbSecret); err != nil {
            return
        }
    }
    for _, dbFinding := range dbFindings {
        if err = f.db.WriteFinding(dbFinding); err != nil {
            return
        }
    }

    return
}

func (f *Finder) buildDBObjects(result *result) (dbCommit *database.Commit, dbFindings []*database.Finding, dbSecrets []*database.Secret, ok bool) {
    var commit *database.Commit
    var secrets []*database.Secret
    var findings []*database.Finding

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

            dbFinding, findingErr := f.buildDBFinding(finding, result.Commit, findingResult.FileChange, dbSecret.ID, commit.ID)
            if findingErr != nil {
                errors.ErrorLogForEntry(log, findingErr).Error("unable to build finding object for database")
                continue
            }

            secrets = append(secrets, dbSecret)
            findings = append(findings, dbFinding)
        }
    }

    if findings != nil {
        dbCommit = commit
        dbFindings = findings
        dbSecrets = secrets
        ok = true
    }

    return
}

func (f *Finder) buildDBCommit(commit *git.Commit, repoID string) *database.Commit {
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

func (f *Finder) buildDBSecret(secret *Secret) *database.Secret {
    return &database.Secret{
        ID:           database.CreateHashID(secret.Value),
        Value:        secret.Value,
        ValueDecoded: secret.Decoded,
    }
}

func (f *Finder) buildDBFinding(finding *Finding, commit *git.Commit, fileChange *git.FileChange, secretID, commitID string) (result *database.Finding, err error) {
    var fileContents string
    fileContents, err = commit.FileContents(fileChange)
    if err != nil {
        return
    }

    // Get code and diff
    code := getExcerpt(fileContents, finding.FileRange.StartLineNum, finding.FileRange.EndLineNum)
    //diffExcerpt := getExcerpt(dfc.Diff, df.FileRange.StartDiffLineNum, df.FileRange.EndDiffLineNum)
    diff := ""
    const maxLength = 1000
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
        CommitID:         commitID,
        SecretID:         secretID,
        Processor:        finding.ProcessorName,
        Path:             fileChange.Path,
        StartLineNum:     finding.FileRange.StartLineNum,
        StartIndex:       finding.FileRange.StartIndex,
        EndLineNum:       finding.FileRange.EndLineNum,
        EndIndex:         finding.FileRange.EndIndex,
        StartDiffLineNum: finding.FileRange.StartDiffLineNum,
        EndDiffLineNum:   finding.FileRange.StartDiffLineNum,
        Code:             code,
        Diff:             diff,
    }

    return
}

func NewFindingFromLineFinding(finding *FindingInLine, fileLineNum, diffLineNum int) *Finding {
    return &Finding{
        ProcessorName: finding.ProcessorName,
        FileRange:     structures.NewFileRangeFromLineRange(finding.LineRange, fileLineNum, diffLineNum),
        Secret:        finding.Secret,
    }
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
