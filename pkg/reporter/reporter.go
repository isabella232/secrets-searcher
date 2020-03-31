package reporter

import (
    "fmt"
    "github.com/pantheon-systems/search-secrets/pkg/database"
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    "github.com/pantheon-systems/search-secrets/pkg/github"
    "github.com/sirupsen/logrus"
    "html/template"
    "os"
    "path"
    "path/filepath"
    "time"
)

type (
    Reporter struct {
        githubAPI *github.API
        dir       string
        db        *database.Database
        log       *logrus.Logger
    }
    reportData struct {
        Secrets []secretData
    }
    secretData struct {
        ID       string
        Value    string
        Findings []findingData
    }
    findingData struct {
        RuleName       string
        RepoFullLink   linkData
        CommitHashLink linkData
        CommitDate     time.Time
        CommitAuthor   string
        FileLineLink   linkData
        Code           string
        Diff           string
    }
    linkData struct {
        Label string
        URL   string
    }
)

func New(githubAPI *github.API, dir string, db *database.Database, log *logrus.Logger) *Reporter {
    return &Reporter{
        githubAPI: githubAPI,
        dir:       dir,
        db:        db,
        log:       log,
    }
}

func (r *Reporter) PrepareReport() (err error) {
    if err := os.MkdirAll(r.dir, 0700); err != nil {
        return errors.Wrapv(err, "unable to create report directory", r.dir)
    }

    var reportFilePath = filepath.Join(r.dir, "report.html")
    var reportFile *os.File
    reportFile, err = os.Create(reportFilePath)
    if err != nil {
        return
    }

    var reportData *reportData
    reportData, err = r.buildReportData()
    if err != nil {
        return
    }

    var tmpl *template.Template
    tmpl, err = template.ParseFiles("layout.gohtml")
    if err != nil {
        return
    }
    err = tmpl.Execute(reportFile, reportData)

    return
}

func (r *Reporter) buildReportData() (result *reportData, err error) {
    var secrets []*database.Secret
    secrets, err = r.db.GetSecrets()
    if err != nil {
        return
    }

    result = &reportData{}
    for _, secret := range secrets {
        var secretData *secretData
        secretData, err = r.buildSecretData(secret)
        if err != nil {
            return
        }

        result.Secrets = append(result.Secrets, *secretData)
    }

    return
}

func (r *Reporter) buildSecretData(secret *database.Secret) (result *secretData, err error) {
    var decs []*database.Decision
    decs, err = r.db.GetDecisionsForSecret(secret)
    if err != nil {
        return
    }

    var findings []findingData
    for _, dec := range decs {
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

func (r *Reporter) buildFindingData(dec *database.Decision) (result findingData, err error) {
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

    //var ghCommit *github2.RepositoryCommit
    //ghCommit, err = r.githubAPI.GetChange("pantheon-systems", repo.Name, commit.CommitHash)
    //fmt.Println(ghCommit)

    commitURL := path.Join(repo.HTMLURL, "commit", commit.CommitHash)
    fileLineURL := getLineURLLink(repo, commit, finding)

    result = findingData{
        RuleName:       finding.Rule,
        RepoFullLink:   linkData{Label: repo.FullName, URL: repo.HTMLURL},
        CommitHashLink: linkData{Label: commit.CommitHash[:7], URL: commitURL},
        CommitDate:     commit.Date,
        CommitAuthor:   "Unknown",
        FileLineLink:   linkData{Label: finding.Path, URL: fileLineURL},
        Code:           finding.Code,
        Diff:           finding.Diff,
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
