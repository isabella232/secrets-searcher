package builtin

//go:generate stringer -type ProcessorName

type ProcessorName int

const (
	URLPathParamValSetter ProcessorName = iota
	URLQueryStringParamValSetter
	PyVarAssignSetter
	PyDictFieldAssignSetter
	PyDictLiteralFieldSetter
	PyTupleSetter
	PHPVarAssignSetter
	PHPAssocArrayFieldAssignSetter
	PHPAssocArrayLiteralFieldSetter
	PHPConstDefineSetter
	JSVarAssignSetter
	JSObjFieldAssignSetter
	JSObjLiteralFieldSetter
	GoVarAssignSetter
	GoHashFieldAssignSetter
	GoHashLiteralFieldSetter
	GoFlagDefaultValSetter
	RubyVarAssignSetter
	RubyHashFieldAssignSetter
	RubyArrowParamSetter
	RubyColonParamSetter
	ConfParamSystemdServiceEnvVarSetter
	ConfParamLogstashStyleSetter
	ConfParamLogstashStyleEnvVarDefaultSetter
	ShellScriptVarAssignSetter
	ShellCmdParamValSetter
	YAMLDictFieldValSetter
	JSONObjFieldValSetter
	XMLTagValSetter
	XMLTagValKeyAsAttrSetter
	XMLAttrValSetter
	HTMLTableRowValSetter
	GenericSetter

	RSAPrivateKeyPEM
	OpenSSHPrivateKeyPEM
	ECPrivateKeyPEM
	PGPPrivateKeyBlockPEM

	SlackTokenRegex
	FacebookOAuthRegex
	GoogleOAuthRegex
	TwitterRegex
	HerokuAPIKeyRegex
	SlackWebhookRegex
	GCPServiceAccountRegex
	TwilioAPIKeyRegex
	URLPasswordRegex
	GenericSecretRegex

	Base64Entropy
	HexEntropy
)
