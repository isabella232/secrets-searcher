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
    "os"
    "path/filepath"
    "regexp"
    "strings"
    "time"

    "github.com/mitchellh/mapstructure"
    apppkg "github.com/pantheon-systems/search-secrets/pkg/app"
    "github.com/pantheon-systems/search-secrets/pkg/app/processor_type"
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    "github.com/pantheon-systems/search-secrets/pkg/logwriter"
    "github.com/pantheon-systems/search-secrets/pkg/structures"
    "github.com/sirupsen/logrus"
    "github.com/spf13/cobra"
    "github.com/spf13/pflag"
    "github.com/spf13/viper"
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
    cfg                     config
    cfgFile                 string
    vpr                     *viper.Viper
    log                     *logrus.Logger
    logWriter               *logwriter.LogWriter
    reportedSecretFileMatch = regexp.MustCompile(`^secret-([0-9a-f]{5,40}).yaml$`)
    exitCode                = 0
)

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
        ChunkSize   int `mapstructure:"chunk-size"`
        WorkerCount int `mapstructure:"worker-count"`

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

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        errors.Fatal(log, errors.Wrap(err, "unable to execute application"))
    }
    os.Exit(exitCode)
}

func init() {
    initArgs()
    initLogging()
    cobra.OnInitialize(initConfig, configureLogging)
}

func runApp(*cobra.Command, []string) {
    dbug.Cnf = cfg.Dbug
    if dbug.Cnf.Enabled {
        log.Warn("DEV MODE ENABLED")
    }

    validateParameters()

    var app *apppkg.App
    var err error

    var processors []finder.Processor
    processors, err = buildProcessors(cfg.ProcessorConfig)
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
        if err = appendSecretsFromWhitelistDir(&whitelistSecretIDSet); err != nil {
            log.Fatal(errors.WithMessage(err, "unable to create app"))
        }
    }

    app, err = apppkg.New(&apppkg.Config{
        SkipSourcePrep:          cfg.SkipSourcePrep,
        Interactive:             cfg.Interactive,
        OutputDir:               cfg.OutputDir,
        Refs:                    cfg.Refs,
        Processors:              processors,
        EarliestTime:            earliestTime,
        LatestTime:              latestTime,
        WhitelistPath:           whitelistPath,
        WhitelistSecretIDSet:    whitelistSecretIDSet,
        SkipReportSecrets:       cfg.SkipReportSecrets,
        AppURL:                  appURL,
        EnableReportDebugOutput: cfg.EnableReportDebugOutput,
        ChunkSize:               cfg.ChunkSize,
        WorkerCount:             cfg.WorkerCount,
        SourceConfig: &apppkg.SourceConfig{
            SourceProvider: cfg.Source.Provider,
            GithubToken:    cfg.Source.APIToken,
            Organization:   cfg.Source.Organization,
            Repos:          cfg.Source.Repos,
            ExcludeRepos:   cfg.Source.ExcludeRepos,
            ExcludeForks:   cfg.Source.SkipForks,
            LocalDir:       cfg.Source.LocalDir,
        },
        LogWriter: logWriter,
        Log:       log,
    })

    if err != nil {
        log.Fatal(errors.WithMessage(err, "unable to create app"))
    }

    if err := app.Execute(); err != nil {
        log.Fatal(errors.WithMessage(err, "unable to execute app"))
    }

    if cfg.NonZero && app.Stats.SecretsFoundCount > 0 {
        exitCode = 3
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

func initLogging() {
    log = logrus.New()
    log.SetOutput(os.Stdout)
    log.SetFormatter(&logrus.TextFormatter{})
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

    // Log file
    logFilePath := filepath.Join(cfg.OutputDir, "run.log")

    if _, statErr := os.Stat(logFilePath); os.IsNotExist(statErr) {
        var empty *os.File
        empty, err = os.Create(logFilePath)
        if err != nil {
            errors.Fatal(log, errors.Wrapv(err, "unable to create log file", logFilePath))
        }
        empty.Close()
    } else if statErr == nil {
        if err = os.Truncate(logFilePath, 0); err != nil {
            errors.Fatal(log, errors.Wrapv(err, "unable to truncate log file", logFilePath))
        }
    }
    logWriter, err = logwriter.New(logFilePath)
    if err != nil {
        errors.Fatal(log, errors.Wrapv(err, "unable to build log writer"))
    }
    log.SetOutput(logWriter)
}

func validateParameters() {

    // Output and workflow
    validationAssertTrue(cfg.LogLevel != "", "log-level", "")
    validationAssertTrue(cfg.OutputDir != "", "output-dir", "")

    validationAssertTrue(cfg.ChunkSize > 0, "chunk-size", "")
    validationAssertTrue(cfg.WorkerCount > 0, "worker-count", "")

    // `source` config
    switch cfg.Source.Provider {
    case source_provider.GitHub{}.New().Value():
        // TODO Implement user for GitHub provider, not just organization
        validationAssertTrue(cfg.Source.User == "", "source.user",
            "currently, only `source.organization` is supported, `%s` is not")
        validationAssertTrue(cfg.Source.APIToken != "", "source.api-token", "")
        validationAssertTrue(cfg.Source.Organization != "", "source.organization", "")
    case source_provider.Local{}.New().Value():
        errors.Fatal(log, errors.New("currently only \"github\" is supported as `source.provider`"))
        validateSourceLocalDirParameter()
    default:
        errors.Fatal(log, errors.New("currently only \"github\" is supported as `source.provider`"))
    }

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

        // Processor
        validationAssertTrue(conf.Processor != "", configNameBase+"processor", "")
        processorType := processor_type.NewProcessorTypeFromValue(conf.Processor)
        validationAssertTrue(processorType != nil, configNameBase+"processor",
            "the value for parameter `%s` is not a valid processor type")

        // Processor-specific validation
        switch conf.Processor {
        case processor_type.Regex{}.New().Value():
            validationAssertTrue(conf.RegexString != "", configNameBase+"regex", "")
        case processor_type.PEM{}.New().Value():
            validationAssertTrue(conf.PEMType != "", configNameBase+"pem-type", "")
            validationAssertTrue(pem.NewPEMTypeFromValue(conf.PEMType) != nil, configNameBase+"pem-type",
                "unknown PEM type in parameter `%s`")
        case processor_type.Entropy{}.New().Value():
            entConfNameBase := configNameBase + "entropy."
            validationAssertTrue(conf.EntropyConfig.Charset != "", entConfNameBase+"charset", "")
        }
    }
}

func validateSourceLocalDirParameter() {
    validationAssertTrue(cfg.Source.LocalDir != "", "source.dir", "")
    localAbs, err := filepath.Abs(cfg.Source.LocalDir)
    if err != nil {
        errors.Fatal(log, errors.Errorv("unable to get abs path", cfg.Source.LocalDir))
    }
    outputDirAbs, err := filepath.Abs(cfg.OutputDir)
    if err != nil {
        errors.Fatal(log, errors.Errorv("unable to get abs path", cfg.OutputDir))
    }
    validationAssertTrue(!strings.HasPrefix(localAbs, outputDirAbs), "source.dir",
        "your `%s` dir cannot be within your output dir")
}

func validationAssertTrue(valid bool, configName string, messageTemplate string) {
    if !valid {
        message := fmt.Sprintf(messageTemplate, configName)
        errors.Fatal(log, errors.New(message))
    }
}

func buildProcessors(procConfs []processorConfig) (result []finder.Processor, err error) {
    for _, conf := range procConfs {
        var proc finder.Processor
        proc, err = buildProcessor(conf)
        if err != nil {
            return
        }

        result = append(result, proc)
    }
    return
}

func buildProcessor(procConf processorConfig) (result finder.Processor, err error) {
    var whitelistCodeRes structures.RegexpSet
    if procConf.WhitelistCodeMatch != nil {
        whitelistCodeRes, err = structures.NewRegexpSetFromStrings(procConf.WhitelistCodeMatch)
        if err != nil {
            return
        }
    }

    switch procConf.Processor {
    case processor_type.URL{}.New().Value():
        result = url.NewProcessor(procConf.Name, &whitelistCodeRes)
    case processor_type.Regex{}.New().Value():
        result, err = regex.NewProcessor(procConf.Name, procConf.RegexString, &whitelistCodeRes)
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

func appendSecretsFromWhitelistDir(secretIDSet *structures.Set) error {
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
