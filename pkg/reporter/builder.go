package reporter

import (
    "fmt"
    "github.com/pantheon-systems/search-secrets/pkg/database"
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    "github.com/pantheon-systems/search-secrets/pkg/structures"
    "github.com/sirupsen/logrus"
    "path"
    "path/filepath"
    "sort"
    "strings"
    "time"
)

type (
    Builder struct {
        appURL            string
        enableDebugOutput bool
        db                *database.Database
        log               logrus.FieldLogger
    }
    reportData struct {
        ReportDate        time.Time
        AppURL            string
        Repos             []string
        DbgEnabled        bool
        Secrets           []secretData
        EnableDebugOutput bool
    }
    secretData struct {
        ID       string        `yaml:"secret-id"`
        Value    string        `yaml:"value"`
        ValueLen int           `yaml:"-"`
        Extras   []extraData   `yaml:"extras"`
        Findings []findingData `yaml:"findings"`
    }
    findingData struct {
        ID                  string      `yaml:"finding-id"`
        ProcessorName       string      `yaml:"processor"`
        RepoName            string      `yaml:"-"`
        RepoFullLink        linkData    `yaml:"repo"`
        CommitHash          string      `yaml:"-"`
        CommitHashLink      linkData    `yaml:"commit"`
        CommitHashLinkShort linkData    `yaml:"-"`
        CommitDate          time.Time   `yaml:"commit-date"`
        CommitAuthorEmail   string      `yaml:"-"`
        CommitAuthorFull    string      `yaml:"commit-author"`
        FilePath            string      `yaml:"-"`
        FileLineLink        linkData    `yaml:"file-location"`
        FileLineLinkShort   linkData    `yaml:"-"`
        ColStartIndex       int         `yaml:"col-start-index"`
        ColEndIndex         int         `yaml:"col-end-index"`
        CodeShort           string      `yaml:"-"`
        Code                string      `yaml:"-"`
        CodeIsFile          bool        `yaml:"code-is-whole-file"`
        CodeTrimmed         string      `yaml:"code"`
        CodeShowGuide       bool        `yaml:"-"`
        Extras              []extraData `yaml:"extras"`
    }
    extraData struct {
        Key    string    `yaml:"key"`
        Header string    `yaml:"-"`
        Value  string    `yaml:"value"`
        Code   bool      `yaml:"-"`
        URL    string    `yaml:"url"`
        Link   *linkData `yaml:"-"`
    }
    linkData struct {
        Label   string `yaml:"label"`
        URL     string `yaml:"url"`
        Tooltip string `yaml:"-"`
    }
)

func NewBuilder(appURL string, enableDebugOutput bool, db *database.Database, log logrus.FieldLogger) *Builder {
    return &Builder{
        appURL:            appURL,
        enableDebugOutput: enableDebugOutput,
        db:                db,
        log:               log,
    }
}

func (b *Builder) buildReportData() (result *reportData, err error) {
    b.log.Debug("getting list of secrets ...")
    var ok bool

    var secrets database.Secrets
    secrets, err = b.db.GetSecretsSorted()
    if err != nil {
        err = errors.WithMessage(err, "unable to get secrets")
        return
    }

    var findingsBySecret database.FindingGroups
    findingsBySecret, err = b.db.GetFindingsSortedGroupedBySecretID()
    if err != nil {
        err = errors.WithMessage(err, "unable to get findings")
        return
    }

    var findingExtrasByFinding database.FindingExtraGroups
    findingExtrasByFinding, err = b.db.GetFindingExtrasSortedGroupedByFindingID()
    if err != nil {
        err = errors.WithMessage(err, "unable to get finding extras")
        return
    }

    var secretExtrasBySecretID database.SecretExtraGroups
    secretExtrasBySecretID, err = b.db.GetSecretExtrasSortedGroupedBySecretID()
    if err != nil {
        err = errors.WithMessage(err, "unable to get secret extras")
        return
    }

    var reportSecrets []secretData
    for _, secret := range secrets {
        var findings []*database.Finding
        findings, ok = findingsBySecret[secret.ID]
        if !ok {
            err = errors.Errorv("unable to find secret for finding group", secret.ID)
            return
        }

        var secretExtras database.SecretExtras
        secretExtras, _ = secretExtrasBySecretID[secret.ID]

        var secretData *secretData
        secretData, err = b.buildSecretData(secret, secretExtras, findings, findingExtrasByFinding)
        if err != nil {
            err = errors.WithMessage(err, "unable to build secret data")
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
        ReportDate:        time.Now(),
        AppURL:            b.appURL,
        Repos:             repoNames,
        EnableDebugOutput: b.enableDebugOutput,
        Secrets:           reportSecrets,
    }

    return
}

func (b *Builder) buildSecretData(secret *database.Secret, secretExtras database.SecretExtras, findings []*database.Finding, findingExtrasByFindingID database.FindingExtraGroups) (result *secretData, err error) {
    var findingDatas []findingData
    for _, finding := range findings {
        var findingExtras database.FindingExtras
        findingExtras, _ = findingExtrasByFindingID[finding.ID]

        var findingData findingData
        findingData, err = b.buildFindingData(finding, findingExtras)
        if err != nil {
            err = errors.WithMessage(err, "unable to build finding data")
            return
        }

        findingDatas = append(findingDatas, findingData)
    }

    var secretExtraDatas []extraData
    for _, secretExtra := range secretExtras {
        secretExtraDatas = append(secretExtraDatas, b.buildSecretExtraData(secretExtra))
    }

    // Sort findings by commit date
    sort.Slice(findingDatas, func(i, j int) bool { return findingDatas[i].CommitDate.Before(findingDatas[j].CommitDate) })

    result = &secretData{
        ID:       secret.ID,
        Value:    secret.Value,
        ValueLen: len(secret.Value),
        Extras:   secretExtraDatas,
        Findings: findingDatas,
    }

    return
}

func (b *Builder) buildFindingData(finding *database.Finding, findingExtras database.FindingExtras) (result findingData, err error) {
    var commit *database.Commit
    commit, err = b.db.GetCommit(finding.CommitID)
    if err != nil {
        err = errors.WithMessage(err, "unable to get commit")
        return
    }

    var repo *database.Repo
    repo, err = b.db.GetRepo(commit.RepoID)
    if err != nil {
        err = errors.WithMessage(err, "unable to get repo")
        return
    }

    var findingExtraDatas []extraData
    for _, findingExtra := range findingExtras {
        findingExtraData := b.buildFindingExtraData(findingExtra)
        findingExtraDatas = append(findingExtraDatas, findingExtraData)
    }
    if b.enableDebugOutput {
        findingExtraDatas = append(findingExtraDatas, extraData{
            Key:    "dbug-config",
            Header: "Debug config",
            Value:  b.buildDbugConfig(repo, commit, finding),
            Code:   true,
        })
    }

    commitURL := path.Join(repo.HTMLURL, "commit", commit.CommitHash)
    commitLink := linkData{Label: commit.CommitHash, URL: commitURL}
    commitLinkShort := linkData{Label: commit.CommitHash[:7], URL: commitURL, Tooltip: commit.CommitHash}

    fileLineURL := getLineURLLink(repo, commit, finding)
    fileLineLabel, fileLineLabelShort := b.getFileLineLabels(finding)
    fileLineLink := linkData{Label: fileLineLabel, URL: fileLineURL}
    fileLineLinkShort := linkData{Label: fileLineLabelShort, URL: fileLineURL, Tooltip: fileLineLabel}

    result = findingData{
        ID:                  finding.ID,
        ProcessorName:       finding.Processor,
        RepoName:            repo.Name,
        RepoFullLink:        linkData{Label: repo.FullName, URL: repo.HTMLURL},
        CommitHash:          commit.CommitHash,
        CommitHashLink:      commitLink,
        CommitHashLinkShort: commitLinkShort,
        CommitDate:          commit.Date,
        CommitAuthorEmail:   commit.AuthorEmail,
        CommitAuthorFull:    commit.AuthorFull,
        FilePath:            finding.Path,
        FileLineLink:        fileLineLink,
        FileLineLinkShort:   fileLineLinkShort,
        ColStartIndex:       finding.StartIndex,
        ColEndIndex:         finding.EndIndex,
        Code:                finding.Code,
        CodeTrimmed:         strings.TrimRight(finding.Code, "\n"),
        CodeShowGuide:       finding.StartLineNum == finding.EndLineNum,
        Extras:              findingExtraDatas,
    }

    return
}

func (b *Builder) getFileLineLabels(finding *database.Finding) (label, labelShort string) {

    // "file.go"
    filePathShort := filepath.Base(finding.Path)

    // ", line 123, col 123"
    lineColSuffix := fmt.Sprintf(", line %d, col %d", finding.StartLineNum, finding.StartIndex+1)

    // "path/to/file.go, line 123, col 123"
    label = finding.Path + lineColSuffix

    // "file.go, line 123, col 123"
    labelShort = filePathShort + lineColSuffix

    return
}

func (b *Builder) buildFindingExtraData(extra *database.FindingExtra) extraData {
    var link *linkData
    if extra.URL != "" {
        link = b.buildExtraLink(extra.Value, extra.URL)
    }

    return extraData{
        Key:    extra.Key,
        Header: extra.Header,
        Value:  extra.Value,
        Code:   extra.Code,
        Link:   link,
    }
}

func (b *Builder) buildSecretExtraData(extra *database.SecretExtra) extraData {
    var link *linkData
    if extra.URL != "" {
        link = b.buildExtraLink(extra.Value, extra.URL)
    }

    return extraData{
        Key:    extra.Key,
        Header: extra.Header,
        Value:  extra.Value,
        Code:   extra.Code,
        Link:   link,
    }
}

func (b *Builder) buildExtraLink(label, url string) (result *linkData) {
    if url != "" {
        return &linkData{label, url, ""}
    }
    return
}

func (b *Builder) buildDbugConfig(repo *database.Repo, commit *database.Commit, finding *database.Finding) (result string) {
    var sb strings.Builder
    fmt.Fprintf(&sb, "  filter:\n")
    fmt.Fprintf(&sb, "    repo: '%s'\n", repo.Name)
    fmt.Fprintf(&sb, "    processor: '%s'\n", finding.Processor)
    fmt.Fprintf(&sb, "    commit: '%s'\n", commit.CommitHash)
    fmt.Fprintf(&sb, "    path: '%s'\n", finding.Path)
    fmt.Fprintf(&sb, "    line: %d\n", finding.StartLineNum)

    return sb.String()
}

func getLineURLLink(repo *database.Repo, commit *database.Commit, finding *database.Finding) (result string) {
    baseURL := path.Join(repo.HTMLURL, "blob", commit.CommitHash, finding.Path)
    result = fmt.Sprintf("%s#L%d", baseURL, finding.StartLineNum)
    if finding.StartLineNum != finding.EndLineNum {
        result = fmt.Sprintf("%s-L%d", result, finding.EndLineNum)
    }

    return
}
