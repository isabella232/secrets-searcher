package cmd

import (
    "fmt"
    "github.com/mitchellh/mapstructure"
    apppkg "github.com/pantheon-systems/search-secrets/pkg/app"
    "github.com/pantheon-systems/search-secrets/pkg/database/enum/processor_type"
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    "github.com/pantheon-systems/search-secrets/pkg/finder/processor"
    "github.com/pantheon-systems/search-secrets/pkg/finder/rule"
    "github.com/pantheon-systems/search-secrets/pkg/structures"
    "github.com/sirupsen/logrus"
    "github.com/spf13/cobra"
    "github.com/spf13/pflag"
    "github.com/spf13/viper"
    "os"
    "path/filepath"
    "regexp"
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
        Run:   runApp,
    }
    cfg                     config
    cfgFile                 string
    vpr                     *viper.Viper
    log                     *logrus.Logger
    reportedSecretFileMatch = regexp.MustCompile(`^secret-([0-9a-f]{5,40}).yaml$`)
)

type (
    // Root config
    config struct {

        // Output and workflow
        LogLevel       string `mapstructure:"log-level"`
        SkipSourcePrep bool   `mapstructure:"skip-source-prep"`
        OutputDir      string `mapstructure:"output-dir"`

        // Source config
        Source sourceConfig `mapstructure:"source"`

        // Finder config
        RuleConfigs        []ruleConfig `mapstructure:"rules"`
        Refs               []string     `mapstructure:"refs"`
        EarliestDate       time.Time    `mapstructure:"earliest-date"`
        LatestDate         time.Time    `mapstructure:"latest-date"`
        EarliestCommit     string       `mapstructure:"earliest-commit"`
        LatestCommit       string       `mapstructure:"latest-commit"`
        WhitelistPathMatch []string     `mapstructure:"whitelist-path-match"`
        WhitelistSecretIDs []string     `mapstructure:"whitelist-secret-id"`
        WhitelistSecretDir string       `mapstructure:"whitelist-secret-dir"`
    }

    // Source config
    sourceConfig struct {
        Provider     string   `mapstructure:"provider"`
        APIToken     string   `mapstructure:"api-token"`
        User         string   `mapstructure:"user"`
        Organization string   `mapstructure:"organization"`
        Repos        []string `mapstructure:"repos"`
    }

    // Rule config
    ruleConfig struct {

        // Setup
        Name      string `mapstructure:"name"`
        Processor string `mapstructure:"processor"`

        // General
        WhitelistCodeMatch []string `mapstructure:"whitelist-code-match"`

        // "regex" processor
        RegexString string `mapstructure:"regex"`

        // "pem" processor
        PEMType string `mapstructure:"pem-type"`

        // "entropy" processor config
        EntropyConfig entropyConfig `mapstructure:"entropy"`
    }

    // Entropy rule config
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

    // Config file
    flags.StringVar(
        &cfgFile,
        "config",
        "",
        "Config file location")

    // Root config
    flags.String(
        "log-level",
        logrus.DebugLevel.String(),
        fmt.Sprintf("How detailed should the log be? Valid values: %s.", strings.Join(validLogLevels(), ", ")))
    flags.Bool(
        "skip-source-prep",
        false,
        "If true, repos will be cloned, and existing repos fetched, before searching.")
    flags.String(
        "output-dir",
        "./output",
        "Output directory.")

    // Source config
    flags.String(
        "source.api-token",
        "",
        "API token to use when querying for a list of repos to clone for the search.")

    // Finder config
    flags.StringSlice(
        "refs",
        []string{},
        "Only search these references (branch names or full references like \"refs/tags/tag1\").")
    flags.String(
        "earliest-date",
        time.Time{}.Format(dateFormat),
        "Only search commits on or after this date.")
    flags.String(
        "latest-date",
        time.Now().Format(dateFormat),
        "Only search commits on or before this date.")
    flags.String(
        "earliest-commit",
        "",
        "Only search this and commits after this commit. Only makes sense when searching a single repo.")
    flags.String(
        "latest-commit",
        "",
        "Only search this and commits before this commit. Only makes sense when searching a single repo")
    flags.StringSlice(
        "whitelist-path-match",
        []string{},
        "Whitelist files with paths that match these patterns.")
    flags.StringSlice(
        "whitelist-secret-ids",
        []string{},
        "Whitelist files with these IDs.")
    flags.StringSlice(
        "whitelist-secret-dir",
        []string{},
        "If a corresponding `secret-[SECRETID].yaml` file is found in this directory, that secret will be whitelisted.")
}

func initLogging() {
    log = logrus.New()
    log.SetOutput(os.Stdout)
    logrus.SetFormatter(&logrus.TextFormatter{})
}

func initConfig() {
    vpr = viper.New()

    // Config file
    if cfgFile == "" {
        errors.Fatal(log, errors.New("`config` parameter is required"))
    }
    vpr.SetConfigFile(cfgFile)

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

func validateParameters() {

    // Output and workflow
    validationAssertTrue(cfg.LogLevel != "", "log-level", "")
    validationAssertTrue(cfg.OutputDir != "", "output-dir", "")

    // `source` config
    // FIXME Implement "local" provider and others if necessary
    validationAssertTrue(cfg.Source.Provider == "github", "source.provider",
        "currently only \"github\" is supported as `%s`")
    // FIXME Implement user for GitHub provider
    validationAssertTrue(cfg.Source.User == "", "source.user",
        "currently, only `source.organization` is supported, `%s` is not")
    validationAssertTrue(cfg.Source.APIToken != "", "source.api-token", "")
    validationAssertTrue(cfg.Source.Organization != "", "source.organization", "")

    // `rules` config
    validationAssertTrue(len(cfg.RuleConfigs) > 0, "rules", "")
    ruleNameRegistry := structures.NewSet(nil)
    for i, ruleConf := range cfg.RuleConfigs {
        configNameBase := fmt.Sprintf("rules.%d.", i)

        validationAssertTrue(ruleConf.Name != "", configNameBase+"name", "")

        // Unique rule name
        validationAssertTrue(!ruleNameRegistry.Contains(ruleConf.Name), configNameBase+"name",
            "parameter `%s` has a duplicate name to a previous rule")
        ruleNameRegistry.Add(ruleConf.Name)

        // Processor
        validationAssertTrue(ruleConf.Processor != "", configNameBase+"processor", "")
        processorType := processor_type.NewProcessorTypeFromValue(ruleConf.Processor)
        validationAssertTrue(processorType != nil, configNameBase+"processor",
            "the value for parameter `%s` is not a valid processor type")

        // Processor-specific validation
        switch ruleConf.Processor {
        case processor_type.Regex{}.New().Value():
            validationAssertTrue(ruleConf.RegexString != "", configNameBase+"regex", "")
        case processor_type.PEM{}.New().Value():
            validationAssertTrue(ruleConf.PEMType != "", configNameBase+"pem-type", "")
        case processor_type.Entropy{}.New().Value():
            entConfNameBase := configNameBase + "entropy."
            validationAssertTrue(ruleConf.EntropyConfig.Charset != "", entConfNameBase+"charset", "")
            validationAssertTrue(ruleConf.EntropyConfig.Charset != "", entConfNameBase+"length-threshold", "")
            validationAssertTrue(ruleConf.EntropyConfig.Charset != "", entConfNameBase+"entropy-threshold", "")
        }
    }
}

func runApp(*cobra.Command, []string) {
    validateParameters()

    var app *apppkg.App
    var err error

    var rules []*rule.Rule
    rules, err = buildRules(cfg.RuleConfigs)
    var earliestTime = cfg.EarliestDate
    var latestTime = cfg.LatestDate.Add(24 * time.Hour).Add(-1 * time.Second)

    var whitelistPath structures.RegexpSet
    whitelistPath, err = structures.NewRegexpSetFromStrings(cfg.WhitelistPathMatch)
    if err != nil {
        log.Fatal(errors.WithMessagev(err, "unable to create regexp set from `whitelist-path-match` parameter", cfg.WhitelistPathMatch))
    }

    // Build set of secrets to whitelist
    whitelistSecretIDSet := structures.NewSet(cfg.WhitelistSecretIDs)
    if cfg.WhitelistSecretDir != "" {
        if err = appendSecretsFromWhitelistDir(&whitelistSecretIDSet, cfg.WhitelistSecretDir); err != nil {
            log.Fatal(errors.WithMessage(err, "unable to create app"))
        }
    }

    app, err = apppkg.New(
        cfg.SkipSourcePrep,
        cfg.Source.APIToken,
        cfg.Source.Organization,
        cfg.OutputDir,
        cfg.Source.Repos,
        cfg.Refs,
        rules,
        earliestTime,
        latestTime,
        cfg.EarliestCommit,
        cfg.LatestCommit,
        whitelistPath,
        whitelistSecretIDSet,
        log,
    )

    if err != nil {
        log.Fatal(errors.WithMessage(err, "unable to create app"))
    }

    if err := app.Execute(); err != nil {
        log.Fatal(errors.WithMessage(err, "unable to execute app"))
    }
}

func validationAssertTrue(valid bool, configName string, messageTemplate string) {
    if !valid {
        message := fmt.Sprintf(messageTemplate, configName)
        errors.Fatal(log, errors.New(message))
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
    var whitelistRes structures.RegexpSet
    if ruleConf.WhitelistCodeMatch != nil {
        whitelistRes, err = structures.NewRegexpSetFromStrings(ruleConf.WhitelistCodeMatch)
        if err != nil {
            return
        }
    }

    switch ruleConf.Processor {
    case processor_type.Regex{}.New().Value():
        result, err = processor.NewRegexProcessor(ruleConf.RegexString, &whitelistRes, log)
    case processor_type.PEM{}.New().Value():
        result = processor.NewPEMProcessor(ruleConf.PEMType, &whitelistRes, log)
    case processor_type.Entropy{}.New().Value():
        ec := ruleConf.EntropyConfig
        result = processor.NewEntropyProcessor(ec.Charset, ec.LengthThreshold, ec.EntropyThreshold, &whitelistRes, true, log)
    default:
        err = errors.Errorv("unknown search type", ruleConf.Processor)
    }
    return
}

func appendSecretsFromWhitelistDir(secretIDSet *structures.Set, whitelistSecretDir string) error {
    return filepath.Walk(cfg.WhitelistSecretDir,
        func(filePath string, info os.FileInfo, err error) error {
            if err != nil {
                return err
            }
            if info.IsDir() {
                return nil
            }
            matches := reportedSecretFileMatch.FindStringSubmatch(info.Name())
            if len(matches) == 0 {
                return nil
            }
            secretIDSet.Add(matches[1])
            return nil
        })
}

func validLogLevels() []string {
    var logLevels []string
    for _, l := range logrus.AllLevels {
        logLevels = append(logLevels, l.String())
    }
    return logLevels
}
