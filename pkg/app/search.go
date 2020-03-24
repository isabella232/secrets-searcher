package app

import (
    codepkg "github.com/pantheon-systems/search-secrets/pkg/code"
    "github.com/pantheon-systems/search-secrets/pkg/database"
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    filterpkg "github.com/pantheon-systems/search-secrets/pkg/filter"
    finderpkg "github.com/pantheon-systems/search-secrets/pkg/finder"
    "github.com/pantheon-systems/search-secrets/pkg/finder/trufflehog"
    reporterpkg "github.com/pantheon-systems/search-secrets/pkg/reporter"
    "github.com/pantheon-systems/search-secrets/pkg/secret"
    "github.com/sirupsen/logrus"
    "path/filepath"
)

type Search struct {
    code         *codepkg.Code
    finder       *finderpkg.Finder
    secretParser *secret.Parser
    log          *logrus.Logger
    reporter     *reporterpkg.Reporter
}

func NewSearch(githubToken, organization, outputDir string, truffleHogCmd, repos, reasons []string, skipEntropy bool, log *logrus.Logger) (*Search, error) {

    // Directories
    outputDirAbs, err := filepath.Abs(outputDir)
    if err != nil {
        errors.Fatal(log, errors.Wrapv(err, "invalid output dir", outputDir))
    }
    codeDir := filepath.Join(outputDirAbs, "code")
    dbDir := filepath.Join(outputDirAbs, "db")
    reportDir := filepath.Join(outputDirAbs, "report")

    // Create database
    db, err := database.New(dbDir)
    if err != nil {
        errors.Fatal(log, errors.WithMessage(err, "unable to create Database"))
    }

    // Create filter
    filter := filterpkg.New(repos, reasons, skipEntropy)

    // Create code
    code, err := codepkg.New(githubToken, organization, codeDir, filter, db, log)
    if err != nil {
        errors.Fatal(log, errors.WithMessage(err, "unable to create code"))
    }

    // Create TruffleHog driver
    driver, err := trufflehog.New(truffleHogCmd, log)
    if err != nil {
        return nil, errors.Wrap(err, "unable to connect to GitHub API")
    }

    // Create finder
    finder, err := finderpkg.New(driver, code, filter, db, log)
    if err != nil {
        errors.Fatal(log, errors.WithMessage(err, "unable to create finder"))
    }

    // Create secret parser
    secretParser := secret.NewParser(filter, db, log)

    // Create reporter
    reporter := reporterpkg.New(reportDir, db, log)

    return &Search{
        code:         code,
        finder:       finder,
        secretParser: secretParser,
        reporter:     reporter,
        log:          log,
    }, nil
}

func (s *Search) Execute() (err error) {
    err = s.code.PrepareRepos()
    if err != nil {
        return
    }

    err = s.finder.PrepareFindings()
    if err != nil {
        return
    }

    err = s.secretParser.PrepareSecrets()
    if err != nil {
        return
    }

    err = s.reporter.PrepareReport()
    if err != nil {
        return
    }

    return
}
