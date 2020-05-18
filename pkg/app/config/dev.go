package config

type (
	DevConfig struct {
		Filter DevFilterConfig `param:"filter"`
	}
	DevFilterConfig struct {
		Processor string `param:"processor"`
		Repo      string `param:"repo"`
		Commit    string `param:"commit"`
		Path      string `param:"path"`
		Line      int    `param:"line"`
	}
)
