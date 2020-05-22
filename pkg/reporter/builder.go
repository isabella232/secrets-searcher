package reporter

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/pantheon-systems/search-secrets/pkg/stats"

	"github.com/pantheon-systems/search-secrets/pkg/database"
	"github.com/pantheon-systems/search-secrets/pkg/errors"
	"github.com/pantheon-systems/search-secrets/pkg/logg"
	"github.com/pantheon-systems/search-secrets/pkg/manip"
	"github.com/pantheon-systems/search-secrets/pkg/source"
)

const (
	defaultGroup = "default"
)

type (
	builder struct {
		appURL            string
		enableDebugOutput bool
		reportDir         string
		secretsDir        string
		groupBy           SecretGrouper
		filter            SecretFilter
		sourceProvider    source.ProviderI
		stats             *stats.Stats
		db                *database.Database
		log               logg.Logg
	}
	reportData struct {
		ReportDate        time.Time
		AppLink           linkData
		Repos             []string
		DbgEnabled        bool
		Secrets           map[string][]*SecretData
		EnableDebugOutput bool
		SecretCountMsg    string
		DefaultGroup      string
	}
	SecretData struct {
		ID            string         `yaml:"secret-id"`
		Value         string         `yaml:"value"`
		ValueLen      int            `yaml:"-"`
		ValueFilePath string         `yaml:"-"`
		Extras        []*extraData   `yaml:"extras"`
		Finding       *findingData   `yaml:"-"`
		Findings      []*findingData `yaml:"findings"`
	}
	findingData struct {
		ID                  string       `yaml:"finding-id"`
		ProcessorName       string       `yaml:"processor"`
		RepoName            string       `yaml:"-"`
		RepoFullLink        linkData     `yaml:"repo"`
		CommitHash          string       `yaml:"-"`
		CommitHashLink      linkData     `yaml:"commit"`
		CommitHashLinkShort linkData     `yaml:"-"`
		CommitDate          time.Time    `yaml:"commit-date"`
		CommitAuthorEmail   string       `yaml:"-"`
		CommitAuthorName    string       `yaml:"commit-author"`
		FilePath            string       `yaml:"-"`
		FileLineLink        linkData     `yaml:"file-location"`
		FileLineLinkShort   linkData     `yaml:"-"`
		ColStartIndex       int          `yaml:"col-start-index"`
		ColEndIndex         int          `yaml:"col-end-index"`
		CodeShort           string       `yaml:"-"`
		Code                string       `yaml:"-"`
		CodeNoBreaks        string       `yaml:"-"`
		BeforeCode          string       `yaml:"-"`
		AfterCode           string       `yaml:"-"`
		FileBasename        string       `yaml:"file-basename"`
		CodeWithContext     string       `yaml:"code"`
		CodeShowGuide       bool         `yaml:"-"`
		Extras              []*extraData `yaml:"extras"`
	}
	extraData struct {
		Key    string    `yaml:"key"`
		Header string    `yaml:"-"`
		Value  string    `yaml:"value"`
		Code   bool      `yaml:"-"`
		URL    string    `yaml:"url"`
		Link   *linkData `yaml:"link"`
		Debug  bool
	}
	linkData struct {
		Label   string `yaml:"label"`
		URL     string `yaml:"url"`
		Tooltip string `yaml:"-"`
	}
)

func newBuilder(appURL string, enableDebugOutput bool, reportDir, secretsDir string, groupBy SecretGrouper, filter SecretFilter, sourceProvider source.ProviderI, stats *stats.Stats, db *database.Database, log logg.Logg) *builder {
	return &builder{
		appURL:            appURL,
		enableDebugOutput: enableDebugOutput,
		reportDir:         reportDir,
		secretsDir:        secretsDir,
		groupBy:           groupBy,
		filter:            filter,
		sourceProvider:    sourceProvider,
		stats:             stats,
		db:                db,
		log:               log,
	}
}

func (b *builder) groupedReportData() (secrets database.Secrets, findingsBySecret database.FindingGroups, findingExtrasByFindingID database.FindingExtraGroups, secretExtrasBySecretID database.SecretExtraGroups, err error) {
	var reportData *database.ReportData
	reportData, err = b.db.GetBaseReportData()
	if err != nil {
		return
	}
	secrets = reportData.Secrets

	// Findings by secret ID
	findingsBySecret = make(database.FindingGroups)
	for _, finding := range reportData.Findings {
		findingsBySecret[finding.SecretID] = append(findingsBySecret[finding.SecretID], finding)
	}

	// Finding extras by finding ID
	findingExtrasByFindingID = make(database.FindingExtraGroups)
	for _, findingExtra := range reportData.FindingExtras {
		findingExtrasByFindingID[findingExtra.FindingID] = append(findingExtrasByFindingID[findingExtra.FindingID], findingExtra)
	}

	// Secret extras by secret ID
	secretExtrasBySecretID = make(database.SecretExtraGroups)
	for _, secretExtra := range reportData.SecretExtras {
		secretExtrasBySecretID[secretExtra.SecretID] = append(secretExtrasBySecretID[secretExtra.SecretID], secretExtra)
	}

	return
}

func (b *builder) buildReportData() (result *reportData, err error) {
	b.log.Debug("getting list of secrets ...")
	var ok bool

	var (
		secrets                  database.Secrets
		findingsBySecret         database.FindingGroups
		findingExtrasByFindingID database.FindingExtraGroups
		secretExtrasBySecretID   database.SecretExtraGroups
	)
	secrets, findingsBySecret, findingExtrasByFindingID, secretExtrasBySecretID, err = b.groupedReportData()

	var secretDatas []*SecretData
	for _, secret := range secrets {
		var findings []*database.Finding
		findings, ok = findingsBySecret[secret.ID]
		if !ok {
			err = errors.Errorv("unable to find secret for finding group", secret.ID)
			return
		}

		var secretExtras database.SecretExtras
		secretExtras, _ = secretExtrasBySecretID[secret.ID]

		var secretData *SecretData
		secretData, err = b.buildSecretData(secret, secretExtras, findings, findingExtrasByFindingID)
		if err != nil {
			err = errors.WithMessage(err, "unable to build secret data")
			return
		}

		if !b.filter(secretData) {
			continue
		}

		secretDatas = append(secretDatas, secretData)
	}
	secretCount := len(secretDatas)

	secretGroups := map[string][]*SecretData{}
	for _, secretData := range secretDatas {
		group := b.groupBy(secretData)
		secretGroups[group] = append(secretGroups[group], secretData)
	}

	repos := manip.NewEmptyBasicSet()
	for _, secrets := range secretGroups {
		for _, reportSecret := range secrets {
			for _, reportFinding := range reportSecret.Findings {
				repos.Add(reportFinding.RepoName)
			}
		}
	}
	repoNames := repos.StringValues()
	sort.Strings(repoNames)

	var secretCountMsg = fmt.Sprintf("%d secrets", secretCount)

	result = &reportData{
		ReportDate:        time.Now(),
		AppLink:           linkData{URL: b.appURL, Label: b.appURL},
		Repos:             repoNames,
		EnableDebugOutput: b.enableDebugOutput,
		Secrets:           secretGroups,
		SecretCountMsg:    secretCountMsg,
		DefaultGroup:      defaultGroup,
	}

	return
}

func (b *builder) buildSecretData(secret *database.Secret, secretExtras database.SecretExtras, findings []*database.Finding, findingExtrasByFindingID database.FindingExtraGroups) (result *SecretData, err error) {
	var findingDatas []*findingData
	for _, finding := range findings {
		var findingExtras database.FindingExtras
		findingExtras, _ = findingExtrasByFindingID[finding.ID]

		var findingData *findingData
		findingData, err = b.buildFindingData(finding, findingExtras)
		if err != nil {
			err = errors.WithMessage(err, "unable to build finding data")
			return
		}

		findingDatas = append(findingDatas, findingData)
	}

	var secretExtraDatas []*extraData
	for _, secretExtra := range secretExtras {
		if !b.enableDebugOutput && secretExtra.Debug {
			continue
		}
		secretExtraDatas = append(secretExtraDatas, b.buildSecretExtraData(secretExtra))
	}

	// Link to the raw file
	var filePath string
	var fileBasename string
	for _, findingData := range findingDatas {
		if findingData.FileBasename != "" {
			fileBasename = findingData.FileBasename
			break
		}
	}
	if fileBasename != "" {
		filePath = filepath.Join(b.secretsDir, secret.ID, fileBasename)
		filePathRel := strings.TrimPrefix(filePath, b.reportDir)[1:]

		secretExtraDatas = append(secretExtraDatas, &extraData{
			Key:    "raw-file",
			Header: "Raw file",
			Link: &linkData{
				Label: fileBasename,
				URL:   filePathRel,
			},
		})
	}

	// Sort findings by commit date
	sort.Slice(findingDatas, func(i, j int) bool { return findingDatas[i].CommitDate.Before(findingDatas[j].CommitDate) })

	result = &SecretData{
		ID:            secret.ID,
		Value:         secret.Value,
		ValueLen:      len(secret.Value),
		ValueFilePath: filePath,
		Extras:        secretExtraDatas,
		Finding:       findingDatas[0],
		Findings:      findingDatas,
	}

	return
}

func (b *builder) buildFindingData(finding *database.Finding, findingExtras database.FindingExtras) (result *findingData, err error) {
	var commit *database.Commit
	commit, err = b.db.GetCommit(finding.CommitID)
	if err != nil {
		err = errors.WithMessage(err, "unable to get commit")
		return
	}

	var repo *database.Repo
	repoID := commit.RepoID
	repo, err = b.db.GetRepo(repoID)
	if err != nil {
		err = errors.WithMessage(err, "unable to get repo")
		return
	}

	var findingExtraDatas []*extraData
	for _, findingExtra := range findingExtras {
		if !b.enableDebugOutput && findingExtra.Debug {
			continue
		}
		findingExtraData := b.buildFindingExtraData(findingExtra)
		findingExtraDatas = append(findingExtraDatas, findingExtraData)
	}
	if b.enableDebugOutput {
		findingExtraDatas = append(findingExtraDatas, &extraData{
			Key:    "dev-filter",
			Header: "Dev filter",
			Value:  b.buildDevConfig(repo, commit, finding),
			Code:   true,
			Debug:  true,
		})
	}

	commitURL := b.sourceProvider.GetCommitURL(repo.Name, commit.CommitHash)
	commitLink := linkData{Label: commit.CommitHash, URL: commitURL}
	commitLinkShort := linkData{Label: commit.CommitHash[:7], URL: commitURL, Tooltip: commit.CommitHash}

	fileLineURL := b.sourceProvider.GetFileLineURL(repo.Name, commit.CommitHash, finding.Path,
		finding.StartLineNum, finding.EndLineNum)
	fileLineLabel, fileLineLabelShort := b.getFileLineLabels(finding)
	fileLineLink := linkData{Label: fileLineLabel, URL: fileLineURL}
	fileLineLinkShort := linkData{Label: fileLineLabelShort, URL: fileLineURL, Tooltip: fileLineLabel}

	repoURL := b.sourceProvider.GetRepoURL(repo.Name)
	repoLink := linkData{Label: repo.Name, URL: repoURL}

	codeNoBreaks := manip.MakeOneLine(finding.Code, "↵")

	result = &findingData{
		ID:                  finding.ID,
		ProcessorName:       finding.Processor,
		RepoName:            repo.Name,
		RepoFullLink:        repoLink,
		CommitHash:          commit.CommitHash,
		CommitHashLink:      commitLink,
		CommitHashLinkShort: commitLinkShort,
		CommitDate:          commit.Date,
		CommitAuthorName:    commit.AuthorName,
		CommitAuthorEmail:   commit.AuthorEmail,
		FilePath:            finding.Path,
		FileLineLink:        fileLineLink,
		FileLineLinkShort:   fileLineLinkShort,
		ColStartIndex:       finding.StartIndex,
		ColEndIndex:         finding.EndIndex,
		BeforeCode:          finding.BeforeCode,
		Code:                finding.Code,
		CodeNoBreaks:        codeNoBreaks,
		AfterCode:           finding.AfterCode,
		FileBasename:        finding.FileBasename,
		CodeWithContext:     finding.BeforeCode + finding.Code + finding.AfterCode,
		CodeShowGuide:       finding.StartLineNum == finding.EndLineNum,
		Extras:              findingExtraDatas,
	}

	return
}

func (b *builder) getFileLineLabels(finding *database.Finding) (label, labelShort string) {

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

func (b *builder) buildFindingExtraData(extra *database.FindingExtra) *extraData {
	var link *linkData
	if extra.URL != "" {
		link = b.buildExtraLink(extra.Value, extra.URL)
	}

	return &extraData{
		Key:    extra.Key,
		Header: extra.Header,
		Value:  extra.Value,
		Code:   extra.Code,
		Link:   link,
		Debug:  extra.Debug,
	}
}

func (b *builder) buildSecretExtraData(extra *database.SecretExtra) *extraData {
	var link *linkData
	if extra.URL != "" {
		link = b.buildExtraLink(extra.Value, extra.URL)
	}

	return &extraData{
		Key:    extra.Key,
		Header: extra.Header,
		Value:  extra.Value,
		Code:   extra.Code,
		Link:   link,
		Debug:  extra.Debug,
	}
}

func (b *builder) buildExtraLink(label, url string) (result *linkData) {
	if url != "" {
		return &linkData{label, url, ""}
	}
	return
}

func (b *builder) buildDevConfig(repo *database.Repo, commit *database.Commit, finding *database.Finding) (result string) {
	var sb strings.Builder
	fmt.Fprintf(&sb, "  filter:\n")
	fmt.Fprintf(&sb, "    repo: '%s'\n", repo.Name)
	fmt.Fprintf(&sb, "    processor: '%s'\n", finding.Processor)
	fmt.Fprintf(&sb, "    commit: '%s'\n", commit.CommitHash)
	fmt.Fprintf(&sb, "    path: '%s'\n", finding.Path)
	fmt.Fprintf(&sb, "    line: %d\n", finding.StartLineNum)

	return sb.String()
}
