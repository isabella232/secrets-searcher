package cmd

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/pantheon-systems/search-secrets/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strings"
)

const (
	appName       = "searchsecrets"
	configFileExt = "yaml"
)

var (
	rootCmd = &cobra.Command{
		Use:   appName,
		Short: "Search repositories for all resources across all configured contexts",
	}
	logLevelValue string
	cfgFile       string
	vpr           *viper.Viper
	log           *logrus.Logger
)

func init() {
	initArgs()
	initLogging()
	cobra.OnInitialize(initConfig, configureLogging)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		errors.Fatal(log, errors.Wrap(err, "unable to execute application"))
	}
}

func initArgs() {
	flags := rootCmd.PersistentFlags()

	flags.StringVar(
		&cfgFile,
		"config",
		"",
		fmt.Sprintf("config file (default is $HOME/.%s.%s)", appName, configFileExt),
	)

	flags.StringVarP(
		&logLevelValue,
		"log-level",
		"l",
		logrus.DebugLevel.String(),
		fmt.Sprintf("How detailed should the log be? Valid values: %s.", strings.Join(validLogLevels(), ", ")),
	)
}

func initLogging() {
	log = logrus.New()
	log.SetOutput(os.Stdout)
	logrus.SetFormatter(&logrus.TextFormatter{})
}

func initConfig() {
	vpr = viper.New()

	// Config file
	if cfgFile != "" {
		vpr.SetConfigFile(cfgFile)
		touchConfigFile()
	} else {
		return
		//vpr.AddConfigPath("$HOME'")
		//vpr.SetConfigName("." + appName)
	}
	vpr.SetConfigType(configFileExt)

	// Bind cobra and viper together
	var flags []*pflag.Flag
	for _, cmd := range append([]*cobra.Command{rootCmd}, rootCmd.Commands()...) {
		cmd.LocalFlags().VisitAll(func(f *pflag.Flag) {
			if f.Name != "config" {
				flags = append(flags, f)
			}
		})
	}
	for _, f := range flags {
		if err := vpr.BindPFlag(f.Name, f); err != nil {
			errors.Fatal(log, errors.Wrapv(err, "unable to bind flag", f.Name))
		}
	}

	// Read config
	if err := vpr.ReadInConfig(); err != nil {
		errors.Fatal(log, errors.Wrap(err, "unable to read config file"))
	}

	// Debug print
	for _, f := range flags {
		log.WithField("key", f.Name).WithField("value", vpr.Get(f.Name)).Info("config flag")
	}
}

func configureLogging() {
	// Log level
	logLevel, err := logrus.ParseLevel(logLevelValue)
	if err != nil {
		errors.Fatal(log, errors.Wrapv(err, "invalid value for log-level: ", logLevelValue))
	}
	log.SetLevel(logLevel)

	// Formatter
	var fm logrus.Formatter = &logrus.TextFormatter{}
	log.SetFormatter(fm)
}

func validLogLevels() []string {
	var logLevels []string
	for _, l := range logrus.AllLevels {
		logLevels = append(logLevels, l.String())
	}
	return logLevels
}

func touchConfigFile() {
	hd, err := homedir.Dir()
	if err != nil {
		errors.Fatal(log, errors.Wrap(err, "unable to find home directory"))
	}
	configFile := filepath.Join(hd, "."+appName)
	if _, err := os.Stat(configFile); err != nil && os.IsNotExist(err) {
		file, err := os.Create(configFile)
		if err != nil {
			errors.Fatal(log, errors.Wrapv(err, "unable to create config file", configFile))
		}
		file.Close()
	} else if err != nil {
		errors.Fatal(log, errors.Wrap(err, "unable to read config file"))
	}
}
