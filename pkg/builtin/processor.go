package builtin

import (
	"github.com/pantheon-systems/secrets-searcher/pkg/app/config"
	"github.com/pantheon-systems/secrets-searcher/pkg/entropy"
	"github.com/pantheon-systems/secrets-searcher/pkg/search"
	. "github.com/pantheon-systems/secrets-searcher/pkg/search/rulebuild"
)

// Target definitions
func processorDefinitions() (result []*config.ProcessorConfig) {
	result = []*config.ProcessorConfig{}
	result = append(result, setterProcessorDefinitions()...)
	result = append(result, pemProcessorDefinitions()...)
	result = append(result, regexProcessorDefinitions()...)
	//result = append(result, entropyProcessorDefinitions()...)
	return
}

// Setter definitions
func setterProcessorDefinitions() (result []*config.ProcessorConfig) {
	return []*config.ProcessorConfig{

		// URL setters

		// URL path elements
		//
		// MATCH -KKKKKKK-VVVV
		//       /api-key/shhh/foo/bar
		//       /api-key/shhh?foo=bar
		{
			Name:      URLPathParamValSetter.String(),
			Processor: search.Setter.String(),
			SetterProcessorConfig: config.SetterProcessorConfig{
				FileExts:     AnyPath(),
				KeyTmpls:     []string{`\/` + VarKey},
				KeyChars:     CommonURLPathChars(),
				Operator:     SlashChar,
				NoWhitespace: true,
				ValTmpls:     []string{JustVal},
				NotValChars:  []string{SlashChar},
			},
		},

		// URL querystring parameters
		//
		// MATCH -KKKKKKK-VVVV
		//       ?api-key=shhh
		//       &api-key=shhh
		{
			Name:      URLQueryStringParamValSetter.String(),
			Processor: search.Setter.String(),
			SetterProcessorConfig: config.SetterProcessorConfig{
				FileExts:     AnyPath(),
				KeyTmpls:     []string{`[?&]` + VarKey},
				Operator:     EqOper,
				NoWhitespace: true,
				ValTmpls:     []string{JustVal},
			},
		},

		//
		// Python setters

		// Python variable assignment & keyword arguments
		//
		// MATCH KKKKKKK----VVVV-
		//       api_key = "shhh"
		//       api_key = 'shhh'
		//       secrets = dict(api_key="shhh")
		//       secrets = get_secrets(api_key="shhh")
		{
			Name:      PyVarAssignSetter.String(),
			Processor: search.Setter.String(),
			SetterProcessorConfig: config.SetterProcessorConfig{
				FileExts: PYExtPaths(),
				KeyTmpls: []string{VarKey},
				Operator: EqOper,
				ValTmpls: []string{SingleDblQuoteVal},
			},
		},

		// Python dict field assignment
		//
		// MATCH        --KKKKKKK------VVVV-
		//       secrets["api_key"] = "shhh"
		//       secrets['api_key'] = 'shhh'
		//       secrets[api_key_name_var] = 'shhh'
		{
			Name:      PyDictFieldAssignSetter.String(),
			Processor: search.Setter.String(),
			SetterProcessorConfig: config.SetterProcessorConfig{
				FileExts: PYExtPaths(),
				KeyTmpls: []string{BracketSingleDblQuoteNoneKey},
				Operator: EqOper,
				ValTmpls: []string{SingleDblQuoteVal},
			},
		},

		// Python dict literal field
		//
		// MATCH  -KKKKKKK----VVVV-
		//       {"api_key": "shhh"}
		//       {'api_key': 'shhh'}
		//       {api_key_name_var: 'shhh'}
		{
			Name:      PyDictLiteralFieldSetter.String(),
			Processor: search.Setter.String(),
			SetterProcessorConfig: config.SetterProcessorConfig{
				FileExts: PYExtPaths(),
				KeyTmpls: []string{SingleDblQuoteNoneKey},
				Operator: ColonOper,
				ValTmpls: []string{SingleDblQuoteVal},
			},
		},

		// Python tuple or list value
		//
		// MATCH -KKKKKKK----VVVV-
		//       'api_key', 'shhh'
		//       ['api_key', 'shhh']
		//       [api_key_name_var, 'shhh']
		{
			Name:      PyTupleSetter.String(),
			Processor: search.Setter.String(),
			SetterProcessorConfig: config.SetterProcessorConfig{
				FileExts: PYExtPaths(),
				KeyTmpls: []string{SingleDblQuoteNoneKey},
				Operator: CommaOper,
				ValTmpls: []string{SingleDblQuoteVal},
			},
		},

		//
		// PHP setters

		// PHP variable assignment
		//
		// MATCH -KKKKKKK----VVVV-
		//       $api_key = 'shhh'
		//       Secrets::$api_key = 'shhh'
		//       Secrets\Important::api_key = 'shhh'
		//       Secrets\Important::$api_key = 'shhh'
		{
			Name:      PHPVarAssignSetter.String(),
			Processor: search.Setter.String(),
			SetterProcessorConfig: config.SetterProcessorConfig{
				FileExts: PHPExtPaths(),
				MainTmpl: `(?:[a-z\d\_\\]+::)?` + KeyTmplVar + OpVar + ValTmplVar,
				KeyTmpls: []string{DollarVarKey},
				Operator: EqOper,
				ValTmpls: []string{SingleDblQuoteVal},
			},
		},

		// PHP assoc array field assignment
		//
		// MATCH         --KKKKKKK------VVVV-
		//       $secrets['api_key'] = 'shhh';
		//       $secrets["api_key"] = "shhh";
		//       $secrets[$api_key_name_var] = "shhh";
		{
			Name:      PHPAssocArrayFieldAssignSetter.String(),
			Processor: search.Setter.String(),
			SetterProcessorConfig: config.SetterProcessorConfig{
				FileExts: PHPExtPaths(),
				KeyTmpls: []string{BracketSingleDblQuoteDollarKey},
				Operator: EqOper,
				ValTmpls: []string{SingleDblQuoteVal},
			},
		},

		// PHP assoc array literal field
		//
		// MATCH -KKKKKKK------VVVV-
		//       'api_key' => 'shhh',
		//       "api_key" => "shhh",
		//       $api_key_name_var => "shhh",
		{
			Name:      PHPAssocArrayLiteralFieldSetter.String(),
			Processor: search.Setter.String(),
			SetterProcessorConfig: config.SetterProcessorConfig{
				FileExts: PHPExtPaths(),
				KeyTmpls: []string{SingleDblQuoteDollarKey},
				Operator: EqArrowOper,
				ValTmpls: []string{SingleDblQuoteVal},
			},
		},

		// PHP tuple or list value
		//
		// MATCH --------KKKKKKK----VVVV-
		//       define('api_key', 'shhh')
		{
			Name:      PHPConstDefineSetter.String(),
			Processor: search.Setter.String(),
			SetterProcessorConfig: config.SetterProcessorConfig{
				FileExts: PHPExtPaths(),
				MainTmpl: `define\(` + KeyTmplVar + OpVar + ValTmplVar + `\)`,
				KeyTmpls: []string{SingleDblQuoteNoneKey},
				Operator: CommaOper,
				ValTmpls: []string{SingleDblQuoteVal},
			},
		},

		//
		// Javascript setters

		// Javascript variable assignment
		//
		// MATCH KKKKKK----VVVV-
		//       apiKey = "shhh";
		//       apiKey = 'shhh';
		//       var apiKey = "shhh";
		//       let apiKey = "shhh";
		//       const apiKey = "shhh";
		{
			Name:      JSVarAssignSetter.String(),
			Processor: search.Setter.String(),
			SetterProcessorConfig: config.SetterProcessorConfig{
				FileExts: JSExtPaths(),
				KeyTmpls: []string{VarKey},
				Operator: EqOper,
				ValTmpls: []string{SingleDblQuoteVal},
			},
		},

		// Javascript object field assignment
		//
		// MATCH        --KKKKKKK------VVVV-
		//       secrets["api_key"] = "shhh"
		//       secrets['api_key'] = 'shhh'
		//       secrets[api_key_name_var] = 'shhh'
		{
			Name:      JSObjFieldAssignSetter.String(),
			Processor: search.Setter.String(),
			SetterProcessorConfig: config.SetterProcessorConfig{
				FileExts: JSExtPaths(),
				KeyTmpls: []string{BracketSingleDblQuoteNoneKey},
				Operator: EqOper,
				ValTmpls: []string{SingleDblQuoteVal},
			},
		},

		// Javascript object literal field
		//
		// MATCH   -KKKKKK----VVVV-
		//       { "apiKey": "shhh" };
		//       { 'apiKey': 'shhh' };
		//       { apiKeyNameVar: "shhh" }
		{
			Name:      JSObjLiteralFieldSetter.String(),
			Processor: search.Setter.String(),
			SetterProcessorConfig: config.SetterProcessorConfig{
				FileExts: JSExtPaths(),
				KeyTmpls: []string{SingleDblQuoteNoneKey},
				Operator: ColonOper,
				ValTmpls: []string{SingleDblQuoteVal},
			},
		},

		//
		// Go setters

		// Go variable assignment
		//
		// MATCH     KKKKKK----VVVV-
		//       var apiKey = "shhh"
		//       apiKey := "shhh"
		//       apiKey = "shhh"
		//       const apiKey = "shhh"
		// MATCH       KKKKKK-----------VVVV-
		//       const apiKey string = "shhh"
		//       var apiKey string = "shhh"
		{
			Name:      GoVarAssignSetter.String(),
			Processor: search.Setter.String(),
			SetterProcessorConfig: config.SetterProcessorConfig{
				FileExts: GoExtPaths(),
				KeyTmpls: []string{DeclaredStrVarKey},
				Operator: GoOper,
				ValTmpls: []string{DblTickQuoteVal},
			},
		},

		// Go hash field assignment
		//
		// MATCH        --KKKKKK------VVVV-
		//       secrets["apiKey"] = "shhh"
		//       secrets[apiKeyNameVar] = "shhh"
		{
			Name:      GoHashFieldAssignSetter.String(),
			Processor: search.Setter.String(),
			SetterProcessorConfig: config.SetterProcessorConfig{
				FileExts: GoExtPaths(),
				KeyTmpls: []string{BracketDblQuoteNoneKey},
				Operator: EqOper,
				ValTmpls: []string{DblQuoteVal},
			},
		},

		// Go hash literal field
		//
		// MATCH                              -KKKKKK----VVVV-
		//       secrets := map[string]string{"apiKey": "shhh"}
		//       secrets := map[string]string{apiKeyNameVar: "shhh"}
		{
			Name:      GoHashLiteralFieldSetter.String(),
			Processor: search.Setter.String(),
			SetterProcessorConfig: config.SetterProcessorConfig{
				FileExts: GoExtPaths(),
				KeyTmpls: []string{DblQuoteNoneKey},
				Operator: ColonOper,
				ValTmpls: []string{DblQuoteVal},
			},
		},

		// Go flag default value
		//
		// MATCH             -KKKKKKK----VVVV-
		//       flag.String("api-key", "shhh", "usage")
		//       flag.StringVar(&name, "api-key", "shhh", "usage")
		//       flag.StringVar(apiKeyNameVar, "shhh", "usage")
		{
			Name:      GoFlagDefaultValSetter.String(),
			Processor: search.Setter.String(),
			SetterProcessorConfig: config.SetterProcessorConfig{
				FileExts: GoExtPaths(),
				KeyTmpls: []string{DblQuoteNoneKey},
				Operator: CommaOper,
				ValTmpls: []string{DblQuoteVal},
			},
		},

		//
		// Ruby setters

		// Ruby variable assignment
		//
		// MATCH KKKKKK----VVVV-
		//       apiKey = "shhh"
		//       apiKey = 'shhh'
		{
			Name:      RubyVarAssignSetter.String(),
			Processor: search.Setter.String(),
			SetterProcessorConfig: config.SetterProcessorConfig{
				FileExts: RubyExtPaths(),
				KeyTmpls: []string{VarKey},
				Operator: EqOper,
				ValTmpls: []string{SingleDblQuoteVal},
			},
		},

		// Ruby hash field assignment
		//
		// MATCH        --KKKKKKK------VVVV-
		//       secrets['api_key'] = 'shhh'
		//       secrets["api_key"] = "shhh"
		//       secrets[:api_key] = "shhh"
		//       secrets[api_key_name_var] = 'shhh'
		{
			Name:      RubyHashFieldAssignSetter.String(),
			Processor: search.Setter.String(),
			SetterProcessorConfig: config.SetterProcessorConfig{
				FileExts: RubyExtPaths(),
				KeyTmpls: []string{BracketSingleDblCommaNoneKey},
				Operator: EqOper,
				ValTmpls: []string{SingleDblQuoteVal},
			},
		},

		// Ruby arrow parameters
		//
		// MATCH  -KKKKKKK------VVVV-
		//       {"api_key" => "shhh", "foo" => "bar"}
		//       {:api_key => "shhh", :foo => 'bar'}
		//       Hash["api_key" => "shhh", "foo" => "bar"]
		//       Hash[:api_key => "shhh", :foo => "bar"]
		//       Hash[api_key_name_var => "shhh", foo => "bar"]
		//       Hash[api_key: "shhh", foo: "bar"]
		//       call_func api_key: "shhh"
		//       call_func :api_key => "shhh"
		//       call_func 'api_key' => "shhh"
		//       call_func "api_key" => "shhh"
		{
			Name:      RubyArrowParamSetter.String(),
			Processor: search.Setter.String(),
			SetterProcessorConfig: config.SetterProcessorConfig{
				FileExts: RubyExtPaths(),
				KeyTmpls: []string{SingleDblColonNoneKey},
				Operator: EqArrowOper,
				ValTmpls: []string{SingleDblQuoteVal},
			},
		},

		// Ruby colon parameters
		//
		// MATCH -KKKKKKK---VVVV-
		//       {api_key: "shhh", foo: "bar"}
		//       Hash[api_key: "shhh", foo: "bar"]
		//       call_func color: color
		{
			Name:      RubyColonParamSetter.String(),
			Processor: search.Setter.String(),
			SetterProcessorConfig: config.SetterProcessorConfig{
				FileExts: RubyExtPaths(),
				KeyTmpls: []string{VarKey},
				Operator: ColonOper,
				ValTmpls: []string{SingleDblQuoteVal},
			},
		},

		//
		// Config setters

		// Systemd service config file environment var
		//
		// MATCH ------------KKKKKKKKKKKKKKKKKKK-VVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVV
		//       Environment=AGGREGATES_PASSWORD=27d09f46d6b94d07a7f803191ef49f81
		{
			Name:      ConfParamSystemdServiceEnvVarSetter.String(),
			Processor: search.Setter.String(),
			SetterProcessorConfig: config.SetterProcessorConfig{
				FileExts: SystemdConfExtPaths(),
				MainTmpl: `Environment=` + KeyTmplVar + EqOper + ValTmplVar,
				KeyTmpls: []string{VarKey},
				ValTmpls: []string{JustVal},
			},
		},

		// Logstash-style config file
		//
		// MATCH                -KKKKKKKKK------VVVV-
		//       add_field => { "api-token" => "shhh"  }
		// MATCH                -KKKKKKKKK---------------------------VVVV--
		//       add_field => { "api-token" => "${API_TOKEN_VAR_NAME:shhh}"  }
		{
			Name:      ConfParamLogstashStyleSetter.String(),
			Processor: search.Setter.String(),
			SetterProcessorConfig: config.SetterProcessorConfig{
				FileExts: ConfExtPaths(),
				KeyTmpls: []string{DblQuoteKey},
				Operator: EqArrowOper,
				ValTmpls: []string{DblQuoteVal},
			},
		},

		// Logstash-style config file env key
		//
		// MATCH                                --KKKKKKKKKKKKKKKKKK-VVVV-
		//       add_field => { "api-token" => "${API_TOKEN_VAR_NAME:shhh}"  }
		{
			Name:      ConfParamLogstashStyleEnvVarDefaultSetter.String(),
			Processor: search.Setter.String(),
			SetterProcessorConfig: config.SetterProcessorConfig{
				FileExts: ConfExtPaths(),
				MainTmpl: `\$` + OpenBrace + KeyTmplVar + `:` + ValTmplVar + CloseBrace,
				KeyTmpls: []string{VarKey},
				ValTmpls: []string{JustVal},
			},
		},

		//
		// Shell setters

		// Shell script variable assignment
		//
		// MATCH KKKKKKK--VVVV-
		//       api_key="shhh"
		//       API_KEY="shhh"
		// MATCH        KKKKKKK--VVVV-
		//       export API_KEY="shhh"
		//       local api_key=shhh
		{
			Name:      ShellScriptVarAssignSetter.String(),
			Processor: search.Setter.String(),
			SetterProcessorConfig: config.SetterProcessorConfig{
				FileExts: ShellScriptExtPaths(),
				KeyTmpls: []string{SingleDblQuoteNoneKey},
				Operator: EqOper,
				ValTmpls: []string{SingleDblQuoteJustVal},
			},
		},

		// Command parameters
		//
		// MATCH             --KKKKKKK--VVVV-
		//       ./script.sh --api-key="shhh"
		//       ./script.sh --api-key "shhh"
		//       ./script.sh --api-key shhh
		//       ./script.sh --api-key  shhh
		{
			Name:      ShellCmdParamValSetter.String(),
			Processor: search.Setter.String(),
			SetterProcessorConfig: config.SetterProcessorConfig{
				FileExts:     AnyPath(),
				KeyTmpls:     []string{SingleDblNoQuotePattern + `-?-` + Key + SingleDblNoQuotePattern},
				Operator:     `(?:=| {1,10})`, // Equals sign or some whitespace
				NoWhitespace: true,            // We're handling the whitespace in the operator where needed
				ValTmpls:     []string{SingleDblQuoteJustVal},
				KeyChars:     []string{AlphaChar, DigitChar, DashChar, UnderscoreChar},
			},
		},

		//
		// YAML setters

		// YAML object literal field value
		//
		//       secrets:
		// MATCH   KKKKKKK--VVVV
		//         api_key: shhh
		//         api_key_doub: 'shhh'
		//         api_key_sing: "shhh"
		{
			Name:      YAMLDictFieldValSetter.String(),
			Processor: search.Setter.String(),
			SetterProcessorConfig: config.SetterProcessorConfig{
				FileExts: YAMLExtPaths(),
				KeyTmpls: []string{SingleDblQuoteNoneKey},
				Operator: ColonOper,
				ValTmpls: []string{SingleDblQuoteJustVal},
			},
		},

		//
		// JSON setters

		// JSON object field value
		//
		// MATCH   -KKKKKKK----VVVV-
		//       { "api_key": "shhh" }
		{
			Name:      JSONObjFieldValSetter.String(),
			Processor: search.Setter.String(),
			SetterProcessorConfig: config.SetterProcessorConfig{
				FileExts: JSONExtPaths(),
				KeyTmpls: []string{DblQuoteKey},
				Operator: ColonOper,
				ValTmpls: []string{DblQuoteVal},
			},
		},

		//
		// XML setters
		//
		// Note: Parsing XML and HTML is not really the point of this processor, and regex is the
		// wrong tool for the job anyway, but I bet we can grab some low hanging fruit..

		// XML tag value
		//
		// MATCH -KKKKKK----------------------VVVV---------
		//       <apiKey unknown="attributes">shhh</apiKey>
		{
			Name:      XMLTagValSetter.String(),
			Processor: search.Setter.String(),
			SetterProcessorConfig: config.SetterProcessorConfig{
				FileExts: append(XMLExtPaths(), HTMLExtPaths()...),
				MainTmpl: `<` + KeyTmplVar + `[^>]*>` + ValTmplVar + `<\/[^>]+>`,
				KeyTmpls: []string{DblQuoteNoneKey},
				Operator: NoOper,
				ValTmpls: []string{JustVal},
			},
		},

		// XML tag value with key as attribute
		//
		// MATCH           --KKKKKK--VVVV---------
		//       <value key="apiKey">shhh</apiKey>
		// MATCH           --KKKKKK----VVVV----------
		//       <value key="apiKey" > shhh </apiKey>
		//       <value key="apiKey">shhh</apiKey>
		{
			Name:      XMLTagValKeyAsAttrSetter.String(),
			Processor: search.Setter.String(),
			SetterProcessorConfig: config.SetterProcessorConfig{
				FileExts: XMLExtPaths(),
				MainTmpl: `=?` + KeyTmplVar + `>\s*?` + ValTmplVar + `\s*?<\/[^>]+>`,
				KeyTmpls: []string{DblQuoteNoneKey},
				Operator: NoOper,
				ValTmpls: []string{JustVal},
			},
		},

		// XML attribute value
		//
		// MATCH         KKKKKK--VVVV-
		//       <secret apiKey="shhh">
		//       <secret apiKey='shhh'>
		//       <secret apiKey=shhh>
		{
			Name:      XMLAttrValSetter.String(),
			Processor: search.Setter.String(),
			SetterProcessorConfig: config.SetterProcessorConfig{
				FileExts: XMLExtPaths(),
				KeyTmpls: []string{VarKey},
				Operator: EqOper,
				ValTmpls: []string{SingleDblQuoteJustVal},
			},
		},

		//
		// HTML setters

		// HTML table row
		//
		// MATCH ----KKKKKKK---------VVVV-----
		//       <th>Api key</th><td>shhh</td>
		//       <td> Api key </td><td> shhh </td>
		// MATCH -------------------------KKKKKKK--------------------------VVVV-----
		//       <td unknown="attributes">Api key</td><td are="everywhere">shhh</td>
		{
			Name:      HTMLTableRowValSetter.String(),
			Processor: search.Setter.String(),
			SetterProcessorConfig: config.SetterProcessorConfig{
				FileExts: HTMLExtPaths(),
				MainTmpl: `` +
					// TH
					`<t[hd][^>]*?>` + // Open TH
					`\s*?` + // Whitespace
					KeyTmplVar + // Key
					`\s*?` + // Whitespace
					`<\/t[hd]>` + // Close TH

					`\s*?` + // Whitespace

					// TD
					`<td[^>]*?>` + // Open TD
					`\s*?` + // Whitespace
					`` + ValTmplVar + // Value
					`\s*?` + // Whitespace
					`<\/td>`, // Close TD
				Operator: NoOper,
				ValTmpls: []string{JustVal},
			},
		},

		// Generic setter
		//
		// This is run on any type of file to
		// catch the the things that might slip through the cracks.
		{
			Name:      GenericSetter.String(),
			Processor: search.Setter.String(),
			SetterProcessorConfig: config.SetterProcessorConfig{
				FileExts: AnyPath(),
				KeyTmpls: []string{SingleDblQuoteNoneKey},
				Operator: MatchAnyOf(EqOper, ColonOper) + `?`,
				ValTmpls: []string{SingleDblQuoteVal},
			},
		},
	}
}

// Regex processor definitions
func pemProcessorDefinitions() (result []*config.ProcessorConfig) {
	return []*config.ProcessorConfig{

		// RSA Private Key PEM processor
		{
			Name:      RSAPrivateKeyPEM.String(),
			Processor: search.PEM.String(),
			PEMProcessorConfig: config.PEMProcessorConfig{
				PEMType: "RSA PRIVATE KEY",
			},
		},

		// OpenSSH Private Key PEM processor
		{
			Name:      OpenSSHPrivateKeyPEM.String(),
			Processor: search.PEM.String(),
			PEMProcessorConfig: config.PEMProcessorConfig{
				PEMType: "OPENSSH PRIVATE KEY",
			},
		},

		// EC Private Key PEM processor
		{
			Name:      ECPrivateKeyPEM.String(),
			Processor: search.PEM.String(),
			PEMProcessorConfig: config.PEMProcessorConfig{
				PEMType: "EC PRIVATE KEY",
			},
		},

		// PGP Private Key Block PEM processor
		{
			Name:      PGPPrivateKeyBlockPEM.String(),
			Processor: search.PEM.String(),
			PEMProcessorConfig: config.PEMProcessorConfig{
				PEMType: "PGP PRIVATE KEY BLOCK",
			},
		},
	}
}

// Regex processor definitions
func regexProcessorDefinitions() (result []*config.ProcessorConfig) {
	return []*config.ProcessorConfig{

		// Slack token regex
		{
			Name:      SlackTokenRegex.String(),
			Processor: search.Regex.String(),
			RegexProcessorConfig: config.RegexProcessorConfig{
				RegexString: `(xox[p|b|o|a]-[0-9]{12}-[0-9]{12}-[0-9]{12}-[a-z0-9]{32})`,
			},
		},

		// Facebook OAuth regex
		{
			Name:      FacebookOAuthRegex.String(),
			Processor: search.Regex.String(),
			RegexProcessorConfig: config.RegexProcessorConfig{
				RegexString: `[f|F][a|A][c|C][e|E][b|B][o|O][o|O][k|K].*['|"][0-9a-f]{32}['|"]`,
			},
		},

		// Google OAuth regex
		{
			Name:      GoogleOAuthRegex.String(),
			Processor: search.Regex.String(),
			RegexProcessorConfig: config.RegexProcessorConfig{
				RegexString: `[t|T][w|W][i|I][t|T][t|T][e|E][r|R].*['|"][0-9a-zA-Z]{35,44}['|"]`,
			},
		},

		// Twitter regex
		{
			Name:      TwitterRegex.String(),
			Processor: search.Regex.String(),
			RegexProcessorConfig: config.RegexProcessorConfig{
				RegexString: `("client_secret":"[a-zA-Z0-9-_]{24}")`,
			},
		},

		// Heroku API Key regex
		{
			Name:      HerokuAPIKeyRegex.String(),
			Processor: search.Regex.String(),
			RegexProcessorConfig: config.RegexProcessorConfig{
				RegexString: `[h|H][e|E][r|R][o|O][k|K][u|U].*[0-9A-F]{8}-[0-9A-F]{4}-[0-9A-F]{4}-[0-9A-F]{4}-[0-9A-F]{12}`,
			},
		},

		// Slack Webhook regex
		{
			Name:      SlackWebhookRegex.String(),
			Processor: search.Regex.String(),
			RegexProcessorConfig: config.RegexProcessorConfig{
				RegexString: `https://hooks.slack.com/services/T[a-zA-Z0-9_]{8}/B[a-zA-Z0-9_]{8}/[a-zA-Z0-9_]{24}`,
			},
		},

		// GCP Service Account regex
		{
			Name:      GCPServiceAccountRegex.String(),
			Processor: search.Regex.String(),
			RegexProcessorConfig: config.RegexProcessorConfig{
				RegexString: `(?s){\s*"type": ?"service_account",.*"private_key_id": ?"([^"]+)"`,
			},
		},

		// Twilio API Key regex
		{
			Name:      TwilioAPIKeyRegex.String(),
			Processor: search.Regex.String(),
			RegexProcessorConfig: config.RegexProcessorConfig{
				RegexString: `SK[a-z0-9]{32}`,
			},
		},

		// URL Password regex
		{
			Name:      URLPasswordRegex.String(),
			Processor: search.Regex.String(),
			RegexProcessorConfig: config.RegexProcessorConfig{
				RegexString: `[a-z](?:[a-z]|\d|\+|-|\.)*://([a-zA-z0-9\-_]{4,20}:[a-zA-z0-9\-_]{4,20})@[a-zA-z0-9:.\-_/]*`,
			},
		},

		// Generic Secret regex
		{
			Name:      GenericSecretRegex.String(),
			Processor: search.Regex.String(),
			RegexProcessorConfig: config.RegexProcessorConfig{
				RegexString: `[s|S][e|E][c|C][r|R][e|E][t|T].*['|"][0-9a-zA-Z]{32,45}['|"]`,
			},
		},
	}
}

// Entropy processor definitions
func entropyProcessorDefinitions() (result []*config.ProcessorConfig) {
	return []*config.ProcessorConfig{

		// Base64 entropy
		{
			Name:      Base64Entropy.String(),
			Processor: search.Entropy.String(),
			EntropyProcessorConfig: config.EntropyProcessorConfig{
				Charset:             entropy.Base64CharsetName,
				WordLengthThreshold: 20,
				Threshold:           4.5,
				SkipPEMs:            true,
				WhitelistCodeMatch:  []string{

					//// Characters and obvious keyboard leans (order from more to less specific)
					//`(ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789\+/=)`,
					//`(ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789\+/)`,
					//`(0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz)`,
					//`(abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789)`,
					//`(abcdefghijklmnoprstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789)`,
					//`(abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ)`,
					//`(abcdefghijklmnopqrstuvwxyz0123456789)`,
					//`(ABCDEFGHIJKLMNOPQRSTUVWXYZ)`,
					//`(abcdefghijklmnopqrstuvwxyz)`,
					//`(BdgGhHiIjLmnOrsStTUwWYyZz)`,
					//`(BdgGhHiIjLmnNsStTUwWYyz)`,
					//`(AaBdgGhHiIjLmnrstTUYyZz)`,
					//
					//// Pantheon bindings
					//`/srv/bindings/[a-fA-F0-9]+`,
					//
					//// TODO Add a file path conditional to these processors, since this is only OK in a .travis.yml
					//`- secure: "([^"]+)"`,
					//`\.dockerconfigjson: secure: "([^"]+)"`,
					//
					//// TODO: These all need to be evaluated, so they are here until we reach that stage
					//`"(?:access_token)": "([^"]+)"`,
					//`(?:target_url|certificate-authority-data): (.+)$`,
					//`''(?:secret|read_device_credentials)'': ''[^"]+''`,
					//`proxy_set_header X-Frontend-Secret ([a-zA-Z0-9+/=]+)?`,
					//`auth0-client-id-dashboard=(.+)`,
					//`\b(?:ssh-rsa|ssh-ed25519) ''?AAAA[0-9A-Za-z+/]+[=]{0,3}\b`,
					//`variables[''PANTHEON_WPVULNDB_API_TOKEN''] = ''([^'']+)''`,
					//`auth0-client-id-dashboard=[A-Za-z0-9]+`,
					//`"audience": "([^"]+)",`,
					//`DNSActionRequiredLegacyNoHTTPSAlertBodyProviderUnknown`,
					//`Secret "pantheon" "([^"]+)"`,
				},
			},
		},

		// Hex entropy
		{
			Name:      HexEntropy.String(),
			Processor: search.Entropy.String(),
			EntropyProcessorConfig: config.EntropyProcessorConfig{
				Charset:             entropy.HexCharsetName,
				WordLengthThreshold: 20,
				Threshold:           3,
				SkipPEMs:            false,
				WhitelistCodeMatch:  []string{

					//`\.git@([^#]+)`,
					//
					//// Git commit hash
					//`git_ref\] = '([^']+)'`,
					//`git [^ ]+ ([0-9a-fA-F]+)`,
					//`>>>>>>> ([0-9a-fA-F]+)`,
					//
					//// Platform resources
					//`/srv/bindings/([a-fA-F0-9]+)`,
					//
					//// Ruby dict assignment
					//`\[:(?:.*checksum|rpm_sha)] = (?:"([^"]+)"|'[^']+')`,
					//
					//// JSON line with a certain key
					//`"(?:thread_hash)": "([^"]+)"`,
					//
					//// Cert auth in JSON
					//`{"ca":"2\.0\$([^\$]+)\$`,
					//`{"ca":"2\.0\$[^\$]+\$([^\$]+)\$`,
					//`{"ca":"2\.0\$[^\$]+\$[^\$]+\$([^\$]+)\$`,
					//
					//// Hashes
					//// FIME Can't remember what this was before but this isn't working now
					////        `* sha1('[^']+') = '([^'])'`,
					//`checksum "([a-fA-F0-9]+)"`,
					//`"[a-zA-Z0-9_-]+(?:_hash|_checksum)": "([A-Fa-f0-9]+)",?`,
					//`<li class=\\"(?:resource)\\" id=\\"([A-Fa-f0-9]+)\\">`,
					//
					//// Misc
					//`PIPEWISE_API_KEY = '([^']+)'`,
					//`\[\:db_insert_placeholder_\d+\] => `,
					//`hash.?salt.{30}`,
					//`md5sum="([^"]+)"`,
				},
			},
		},
	}
}
