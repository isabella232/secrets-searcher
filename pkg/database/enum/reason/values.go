package reason

//go:generate sh -c "go-genums Reason value string pkg/database/enum/reason/values.go > pkg/database/enum/reason/reason.go"/

const (
	valueEntropy              = "entropy"
	valueSlackToken           = "slack-token"
	valueRSAPrivateKey        = "rsa-private-key"
	valueSSHPrivateKeyOpenSSH = "ssh-private-key-openssh"
	valueSSHPrivateKeyDSA     = "ssh-private-key-dsa"
	valueSSHPrivateKeyEC      = "ssh-private-key-ec"
	valuePGPPrivateKeyBlock   = "pgp-private-key-block"
	valueFacebookOauth        = "facebook-oauth"
	valueTwitterOauth         = "twitter-oauth"
	valueGitHub               = "github"
	valueGoogleOauth          = "google-oauth"
	valueAWSAPIKey            = "aws-api-key"
	valueHerokuAPIKey         = "heroku-api-key"
	valueGenericSecret        = "generic-secret"
	valueGenericAPIKey        = "generic-api-key"
	valueSlackWebhook         = "slack-webhook"
	valueGCPServiceAccount    = "gcp-service-account"
	valueTwilioApiKey         = "twilio-api-key"
	valuePasswordInUrl        = "password-in-url"
)
