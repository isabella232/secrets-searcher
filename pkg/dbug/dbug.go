package dbug

type (
    Config struct {
        Enabled bool `mapstructure:"enabled"`

        EnableInteract bool `mapstructure:"enable-interact"`

        EnableCodePhase   bool `mapstructure:"enable-code-phase"`
        EnableSearchPhase bool `mapstructure:"enable-search-phase"`
        EnableReportPhase bool `mapstructure:"enable-report-phase"`

        Filter FilterConfig `mapstructure:"filter"`
    }
    FilterConfig struct {
        Processor string `mapstructure:"processor"`
        Repo      string `mapstructure:"repo"`
        Commit    string `mapstructure:"commit"`
        Path      string `mapstructure:"path"`
        Line      int    `mapstructure:"line"`
    }
)

var Cnf Config
