package config

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/mitchellh/mapstructure"
	"github.com/pantheon-systems/secrets-searcher/pkg/app/vars"
	"github.com/pantheon-systems/secrets-searcher/pkg/errors"
	"github.com/pantheon-systems/secrets-searcher/pkg/manip"
	"github.com/pantheon-systems/secrets-searcher/pkg/valid"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Pull config information from config file and command line flags and save to "cfg" var
func BuildConfig(args, envVars []string) (result *AppConfig, err error) {

	// Parse flags
	var cfgFiles []string
	if cfgFiles, err = parseFlags(args[1:]); err != nil {
		err = errors.Wrap(err, "unable to parse flags")
		return
	}

	// Build config object
	result = NewAppConfig()

	// Merge in config file data
	if err = MergeInConfigFileData(result, cfgFiles); err != nil {
		err = errors.WithMessage(err, "unable to merge in config file data")
		return
	}

	// Merge in config file data
	if err = MergeInEnvVars(result, envVars); err != nil {
		err = errors.WithMessage(err, "unable to merge in env vars")
		return
	}

	return
}

// Pull config information from config file and command line flags and save to "cfg" var
func MergeInConfigFileData(cfg *AppConfig, cfgFiles []string) (err error) {

	// Bind config files to viper
	vpr := viper.New()
	for i, cfgFile := range cfgFiles {
		if err = validation.Validate(cfgFile, validation.Required, valid.ExistingFile); err != nil {
			err = errors.WithMessagef(err, "invalid value for \"config\" number %d: %s", i+1, cfgFile)
			return
		}

		// Merge in more config files
		vpr.SetConfigFile(cfgFile)
		if err = vpr.MergeInConfig(); err != nil {
			err = errors.Wrapv(err, "unable to merge config file", cfgFile)
			return
		}
	}

	// Set config value
	var metadata mapstructure.Metadata
	if err = vpr.Unmarshal(&cfg, configureConfigFileDecode(&metadata)); err != nil {
		err = errors.Wrap(err, "unable to unmarshal config")
		return
	}

	if len(metadata.Unused) > 0 {
		err = errors.Errorv("there are extra values in your config", metadata.Unused)
		return
	}

	return
}

func MergeInEnvVars(cfg *AppConfig, envVars []string) (err error) {
	if err = ParseEnvVars(cfg, envVars); err != nil {
		err = errors.Wrap(err, "unable to parse env vars")
	}
	return
}

func NewConfigParam(appCfg, leafFieldPtr interface{}) (result *manip.Param) {
	if _, ok := appCfg.(*AppConfig); !ok {
		panic("must be a *AppConfig for NewConfigParam")
	}
	return manip.NewParam(appCfg, leafFieldPtr, vars.ConfigParamTag, structValidatable)
}

func structValidatable(cfgVal reflect.Value) (result bool) {
	t := cfgVal.Type()
	return t.Implements(reflect.TypeOf(new(validation.Validatable)).Elem()) ||
		t.Implements(reflect.TypeOf(new(validation.ValidatableWithContext)).Elem())
}

func printHelp() {
	div := strings.Repeat("=", len(vars.Description))
	fmt.Println("")
	fmt.Println(div)
	fmt.Println(vars.Name)
	fmt.Println(vars.Description)
	fmt.Println(vars.URL)
	fmt.Println(div)
	fmt.Println("")
	fmt.Println("To configure this command, pass --config=\"config.yaml\"")
}

func parseFlags(args []string) (cfgFiles []string, err error) {
	var help bool

	flags := pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
	flags.StringSliceVarP(&cfgFiles, "config", "c", nil, "config files")
	flags.BoolVarP(&help, "help", "h", false, "show command usage")
	if err = flags.Parse(args); err != nil {
		err = errors.Wrap(err, "unable to parse flags")
		return
	}

	if err = validation.Validate(cfgFiles, validation.Required); err != nil {
		err = errors.WithMessage(err, "invalid value for \"config\"")
		return
	}

	// Show help message if requested
	if help {
		printHelp()
		os.Exit(0)
	}

	return
}

func configureConfigFileDecode(metadata *mapstructure.Metadata) func(c *mapstructure.DecoderConfig) {
	return func(c *mapstructure.DecoderConfig) {
		c.TagName = vars.ConfigParamTag
		c.Metadata = metadata
		c.DecodeHook = mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeHookFunc("2006-01-02T15:04:05"),
			mapstructure.StringToSliceHookFunc(","),
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.DecodeHookFunc(StringToTimeDurationHookFunc),
		)
	}
}

func StringToTimeDurationHookFunc(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f.Kind() != reflect.String {
		return data, nil
	}
	if f.Kind() != reflect.String {
		return data, nil
	}
	if t != reflect.TypeOf(time.Duration(5)) {
		return data, nil
	}

	// Convert it by parsing
	return time.ParseDuration(data.(string))
}

type SetsDefaults interface {
	SetDefaults()
}
