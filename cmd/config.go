package cmd

import (
    "fmt"
    apppkg "github.com/pantheon-systems/search-secrets/pkg/app"
    "github.com/pantheon-systems/search-secrets/pkg/app/source_provider"
    "github.com/pantheon-systems/search-secrets/pkg/dbug"
    "github.com/pantheon-systems/search-secrets/pkg/finder"
    "github.com/pantheon-systems/search-secrets/pkg/finder/processor/entropy"
    "github.com/pantheon-systems/search-secrets/pkg/finder/processor/pem"
    "github.com/pantheon-systems/search-secrets/pkg/finder/processor/regex"
    "github.com/pantheon-systems/search-secrets/pkg/finder/processor/url"
    "os"
    "path/filepath"
    "regexp"
    "strings"
    "time"

    "github.com/mitchellh/mapstructure"
    "github.com/pantheon-systems/search-secrets/pkg/app/processor_type"
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    "github.com/pantheon-systems/search-secrets/pkg/structures"
    "github.com/sirupsen/logrus"
    "github.com/spf13/pflag"
    "github.com/spf13/viper"
)

const dateFormat = "2006-01-02"

var cfgUnmarshalOpt = viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
    mapstructure.StringToTimeHookFunc(dateFormat),
    mapstructure.StringToSliceHookFunc(","),
    mapstructure.StringToTimeDurationHookFunc(),
))

type (

    // Stores all config information from config file and command line flags
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
        ShowWorkersBars     bool          `mapstructure:"show-worker-bars"`

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

// Define command line flags
// Only a subset of config values can be overwritten with command line flags
func defineFlags() *pflag.FlagSet {
    flags := pflag.CommandLine

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
    flags.Bool(
        "show-worker-bars",
        false,
        "Show one progress bar per worker instead of one per repo.")

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
        time.Time{}.Format(dateFormat),
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
        "If true, the `./output/report/secrets/*/secret-*.yaml` files will not be generated.")
    flags.Bool(
        "enable-report-debug-output",
        false,
        "If true, the report will include some debug output (works even if `dbug.enabled` is false).")

    // Debug config
    flags.Bool(
        "dbug.enabled",
        false,
        "If true, certain development features are enabled.")

    // Show usage
    flags.Bool(
        "help",
        false,
        "Show command usage.")

    return flags
}

// Pull config information from config file and command line flags and save to `cfg`
func initConfig(args []string) (result config, err error) {
    vpr := viper.New()

    // Parse flags
    flags := defineFlags()
    _ = flags.Parse(args[1:]) // Will exit on error

    // Show help message
    if help, _ := flags.GetBool("help"); help {
        printHelp(flags)
        os.Exit(0)
    }

    // Bind CLI flags to viper
    if err = vpr.BindPFlags(flags); err != nil {
        err = errors.Wrap(err, "unable to bind flags")
    }

    // Bind config file to viper
    var cfgFile string
    if cfgFile = vpr.GetString("config"); cfgFile == "" {
        err = errors.Wrap(err, "\"config\" parameter required")
        return
    }
    vpr.SetConfigFile(cfgFile)

    // Parse all config
    if err = vpr.ReadInConfig(); err != nil {
        err = errors.Wrap(err, "unable to read config file")
        return
    }

    // Build config object
    if err = vpr.Unmarshal(&result, cfgUnmarshalOpt); err != nil {
        err = errors.Wrap(err, "unable to unmarshal config")
        return
    }

    // Validate parameters
    validateConfig(result)

    // Set dbug config into a singleton for easy access
    dbug.Cnf = result.Dbug

    return
}

// Validate config
func validateConfig(cfg config) {

    // Output and workflow
    validationAssertTrue(cfg.LogLevel != "", "log-level", "")
    validationAssertTrue(cfg.OutputDir != "", "output-dir", "")

    validationAssertTrue(cfg.ChunkSize > 0, "chunk-size", "")
    validationAssertTrue(cfg.WorkerCount > 0, "worker-count", "")

    // `source` config
    validateSourceParameter(cfg)

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

// Translate viper config into application config
func buildAppConfig(cfg config, log *logrus.Entry) (result *apppkg.Config, err error) {
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
    var latestTime time.Time
    if !cfg.LatestDate.IsZero() {
        latestTime = cfg.LatestDate.Add(24 * time.Hour).Add(-1 * time.Second)
    }

    // Whitelist paths
    var whitelistPath structures.RegexpSet
    whitelistPath, err = structures.NewRegexpSetFromStrings(cfg.WhitelistPathMatch)
    if err != nil {
        err = errors.WithMessagev(err, "unable to create regexp set from `whitelist-path-match` parameter", cfg.WhitelistPathMatch)
        return
    }

    // Build set of secrets to whitelist
    whitelistSecretIDSet := structures.NewSet(cfg.WhitelistSecretIDs)
    if cfg.WhitelistSecretDir != "" {
        if err = appendSecretsFromWhitelistDir(cfg, &whitelistSecretIDSet); err != nil {
            err = errors.WithMessage(err, "unable to appent whitelist secret IDs")
            return
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
        Processors:              processors,
        EarliestTime:            cfg.EarliestDate,
        LatestTime:              latestTime,
        WhitelistPath:           whitelistPath,
        WhitelistSecretIDSet:    whitelistSecretIDSet,
        AppURL:                  appURL,
        EnableReportDebugOutput: cfg.EnableReportDebugOutput,
        ChunkSize:               cfg.ChunkSize,
        WorkerCount:             cfg.WorkerCount,
        ShowWorkersBars:         cfg.ShowWorkersBars,
        CommitSearchTimeout:     cfg.CommitSearchTimeout,
        SourceConfig:            sourceConfig,
        NonZero:                 cfg.NonZero,
        Log:                     log,
    }

    return
}

func validateSourceParameter(cfg config) {
    switch cfg.Source.Provider {
    case source_provider.GitHub{}.New().Value():
        // TODO Implement user for GitHub provider, not just organization
        validationAssertTrue(cfg.Source.User == "", "source.user",
            "currently, only `source.organization` is supported, `%s` is not")
        validationAssertTrue(cfg.Source.APIToken != "", "source.api-token", "")
        validationAssertTrue(cfg.Source.Organization != "", "source.organization", "")
    case source_provider.Local{}.New().Value():
        fatal("currently only \"github\" is supported as `source.provider`")
        validateSourceLocalDirParameter(cfg)
    default:
        fatal("currently only \"github\" is supported as `source.provider`")
    }
}

func validateSourceLocalDirParameter(cfg config) {
    validationAssertTrue(cfg.Source.LocalDir != "", "source.dir", "")

    localAbs, err := filepath.Abs(cfg.Source.LocalDir)
    if err != nil {
        fatal(fmt.Sprintf("unable to get abs path of %s", cfg.Source.LocalDir))
    }

    outputDirAbs, err := filepath.Abs(cfg.OutputDir)
    if err != nil {
        fatal(fmt.Sprintf("unable to get abs path of %s", cfg.OutputDir))
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
        noExtraValues := conf.EntropyConfig == entropyConfig{}
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
        result = pem.NewProcessor(procConf.Name, pemType, &whitelistCodeRes)
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
        fatal(message)
    }
}

func appendSecretsFromWhitelistDir(cfg config, secretIDSet *structures.Set) error {
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

func printHelp(flags *pflag.FlagSet) {
    div := strings.Repeat("=", len(appDescription))
    fmt.Println("")
    fmt.Println(div)
    fmt.Println(appName)
    fmt.Println(appDescription)
    fmt.Println(appURL)
    fmt.Println(div)
    fmt.Println("")
    fmt.Println(flags.FlagUsages())
}

func logConfig(flags *pflag.FlagSet, log logrus.FieldLogger) {
    div := strings.Repeat("=", len(appDescription))
    fmt.Println("")
    fmt.Println(div)
    fmt.Println(appName)
    fmt.Println(appDescription)
    fmt.Println(appURL)
    fmt.Println(div)
    fmt.Println("")
    fmt.Println(flags.FlagUsages())
}

func fatal(message string) {
    fmt.Println(message)
    os.Exit(exitCodeErr)
}
