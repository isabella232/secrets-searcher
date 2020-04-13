package finder

import (
    "github.com/pantheon-systems/search-secrets/pkg/database"
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    "github.com/pantheon-systems/search-secrets/pkg/finder/rule"
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
        rules                []rule.Rule
        earliestTime         time.Time
        latestTime           time.Time
        whitelistPath        structures.RegexpSet
        whitelistSecretIDSet structures.Set
        secretTracker        structures.Set
        interact             interact.Interactish
        db                   *database.Database
        log                  *logrus.Logger
    }
    DriverResult struct {
        Commit *DriverCommit
    }
    DriverCommit struct {
        RepoID      string
        RepoName    string
        Commit      string
        CommitHash  string
        Date        time.Time
        AuthorFull  string
        AuthorEmail string
        FileChanges []*DriverFileChange
    }
    DriverFileChange struct {
        Path         string
        FileContents string
        Findings     []*DriverFinding
        Diff         string
    }
    DriverFinding struct {
        RuleName  string
        FileRange *structures.FileRange
        Secrets   []*DriverSecret
    }
    DriverSecret struct {
        Value   string
        Decoded string
    }
)

func New(repoFilter *structures.Filter, refFilter *structures.Filter, rules []rule.Rule, earliestTime, latestTime time.Time, whitelistPath structures.RegexpSet, whitelistSecretIDSet structures.Set, interact interact.Interactish, db *database.Database, log *logrus.Logger) *Finder {
    return &Finder{
        repoFilter:           repoFilter,
        refFilter:            refFilter,
        rules:                rules,
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

    out := make(chan *DriverResult)

    pl := pool.NewPool(len(repos), numCPU*2)
    pl.Start()

    prog := f.interact.NewProgress()

    for _, repo := range repos {
        log := f.log.WithField("repo", repo.Name)

        log.Debug("adding find worker for repo")
        pl.Add(worker{
            repo:          repo,
            cloneDir:      repo.CloneDir,
            refs:          f.refFilter.Values(),
            rules:         f.rules,
            earliestTime:  f.earliestTime,
            latestTime:    f.latestTime,
            whitelistPath: f.whitelistPath,
            prog:          prog,
            out:           out,
            searched:      structures.NewSet(nil),
            db:            f.db,
            log:           f.log.WithField("repo", repo.Name),
        })
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
        log := f.log.WithField("repo", dr.Commit.RepoName)
        log.Debug("received finding from channel")

        if err = f.processCommit(dr.Commit); err != nil {
            errors.ErrorLog(f.log, errors.WithMessage(err, "error processing commit"))
            continue
        }
    }

    secretCount = f.secretTracker.Len()
    f.log.Infof("completed, found %d secrets", secretCount)

    return
}

func (f *Finder) processCommit(dc *DriverCommit) (err error) {
    commitID := database.CreateHashID(dc.RepoID, dc.CommitHash)

    var saved bool
    var oneSaved bool
    for _, dfc := range dc.FileChanges {
        for _, df := range dfc.Findings {
            newLog := f.log.WithFields(logrus.Fields{
                "repo":       dc.RepoName,
                "commitHash": dc.CommitHash,
                "rule":       df.RuleName,
                "path":       dfc.Path,
            })

            saved, err = f.processFinding(commitID, dfc, df, newLog)
            if err != nil {
                return
            }
            if saved {
                oneSaved = true
            }
        }
    }

    if !oneSaved {
        return
    }

    commit := &database.Commit{
        ID:          commitID,
        RepoID:      dc.RepoID,
        Commit:      dc.Commit,
        CommitHash:  dc.CommitHash,
        Date:        dc.Date,
        AuthorFull:  dc.AuthorFull,
        AuthorEmail: dc.AuthorEmail,
    }
    err = f.db.WriteCommitIfNotExists(commit)

    return
}

func (f *Finder) processFinding(commitID string, dfc *DriverFileChange, df *DriverFinding, log *logrus.Entry) (saved bool, err error) {
    findingID := database.CreateHashID(commitID, df.RuleName, dfc.Path,
        df.FileRange.StartLineNum, df.FileRange.StartIndex, df.FileRange.EndLineNum, df.FileRange.EndIndex)
    log = log.WithFields(logrus.Fields{
        "finding": findingID,
    })

    // Collect secrets
    secrets := f.getSecretsFromFinding(df)
    if secrets == nil {
        log.Debug("no secrets found for finding, not saving")
        return
    }

    // Get code excerpt
    codePadding := 0 //TODO Add some padding that will show up in the report
    codeExcerpt := getExcerpt(dfc.FileContents, df.FileRange.StartLineNum, df.FileRange.EndLineNum)

    // Get diff excerpt
    diffPadding := 0 //TODO Add some padding that will show up in the report
    //diffExcerpt := getExcerpt(dfc.Diff, df.FileRange.StartDiffLineNum, df.FileRange.EndDiffLineNum)
    diffExcerpt := ""

    const maxLength = 1000
    if len(codeExcerpt) > maxLength {
        log.Warn("truncating code output")
        codeExcerpt = codeExcerpt[:maxLength] + " [...]"
    }

    log.Debug("saving finding")

    // Save finding
    finding := &database.Finding{
        ID:               findingID,
        CommitID:         commitID,
        Rule:             df.RuleName,
        Path:             dfc.Path,
        StartLineNum:     df.FileRange.StartLineNum,
        StartIndex:       df.FileRange.StartIndex,
        EndLineNum:       df.FileRange.EndLineNum,
        EndIndex:         df.FileRange.EndIndex,
        StartDiffLineNum: df.FileRange.StartDiffLineNum,
        EndDiffLineNum:   df.FileRange.StartDiffLineNum,
        Code:             codeExcerpt,
        CodePadding:      codePadding,
        Diff:             diffExcerpt,
        DiffPadding:      diffPadding,
    }
    if err = f.db.WriteFinding(finding); err != nil {
        return
    }

    for _, secret := range secrets {
        log = log.WithField("secret", secret.ID)
        log.Debug("saving secret")

        if err = f.db.WriteSecretIfNotExists(secret); err != nil {
            return
        }
        f.secretTracker.Add(secret.ID)

        secretFinding := &database.SecretFinding{
            ID:        database.CreateHashID(findingID, secret.ID),
            FindingID: findingID,
            SecretID:  secret.ID,
        }
        if err = f.db.WriteSecretFinding(secretFinding); err != nil {
            return
        }
    }

    saved = true

    return
}

func (f *Finder) getSecretsFromFinding(df *DriverFinding) (secrets []*database.Secret) {
    for _, secret := range df.Secrets {
        secretID := database.CreateHashID(secret.Value)

        // Check whitelist
        if f.whitelistSecretIDSet.Contains(secretID) {
            f.log.WithField("secret", secretID).Debug("secret whitelisted by ID, skipping secret")
            continue
        }

        secrets = append(secrets, &database.Secret{
            ID:           secretID,
            Value:        secret.Value,
            ValueDecoded: secret.Decoded,
        })
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
