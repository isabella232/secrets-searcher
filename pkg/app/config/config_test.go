package config_test

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	searchpkg "github.com/pantheon-systems/secrets-searcher/pkg/search"
	"github.com/pantheon-systems/secrets-searcher/pkg/source"

	va "github.com/go-ozzo/ozzo-validation/v4"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	. "github.com/pantheon-systems/secrets-searcher/pkg/app/config"
	"github.com/pantheon-systems/secrets-searcher/pkg/dev"
	"github.com/pantheon-systems/secrets-searcher/pkg/logg"
	"github.com/pantheon-systems/secrets-searcher/pkg/valid"
	. "github.com/pantheon-systems/secrets-searcher/pkg/valid/testing"
)

func TestConfig(t *testing.T) {
	dev.RunningTests = true
	RegisterFailHandler(Fail)
	RunSpecs(t, "Config Test Suite")
}

func runConfigTest(getCfg getCfgAndField, expected va.Error) {

	// Function should return pointers to appConfig, and another to a field to test for errors on
	appConfig, fieldPtr := getCfg()
	param := NewConfigParam(appConfig, fieldPtr)

	// Fire test
	err := va.Validate(appConfig)

	// Assertions
	var matcher types.GomegaMatcher
	switch expected.Message() {
	case "no_error":
		// Expect that no error is on the field
		matcher = Not(HaveError(param))
	case "reverse_error":
		// Expect that a certain error is not on the field
		matcher = Not(HaveErrorType(param, expected))
	default:
		// Expect that a certain error is on the field
		matcher = HaveErrorType(param, expected)
	}

	Expect(err).To(matcher)
}

var _ = Describe("Config tests", func() {
	Describe("Defaults", func() {

		Context("If I pass in default config", func() {

			It("default values are set", func() {

				// Fire
				appCfg := buildAppConfigObjectFromFile("config-empty.yaml")

				Expect(appCfg.LogLevel).To(Equal(logg.Info.Value()))
				Expect(appCfg.OutputDir).To(Not(BeEmpty()))
			})
		})
	})

	Describe("Environment variables", func() {

		Context("If I pass an environment variable", func() {

			It("they are set in a new config object", func() {
				const outputDirValue = "output-dir"
				const sourceAPITokenValue = "source-api-token"
				args := []string{"", "--config=" + testConfigPath("config-empty.yaml")}
				env := []string{
					"SECRETS_LOG_LEVEL=" + logg.Trace.Value(),
					"SECRETS_OUTPUT_DIR=" + outputDirValue,
					"SECRETS_SOURCE_API_TOKEN=" + sourceAPITokenValue,
					"SECRETS_DEV_ENABLED=true",
				}

				// Fire
				appCfg := builtConfigWithArgs(args, env)

				Expect(appCfg.LogLevel).To(Equal(logg.Trace.Value()))
				Expect(appCfg.OutputDir).To(Equal(outputDirValue))
				Expect(appCfg.SourceConfig.APIToken).To(Equal(sourceAPITokenValue))
			})
		})
	})

	Describe("Validation", func() {

		Context("If I pass in default config", func() {

			It("validation errors are returned for required fields", func() {
				appCfg := buildAppConfigObjectFromFile("config-empty.yaml")

				// Fire
				err := va.Validate(appCfg)

				Expect(err).To(HaveErrorType(NewConfigParam(appCfg, &appCfg.SourceConfig.LocalDir), va.ErrRequired))
			})
		})

		Context("If I use minimal config with required fields", func() {

			It("we should not get a validation error", func() {
				appCfg := validConfig()

				// Fire
				err := va.Validate(appCfg)

				Expect(err).To(BeNil())
			})
		})

		Context("If I use minimal config with required fields", func() {

			It("we should not get a validation error", func() {
				appCfg := validConfig()

				// Fire
				err := va.Validate(appCfg)

				Expect(err).To(BeNil())
			})
		})
	})

	DescribeTable("valid values should not cause validation errors",
		runConfigTest,

		// AppConfig
		Entry("log-level from enum", setLogLevel(logg.Debug.Value()), noErr()),
		Entry("output-dir string value not empty", setOutputDir("anystring"), noErr()),
		Entry("output-dir is an existing directory should be ok", setOutputDir(cwd()), noErr()),

		// SourceConfig
		Entry("source.provider from enum", setSourceProvider(source.Local.Value()), noErr()),
		Entry("source.local-dir string not empty", setSourceLocalDir("anystring"), notErr(va.ErrRequired)),
		Entry("source.local-dir outside of output directory", setSourceLocalDirOutsideOutputDir(), noErr()),
	)

	DescribeTable("invalid values should cause validation errors",
		runConfigTest,

		// AppConfig
		Entry("log-level value empty", setLogLevel(""), va.ErrRequired),
		Entry("log-level not from enum list", setLogLevel("invalid"), va.ErrInInvalid),
		Entry("output-dir empty", setOutputDir(""), va.ErrRequired),

		// SourceConfig
		Entry("source.provider empty", setSourceProvider(""), va.ErrRequired),
		Entry("source.provider not from enum", setSourceProvider("not-in-enum"), va.ErrInInvalid),
		Entry("source.local-dir inside of output directory",
			setSourceLocalDirInsideOutputDir(), valid.ErrPathNotWithinParam),
	)
})

// Config object builders

// Config with defaults and env vars set
func testConfigPath(file string) string {
	return filepath.Join("testdata", file)
}

// Config with defaults and env vars set
func buildAppConfigObjectFromFile(file string) (result *AppConfig) {
	path := testConfigPath(file)
	args := []string{"", "--config=" + path}
	return builtConfigWithArgs(args, nil)
}

// Config with defaults and env vars set
func builtConfigWithArgs(args, env []string) (result *AppConfig) {
	var err error
	if result, err = BuildConfig(args, env); err != nil {
		panic("cannot build config: " + err.Error())
	}
	return result
}

func validConfig() (result *AppConfig) {
	appConfig := buildAppConfigObjectFromFile("config-empty.yaml")
	appConfig.SourceConfig.Provider = source.Local.Value()
	appConfig.SourceConfig.LocalDir = os.TempDir() // Needed with provider "local"
	appConfig.SearchConfig.ProcessorConfigs = []*ProcessorConfig{{
		Name:      "my very easy going regex processor",
		Processor: searchpkg.Regex.String(),
		RegexProcessorConfig: RegexProcessorConfig{
			RegexString:        `.*`,
			WhitelistCodeMatch: nil,
		},
	}}
	return appConfig
}

// Fields

func setLogLevel(value string) getCfgAndField {
	return func() (*AppConfig, interface{}) {
		c := buildAppConfigObjectFromFile("config-empty.yaml")
		c.LogLevel = value
		return c, &c.LogLevel
	}
}

func setOutputDir(value string) getCfgAndField {
	return func() (*AppConfig, interface{}) {
		c := buildAppConfigObjectFromFile("config-empty.yaml")
		c.OutputDir = value
		return c, &c.OutputDir
	}
}

func setSourceProvider(value string) getCfgAndField {
	return func() (*AppConfig, interface{}) {
		c := buildAppConfigObjectFromFile("config-empty.yaml")
		c.SourceConfig.Provider = value
		return c, &c.SourceConfig.Provider
	}
}

func setSourceLocalDir(value string) getCfgAndField {
	return func() (*AppConfig, interface{}) {
		c := buildAppConfigObjectFromFile("config-empty.yaml")
		c.SourceConfig.LocalDir = value
		return c, &c.SourceConfig.LocalDir
	}
}

func setSourceLocalDirOutsideOutputDir() getCfgAndField {
	return func() (*AppConfig, interface{}) {
		c := buildAppConfigObjectFromFile("config-empty.yaml")
		c.SourceConfig.Provider = source.Local.Value()
		c.OutputDir = mkTempDir("output")
		c.SourceConfig.LocalDir = mkTempDir("outside-output-dir")
		return c, &c.SourceConfig.LocalDir
	}
}

func setSourceLocalDirInsideOutputDir() getCfgAndField {
	return func() (*AppConfig, interface{}) {
		c := buildAppConfigObjectFromFile("config-empty.yaml")
		c.SourceConfig.Provider = source.Local.Value()
		c.OutputDir = mkTempDir("output")
		c.SourceConfig.LocalDir = mkTempDirIn(c.OutputDir, "inside-output")
		return c, &c.SourceConfig.LocalDir
	}
}

func setSearchTarget(value string) getCfgAndField {
	return func() (*AppConfig, interface{}) {
		c := buildAppConfigObjectFromFile("config-empty.yaml")
		c.SearchConfig.CustomTargetConfigs = []*TargetConfig{
			{
				Name:               "test-target",
				KeyPatterns:        nil,
				ExcludeKeyPatterns: nil,
				ValChars:           nil,
				ValLenMin:          0,
				ValLenMax:          0,
				ValEntropyMin:      0,
			},
		}
		return c, &c.SearchConfig.CustomTargetConfigs
	}
}

// Run a single table entry

type getCfgAndField func() (cfg *AppConfig, field interface{})

//
// Helpers

// Create a unique temp directory (automatically gets cleaned up)
func mkTempDir(pattern string) (result string) {
	return mkTempDirIn(tempDir, pattern)
}

// Create a unique temp directory within another.
// Not automatically cleaned up unless you specify a temp dir as the base.
func mkTempDirIn(base, pattern string) (result string) {
	var err error
	if result, err = ioutil.TempDir(base, pattern); err != nil {
		log.Fatal(err)
	}
	return result
}

func cwd() (result string) {
	result, _ = os.Getwd()
	return
}

// Matchers
func notErr(errObj va.Error) va.Error { return errObj.SetMessage("reverse_error") }
func noErr() va.Error                 { return va.NewError("", "no_error") }

//
// Plumbing

var tempDir string

var _ = BeforeSuite(func() {
	var err error
	if tempDir, err = ioutil.TempDir(os.TempDir(), "secrets-testing"); err != nil {
		log.Fatal(err)
	}
})

var _ = AfterSuite(func() {
	_ = os.RemoveAll(tempDir)
})
