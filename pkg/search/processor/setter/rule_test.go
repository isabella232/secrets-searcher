package setter_test

import (
	"testing"

	"github.com/pantheon-systems/secrets-searcher/pkg/app/config"
	"github.com/pantheon-systems/secrets-searcher/pkg/builtin"
	"github.com/pantheon-systems/secrets-searcher/pkg/dev"
	"github.com/pantheon-systems/secrets-searcher/pkg/search"
	. "github.com/pantheon-systems/secrets-searcher/pkg/search/processor/setter"
	"github.com/pantheon-systems/secrets-searcher/pkg/search/rulebuild"
	"github.com/pantheon-systems/secrets-searcher/pkg/search/searchtest"
	"github.com/stretchr/testify/require"
)

type ruleTest struct {
	coreProcessor    builtin.ProcessorName
	coreTargets      []builtin.TargetName
	customProcessors []*config.ProcessorConfig
	customTargets    []*config.TargetConfig
	path             string
	line             string
	expKey           string
	expVal           string
	expCtx           string
	expFail          bool
}

// Run rule test
func runRuleTest(t *testing.T, test ruleTest) {
	dev.RunningTests = true

	var targets *search.TargetSet
	if test.customTargets != nil {
		targets = searchtest.CustomTargets(test.customTargets)
	} else {
		targets = searchtest.CoreTargets(test.coreTargets...)
	}

	var rule *Rule
	if test.customProcessors != nil {
		rule = searchtest.CustomRule(test.customProcessors[0], targets)
	} else {
		rule = searchtest.CoreRule(test.coreProcessor, targets)
	}

	// Fire
	contextValue, keyValue, secretValue, matchingRe, ok := rule.FindNextSecret(test.line)

	// Info message
	var info string
	if !ok {
		matchingRe = nil
	}
	info = searchtest.RuleMatchInfo(rule, keyValue, secretValue, matchingRe, ok, test.line)

	// Test failure
	if test.expFail {
		require.Falsef(t, ok, "Rule \"%s\" found secrets but we expected it not to.\n%s", rule.GetName(), info)
		return
	}

	// Test secret value
	require.Truef(t, ok,
		"Rule \"%s\" should have matched.\n%s", rule.GetName(), info)
	require.Equalf(t, test.expVal, secretValue.Value,
		"Rule \"%s\" submitted a non-matching secret value.\n%s", rule.GetName(), info)
	require.Equalf(t, test.expCtx, contextValue.Value,
		"Rule \"%s\" submitted a non-matching context value.\n%s", rule.GetName(), info)

	// Test key value
	if test.expKey != "" {
		require.Equalf(t, test.expKey, keyValue.Value,
			"Rule \"%s\" submitted a non-matching key.\n%s", rule.GetName(), info)
	}
}

func TestFind_CharacterNegation(t *testing.T) {
	runRuleTest(t, ruleTest{
		customProcessors: []*config.ProcessorConfig{{
			Name:      "NegationSetterProcessor",
			Processor: search.Setter.String(),
			SetterProcessorConfig: config.SetterProcessorConfig{
				FileExts: rulebuild.ShellScriptExtPaths(),
				KeyTmpls: []string{rulebuild.SingleDblQuoteNoneKey},
				Operator: rulebuild.EqOper,
				ValTmpls: []string{rulebuild.SingleDblQuoteJustVal},
			},
		}},
		customTargets: []*config.TargetConfig{{
			Name:        "GreedyTarget",
			KeyPatterns: []string{`pass`},
			ValChars:    rulebuild.AnyChars(),
			ValLenMin:   5,
			ValLenMax:   64,
		}},
		path:   "file.sh",
		line:   "pass = 8d9b08206d56012f52f91231390e3932-END # Here are some comments that shouldn't be matched",
		expKey: "pass",
		expVal: "8d9b08206d56012f52f91231390e3932-END",
		expCtx: "pass = 8d9b08206d56012f52f91231390e3932-END",
	})
}

func TestFind_JSONObjFieldVal_DblQuoteKey(t *testing.T) {
	runRuleTest(t, ruleTest{
		coreProcessor: builtin.JSONObjFieldValSetter,
		coreTargets:   []builtin.TargetName{builtin.APIKeysAndTokens},
		path:          "file.json",
		line:          "\"api_key\": \"8d9b08206d56012f52f91231390e3932\",",
		expKey:        "api_key",
		expVal:        "8d9b08206d56012f52f91231390e3932",
		expCtx:        "\"api_key\": \"8d9b08206d56012f52f91231390e3932\"",
	})
}

func TestFind_JSONObjFieldVal_InlinePassword(t *testing.T) {
	runRuleTest(t, ruleTest{
		coreProcessor: builtin.JSONObjFieldValSetter,
		coreTargets:   []builtin.TargetName{builtin.Passwords},
		path:          "file.json",
		line: "0\",\"password\":\"dd41388c67e355fa0b03c4982adef4deb265950a\"," +
			"\"enabledDatabaseCustomization\":true,\"customScripts\":{\"login\":\"function login (email, " +
			"password, callback) {\\n  \\n  request.po",
		expKey: "password",
		expVal: "dd41388c67e355fa0b03c4982adef4deb265950a",
		expCtx: "\"password\":\"dd41388c67e355fa0b03c4982adef4deb265950a\"",
	})
}

func TestFind_JSONObjFieldVal(t *testing.T) {
	runRuleTest(t, ruleTest{
		coreProcessor: builtin.JSONObjFieldValSetter,
		coreTargets:   []builtin.TargetName{builtin.APIKeysAndTokens},
		path:          ".libsonnet",
		line: ",{\"name\":\"ingressgateway-titan-certs\",\"secret\":{\"optional\":true," +
			"\"secretName\":\"istio-ingressgateway-titan-certs\"}}]}}}}",
		expKey: "secretName",
		expVal: "istio-ingressgateway-titan-certs",
		expCtx: "\"secretName\":\"istio-ingressgateway-titan-certs\"",
	})
}

func TestFind_YAMLDictFieldVal(t *testing.T) {
	runRuleTest(t, ruleTest{
		coreProcessor: builtin.YAMLDictFieldValSetter,
		coreTargets:   []builtin.TargetName{builtin.APIKeysAndTokens},
		path:          "file.yaml",
		line:          `repo_token: 2jubd65kcsQsOa6kjaMAmWJ1wEqmqqi0E`,
		expKey:        "repo_token",
		expVal:        "2jubd65kcsQsOa6kjaMAmWJ1wEqmqqi0E",
		expCtx:        "repo_token: 2jubd65kcsQsOa6kjaMAmWJ1wEqmqqi0E",
	})
}

func TestFind_JSObjLiteralField_NoQuoteKey(t *testing.T) {
	runRuleTest(t, ruleTest{
		coreProcessor: builtin.JSObjLiteralFieldSetter,
		coreTargets:   []builtin.TargetName{builtin.APIKeysAndTokens},
		path:          "file.jade",
		line:          "token: '80e1c56259bf32235ef432e811bbf86e',",
		expKey:        "token",
		expVal:        "80e1c56259bf32235ef432e811bbf86e",
		expCtx:        "token: '80e1c56259bf32235ef432e811bbf86e'",
	})
}

func TestFind_URLQueryStringParamVal(t *testing.T) {
	runRuleTest(t, ruleTest{
		coreProcessor: builtin.URLQueryStringParamValSetter,
		coreTargets:   []builtin.TargetName{builtin.APIKeysAndTokens},
		path:          "README-deprecated.md",
		line: "[![Build Status](https://circleci.com/gh/pantheon-systems/dashboard.svg?style=shield" +
			"&circle-token=bed57e834f553491febfcf21fe8b632d6626f60a)](https://cir",
		expKey: "circle-token",
		expVal: "bed57e834f553491febfcf21fe8b632d6626f60a",
		expCtx: "&circle-token=bed57e834f553491febfcf21fe8b632d6626f60a",
	})
}

func TestFind_URLPathParamVal_ShouldntMatchWithPathPieceInBetween(t *testing.T) {
	runRuleTest(t, ruleTest{
		coreProcessor: builtin.URLPathParamValSetter,
		coreTargets:   []builtin.TargetName{builtin.APIKeysAndTokens},
		path:          "file.coffee",
		line: "    image: 'https://secure.gravatar.com/avatar/e567aa8adbd2d49cd9990ea1ed19d4eb?s=40&d=" +
			"https%3A%2F%2Fpantheon-content.s3.amazonaws.com%2Fblank_user.png'",
		expFail: true,
	})
}

func TestFind_URLPathParamVal_NotAURL(t *testing.T) {
	runRuleTest(t, ruleTest{
		coreProcessor: builtin.URLPathParamValSetter,
		coreTargets:   []builtin.TargetName{builtin.APIKeysAndTokens},
		path:          "Makefile",
		line:          "@mkdir -p ./devops/k8s/secrets/non-prod/webhook-receiver-service-account",
		expFail:       true,
	})
}

func TestFind_ShellScriptVarAssign_EnvFile(t *testing.T) {
	runRuleTest(t, ruleTest{
		// 	Entry("ShellScriptVarAssign env var, .env file", &ruleTest{
		coreProcessor: builtin.ShellScriptVarAssignSetter,
		coreTargets:   []builtin.TargetName{builtin.APIKeysAndTokens},
		path:          "file.env",
		line:          "HUBOT_TRELLO_TOKEN=cc00c867e89a7f17478a817d6a745031c1fa2cb9000eb5720bd94fdce42581f1",
		expKey:        "HUBOT_TRELLO_TOKEN",
		expVal:        "cc00c867e89a7f17478a817d6a745031c1fa2cb9000eb5720bd94fdce42581f1",
		expCtx:        "HUBOT_TRELLO_TOKEN=cc00c867e89a7f17478a817d6a745031c1fa2cb9000eb5720bd94fdce42581f1",
	})
}

func TestFind_ShellCmdParamVal_LongOptHexVal(t *testing.T) {
	runRuleTest(t, ruleTest{
		coreProcessor: builtin.ShellCmdParamValSetter,
		coreTargets:   []builtin.TargetName{builtin.APIKeysAndTokens},
		path:          "file.json",
		line:          "\"command\": \"/etc/sensu/handlers/pantheon_pagerduty.rb --api-key 8d9b08206d56012f52f91231390e3932\",",
		expKey:        "api-key",
		expVal:        "8d9b08206d56012f52f91231390e3932",
		// This would be better if we didn't have the trailing quote but it's not a big deal
		// but the regex would need to be too much more complicated, for such a tiny concern.
		expCtx: "--api-key 8d9b08206d56012f52f91231390e3932\"",
	})
}

func TestFind_ShellCmdParamVal_RegressionFix(t *testing.T) {
	runRuleTest(t, ruleTest{
		coreProcessor: builtin.ShellCmdParamValSetter,
		coreTargets:   []builtin.TargetName{builtin.APIKeysAndTokens},
		path:          "Makefile",
		line:          "\tconda build -q --user pantheon --token pa-07381fe0-4587-4ca4-9d9c-73a00bcbe869 recipe",
		expKey:        "token",
		expVal:        "pa-07381fe0-4587-4ca4-9d9c-73a00bcbe869",
		expCtx:        "--token pa-07381fe0-4587-4ca4-9d9c-73a00bcbe869",
	})
}

func TestFind_ShellCmdParamVal_LongShortArgsMixed(t *testing.T) {
	runRuleTest(t, ruleTest{
		coreProcessor: builtin.ShellCmdParamValSetter,
		coreTargets:   []builtin.TargetName{builtin.Passwords},
		path:          "file.yaml",
		line:          "/xy --dir=/cloudsql -credential_file=/etc/subdir/service-account.json -max_connections=65 &",
		expKey:        "credential_file",
		expVal:        "/etc/subdir/service-account.json",
		expCtx:        "-credential_file=/etc/subdir/service-account.json",
	})
}

func TestFind_PHPVarAssign_MatchClassNameForContext(t *testing.T) {
	runRuleTest(t, ruleTest{
		coreProcessor: builtin.PHPVarAssignSetter,
		coreTargets:   []builtin.TargetName{builtin.APIKeysAndTokens},
		path:          "file.php",
		line:          `Recurly_Client::$apiKey = 'a52b0d69401e4fa483af274c5da1ea9a';`,
		expKey:        `apiKey`,
		expVal:        `a52b0d69401e4fa483af274c5da1ea9a`,
		expCtx:        `Recurly_Client::$apiKey = 'a52b0d69401e4fa483af274c5da1ea9a'`,
	})
}

func TestFind_PHPConstDefine_MatchDefineForContext(t *testing.T) {
	runRuleTest(t, ruleTest{
		coreProcessor: builtin.PHPConstDefineSetter,
		coreTargets:   []builtin.TargetName{builtin.APIKeysAndTokens},
		path:          "file.php",
		line:          "define('DESK_OAUTH_SECRET', '3EnbzNCdfSCwmIYqMP4aiVIEUsKK5QaxNvBfSUjC');",
		expKey:        "DESK_OAUTH_SECRET",
		expVal:        "3EnbzNCdfSCwmIYqMP4aiVIEUsKK5QaxNvBfSUjC",
		expCtx:        "define('DESK_OAUTH_SECRET', '3EnbzNCdfSCwmIYqMP4aiVIEUsKK5QaxNvBfSUjC')",
	})
}

func TestFind_PyVarAssign_RSTFile(t *testing.T) {
	runRuleTest(t, ruleTest{
		coreProcessor: builtin.PyVarAssignSetter,
		coreTargets:   []builtin.TargetName{builtin.APIKeysAndTokens},
		path:          "file.rst",
		line:          "session = consulate.Session(token='5d24c96b4f6a4aefb99602ce9b60d16b')",
		expKey:        "token",
		expVal:        "5d24c96b4f6a4aefb99602ce9b60d16b",
		expCtx:        "token='5d24c96b4f6a4aefb99602ce9b60d16b'",
	})
}

func TestFind_ConfParamLogstashStyleEnvVarDefault_GetDefaultValue(t *testing.T) {
	runRuleTest(t, ruleTest{
		coreProcessor: builtin.ConfParamLogstashStyleEnvVarDefaultSetter,
		coreTargets:   []builtin.TargetName{builtin.APIKeysAndTokens},
		path:          "file.conf",
		line: "mutate { add_field => { \"token\" => \"" +
			"${LOGS_GW_ACCOUNT_TOKEN:cc00c867e89a7f17478a817d6a745031c1fa2cb9000eb5720bd94fdce42581f1}\"  }  }",
		expKey: "LOGS_GW_ACCOUNT_TOKEN",
		expVal: "cc00c867e89a7f17478a817d6a745031c1fa2cb9000eb5720bd94fdce42581f1",
		expCtx: "${LOGS_GW_ACCOUNT_TOKEN:cc00c867e89a7f17478a817d6a745031c1fa2cb9000eb5720bd94fdce42581f1}",
	})
}

func TestFind_ConfParamLogstashStyle_IgnoreTemplatedValues(t *testing.T) {
	runRuleTest(t, ruleTest{
		coreProcessor: builtin.ConfParamLogstashStyleSetter,
		coreTargets:   []builtin.TargetName{builtin.APIKeysAndTokens},
		path:          "file.conf",
		line: "mutate { add_field => " +
			"{ \"token\" => \"${LOGS_GW_ACCOUNT_TOKEN:" +
			"cc00c867e89a7f17478a817d6a745031c1fa2cb9000eb5720bd94fdce42581f1}\"  }  }",
		expFail: true,
	})
}

func TestFind_ConfParamSystemdServiceEnvVar(t *testing.T) {
	runRuleTest(t, ruleTest{
		coreProcessor: builtin.ConfParamSystemdServiceEnvVarSetter,
		coreTargets:   []builtin.TargetName{builtin.Passwords},
		path:          "file.service",
		line:          "Environment=AGGREGATES_PASSWORD=27d09f46d6b94d07a7f803191ef49f81",
		expKey:        "AGGREGATES_PASSWORD",
		expVal:        "27d09f46d6b94d07a7f803191ef49f81",
		expCtx:        "Environment=AGGREGATES_PASSWORD=27d09f46d6b94d07a7f803191ef49f81",
	})
}
