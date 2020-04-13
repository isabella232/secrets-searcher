package reporter

import (
    "fmt"
    "github.com/pantheon-systems/search-secrets/pkg/database"
    "github.com/pantheon-systems/search-secrets/pkg/dev"
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    "github.com/pantheon-systems/search-secrets/pkg/structures"
    "github.com/sirupsen/logrus"
    "path"
    "sort"
    "strings"
    "time"
)

type (
    Builder struct {
        appURL string
        db     *database.Database
        log    *logrus.Logger
    }
    reportData struct {
        ReportDate time.Time
        AppURL     string
        Repos      []string
        DevEnabled bool
        Secrets    []secretData
    }
    secretData struct {
        ID           string        `yaml:"secret-id"`
        Value        string        `yaml:"value"`
        ValueLen     int           `yaml:"-"`
        ValueDecoded string        `yaml:"value-decoded"`
        Findings     []findingData `yaml:"findings"`
    }
    findingData struct {
        ID                  string    `yaml:"finding-id"`
        RuleName            string    `yaml:"rule"`
        RepoName            string    `yaml:"-"`
        RepoFullLink        linkData  `yaml:"repo"`
        CommitHash          string    `yaml:"-"`
        CommitHashLink      linkData  `yaml:"commit"`
        CommitHashLinkShort linkData  `yaml:"-"`
        CommitDate          time.Time `yaml:"commit-date"`
        CommitAuthorEmail   string    `yaml:"-"`
        CommitAuthorFull    string    `yaml:"commit-author"`
        FilePath            string    `yaml:"-"`
        FileLineLink        linkData  `yaml:"file-location"`
        FileLineLinkShort   linkData  `yaml:"-"`
        StartLineNumDiff    int       `yaml:"-"`
        ColStartIndex       int       `yaml:"col-start-index"`
        ColEndIndex         int       `yaml:"col-end-index"`
        ColIndexDiff        int
        RawFileLink         linkData `yaml:"raw-file-location"`
        CodeShort           string   `yaml:"-"`
        Code                string   `yaml:"-"`
        CodeTrimmed         string   `yaml:"code"`
        CodeShowGuide       bool
        Diff                string `yaml:"diff"`
    }
    linkData struct {
        Label   string `yaml:"label"`
        URL     string `yaml:"url"`
        Tooltip string `yaml:"tooltip"`
    }
)

func NewBuilder(appURL string, db *database.Database, log *logrus.Logger) *Builder {
    return &Builder{
        appURL: appURL,
        db:     db,
        log:    log,
    }
}

func (r *Builder) buildReportData() (result *reportData, err error) {
    r.log.Debug("getting list of secrets ...")

    var sfsBySecret map[*database.Secret][]*database.SecretFinding
    sfsBySecret, err = r.db.GetSecretFindingsGroupedBySecret()
    if err != nil {
        return
    }

    // Sort secrets since they were just in a map and lost order
    var secrets []*database.Secret
    for secret := range sfsBySecret {
        secrets = append(secrets, secret)
    }
    r.db.SortSecrets(secrets)

    var reportSecrets []secretData
    var ok bool
    for _, secret := range secrets {
        var sfs []*database.SecretFinding
        sfs, ok = sfsBySecret[secret]
        if !ok {
            err = errors.New("secret does not exist as index")
            return
        }

        var secretData *secretData
        secretData, err = r.buildSecretData(secret, sfs)
        if err != nil {
            return
        }

        reportSecrets = append(reportSecrets, *secretData)
    }

    repos := structures.NewSet(nil)
    for _, reportSecret := range reportSecrets {
        for _, reportFinding := range reportSecret.Findings {
            repos.Add(reportFinding.RepoName)
        }
    }
    repoNames := repos.Values()
    sort.Strings(repoNames)

    result = &reportData{
        ReportDate: time.Now(),
        AppURL:     r.appURL,
        Repos:      repoNames,
        DevEnabled: dev.Enabled,
        Secrets:    reportSecrets,
    }

    return
}

func (r *Builder) buildSecretData(secret *database.Secret, sfs []*database.SecretFinding) (result *secretData, err error) {
    var findings []findingData
    for _, dec := range sfs {
        var findingData findingData
        findingData, err = r.buildFindingData(dec)
        if err != nil {
            return
        }

        findings = append(findings, findingData)
    }

    result = &secretData{
        ID:           secret.ID,
        Value:        secret.Value,
        ValueLen:     len(secret.Value),
        ValueDecoded: secret.ValueDecoded,

        Findings: findings,
    }

    return
}

func (r *Builder) buildFindingData(dec *database.SecretFinding) (result findingData, err error) {
    var finding *database.Finding
    finding, err = r.db.GetFinding(dec.FindingID)
    if err != nil {
        return
    }

    var commit *database.Commit
    commit, err = r.db.GetCommit(finding.CommitID)
    if err != nil {
        return
    }

    var repo *database.Repo
    repo, err = r.db.GetRepo(commit.RepoID)
    if err != nil {
        return
    }

    commitURL := path.Join(repo.HTMLURL, "commit", commit.CommitHash)

    fileLineLabelFormat := "%s, line %d, col %d"
    if dev.Enabled {
        fileLineLabelFormat += " (diff line %d)"
    }

    fileLineURL := getLineURLLink(repo, commit, finding)
    fileLineLabel := fmt.Sprintf(fileLineLabelFormat, finding.Path, finding.StartLineNum, finding.StartIndex+1, finding.StartDiffLineNum)
    fileLineLink := linkData{Label: fileLineLabel, URL: fileLineURL}

    fileLineShortLabel := fmt.Sprintf(fileLineLabelFormat, path.Base(finding.Path), finding.StartLineNum, finding.StartIndex+1, finding.StartDiffLineNum)
    fileLineShortLink := linkData{Label: fileLineShortLabel, URL: fileLineURL, Tooltip: fileLineLabel}

    rawFileURL := getRawURLLink(repo, commit, finding)
    rawFileLabel := fmt.Sprintf("%s, line %d", finding.Path, finding.StartLineNum)
    rawFileLink := linkData{Label: rawFileLabel, URL: rawFileURL}

    result = findingData{
        ID:                  finding.ID,
        RuleName:            finding.Rule,
        RepoName:            repo.Name,
        RepoFullLink:        linkData{Label: repo.FullName, URL: repo.HTMLURL},
        CommitHash:          commit.CommitHash,
        CommitHashLink:      linkData{Label: commit.CommitHash, URL: commitURL},
        CommitHashLinkShort: linkData{Label: commit.CommitHash[:7], URL: commitURL, Tooltip: commit.CommitHash},
        CommitDate:          commit.Date,
        CommitAuthorEmail:   commit.AuthorEmail,
        CommitAuthorFull:    commit.AuthorFull,
        FilePath:            finding.Path,
        FileLineLink:        fileLineLink,
        FileLineLinkShort:   fileLineShortLink,
        StartLineNumDiff:    finding.StartDiffLineNum,
        ColStartIndex:       finding.StartIndex,
        ColEndIndex:         finding.EndIndex,
        ColIndexDiff:        finding.EndIndex - finding.StartIndex,
        RawFileLink:         rawFileLink,
        Code:                finding.Code,
        CodeTrimmed:         strings.TrimRight(finding.Code, "\n"),
        CodeShowGuide:       finding.StartLineNum == finding.EndLineNum,
        Diff:                finding.Diff,
    }

    return
}

func getLineURLLink(repo *database.Repo, commit *database.Commit, finding *database.Finding) (result string) {
    baseURL := path.Join(repo.HTMLURL, "blob", commit.CommitHash, finding.Path)
    result = fmt.Sprintf("%s#L%d", baseURL, finding.StartLineNum)
    if finding.StartLineNum != finding.EndLineNum {
        result = fmt.Sprintf("%s-%d", result, finding.EndLineNum)
    }

    return
}

func getRawURLLink(repo *database.Repo, commit *database.Commit, finding *database.Finding) (result string) {
    baseURL := path.Join(repo.HTMLURL, "blob", commit.CommitHash, finding.Path)
    result = fmt.Sprintf("%s#L%d", baseURL, finding.StartLineNum)
    if finding.StartLineNum != finding.EndLineNum {
        result = fmt.Sprintf("%s-%d", result, finding.EndLineNum)
    }

    return
}
