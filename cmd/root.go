package cmd

import (
    "fmt"
    "github.com/mitchellh/go-homedir"
    "github.com/pantheon-systems/search-secrets/pkg/app"
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
    appName            = "search-secrets"
    configFileName     = "." + appName
    configFileExt      = "yaml"
    configFileBasename = configFileName + "." + configFileExt
)

var (
    rootCmd = &cobra.Command{
        Use:   appName,
        Short: "Search for sensitive information stored in Pantheon git repositories.",
        Run:   run,
    }
    logLevelValue string
    cfgFile       string
    vpr           *viper.Viper
    log           *logrus.Logger
)

func init() {
    cobra.OnInitialize(initConfig, configureLogging)

    initArgs()
    initLogging()
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        errors.Fatal(log, errors.Wrap(err, "unable to execute application"))
    }
}

func run(*cobra.Command, []string) {
    githubToken := vpr.GetString("github-token")
    outputDir := vpr.GetString("output-dir")
    organization := vpr.GetString("organization")
    truffleHogCmd := vpr.GetStringSlice("trufflehog-cmd")
    reposFilter := vpr.GetStringSlice("repos")
    reasonsFilter := vpr.GetStringSlice("reasons")
    skipEntropy := vpr.GetBool("skip-entropy")

    // Validate
    if organization == "" {
        errors.Fatal(log, errors.New("organization is required"))
    }
    if githubToken == "" {
        errors.Fatal(log, errors.New("github-token is required"))
    }

    search, err := app.NewSearch(githubToken, organization, outputDir, truffleHogCmd, reposFilter, reasonsFilter, skipEntropy, log)
    if err != nil {
        log.Fatal(errors.WithMessage(err, "unable to create search app"))
    }

    if err := search.Execute(); err != nil {
        log.Fatal(errors.WithMessage(err, "unable to execute search app"))
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

    flags.String(
        "github-token",
        "",
        "GitHub API token.",
    )

    flags.String(
        "output-dir",
        "./output",
        "Output directory.",
    )

    flags.String(
        "organization",
        "",
        "Organization to search.",
    )

    flags.StringSlice(
        "trufflehog-cmd",
        []string{"./thog.sh"},
        "TruffleHog command.",
    )

    flags.StringSlice(
        "repos",
        []string{},
        "Only search these repos.",
    )

    flags.StringSlice(
        "reasons",
        []string{},
        "Only search these reasons.",
    )

    flags.Bool(
        "skip-entropy",
        false,
        "Use every reason except for \"entropy\". If \"reasons\" is passed, this argument is ignored.",
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
    } else {
        touchConfigFile()
        vpr.AddConfigPath("$HOME")
        vpr.SetConfigName(configFileName)
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
    configFile := filepath.Join(hd, configFileBasename)
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
