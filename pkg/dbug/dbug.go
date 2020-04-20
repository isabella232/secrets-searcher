package dbug

var Cnf Config

type (
    Config struct {
        Enabled           bool `mapstructure:"enabled"`
        EnableInteract    bool `mapstructure:"enable-interact"`
        EnableCodePhase   bool `mapstructure:"enable-code-phase"`
        EnableSearchPhase bool `mapstructure:"enable-search-phase"`
        EnableReportPhase bool `mapstructure:"enable-report-phase"`
        FilterConfig      `mapstructure:"filter"`
    }
    FilterConfig struct {
        Processor string `mapstructure:"processor"`
        Repo      string `mapstructure:"repo"`
        Commit    string `mapstructure:"commit"`
        Path      string `mapstructure:"path"`
        Line      int    `mapstructure:"line"`
    }
)
