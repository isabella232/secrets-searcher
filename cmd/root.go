package cmd

import (
    apppkg "github.com/pantheon-systems/search-secrets/pkg/app"
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    "github.com/sirupsen/logrus"
    "os"
)

const (
    appName        = "search-secrets"
    appDescription = "Search for sensitive information stored in Pantheon git repositories."
    appURL         = "https://github.com/pantheon-systems/search-secrets"

    exitCodeOK   = 0
    exitCodeErr  = 2
    exitCodeFail = 3
)

func Execute() {
    logger := initLogging()

    // EXECUTE APP
    passed, err := ExecuteArgs(os.Args, logger)

    if err != nil {
        log := logger.WithField("prefix", "init")
        errors.ErrLog(log, err).Error("unable to execute application")
        os.Exit(exitCodeErr)
    }

    if !passed {
        os.Exit(exitCodeFail)
    }

    os.Exit(exitCodeOK)
}

func ExecuteArgs(args []string, logger *logrus.Logger) (passed bool, err error) {
    errors.CatchPanicSetErr(&err, "unable to initialize application")

    // Parse config
    var cfg0 config
    cfg0, err = initConfig(args)
    if err != nil {
        err = errors.WithMessage(err, "unable to create config")
        return
    }

    // Configure logging
    var log *logrus.Entry
    log, err = configureLogging(logger, cfg0)
    if err != nil {
        err = errors.WithMessage(err, "unable to create config")
        return
    }

    // Build app config
    var appConfig *apppkg.Config
    appConfig, err = buildAppConfig(cfg0, log)
    if err != nil {
        err = errors.WithMessage(err, "unable to create app config")
        return
    }

    // Build app
    var app *apppkg.App
    app, err = apppkg.New(appConfig)
    if err != nil {
        err = errors.WithMessage(err, "unable to create app")
        return
    }

    // EXECUTE APP
    passed, err = app.Execute()
    if err != nil {
        err = errors.WithMessage(err, "unable to execute app")
        return
    }

    return
}
