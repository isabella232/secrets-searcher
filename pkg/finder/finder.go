package finder

import (
	"fmt"
	"github.com/pantheon-systems/search-secrets/pkg/code"
	"github.com/pantheon-systems/search-secrets/pkg/database"
	"github.com/pantheon-systems/search-secrets/pkg/database/enum/reason"
	"github.com/pantheon-systems/search-secrets/pkg/errors"
	"github.com/pantheon-systems/search-secrets/pkg/filter"
	diffpkg "github.com/pantheon-systems/search-secrets/pkg/secret/diff"
	"github.com/sirupsen/logrus"
	"strings"
)

var regexps = map[reason.ReasonEnum]string{
	reason.SlackToken{}.New():           "(xox[p|b|o|a]-[0-9]{12}-[0-9]{12}-[0-9]{12}-[a-z0-9]{32})",
	reason.RSAPrivateKey{}.New():        "-----BEGIN RSA PRIVATE KEY-----",
	reason.SSHPrivateKeyOpenSSH{}.New(): "-----BEGIN OPENSSH PRIVATE KEY-----",
	reason.SSHPrivateKeyDSA{}.New():     "-----BEGIN DSA PRIVATE KEY-----",
	reason.SSHPrivateKeyEC{}.New():      "-----BEGIN EC PRIVATE KEY-----",
	reason.PGPPrivateKeyBlock{}.New():   "-----BEGIN PGP PRIVATE KEY BLOCK-----",
	reason.FacebookOauth{}.New():        "[f|F][a|A][c|C][e|E][b|B][o|O][o|O][k|K].*['|\"][0-9a-f]{32}['|\"]",
	reason.TwitterOauth{}.New():         "[t|T][w|W][i|I][t|T][t|T][e|E][r|R].*['|\"][0-9a-zA-Z]{35,44}['|\"]",
	reason.GitHub{}.New():               "[g|G][i|I][t|T][h|H][u|U][b|B].*['|\"][0-9a-zA-Z]{35,40}['|\"]",
	reason.GoogleOauth{}.New():          "(\"client_secret\":\"[a-zA-Z0-9-_]{24}\")",
	reason.AWSAPIKey{}.New():            "AKIA[0-9A-Z]{16}",
	reason.HerokuAPIKey{}.New():         "[h|H][e|E][r|R][o|O][k|K][u|U].*[0-9A-F]{8}-[0-9A-F]{4}-[0-9A-F]{4}-[0-9A-F]{4}-[0-9A-F]{12}",
	reason.GenericSecret{}.New():        "[s|S][e|E][c|C][r|R][e|E][t|T].*['|\"][0-9a-zA-Z]{32,45}['|\"]",
	reason.GenericAPIKey{}.New():        "[a|A][p|P][i|I][_]?[k|K][e|E][y|Y].*['|\"][0-9a-zA-Z]{32,45}['|\"]",
	reason.SlackWebhook{}.New():         "https://hooks.slack.com/services/T[a-zA-Z0-9_]{8}/B[a-zA-Z0-9_]{8}/[a-zA-Z0-9_]{24}",
	reason.GCPServiceAccount{}.New():    "\"type\": \"service_account\"",
	reason.TwilioApiKey{}.New():         "SK[a-z0-9]{32}",
	reason.PasswordInUrl{}.New():        "[a-zA-Z]{3,10}://[^/\\s:@]{3,20}:[^/\\s:@]{3,20}@.{1,100}[\"'\\s]",
}

type (
	Finder struct {
		driver Driver
		code   *code.Code
		filter *filter.Filter
		db     *database.Database
		log    *logrus.Logger
	}
	Driver interface {
		GetFindings(cloneDir string, useEntropy bool, reasonRegexps *ReasonRegexps, out chan *DriverFinding) error
	}
	DriverFinding struct {
		Branch       string
		Commit       string
		CommitHash   string
		Diff         string
		Path         string
		PrintDiff    string
		Reason       string
		StringsFound []string
	}
	ReasonRegexps map[string]string
)

func New(driver Driver, code *code.Code, filter *filter.Filter, db *database.Database, log *logrus.Logger) (finder *Finder, err error) {
	finder = &Finder{
		driver: driver,
		code:   code,
		filter: filter,
		db:     db,
		log:    log,
	}

	return
}

func (f *Finder) PrepareFindings() (err error) {
	if f.db.TableExists(database.FindingTable) {
		f.log.Warn("finding table already exists, skipping")
		return
	}

	repos, err := f.db.GetRepos()
	if err != nil {
		return err
	}

	for _, repo := range repos {
		if ! f.filter.Repos.Include(repo.Name) {
			continue
		}
		err = f.prepareFindingsForRepo(repo)
		if err != nil {
			return err
		}
	}

	return
}


func (f *Finder) prepareFindingsForRepo(repo *database.Repo) (err error) {
	var reasons []reason.ReasonEnum
	if f.filter.Reasons.Enabled {
		for _, reasonValue := range f.filter.Reasons.Values {
			r := reason.NewReasonFromValue(reasonValue)
			if r == nil {
				return errors.Errorv("unknown commit reason in filter", reasonValue)
			}
			reasons = append(reasons, r)
		}
	} else {
		reasons = reason.ReasonEnumValues()
	}

	// Build regexp map and entropy flag
	reasonRegexps := ReasonRegexps{}
	useEntropy := false
	for _, r := range reasons {
		// Entropy doesn't have regexp
		entropyReason := reason.Entropy{}.New()
		if r == entropyReason {
			useEntropy = true
			continue
		}

		regexp, ok := regexps[r]
		if ! ok {
			return errors.Errorv("unable to find regular expression for reason", r.Value())
		}

		reasonRegexps[r.Value()] = regexp
	}

	cloneDir := f.code.CloneDir(repo)

	out := make(chan *DriverFinding)
	errs := make(chan error, 1)
	go func() {
		defer close(out)
		defer close(errs)
		err := f.driver.GetFindings(cloneDir, useEntropy, &reasonRegexps, out)
		if err != nil {
			errs <- err
		}
	}()

	for driverFinding := range out {
		reason := reason.NewReasonFromValue(driverFinding.Reason)
		if reason == nil {
			return errors.Errorv("unknown commit reason from driver", reason)
		}

		finding := &database.Finding{
			ID:          database.CreateHashID(fmt.Sprintf("%+v\n", driverFinding)),
			RepoID:      repo.ID,
			Branch:      driverFinding.Branch,
			Commit:      driverFinding.Commit,
			CommitHash:  driverFinding.CommitHash,
			Diff:        driverFinding.Diff,
			Path:        driverFinding.Path,
			PrintDiff:   driverFinding.PrintDiff,
			ReasonValue: reason.Value(),
			Reason:      reason,
		}
		if err = f.db.Write(database.FindingTable, finding.ID, finding); err != nil {
			return
		}

		diff := diffpkg.New(finding.Diff)
		for i, stringFound := range driverFinding.StringsFound {
			stringFoundTrimmed := strings.TrimSpace(stringFound)

			// Find line
			diff.IncrementUntil(func(line *diffpkg.Line) bool { return line.CodeContains(stringFoundTrimmed) })
			lineI := diff.LineI

			id := database.CreateHashID(fmt.Sprintf("%s-%d", finding.ID, i))
			findingString := &database.FindingString{
				ID:        id,
				FindingID: finding.ID,
				String:    stringFoundTrimmed,
				Index:     i,
				StartLine: lineI,
			}
			if err = f.db.Write(database.FindingStringTable, findingString.ID, findingString); err != nil {
				return
			}
		}
	}

	return
}
