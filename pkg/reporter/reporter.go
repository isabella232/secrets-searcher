package reporter

//go:generate templify -p reporter -o report_template.go source/report.gohtml

import (
    "fmt"
    "github.com/pantheon-systems/search-secrets/pkg/database"
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    "github.com/sirupsen/logrus"
    "gopkg.in/yaml.v2"
    "html/template"
    "io/ioutil"
    "os"
    "path"
    "path/filepath"
    "time"
)

type (
    Reporter struct {
        dir            string
        secretsDir     string
        reportFilePath string
        db             *database.Database
        log            *logrus.Logger
    }
    reportData struct {
        ReportDate time.Time
        Secrets    []secretData
    }
    secretData struct {
        ID       string        `yaml:"secret-id"`
        Value    string        `yaml:"value"`
        Findings []findingData `yaml:"findings"`
    }
    findingData struct {
        ID                  string    `yaml:"finding-id"`
        RuleName            string    `yaml:"rule"`
        RepoFullLink        linkData  `yaml:"repo"`
        CommitHashLink      linkData  `yaml:"commit"`
        CommitHashLinkShort linkData  `yaml:"-"`
        CommitDate          time.Time `yaml:"commit-date"`
        CommitAuthorEmail   string    `yaml:"-"`
        CommitAuthorFull    string    `yaml:"commit-author"`
        FileLineLink        linkData  `yaml:"file-location"`
        FileLineLinkShort   linkData  `yaml:"-"`
        CodeShort           string    `yaml:"-"`
        Code                string    `yaml:"code"`
        Diff                string    `yaml:"diff"`
    }
    linkData struct {
        Label   string `yaml:"label"`
        URL     string `yaml:"url"`
        Tooltip string `yaml:"tooltip"`
    }
)

func New(dir string, db *database.Database, log *logrus.Logger) *Reporter {
    return &Reporter{
        dir:            dir,
        secretsDir:     filepath.Join(dir, "secrets"),
        reportFilePath: filepath.Join(dir, "report.html"),
        db:             db,
        log:            log,
    }
}

func (r *Reporter) PrepareReport() (err error) {
    if _, err = os.Stat(r.dir); !os.IsNotExist(err) {
        return errors.Errorv("report directory already exists, cannot prepare report", r.dir)
    }

    if err := os.MkdirAll(r.dir, 0700); err != nil {
        return errors.Wrapv(err, "unable to create report directory", r.dir)
    }
    if err := os.MkdirAll(r.secretsDir, 0700); err != nil {
        return errors.Wrapv(err, "unable to create secrets directory", r.secretsDir)
    }

    var reportFile *os.File
    reportFile, err = os.Create(r.reportFilePath)
    if err != nil {
        return errors.Wrapv(err, "unable to create report file", r.reportFilePath)
    }

    var reportData *reportData
    reportData, err = r.buildReportData()
    if err != nil {
        return
    }

    var tmpl *template.Template
    tmpl = template.New("report")
    tmpl, err = tmpl.Parse(report_templateTemplate())
    if err != nil {
        return err
    }
    err = tmpl.Execute(reportFile, reportData)

    err = r.outputSecrets(reportData)
    if err != nil {
        return
    }

    return
}

func (r *Reporter) buildReportData() (result *reportData, err error) {
    r.log.Debug("getting list of secrets ...")

    var sfsBySecret map[*database.Secret][]*database.SecretFinding
    sfsBySecret, err = r.db.GetSecretFindingsGroupedBySecret()
    if err != nil {
        return
    }

    var reportSecrets []secretData
    for secret, sfs := range sfsBySecret {
        var secretData *secretData
        secretData, err = r.buildSecretData(secret, sfs)
        if err != nil {
            return
        }

        reportSecrets = append(reportSecrets, *secretData)
    }

    result = &reportData{
        ReportDate: time.Now(),
        Secrets:    reportSecrets,
    }

    return
}

func (r *Reporter) buildSecretData(secret *database.Secret, sfs []*database.SecretFinding) (result *secretData, err error) {
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
        ID:       secret.ID,
        Value:    secret.Value,
        Findings: findings,
    }

    return
}

func (r *Reporter) buildFindingData(dec *database.SecretFinding) (result findingData, err error) {
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

    fileLineURL := getLineURLLink(repo, commit, finding)
    fileLineLabel := fmt.Sprintf("%s, line %d", finding.Path, finding.StartLineNum)
    fileLineLink := linkData{Label: fileLineLabel, URL: fileLineURL}

    fileLineShortLabel := fmt.Sprintf("%s, line %d", path.Base(finding.Path), finding.StartLineNum)
    fileLineShortLink := linkData{Label: fileLineShortLabel, URL: fileLineURL, Tooltip: fileLineLabel}

    result = findingData{
        ID:                  finding.ID,
        RuleName:            finding.Rule,
        RepoFullLink:        linkData{Label: repo.FullName, URL: repo.HTMLURL},
        CommitHashLink:      linkData{Label: commit.CommitHash, URL: commitURL},
        CommitHashLinkShort: linkData{Label: commit.CommitHash[:7], URL: commitURL, Tooltip: commit.CommitHash},
        CommitDate:          commit.Date,
        CommitAuthorEmail:   commit.AuthorEmail,
        CommitAuthorFull:    commit.AuthorFull,
        FileLineLink:        fileLineLink,
        FileLineLinkShort:   fileLineShortLink,
        Code:                finding.Code,
        Diff:                finding.Diff,
    }

    return
}

func (r *Reporter) outputSecrets(data *reportData) (err error) {
    for _, secretData := range data.Secrets {
        filePath := filepath.Join(r.secretsDir, fmt.Sprintf("secret-%s.yaml", secretData.ID))

        var bytes []byte
        bytes, err = yaml.Marshal(secretData)
        if err != nil {
            return
        }

        err = ioutil.WriteFile(filePath, bytes, 0644)
        if err != nil {
            return
        }
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
