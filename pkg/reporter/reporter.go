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
    "path"
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
    dir            string
    archiveDir     string
    secretsDir     string
    reportFilePath string
    builder        *Builder
    db             *database.Database
    log            logrus.FieldLogger
}

func New(dir, archiveDir, appURL string, enableDebugOutput bool, db *database.Database, log logrus.FieldLogger) *Reporter {
    secretsDir := filepath.Join(dir, "secrets")
    reportFilePath := filepath.Join(dir, "report.html")

    return &Reporter{
        dir:            dir,
        archiveDir:     archiveDir,
        secretsDir:     secretsDir,
        reportFilePath: reportFilePath,
        builder:        NewBuilder(appURL, enableDebugOutput, db, log),
        db:             db,
        log:            log,
    }
}

func (r *Reporter) PrepareReport() (err error) {

    var data *reportData
    data, err = r.builder.buildReportData()
    if err != nil {
        return errors.WithMessage(err, "unable to build report data")
    }

    // Create report HTML file
    if err = r.createReportFile(data); err != nil {
        return errors.WithMessage(err, "unable to create report file")
    }

    // Output debug message
    if data.Secrets != nil {
        r.log.Debugf("found %d secrets", len(data.Secrets))
    } else {
        r.log.Debug("found no secrets")
        return
    }

    // Create secrets directory
    if err = r.outputSecretFiles(data); err != nil {
        return errors.WithMessage(err, "unable to create secret files")
    }

    // Copy report directory archive
    r.log.Debugf("copying %s to %s ...", r.dir, r.archiveDir)
    if err = copy.Copy(r.dir, r.archiveDir); err != nil {
        return errors.Wrapv(err, "", r.dir, r.archiveDir)
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

func (r *Reporter) outputSecretFiles(data *reportData) (err error) {
    if err = os.MkdirAll(r.secretsDir, 0700); err != nil {
        return errors.Wrapv(err, "unable to create secrets directory", r.secretsDir)
    }

    for _, sData := range data.Secrets {

        // Paths
        secretDir := filepath.Join(r.secretsDir, sData.ID)

        // Create directory if necenssary
        if err = os.MkdirAll(secretDir, 0700); err != nil {
            return errors.Wrapv(err, "unable to create secret directory", secretDir)
        }

        err = r.outputSecretValueFile(sData)
        if err != nil {
            err = errors.Wrap(err, "unable to create file")
            return
        }

        err = r.outputSecretMetadataFiles(sData)
        if err != nil {
            err = errors.Wrap(err, "unable to create file")
            return
        }
    }

    return
}

func (r *Reporter) outputSecretValueFile(sData secretData) (err error) {
    var savePath string
    var fileContents string
    savePath, fileContents, err = r.getSecretValueContents(sData)
    if err != nil {
        err = errors.WithMessage(err, "unable to get file contents")
    }

    if savePath == "" {
        // We don't have a finding that has the code taking up the whole file
        return
    }

    // Paths
    secretDir := filepath.Join(r.secretsDir, sData.ID)
    filePath := filepath.Join(secretDir, path.Base(savePath))

    // Create file
    var f *os.File
    f, err = os.Create(filePath)
    if err != nil {
        err = errors.Wrapv(err, "unable to create file", filePath)
        return
    }
    defer f.Close()

    _, err = f.WriteString(fileContents)
    if err != nil {
        err = errors.Wrapv(err, "unable to write to file", filePath)
        return
    }

    return
}

func (r *Reporter) outputSecretMetadataFiles(sData secretData) (err error) {
    dirPath := filepath.Join(r.secretsDir, sData.ID)
    filePath := filepath.Join(dirPath, fmt.Sprintf("secret-%s.yaml", sData.ID))

    // Create directory if necenssary
    if err = os.MkdirAll(dirPath, 0700); err != nil {
        return errors.Wrapv(err, "unable to create secret directory", dirPath)
    }

    var bytes []byte
    bytes, err = yaml.Marshal(sData)
    if err != nil {
        return errors.WithMessage(err, "unable to marshal secret into YAML")
    }

    err = ioutil.WriteFile(filePath, bytes, 0644)
    if err != nil {
        return errors.WithMessagev(err, "unable to write secret", sData.ID, filePath)
    }

    return
}

func (r *Reporter) getSecretValueContents(sData secretData) (fileContents, savePath string, err error) {
    for _, findingData := range sData.Findings {
        if findingData.CodeIsFile {
            fileContents = findingData.Code
            savePath = findingData.FilePath
            break
        }
    }
    return
}
