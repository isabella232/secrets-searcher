package cmd

import (
    "fmt"
    "github.com/pantheon-systems/search-secrets/pkg/app/source_provider"
    "github.com/pantheon-systems/search-secrets/pkg/dbug"
    "github.com/pantheon-systems/search-secrets/pkg/finder"
    "github.com/pantheon-systems/search-secrets/pkg/finder/processor/entropy"
    "github.com/pantheon-systems/search-secrets/pkg/finder/processor/pem"
    "github.com/pantheon-systems/search-secrets/pkg/finder/processor/regex"
    "github.com/pantheon-systems/search-secrets/pkg/finder/processor/url"
    "github.com/spf13/cobra"
    "os"
    "path/filepath"
    "regexp"
    "strings"
    "time"

    "github.com/mitchellh/mapstructure"
    apppkg "github.com/pantheon-systems/search-secrets/pkg/app"
    "github.com/pantheon-systems/search-secrets/pkg/app/processor_type"
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    "github.com/pantheon-systems/search-secrets/pkg/structures"
    "github.com/sirupsen/logrus"
    "github.com/spf13/pflag"
    "github.com/spf13/viper"
)

var cfg config

type (
    // Root config
    config struct {

        // Output and workflow
        LogLevel    string `mapstructure:"log-level"`
        OutputDir   string `mapstructure:"output-dir"`
        NonZero     bool   `mapstructure:"non-zero"`
        Interactive bool   `mapstructure:"interactive"`

        // Source config
        SkipSourcePrep bool         `mapstructure:"skip-source-prep"`
        Source         sourceConfig `mapstructure:"source"`

        // Finder config
        ProcessorConfig    []processorConfig `mapstructure:"processors"`
        Refs               []string          `mapstructure:"refs"`
        EarliestDate       time.Time         `mapstructure:"earliest-date"`
        LatestDate         time.Time         `mapstructure:"latest-date"`
        WhitelistPathMatch []string          `mapstructure:"whitelist-path-match"`
        WhitelistSecretIDs []string          `mapstructure:"whitelist-secret-ids"`
        WhitelistSecretDir string            `mapstructure:"whitelist-secret-dir"`

        // Reporting config
        SkipReportSecrets       bool `mapstructure:"skip-report-secrets"`
        EnableReportDebugOutput bool `mapstructure:"enable-report-debug-output"`

        // Finder config
        ChunkSize           int           `mapstructure:"chunk-size"`
        WorkerCount         int           `mapstructure:"worker-count"`
        CommitSearchTimeout time.Duration `mapstructure:"commit-search-timeout"`

        // Dev config
        Dbug dbug.Config `mapstructure:"dbug"`
    }

    // Source config
    sourceConfig struct {
        Provider string `mapstructure:"provider"`

        // Common
        Repos        []string `mapstructure:"repos"`
        ExcludeRepos []string `mapstructure:"exclude-repos"`

        // Local
        LocalDir string `mapstructure:"dir"`

        // GitHub
        APIToken     string `mapstructure:"api-token"`
        User         string `mapstructure:"user"`
        Organization string `mapstructure:"organization"`
        SkipForks    bool   `mapstructure:"exclude-forks"`
    }

    // Processor config
    processorConfig struct {

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

    // Entropy processor config
    entropyConfig struct {
        Charset          string  `mapstructure:"charset"`
        LengthThreshold  int     `mapstructure:"length-threshold"`
        EntropyThreshold float64 `mapstructure:"entropy-threshold"`
    }
)

func buildAppConfig() (result *apppkg.Config, err error) {
    // Processor
    var processors []finder.Processor
    for _, conf := range cfg.ProcessorConfig {
        var proc finder.Processor
        proc, err = buildProcessor(conf)
        if err != nil {
            err = errors.WithMessagev(err, "unable to build processor", proc.Name())
            return
        }

        processors = append(processors, proc)
    }

    // Latest time - convert date to datetime
    var latestTime = cfg.LatestDate.Add(24 * time.Hour).Add(-1 * time.Second)

    // Whitelist paths
    var whitelistPath structures.RegexpSet
    whitelistPath, err = structures.NewRegexpSetFromStrings(cfg.WhitelistPathMatch)
    if err != nil {
        log.Fatal(errors.WithMessagev(err, "unable to create regexp set from `whitelist-path-match` parameter", cfg.WhitelistPathMatch))
    }

    // Build set of secrets to whitelist
    whitelistSecretIDSet := structures.NewSet(cfg.WhitelistSecretIDs)
    if cfg.WhitelistSecretDir != "" {
        if err = appendSecretsFromWhitelistDir(&whitelistSecretIDSet); err != nil {
            log.Fatal(errors.WithMessage(err, "unable to create app"))
        }
    }

    sourceConfig := &apppkg.SourceConfig{
        SourceProvider: cfg.Source.Provider,
        GithubToken:    cfg.Source.APIToken,
        Organization:   cfg.Source.Organization,
        Repos:          cfg.Source.Repos,
        ExcludeRepos:   cfg.Source.ExcludeRepos,
        ExcludeForks:   cfg.Source.SkipForks,
        LocalDir:       cfg.Source.LocalDir,
    }
    result = &apppkg.Config{
        SkipSourcePrep:          cfg.SkipSourcePrep,
        Interactive:             cfg.Interactive,
        OutputDir:               cfg.OutputDir,
        Refs:                    cfg.Refs,
        Processors:              processors,
        EarliestTime:            cfg.EarliestDate,
        LatestTime:              latestTime,
        WhitelistPath:           whitelistPath,
        WhitelistSecretIDSet:    whitelistSecretIDSet,
        AppURL:                  appURL,
        EnableReportDebugOutput: cfg.EnableReportDebugOutput,
        ChunkSize:               cfg.ChunkSize,
        WorkerCount:             cfg.WorkerCount,
        CommitSearchTimeout:     cfg.CommitSearchTimeout,
        SourceConfig:            sourceConfig,
        LogWriter:               logWriter,
        Log:                     log,
    }

    return
}

// A subset of command parameters that can overwrite configuration values
func initArgs() {
    flags := rootCmd.PersistentFlags()

    // Config file
    flags.String(
        "config",
        "",
        "Config file location")

    // Root config
    flags.String(
        "log-level",
        logrus.InfoLevel.String(),
        fmt.Sprintf("How detailed should the log be? Valid values: %s.", strings.Join(validLogLevels(), ", ")))
    flags.Bool(
        "skip-source-prep",
        false,
        "If true, repos will be cloned, and existing repos fetched, before searching.")
    flags.String(
        "output-dir",
        "./output",
        "Output directory.")
    flags.Bool(
        "non-zero",
        false,
        "If set to true, the command will exit with a non-zero exit code if secrets are found.")
    flags.Bool(
        "interactive",
        true,
        "If false, progress bars will not appear, only log messages.")

    // Source config
    flags.String(
        "source.api-token",
        "",
        "API token to use when querying for a list of repos to clone for the search.")

    flags.Int(
        "chunk-size",
        500,
        "Number of commits each worker handles.")
    flags.Int(
        "worker-count",
        8,
        "How many concurrent search workers.")
    flags.Duration(
        "commit-search-timeout",
        5*time.Second,
        "How long to wait for a single commit to be searched.")

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

    // Report config
    flags.Bool(
        "skip-report-secrets",
        false,
        "If true, the `./output/report/secret-[SECRETID].yaml` files will not be generated.")
    flags.Bool(
        "enable-report-debug-output",
        false,
        "If true, the report will include some debug output (works even if `dbug.enabled` is false).")

    // Debug config
    flags.Bool(
        "dbug.enabled",
        false,
        "If true, certain development features are enabled.")
}

// Build the cfg variable
func initConfig() (err error) {
    vpr := viper.New()

    // Config file
    var cfgFile string
    cfgFile, err = rootCmd.PersistentFlags().GetString("config")
    if err != nil {
        err = errors.Wrap(err, "unable to get \"config\" command parameter value")
        return
    }
    if cfgFile == "" {
        err = errors.Wrap(err, "\"config\" parameter required")
        return
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
        if err = vpr.BindPFlag(f.Name, f); err != nil {
            err = errors.Wrapv(err, "unable to bind flag", f.Name)
            return
        }
    }

    // Read config file
    if err = vpr.ReadInConfig(); err != nil {
        err = errors.Wrap(err, "unable to read config file")
        return
    }

    // Unmarshal config into object
    opts := viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
        mapstructure.StringToTimeHookFunc(dateFormat),
        mapstructure.StringToSliceHookFunc(","),
        mapstructure.StringToTimeDurationHookFunc(),
    ))
    if err = vpr.Unmarshal(&cfg, opts); err != nil {
        err = errors.Wrap(err, "unable to unmarshal config")
        return
    }

    // debug config
    dbug.Cnf = cfg.Dbug

    // Validate parameters
    validateParameters()

    return
}

// Validate config
func validateParameters() {

    // Output and workflow
    validationAssertTrue(cfg.LogLevel != "", "log-level", "")
    validationAssertTrue(cfg.OutputDir != "", "output-dir", "")

    validationAssertTrue(cfg.ChunkSize > 0, "chunk-size", "")
    validationAssertTrue(cfg.WorkerCount > 0, "worker-count", "")

    // `source` config
    validateSourceParameter()

    // `processors` config
    validationAssertTrue(len(cfg.ProcessorConfig) > 0, "processors", "")
    procNameRegistry := structures.NewSet(nil)
    for i, conf := range cfg.ProcessorConfig {
        configNameBase := fmt.Sprintf("processors.%d.", i)
        validationAssertTrue(conf.Name != "", configNameBase+"name", "")

        // Unique processor name
        validationAssertTrue(!procNameRegistry.Contains(conf.Name), configNameBase+"name",
            "parameter `%s` has a duplicate name to a previous processor")
        procNameRegistry.Add(conf.Name)

        validationAssertTrue(conf.Processor != "", configNameBase+"processor", "")
        processorType := processor_type.NewProcessorTypeFromValue(conf.Processor)
        validationAssertTrue(processorType != nil, configNameBase+"processor",
            "the value for parameter `%s` is not a valid processor type")

        validateProcessorParameter(conf, configNameBase)
    }
}

func validateSourceParameter() {
    switch cfg.Source.Provider {
    case source_provider.GitHub{}.New().Value():
        // TODO Implement user for GitHub provider, not just organization
        validationAssertTrue(cfg.Source.User == "", "source.user",
            "currently, only `source.organization` is supported, `%s` is not")
        validationAssertTrue(cfg.Source.APIToken != "", "source.api-token", "")
        validationAssertTrue(cfg.Source.Organization != "", "source.organization", "")
    case source_provider.Local{}.New().Value():
        errors.LogErrorThenDie(log, errors.New("currently only \"github\" is supported as `source.provider`"))
        validateSourceLocalDirParameter()
    default:
        errors.LogErrorThenDie(log, errors.New("currently only \"github\" is supported as `source.provider`"))
    }
}

func validateSourceLocalDirParameter() {
    validationAssertTrue(cfg.Source.LocalDir != "", "source.dir", "")
    localAbs, err := filepath.Abs(cfg.Source.LocalDir)
    if err != nil {
        errors.LogErrorThenDie(log, errors.Errorv("unable to get abs path", cfg.Source.LocalDir))
    }
    outputDirAbs, err := filepath.Abs(cfg.OutputDir)
    if err != nil {
        errors.LogErrorThenDie(log, errors.Errorv("unable to get abs path", cfg.OutputDir))
    }
    validationAssertTrue(!strings.HasPrefix(localAbs, outputDirAbs), "source.dir",
        "your `%s` dir cannot be within your output dir")
}

func validateProcessorParameter(conf processorConfig, configNameBase string) {
    switch conf.Processor {
    case processor_type.Regex{}.New().Value():
        noExtraValues := conf.EntropyConfig == entropyConfig{} && conf.PEMType == ""
        validationAssertTrue(noExtraValues, configNameBase+"regex", "invalid config on processor")
        validationAssertTrue(conf.RegexString != "", configNameBase+"regex", "")
    case processor_type.PEM{}.New().Value():
        noExtraValues := conf.EntropyConfig == entropyConfig{} && len(conf.WhitelistCodeMatch) == 0
        validationAssertTrue(noExtraValues, configNameBase+"pem", "invalid config on processor")
        validationAssertTrue(conf.PEMType != "", configNameBase+"pem-type", "")
        validationAssertTrue(pem.NewPEMTypeFromValue(conf.PEMType) != nil, configNameBase+"pem-type",
            "unknown PEM type in parameter `%s`")
    case processor_type.Entropy{}.New().Value():
        noExtraValues := conf.PEMType == "" && conf.RegexString == ""
        validationAssertTrue(noExtraValues, configNameBase+"pem", "invalid config on processor")
        entConfNameBase := configNameBase + "entropy."
        validationAssertTrue(conf.EntropyConfig.Charset != "", entConfNameBase+"charset", "")
    }
}

func buildProcessor(procConf processorConfig) (result finder.Processor, err error) {
    var whitelistCodeRes structures.RegexpSet
    if procConf.WhitelistCodeMatch != nil {
        whitelistCodeRes, err = structures.NewRegexpSetFromStrings(procConf.WhitelistCodeMatch)
        if err != nil {
            err = errors.WithMessagev(err, "unable to build whitelist code expressions for processor", procConf.Name)
            return
        }
    }

    switch procConf.Processor {
    case processor_type.URL{}.New().Value():
        result = url.NewProcessor(procConf.Name, &whitelistCodeRes)
    case processor_type.Regex{}.New().Value():
        re := regexp.MustCompile(procConf.RegexString)
        result = regex.NewProcessor(procConf.Name, re, &whitelistCodeRes)
    case processor_type.PEM{}.New().Value():
        pemType := pem.NewPEMTypeFromValue(procConf.PEMType)
        result = pem.NewProcessor(procConf.Name, pemType)
    case processor_type.Entropy{}.New().Value():
        ec := procConf.EntropyConfig
        result = entropy.NewProcessor(procConf.Name, ec.Charset, ec.LengthThreshold, ec.EntropyThreshold, &whitelistCodeRes, true)
    default:
        err = errors.Errorv("unknown search type", procConf.Processor)
    }
    return
}

func validationAssertTrue(valid bool, configName string, messageTemplate string) {
    if !valid {
        message := fmt.Sprintf(messageTemplate, configName)
        errors.LogErrorThenDie(log, errors.New(message))
    }
}

func appendSecretsFromWhitelistDir(secretIDSet *structures.Set) error {
    var reportedSecretFileMatch = regexp.MustCompile(`^secret-([0-9a-f]{5,40}).yaml$`)
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
