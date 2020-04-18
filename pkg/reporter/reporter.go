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
)

var (
    templateFuncs = template.FuncMap{
        "stringRepeat": func(width int, str string) template.HTML {
            return template.HTML(strings.Repeat(str, width))
        },
    }
)

type Reporter struct {
    dir               string
    archiveDir        string
    secretsDir        string
    reportFilePath    string
    skipReportSecrets bool
    builder           *Builder
    db                *database.Database
    log               *logrus.Logger
}

func New(dir, archiveDir string, skipReportSecrets bool, appURL string, enableDebugOutput bool, db *database.Database, log *logrus.Logger) *Reporter {
    return &Reporter{
        dir:               dir,
        archiveDir:        archiveDir,
        secretsDir:        filepath.Join(dir, "secrets"),
        reportFilePath:    filepath.Join(dir, "report.html"),
        skipReportSecrets: skipReportSecrets,
        builder:           NewBuilder(appURL, enableDebugOutput, db, log),
        db:                db,
        log:               log,
    }
}

func (r *Reporter) PrepareReport() (err error) {
    if _, err = os.Stat(r.dir); !os.IsNotExist(err) {
        return errors.Errorv("report directory already exists, cannot prepare report", r.dir)
    }

    if err := os.MkdirAll(r.dir, 0700); err != nil {
        return errors.Wrapv(err, "unable to create report directory", r.dir)
    }
    if !r.skipReportSecrets {
        if err := os.MkdirAll(r.secretsDir, 0700); err != nil {
            return errors.Wrapv(err, "unable to create secrets directory", r.secretsDir)
        }
    }

    var reportFile *os.File
    reportFile, err = os.Create(r.reportFilePath)
    if err != nil {
        return errors.Wrapv(err, "unable to create report file", r.reportFilePath)
    }

    var reportData *reportData
    reportData, err = r.builder.buildReportData()
    if err != nil {
        return
    }
    if reportData.Secrets != nil {
        r.log.Infof("found %d secrets", len(reportData.Secrets))
    } else {
        r.log.Info("found no secrets")
    }

    var tmpl *template.Template
    tmpl = template.New("report").Funcs(templateFuncs)
    tmpl, err = tmpl.Parse(template_reportTemplate())
    if err != nil {
        return err
    }
    err = tmpl.Execute(reportFile, reportData)

    if !r.skipReportSecrets && reportData.Secrets != nil {
        err = r.outputSecrets(reportData)
        if err != nil {
            return
        }
    }

    r.log.Debugf("copying %s to %s ...", r.dir, r.archiveDir)
    if err = copy.Copy(r.dir, r.archiveDir); err != nil {
        return errors.Wrapv(err, "", r.dir, r.archiveDir)
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
