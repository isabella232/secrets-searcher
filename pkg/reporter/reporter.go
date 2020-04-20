package reporter

//go:generate go run github.com/wlbr/templify -p reporter -o template_report.go source/report.gohtml

import (
    "fmt"
    "github.com/otiai10/copy"
    "github.com/pantheon-systems/search-secrets/pkg/database"
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    "github.com/sirupsen/logrus"
    "gopkg.in/yaml.v2"
    "html/template"
    "io/ioutil"
    "os"
    "path/filepath"
    "strings"
    "time"
)

var (
    templateFuncs = template.FuncMap{
        "stringRepeat": func(width int, str string) template.HTML {
            return template.HTML(strings.Repeat(str, width))
        },
    }
)

type Reporter struct {
    reportDir         string
    reportArchivesDir string
    secretsDir        string
    reportFilePath    string
    builder           *builder
    db                *database.Database
    log               logrus.FieldLogger
}

func New(reportDir, reportArchivesDir, appURL string, enableDebugOutput bool, db *database.Database, log logrus.FieldLogger) *Reporter {
    secretsDir := filepath.Join(reportDir, "secrets")
    reportFilePath := filepath.Join(reportDir, "report.html")

    return &Reporter{
        reportDir:         reportDir,
        reportArchivesDir: reportArchivesDir,
        secretsDir:        secretsDir,
        reportFilePath:    reportFilePath,
        builder:           newBuilder(appURL, enableDebugOutput, reportDir, secretsDir, db, log),
        db:                db,
        log:               log,
    }
}

func (r *Reporter) PrepareReport() (err error) {
    r.log.Info("creating report ... ")

    // Prepare fs
    if err = r.prepareFilesystem(); err != nil {
        return errors.WithMessage(err, "unable to create prepare filesystem for report")
    }

    var data *reportData
    data, err = r.builder.buildReportData()
    if err != nil {
        return errors.WithMessage(err, "unable to build report data")
    }

    // Create report HTML file
    if err = r.createReportFile(data); err != nil {
        return errors.WithMessage(err, "unable to create report file")
    }

    if data.Secrets == nil {
        return
    }

    // Create secrets directory
    if err = r.createSecretFiles(data); err != nil {
        return errors.WithMessage(err, "unable to create secret files")
    }

    // Copy report directory archive
    archiveDirname := fmt.Sprintf("report-%s", time.Now().Format("2006-01-02_15-04-05"))
    archiveDir := filepath.Join(r.reportArchivesDir, archiveDirname)
    r.log.Debugf("copying %s to %s ...", r.reportDir, archiveDir)
    if err = copy.Copy(r.reportDir, archiveDir); err != nil {
        return errors.Wrapv(err, "", r.reportDir, r.reportArchivesDir)
    }

    return
}

func (r *Reporter) prepareFilesystem() (err error) {
    r.log.Debug("resetting report directory ... ")

    if err = os.RemoveAll(r.reportDir); err != nil {
        return errors.Wrapv(err, "unable to delete report directory", r.reportDir)
    }
    if err = os.MkdirAll(r.reportDir, 0700); err != nil {
        return errors.Wrapv(err, "unable to create report directory", r.reportDir)
    }
    if err = os.MkdirAll(r.reportArchivesDir, 0700); err != nil {
        return errors.Wrapv(err, "unable to create report archives directory", r.reportDir)
    }
    return
}

func (r *Reporter) createReportFile(data *reportData) (err error) {
    var reportFile *os.File
    reportFile, err = os.Create(r.reportFilePath)
    if err != nil {
        return errors.Wrapv(err, "unable to create report file", r.reportFilePath)
    }

    var tmpl *template.Template
    tmpl = template.New("report").Funcs(templateFuncs)
    tmpl, err = tmpl.Parse(template_reportTemplate())
    if err != nil {
        return err
    }
    err = tmpl.Execute(reportFile, data)

    return
}

func (r *Reporter) createSecretFiles(data *reportData) (err error) {
    if err = os.MkdirAll(r.secretsDir, 0700); err != nil {
        return errors.Wrapv(err, "unable to create secrets directory", r.secretsDir)
    }

    for _, sData := range data.Secrets {

        // Paths
        secretDir := filepath.Join(r.secretsDir, sData.ID)

        // Create secret directory if necenssary
        if err = os.MkdirAll(secretDir, 0700); err != nil {
            return errors.Wrapv(err, "unable to create secret directory", secretDir)
        }

        err = r.outputSecretValueFileAndAddLink(sData, secretDir)
        if err != nil {
            err = errors.Wrap(err, "unable to create file")
            return
        }

        err = r.outputSecretMetadataFile(sData, secretDir)
        if err != nil {
            err = errors.Wrap(err, "unable to create file")
            return
        }
    }

    return
}

func (r *Reporter) outputSecretValueFileAndAddLink(sData *secretData, secretDir string) (err error) {
    if sData.ValueFilePath == "" {
        return
    }

    // Create file
    var f *os.File
    f, err = os.Create(sData.ValueFilePath)
    if err != nil {
        err = errors.Wrapv(err, "unable to create file", sData.ValueFilePath)
        return
    }
    defer f.Close()

    _, err = f.WriteString(sData.Value)
    if err != nil {
        err = errors.Wrapv(err, "unable to write to file", sData.ValueFilePath)
        return
    }

    return
}

func (r *Reporter) outputSecretMetadataFile(sData *secretData, secretDir string) (err error) {

    // File path
    filePath := filepath.Join(secretDir, fmt.Sprintf("secret-%s.yaml", sData.ID))

    // Marshal secret data to YAML bytes
    var bytes []byte
    bytes, err = yaml.Marshal(sData)
    if err != nil {
        return errors.WithMessage(err, "unable to marshal secret into YAML")
    }

    // Write to file
    err = ioutil.WriteFile(filePath, bytes, 0644)
    if err != nil {
        return errors.WithMessagev(err, "unable to write secret", sData.ID, filePath)
    }

    return
}
