package cmd

import (
    "fmt"
    "github.com/mitchellh/go-homedir"
    "github.com/mitchellh/mapstructure"
    "github.com/pantheon-systems/search-secrets/pkg/app"
    "github.com/pantheon-systems/search-secrets/pkg/database/enum/processor_type"
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    "github.com/pantheon-systems/search-secrets/pkg/rule"
    "github.com/pantheon-systems/search-secrets/pkg/rule/processor"
    "github.com/pantheon-systems/search-secrets/pkg/structures"
    "github.com/sirupsen/logrus"
    "github.com/spf13/cobra"
    "github.com/spf13/pflag"
    "github.com/spf13/viper"
    "os"
    "path/filepath"
    "strings"
    "time"
)

const (
    appName            = "search-secrets"
    configFileName     = "." + appName
    configFileExt      = "yaml"
    configFileBasename = configFileName + "." + configFileExt

    dateFormat = "2006-01-02"
)

var (
    rootCmd = &cobra.Command{
        Use:   appName,
        Short: "Search for sensitive information stored in Pantheon git repositories.",
        Run:   run,
    }
    cfg     config
    cfgFile string
    vpr     *viper.Viper
    log     *logrus.Logger
)

type (
    config struct {
        LogLevel           string       `mapstructure:"log-level"`
        GithubToken        string       `mapstructure:"github-token"`
        OutputDir          string       `mapstructure:"output-dir"`
        Organization       string       `mapstructure:"organization"`
        Repos              []string     `mapstructure:"repos"`
        Refs               []string     `mapstructure:"refs"`
        EarliestDate       time.Time    `mapstructure:"earliest-date"`
        LatestDate         time.Time    `mapstructure:"latest-date"`
        EarliestCommit     string       `mapstructure:"earliest-commit"`
        LatestCommit       string       `mapstructure:"latest-commit"`
        RuleConfigs        []ruleConfig `mapstructure:"rules"`
        WhitelistPathMatch []string     `mapstructure:"whitelist-path-match"`
        WhitelistSecretIDs []string     `mapstructure:"whitelist-secret-id"`
    }
    ruleConfig struct {
        Name               string        `mapstructure:"name"`
        Processor          string        `mapstructure:"processor"`
        RegexString        string        `mapstructure:"regex"`
        PEMType            string        `mapstructure:"pem-type"`
        EntropyConfig      entropyConfig `mapstructure:"entropy"`
        WhitelistCodeMatch []string      `mapstructure:"whitelist-code-match"`
    }
    entropyConfig struct {
        Charset          string  `mapstructure:"charset"`
        LengthThreshold  int     `mapstructure:"length-threshold"`
        EntropyThreshold float64 `mapstructure:"entropy-threshold"`
    }
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

func initArgs() {
    flags := rootCmd.PersistentFlags()

    flags.StringVar(
        &cfgFile,
        "config",
        "",
        fmt.Sprintf("config file (default is $HOME/.%s.%s)", appName, configFileExt),
    )

    flags.String(
        "log-level",
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
        "repos",
        []string{},
        "Only search these repos.",
    )

    flags.StringSlice(
        "branches",
        []string{},
        "Only search these references (branch names or full references like \"refs/tags/tag1\").",
    )

    flags.String(
        "earliest-date",
        time.Time{}.Format(dateFormat),
        "Only search commits on or after this date.",
    )

    flags.String(
        "latest-date",
        time.Now().Format(dateFormat),
        "Only search commits on or before this date.",
    )

    flags.String(
        "earliest-commit",
        "",
        "Only search this and commits after this commit. Only makes sense when searching a single repo.",
    )

    flags.String(
        "latest-commit",
        "",
        "Only search this and commits before this commit. Only makes sense when searching a single repo",
    )

    flags.StringSlice(
        "whitelist-path-match",
        []string{},
        "Whitelist files with paths that match these patterns.",
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

    opts := viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
        mapstructure.StringToTimeHookFunc(dateFormat),
        mapstructure.StringToSliceHookFunc(","),
    ))
    if err := vpr.Unmarshal(&cfg, opts); err != nil {
        errors.Fatal(log, errors.Wrap(err, "unable to unmarshal config"))
    }
}

func configureLogging() {
    // Log level
    logLevel, err := logrus.ParseLevel(cfg.LogLevel)
    if err != nil {
        errors.Fatal(log, errors.Wrapv(err, "invalid value for log-level: ", cfg.LogLevel))
    }
    log.SetLevel(logLevel)

    // Formatter
    var fm logrus.Formatter = &logrus.TextFormatter{}
    log.SetFormatter(fm)
}

func run(*cobra.Command, []string) {
    if cfg.Organization == "" {
        errors.Fatal(log, errors.New("organization is required"))
    }
    if cfg.GithubToken == "" {
        errors.Fatal(log, errors.New("github-token is required"))
    }

    var rules, err = buildRules(cfg.RuleConfigs)
    var earliestTime = cfg.EarliestDate
    var latestTime = cfg.LatestDate.Add(24 * time.Hour).Add(-1 * time.Second)

    var whitelistPath structures.RegexpSet
    whitelistPath, err = structures.NewRegexpSetFromStrings(cfg.WhitelistPathMatch)
    if err != nil {
        return
    }

    whitelistSecretIDSet := structures.NewSet(cfg.WhitelistSecretIDs)

    search, err := app.NewSearch(cfg.GithubToken, cfg.Organization, cfg.OutputDir, cfg.Repos, cfg.Refs, rules, earliestTime, latestTime, cfg.EarliestCommit, cfg.LatestCommit, whitelistPath, whitelistSecretIDSet, log)
    if err != nil {
        log.Fatal(errors.WithMessage(err, "unable to create search app"))
    }

    if err := search.Execute(); err != nil {
        log.Fatal(errors.WithMessage(err, "unable to execute search app"))
    }
}

func buildRules(ruleConfigs []ruleConfig) (result []*rule.Rule, err error) {
    for _, ruleConf := range ruleConfigs {
        var proc rule.Processor
        proc, err = buildProcessor(ruleConf)
        if err != nil {
            return
        }

        result = append(result, rule.New(ruleConf.Name, proc))
    }
    return
}

func buildProcessor(ruleConf ruleConfig) (result rule.Processor, err error) {
    switch ruleConf.Processor {
    case processor_type.Regex{}.New().Value():
        result, err = processor.NewRegexProcessor(ruleConf.RegexString, ruleConf.WhitelistCodeMatch)
    case processor_type.PEM{}.New().Value():
        result = processor.NewPEMProcessor(ruleConf.PEMType, log)
    case processor_type.Entropy{}.New().Value():
        ec := ruleConf.EntropyConfig
        result = processor.NewEntropyProcessor(ec.Charset, ec.LengthThreshold, ec.EntropyThreshold)
    default:
        err = errors.Errorv("unknown search type", ruleConf.Processor)
    }
    return
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
