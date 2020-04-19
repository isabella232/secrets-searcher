package cmd

import (
    apppkg "github.com/pantheon-systems/search-secrets/pkg/app"
    "github.com/pantheon-systems/search-secrets/pkg/dbug"
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    "github.com/pantheon-systems/search-secrets/pkg/stats"
    "github.com/spf13/cobra"
    "os"
)

const (
    appName    = "search-secrets"
    appURL     = "https://github.com/pantheon-systems/search-secrets"
    dateFormat = "2006-01-02"
)

var (
    rootCmd = &cobra.Command{
        Use:   appName,
        Short: "Search for sensitive information stored in Pantheon git repositories.",
        Run:   runApp,
    }
    exitCode = 0
)

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        log.Error("unable to execute application")
        exitCode = 2
    }

    os.Exit(exitCode)
}

func init() {
    initArgs()
    initLogging()
    cobra.OnInitialize(must(initConfig), must(configureLogging))
}

func runApp(*cobra.Command, []string) {
    errors.CatchPanicAndLogIt(log)

    var app *apppkg.App
    var appConfig *apppkg.Config
    var err error

    log.Info("=== Search Secrets is starting")
    if dbug.Cnf.Enabled {
        log.Info("DEV MODE ENABLED")
    }

    appConfig, err = buildAppConfig()
    if err != nil {
        log.Fatal(errors.WithMessage(err, "unable to create app config"))
    }

    app, err = apppkg.New(appConfig)
    if err != nil {
        log.Fatal(errors.WithMessage(err, "unable to create app"))
    }

    // EXECUTE APP
    err = app.Execute()
    if err != nil {
        log.Fatal(errors.WithMessage(err, "unable to execute app"))
    }

    if cfg.NonZero && stats.SecretsFoundCount > 0 {
        exitCode = 3
    }
}

func must(do func() error) func() {
    return func() {
        err := do()
        if err != nil {
            errors.ErrorLogger(log, err).Fatal("initialization error")
        }
    }
}
