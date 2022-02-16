package cmd

import (
	"fmt"
	"os"

	apppkg "github.com/pantheon-systems/secrets-searcher/pkg/app"
	"github.com/pantheon-systems/secrets-searcher/pkg/app/config"
	"github.com/pantheon-systems/secrets-searcher/pkg/errors"
)

const (
	exitCodeOK   = 0
	exitCodeErr  = 2
	exitCodeFail = 3
)

func Execute() {
	if passed, err := execute(); err != nil {
		fmt.Printf("ERROR: %s\n", err.Error())
		os.Exit(exitCodeErr)
	} else if !passed {
		os.Exit(exitCodeFail)
	}
	os.Exit(exitCodeOK)
}

func execute() (passed bool, err error) {
	defer errors.CatchPanicSetErr(&err, "unable to run application")

	// Build app config
	var appCfg *config.AppConfig
	appCfg, err = config.BuildConfig(os.Args, os.Environ())
	if err != nil {
		err = errors.WithMessage(err, "unable to create config")
		return
	}

	// Build app
	var app *apppkg.App
	app, err = apppkg.New(appCfg)
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
