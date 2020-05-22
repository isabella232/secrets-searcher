package reporter

//go:generate go run github.com/wlbr/templify -p reporter -o template_report.go source/report.gohtml

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/pantheon-systems/search-secrets/pkg/stats"

	"github.com/pantheon-systems/search-secrets/pkg/manip"

	"github.com/otiai10/copy"
	"github.com/pantheon-systems/search-secrets/pkg/database"
	"github.com/pantheon-systems/search-secrets/pkg/errors"
	"github.com/pantheon-systems/search-secrets/pkg/logg"
	"github.com/pantheon-systems/search-secrets/pkg/source"
	"gopkg.in/yaml.v2"
)

var (
	templateFuncs = template.FuncMap{
		"stringRepeat": func(width int, str string) template.HTML {
			return template.HTML(strings.Repeat(str, width))
		},
	}
)

type (
	Reporter struct {
		ReportDir         string
		ReportArchivesDir string
		secretsDir        string
		reportFilePath    string
		enablePreReports  bool
		preReportInterval time.Duration
		builder           *builder
		db                *database.Database
		log               logg.Logg
		reporterState
	}
	reporterState struct {
		prepLock         *sync.Mutex
		stopPreReporting chan struct{}
		preparedFS       bool
	}
	SecretGrouper func(secretData *SecretData) (result string)
	SecretFilter  func(secretData *SecretData) (result bool)
)

func New(reportDir, reportArchivesDir, appURL string, enableDebugOutput, enablePreReports bool, preReportInterval time.Duration, secretIDFilter *manip.SliceFilter, metadataProvider source.ProviderI, stats *stats.Stats, db *database.Database, log logg.Logg) *Reporter {
	secretsDir := filepath.Join(reportDir, "secrets")
	reportFilePath := filepath.Join(reportDir, "report.html")

	builderGroupBy := defaultGroupBy
	builderFilter := defaultFilter(secretIDFilter)
	builder := newBuilder(appURL, enableDebugOutput, reportDir, secretsDir, builderGroupBy, builderFilter, metadataProvider, stats, db, log)

	return &Reporter{
		ReportDir:         reportDir,
		ReportArchivesDir: reportArchivesDir,
		secretsDir:        secretsDir,
		reportFilePath:    reportFilePath,
		enablePreReports:  enablePreReports,
		preReportInterval: preReportInterval,
		builder:           builder,
		db:                db,
		log:               log,
		reporterState: reporterState{
			prepLock:         &sync.Mutex{},
			stopPreReporting: make(chan struct{}),
		},
	}
}

func (r *Reporter) PrepareFilesystem() (err error) {
	if r.preparedFS {
		return
	}

	r.log.Debug("resetting report directory ...")

	if err = os.RemoveAll(r.ReportDir); err != nil {
		return errors.Wrapv(err, "unable to delete report directory", r.ReportDir)
	}
	if err = os.MkdirAll(r.ReportDir, 0700); err != nil {
		return errors.Wrapv(err, "unable to create report directory", r.ReportDir)
	}
	if err = os.MkdirAll(r.ReportArchivesDir, 0700); err != nil {
		return errors.Wrapv(err, "unable to create report archives directory", r.ReportDir)
	}
	return
}

func (r *Reporter) RunPreReporting() {
	if !r.enablePreReports {
		return
	}

	// The loop
	go func() {
		for {
			select {
			case <-r.stopPreReporting:
				return
			default:
				r.log.Debug("creating final report ...")
				if err := r.PrepareReport(true, false, false); err != nil {
					panic(err.Error())
				}

				r.log.Infof("created final report, will again in %s ...", r.preReportInterval)

				time.Sleep(r.preReportInterval)
			}
		}
	}()
}

func (r *Reporter) PrepareFinalReport() (err error) {
	r.log.Info("creating final report ...")

	if err = r.PrepareReport(true, true, true); err != nil {
		err = errors.WithMessage(err, "unable to prepare final report")
		return
	}

	return
}

func (r *Reporter) PrepareReport(createFile, createSecretFiles, createArchive bool) (err error) {
	r.prepLock.Lock()
	defer r.prepLock.Unlock()

	var data *reportData
	data, err = r.builder.buildReportData()
	if err != nil {
		err = errors.WithMessage(err, "unable to build report data")
		return
	}

	// Create report HTML file
	if createFile {
		if err = r.createReportFile(data); err != nil {
			err = errors.WithMessage(err, "unable to create report file")
			return
		}
	}

	if data.Secrets == nil {
		return
	}

	// Create secrets directory
	if createSecretFiles {
		if err = r.createSecretFiles(data); err != nil {
			return errors.WithMessage(err, "unable to create secret files")
		}
	}

	// Copy report directory archive
	if createArchive {
		now := time.Now()
		nowString := now.Format("2006-01-02_15-04-05")
		archiveDirname := fmt.Sprintf("report-%s", nowString)
		archiveDir := filepath.Join(r.ReportArchivesDir, archiveDirname)
		r.log.Debugf("copying %s to %s ...", r.ReportDir, archiveDir)
		if err = copy.Copy(r.ReportDir, archiveDir); err != nil {
			return errors.Wrapv(err, "", r.ReportDir, r.ReportArchivesDir)
		}
	}

	return
}

func (r *Reporter) Filter(fnc SecretFilter) {
	r.builder.filter = fnc
}

func (r *Reporter) GroupBy(fnc SecretGrouper) {
	r.builder.groupBy = fnc
}

func (r *Reporter) createReportFile(data *reportData) (err error) {
	tmpFilePath := r.reportFilePath + ".tmp"
	var saveFile *os.File
	if saveFile, err = os.Create(tmpFilePath); err != nil {
		return errors.Wrapv(err, "unable to create report file", tmpFilePath)
	}

	var tmpl *template.Template
	tmpl = template.New("report").Funcs(templateFuncs)
	if tmpl, err = tmpl.Parse(template_reportTemplate()); err != nil {
		return err
	}
	if err = tmpl.Execute(saveFile, data); err != nil {
		return errors.Wrapv(err, "unable to execute template and save report", tmpFilePath)
	}

	if err = os.Rename(tmpFilePath, r.reportFilePath); err != nil {
		return errors.Wrapv(err, "unable to move save file into position to finish report",
			tmpFilePath, r.reportFilePath)
	}

	return
}

func (r *Reporter) createSecretFiles(data *reportData) (err error) {
	if err = os.MkdirAll(r.secretsDir, 0700); err != nil {
		return errors.Wrapv(err, "unable to create secrets directory", r.secretsDir)
	}

	for _, secrets := range data.Secrets {
		for _, sData := range secrets {

			// Paths
			secretDir := filepath.Join(r.secretsDir, sData.ID)

			// Create secret directory if necenssary
			if err = os.MkdirAll(secretDir, 0700); err != nil {
				return errors.Wrapv(err, "unable to create secret directory", secretDir)
			}

			err = r.outputSecretValueFileAndAddLink(sData)
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
	}

	return
}

func (r *Reporter) outputSecretValueFileAndAddLink(sData *SecretData) (err error) {
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
	defer func() { _ = f.Close() }()

	_, err = f.WriteString(sData.Value)
	if err != nil {
		err = errors.Wrapv(err, "unable to write to file", sData.ValueFilePath)
		return
	}

	return
}

func (r *Reporter) outputSecretMetadataFile(sData *SecretData, secretDir string) (err error) {

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

func defaultFilter(secretIDFilter *manip.SliceFilter) (result SecretFilter) {
	return func(secretData *SecretData) (result bool) {
		return secretIDFilter.Includes(secretData.ID)
	}
}

func defaultGroupBy(secretData *SecretData) (result string) {
	var keyVal string
	for _, findingExtra := range secretData.Findings[0].Extras {
		if findingExtra.Key == "setter-key-value" {
			keyVal = findingExtra.Value
			keyVal = strings.ReplaceAll(keyVal, "-", "")
			keyVal = strings.ReplaceAll(keyVal, "_", "")
			keyVal = strings.ReplaceAll(keyVal, ".", "")
			keyVal = strings.ToLower(keyVal)
			break
		}
	}

	var pieces []string
	pieces = append(pieces, fmt.Sprintf("Rule: \"%s\"", secretData.Findings[0].ProcessorName))
	if keyVal != "" {
		pieces = append(pieces, fmt.Sprintf("Variable Name: \"%s\"", keyVal))
	}

	return strings.Join(pieces, " / ")
}
