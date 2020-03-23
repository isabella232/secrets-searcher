package reason

// *** generated with go-genums ***

// ReasonEnum is the the enum interface that can be used
type ReasonEnum interface {
	String() string
	Value() string
	uniqueReasonMethod()
}

// reasonEnumBase is the internal, non-exported type
type reasonEnumBase struct{ value string }

// Value() returns the enum value
func (eb reasonEnumBase) Value() string { return eb.value }

// String() returns the enum name as you use it in Go code,
// needs to be overriden by inheriting types
func (eb reasonEnumBase) String() string { return "" }

// SlackToken is the enum type for 'valueSlackToken' value
type SlackToken struct{ reasonEnumBase }

// New is the constructor for a brand new ReasonEnum with value 'valueSlackToken'
func (SlackToken) New() ReasonEnum { return SlackToken{reasonEnumBase{valueSlackToken}} }

// String returns always "SlackToken" for this enum type
func (SlackToken) String() string { return "SlackToken" }

// uniqueReasonMethod() guarantees that the enum interface cannot be mis-assigned with others defined with an otherwise identical signature
func (SlackToken) uniqueReasonMethod() {}

// PGPPrivateKeyBlock is the enum type for 'valuePGPPrivateKeyBlock' value
type PGPPrivateKeyBlock struct{ reasonEnumBase }

// New is the constructor for a brand new ReasonEnum with value 'valuePGPPrivateKeyBlock'
func (PGPPrivateKeyBlock) New() ReasonEnum { return PGPPrivateKeyBlock{reasonEnumBase{valuePGPPrivateKeyBlock}} }

// String returns always "PGPPrivateKeyBlock" for this enum type
func (PGPPrivateKeyBlock) String() string { return "PGPPrivateKeyBlock" }

// uniqueReasonMethod() guarantees that the enum interface cannot be mis-assigned with others defined with an otherwise identical signature
func (PGPPrivateKeyBlock) uniqueReasonMethod() {}

// GitHub is the enum type for 'valueGitHub' value
type GitHub struct{ reasonEnumBase }

// New is the constructor for a brand new ReasonEnum with value 'valueGitHub'
func (GitHub) New() ReasonEnum { return GitHub{reasonEnumBase{valueGitHub}} }

// String returns always "GitHub" for this enum type
func (GitHub) String() string { return "GitHub" }

// uniqueReasonMethod() guarantees that the enum interface cannot be mis-assigned with others defined with an otherwise identical signature
func (GitHub) uniqueReasonMethod() {}

// AWSAPIKey is the enum type for 'valueAWSAPIKey' value
type AWSAPIKey struct{ reasonEnumBase }

// New is the constructor for a brand new ReasonEnum with value 'valueAWSAPIKey'
func (AWSAPIKey) New() ReasonEnum { return AWSAPIKey{reasonEnumBase{valueAWSAPIKey}} }

// String returns always "AWSAPIKey" for this enum type
func (AWSAPIKey) String() string { return "AWSAPIKey" }

// uniqueReasonMethod() guarantees that the enum interface cannot be mis-assigned with others defined with an otherwise identical signature
func (AWSAPIKey) uniqueReasonMethod() {}

// RSAPrivateKey is the enum type for 'valueRSAPrivateKey' value
type RSAPrivateKey struct{ reasonEnumBase }

// New is the constructor for a brand new ReasonEnum with value 'valueRSAPrivateKey'
func (RSAPrivateKey) New() ReasonEnum { return RSAPrivateKey{reasonEnumBase{valueRSAPrivateKey}} }

// String returns always "RSAPrivateKey" for this enum type
func (RSAPrivateKey) String() string { return "RSAPrivateKey" }

// uniqueReasonMethod() guarantees that the enum interface cannot be mis-assigned with others defined with an otherwise identical signature
func (RSAPrivateKey) uniqueReasonMethod() {}

// GoogleOauth is the enum type for 'valueGoogleOauth' value
type GoogleOauth struct{ reasonEnumBase }

// New is the constructor for a brand new ReasonEnum with value 'valueGoogleOauth'
func (GoogleOauth) New() ReasonEnum { return GoogleOauth{reasonEnumBase{valueGoogleOauth}} }

// String returns always "GoogleOauth" for this enum type
func (GoogleOauth) String() string { return "GoogleOauth" }

// uniqueReasonMethod() guarantees that the enum interface cannot be mis-assigned with others defined with an otherwise identical signature
func (GoogleOauth) uniqueReasonMethod() {}

// GenericSecret is the enum type for 'valueGenericSecret' value
type GenericSecret struct{ reasonEnumBase }

// New is the constructor for a brand new ReasonEnum with value 'valueGenericSecret'
func (GenericSecret) New() ReasonEnum { return GenericSecret{reasonEnumBase{valueGenericSecret}} }

// String returns always "GenericSecret" for this enum type
func (GenericSecret) String() string { return "GenericSecret" }

// uniqueReasonMethod() guarantees that the enum interface cannot be mis-assigned with others defined with an otherwise identical signature
func (GenericSecret) uniqueReasonMethod() {}

// GCPServiceAccount is the enum type for 'valueGCPServiceAccount' value
type GCPServiceAccount struct{ reasonEnumBase }

// New is the constructor for a brand new ReasonEnum with value 'valueGCPServiceAccount'
func (GCPServiceAccount) New() ReasonEnum { return GCPServiceAccount{reasonEnumBase{valueGCPServiceAccount}} }

// String returns always "GCPServiceAccount" for this enum type
func (GCPServiceAccount) String() string { return "GCPServiceAccount" }

// uniqueReasonMethod() guarantees that the enum interface cannot be mis-assigned with others defined with an otherwise identical signature
func (GCPServiceAccount) uniqueReasonMethod() {}

// TwilioApiKey is the enum type for 'valueTwilioApiKey' value
type TwilioApiKey struct{ reasonEnumBase }

// New is the constructor for a brand new ReasonEnum with value 'valueTwilioApiKey'
func (TwilioApiKey) New() ReasonEnum { return TwilioApiKey{reasonEnumBase{valueTwilioApiKey}} }

// String returns always "TwilioApiKey" for this enum type
func (TwilioApiKey) String() string { return "TwilioApiKey" }

// uniqueReasonMethod() guarantees that the enum interface cannot be mis-assigned with others defined with an otherwise identical signature
func (TwilioApiKey) uniqueReasonMethod() {}

// SlackWebhook is the enum type for 'valueSlackWebhook' value
type SlackWebhook struct{ reasonEnumBase }

// New is the constructor for a brand new ReasonEnum with value 'valueSlackWebhook'
func (SlackWebhook) New() ReasonEnum { return SlackWebhook{reasonEnumBase{valueSlackWebhook}} }

// String returns always "SlackWebhook" for this enum type
func (SlackWebhook) String() string { return "SlackWebhook" }

// uniqueReasonMethod() guarantees that the enum interface cannot be mis-assigned with others defined with an otherwise identical signature
func (SlackWebhook) uniqueReasonMethod() {}

// Entropy is the enum type for 'valueEntropy' value
type Entropy struct{ reasonEnumBase }

// New is the constructor for a brand new ReasonEnum with value 'valueEntropy'
func (Entropy) New() ReasonEnum { return Entropy{reasonEnumBase{valueEntropy}} }

// String returns always "Entropy" for this enum type
func (Entropy) String() string { return "Entropy" }

// uniqueReasonMethod() guarantees that the enum interface cannot be mis-assigned with others defined with an otherwise identical signature
func (Entropy) uniqueReasonMethod() {}

// SSHPrivateKeyOpenSSH is the enum type for 'valueSSHPrivateKeyOpenSSH' value
type SSHPrivateKeyOpenSSH struct{ reasonEnumBase }

// New is the constructor for a brand new ReasonEnum with value 'valueSSHPrivateKeyOpenSSH'
func (SSHPrivateKeyOpenSSH) New() ReasonEnum { return SSHPrivateKeyOpenSSH{reasonEnumBase{valueSSHPrivateKeyOpenSSH}} }

// String returns always "SSHPrivateKeyOpenSSH" for this enum type
func (SSHPrivateKeyOpenSSH) String() string { return "SSHPrivateKeyOpenSSH" }

// uniqueReasonMethod() guarantees that the enum interface cannot be mis-assigned with others defined with an otherwise identical signature
func (SSHPrivateKeyOpenSSH) uniqueReasonMethod() {}

// SSHPrivateKeyDSA is the enum type for 'valueSSHPrivateKeyDSA' value
type SSHPrivateKeyDSA struct{ reasonEnumBase }

// New is the constructor for a brand new ReasonEnum with value 'valueSSHPrivateKeyDSA'
func (SSHPrivateKeyDSA) New() ReasonEnum { return SSHPrivateKeyDSA{reasonEnumBase{valueSSHPrivateKeyDSA}} }

// String returns always "SSHPrivateKeyDSA" for this enum type
func (SSHPrivateKeyDSA) String() string { return "SSHPrivateKeyDSA" }

// uniqueReasonMethod() guarantees that the enum interface cannot be mis-assigned with others defined with an otherwise identical signature
func (SSHPrivateKeyDSA) uniqueReasonMethod() {}

// SSHPrivateKeyEC is the enum type for 'valueSSHPrivateKeyEC' value
type SSHPrivateKeyEC struct{ reasonEnumBase }

// New is the constructor for a brand new ReasonEnum with value 'valueSSHPrivateKeyEC'
func (SSHPrivateKeyEC) New() ReasonEnum { return SSHPrivateKeyEC{reasonEnumBase{valueSSHPrivateKeyEC}} }

// String returns always "SSHPrivateKeyEC" for this enum type
func (SSHPrivateKeyEC) String() string { return "SSHPrivateKeyEC" }

// uniqueReasonMethod() guarantees that the enum interface cannot be mis-assigned with others defined with an otherwise identical signature
func (SSHPrivateKeyEC) uniqueReasonMethod() {}

// FacebookOauth is the enum type for 'valueFacebookOauth' value
type FacebookOauth struct{ reasonEnumBase }

// New is the constructor for a brand new ReasonEnum with value 'valueFacebookOauth'
func (FacebookOauth) New() ReasonEnum { return FacebookOauth{reasonEnumBase{valueFacebookOauth}} }

// String returns always "FacebookOauth" for this enum type
func (FacebookOauth) String() string { return "FacebookOauth" }

// uniqueReasonMethod() guarantees that the enum interface cannot be mis-assigned with others defined with an otherwise identical signature
func (FacebookOauth) uniqueReasonMethod() {}

// HerokuAPIKey is the enum type for 'valueHerokuAPIKey' value
type HerokuAPIKey struct{ reasonEnumBase }

// New is the constructor for a brand new ReasonEnum with value 'valueHerokuAPIKey'
func (HerokuAPIKey) New() ReasonEnum { return HerokuAPIKey{reasonEnumBase{valueHerokuAPIKey}} }

// String returns always "HerokuAPIKey" for this enum type
func (HerokuAPIKey) String() string { return "HerokuAPIKey" }

// uniqueReasonMethod() guarantees that the enum interface cannot be mis-assigned with others defined with an otherwise identical signature
func (HerokuAPIKey) uniqueReasonMethod() {}

// GenericAPIKey is the enum type for 'valueGenericAPIKey' value
type GenericAPIKey struct{ reasonEnumBase }

// New is the constructor for a brand new ReasonEnum with value 'valueGenericAPIKey'
func (GenericAPIKey) New() ReasonEnum { return GenericAPIKey{reasonEnumBase{valueGenericAPIKey}} }

// String returns always "GenericAPIKey" for this enum type
func (GenericAPIKey) String() string { return "GenericAPIKey" }

// uniqueReasonMethod() guarantees that the enum interface cannot be mis-assigned with others defined with an otherwise identical signature
func (GenericAPIKey) uniqueReasonMethod() {}

// TwitterOauth is the enum type for 'valueTwitterOauth' value
type TwitterOauth struct{ reasonEnumBase }

// New is the constructor for a brand new ReasonEnum with value 'valueTwitterOauth'
func (TwitterOauth) New() ReasonEnum { return TwitterOauth{reasonEnumBase{valueTwitterOauth}} }

// String returns always "TwitterOauth" for this enum type
func (TwitterOauth) String() string { return "TwitterOauth" }

// uniqueReasonMethod() guarantees that the enum interface cannot be mis-assigned with others defined with an otherwise identical signature
func (TwitterOauth) uniqueReasonMethod() {}

// PasswordInUrl is the enum type for 'valuePasswordInUrl' value
type PasswordInUrl struct{ reasonEnumBase }

// New is the constructor for a brand new ReasonEnum with value 'valuePasswordInUrl'
func (PasswordInUrl) New() ReasonEnum { return PasswordInUrl{reasonEnumBase{valuePasswordInUrl}} }

// String returns always "PasswordInUrl" for this enum type
func (PasswordInUrl) String() string { return "PasswordInUrl" }

// uniqueReasonMethod() guarantees that the enum interface cannot be mis-assigned with others defined with an otherwise identical signature
func (PasswordInUrl) uniqueReasonMethod() {}

var internalReasonEnumValues = []ReasonEnum{
	SlackToken{}.New(),
	PGPPrivateKeyBlock{}.New(),
	GitHub{}.New(),
	AWSAPIKey{}.New(),
	RSAPrivateKey{}.New(),
	GoogleOauth{}.New(),
	GenericSecret{}.New(),
	GCPServiceAccount{}.New(),
	TwilioApiKey{}.New(),
	SlackWebhook{}.New(),
	Entropy{}.New(),
	SSHPrivateKeyOpenSSH{}.New(),
	SSHPrivateKeyDSA{}.New(),
	SSHPrivateKeyEC{}.New(),
	FacebookOauth{}.New(),
	HerokuAPIKey{}.New(),
	GenericAPIKey{}.New(),
	TwitterOauth{}.New(),
	PasswordInUrl{}.New(),
}

// ReasonEnumValues will return a slice of all allowed enum value types
func ReasonEnumValues() []ReasonEnum { return internalReasonEnumValues[:] }

// NewReasonFromValue will generate a valid enum from a value, or return nil in case of invalid value
func NewReasonFromValue(v string) (result ReasonEnum) {
	switch v {
	case valueSlackToken:
		result = SlackToken{}.New()
	case valuePGPPrivateKeyBlock:
		result = PGPPrivateKeyBlock{}.New()
	case valueGitHub:
		result = GitHub{}.New()
	case valueAWSAPIKey:
		result = AWSAPIKey{}.New()
	case valueRSAPrivateKey:
		result = RSAPrivateKey{}.New()
	case valueGoogleOauth:
		result = GoogleOauth{}.New()
	case valueGenericSecret:
		result = GenericSecret{}.New()
	case valueGCPServiceAccount:
		result = GCPServiceAccount{}.New()
	case valueTwilioApiKey:
		result = TwilioApiKey{}.New()
	case valueSlackWebhook:
		result = SlackWebhook{}.New()
	case valueEntropy:
		result = Entropy{}.New()
	case valueSSHPrivateKeyOpenSSH:
		result = SSHPrivateKeyOpenSSH{}.New()
	case valueSSHPrivateKeyDSA:
		result = SSHPrivateKeyDSA{}.New()
	case valueSSHPrivateKeyEC:
		result = SSHPrivateKeyEC{}.New()
	case valueFacebookOauth:
		result = FacebookOauth{}.New()
	case valueHerokuAPIKey:
		result = HerokuAPIKey{}.New()
	case valueGenericAPIKey:
		result = GenericAPIKey{}.New()
	case valueTwitterOauth:
		result = TwitterOauth{}.New()
	case valuePasswordInUrl:
		result = PasswordInUrl{}.New()
	}
	return
}

// MustGetReasonFromValue is the same as NewReasonFromValue, but will panic in case of conversion failure
func MustGetReasonFromValue(v string) ReasonEnum {
	result := NewReasonFromValue(v)
	if result == nil {
		panic("invalid ReasonEnum value cast")
	}
	return result
}
