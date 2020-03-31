package finder

import (
    "github.com/pantheon-systems/search-secrets/pkg/code"
    "github.com/pantheon-systems/search-secrets/pkg/database"
    decisionenum "github.com/pantheon-systems/search-secrets/pkg/database/enum/decision"
    "github.com/pantheon-systems/search-secrets/pkg/rule"
    "sync"

    "github.com/pantheon-systems/search-secrets/pkg/structures"
    "github.com/sirupsen/logrus"
    "strings"
    "time"
)

type (
    Finder struct {
        driver               Driver
        code                 *code.Code
        repoFilter           *structures.Filter
        refFilter            *structures.Filter
        rules                []*rule.Rule
        earliestTime         time.Time
        latestTime           time.Time
        earliestCommit       string
        latestCommit         string
        whitelistPath        structures.RegexpSet
        whitelistSecretIDSet structures.Set
        db                   *database.Database
        log                  *logrus.Logger
    }
    Driver interface {
        Find(repoID, repoName, cloneDir string, refs []string, rules []*rule.Rule, earliestTime, latestTime time.Time, earliestCommit, latestCommit string, whitelistPath structures.RegexpSet, out chan *DriverResult)
    }
    DriverResult struct {
        Err    error
        Commit *DriverCommit
    }
    DriverCommit struct {
        RepoID      string
        Commit      string
        CommitHash  string
        Date        time.Time
        AuthorEmail string
        FileChanges []*DriverFileChange
    }
    DriverFileChange struct {
        Path         string
        FileContents string
        Diff         string
        Findings     []*DriverFinding
    }
    DriverFinding struct {
        Rule             *rule.Rule
        FileRange        *structures.FileRange
        SecretsProcessed bool
        SecretValues     []string
    }
)

func New(driver Driver, code *code.Code, repoFilter, refFilter *structures.Filter, rules []*rule.Rule, earliestTime, latestTime time.Time, earliestCommit, latestCommit string, whitelistPath structures.RegexpSet, whitelistSecretIDSet structures.Set, db *database.Database, log *logrus.Logger) *Finder {
    return &Finder{
        driver:               driver,
        code:                 code,
        repoFilter:           repoFilter,
        refFilter:            refFilter,
        rules:                rules,
        earliestTime:         earliestTime,
        latestTime:           latestTime,
        earliestCommit:       earliestCommit,
        latestCommit:         latestCommit,
        whitelistPath:        whitelistPath,
        whitelistSecretIDSet: whitelistSecretIDSet,
        db:                   db,
        log:                  log,
    }
}

func (f *Finder) PrepareFindings() (err error) {
    if f.db.TableExists(database.FindingTable) {
        //f.log.Warn("finding table already exists, skipping")
        //return
    }
    var repos []*database.Repo
    repos, err = f.db.GetReposFiltered(f.repoFilter)
    if err != nil {
        return
    }

    var wg sync.WaitGroup
    out := make(chan *DriverResult)
    defer close(out)

    // Create goroutines for repo that push findings into the channel
    for _, repo := range repos {
        cloneDir := f.code.CloneDir(repo.Name)
        refs := f.refFilter.Values()

        wg.Add(1)
        go func() {
            defer wg.Done()
            f.driver.Find(repo.ID, repo.Name, cloneDir, refs, f.rules, f.earliestTime, f.latestTime, f.earliestCommit, f.latestCommit, f.whitelistPath, out)
        }()
    }

    // Process findings from channel
    go func() {
        var err error
        err = f.ProcessResult(&out)
        if err != nil {
            f.log.Error(err)
            return
        }
    }()

    wg.Wait()

    return
}

func (f *Finder) ProcessResult(out *chan *DriverResult) (err error) {
    for dr := range *out {
        if dr.Err != nil {
            err = dr.Err
            return
        }

        err = f.processCommit(dr)
        if err != nil {
            return
        }
    }

    return
}

func (f *Finder) processCommit(dr *DriverResult) (err error) {
    dc := dr.Commit

    commit := &database.Commit{
        ID:          dc.CommitHash,
        RepoID:      dc.RepoID,
        Commit:      dc.Commit,
        CommitHash:  dc.CommitHash,
        Date:        dc.Date,
        AuthorEmail: dc.AuthorEmail,
    }
    if err = f.db.WriteCommit(commit); err != nil {
        return
    }

    for _, dfc := range dc.FileChanges {
        for _, df := range dfc.Findings {
            err = f.processFinding(dc, dfc, df)
            if err != nil {
                return
            }
        }
    }

    return
}

func (f *Finder) processFinding(dc *DriverCommit, dfc *DriverFileChange, df *DriverFinding) (err error) {

    findingID := database.CreateHashID(dc.CommitHash, df.Rule.Name, dfc.Path,
        df.FileRange.StartLineNum, df.FileRange.StartIndex, df.FileRange.EndLineNum, df.FileRange.EndIndex)

    // Collect secrets
    secrets := f.getSecretsFromFinding(df)
    if secrets == nil {
        f.log.WithField("findingID", findingID).Debug("no secrets found for finding, not saving")
        return
    }

    // Get code excerpt
    codeExcerpt := getExcerpt(dfc.FileContents, df.FileRange.StartLineNum, df.FileRange.EndLineNum-df.FileRange.StartLineNum+1)

    // Save finding
    finding := &database.Finding{
        ID:               findingID,
        CommitID:         dc.CommitHash,
        Rule:             df.Rule.Name,
        Path:             dfc.Path,
        StartLineNum:     df.FileRange.StartLineNum,
        StartIndex:       df.FileRange.StartIndex,
        EndLineNum:       df.FileRange.EndLineNum,
        EndIndex:         df.FileRange.EndIndex,
        Code:             codeExcerpt,
        Diff:             dfc.Diff,
        SecretsProcessed: df.SecretsProcessed,
    }
    if err = f.db.WriteFinding(finding); err != nil {
        return
    }

    for _, secret := range secrets {
        if err = f.db.WriteSecret(secret); err != nil {
            return
        }

        decision := &database.Decision{
            ID:        database.CreateHashID(findingID, secret.ID),
            FindingID: findingID,
            SecretID:  secret.ID,
            Decision:  decisionenum.DoNotKnowYet{}.New(),
        }
        if err = f.db.WriteDecision(decision); err != nil {
            return
        }
    }

    return
}

func (f *Finder) getSecretsFromFinding(df *DriverFinding) (secrets []*database.Secret) {
    if df.SecretsProcessed && df.SecretValues != nil {
        for _, secretValue := range df.SecretValues {
            secretID := database.CreateHashID(secretValue)
            if ! f.whitelistSecretIDSet.Contains(secretID) {
                f.log.WithField("secretID", secretID).Debug("secrets whitelisted by ID, skipping secret")
            }
            secret := &database.Secret{
                ID:    secretID,
                Value: secretValue,
            }
            secrets = append(secrets, secret)
        }
    }
    return
}

func getExcerpt(contents string, fromLineNum int, lineCount int) (result string) {
    lineNum := 1
    theRest := contents
    toLineNum := fromLineNum + lineCount
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
        if lineNum == toLineNum {
            return
        }
    }
}
