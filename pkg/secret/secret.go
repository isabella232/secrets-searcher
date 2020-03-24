package secret

import (
    "fmt"
    "github.com/pantheon-systems/search-secrets/pkg/database"
    "github.com/pantheon-systems/search-secrets/pkg/database/enum/decision"
    "github.com/pantheon-systems/search-secrets/pkg/database/enum/reason"
    "github.com/pantheon-systems/search-secrets/pkg/errors"
    "github.com/pantheon-systems/search-secrets/pkg/filter"
    "github.com/sirupsen/logrus"
)

type (
    Parser struct {
        filter *filter.Filter
        db     *database.Database
        log    *logrus.Logger
    }
    parsedSecret struct {
        Value    string
        Decision decision.DecisionEnum
    }
)

func NewParser(filter *filter.Filter, db *database.Database, log *logrus.Logger) *Parser {
    return &Parser{
        filter: filter,
        db:     db,
        log:    log,
    }
}

type DiffParser interface {
    Parse(finding *database.Finding, findingString *database.FindingString) (parsedSecrets []*parsedSecret, err error)
}

func (p *Parser) PrepareSecrets() (err error) {
    if p.db.TableExists(database.SecretTable) {
        //p.log.Warn("secret table already exists, skipping")
        //return
    }

    //return !strings.Contains(finding.Path, "/test/") &&
    //    !strings.Contains(finding.Path, "/tests/") &&
    //    !strings.Contains(finding.Path, "/spec/")

    findings, err := p.db.GetFindings()
    if err != nil {
        return errors.WithMessage(err, "unable to get findings")
    }

    for _, finding := range findings {
        findingStrings, err := p.db.GetFindingStringsForFinding(finding)
        for _, findingString := range findingStrings {
            var parser DiffParser

            switch finding.Reason.(type) {
            case reason.PasswordInUrl:
                parser = &PasswordInURLParser{}
            case reason.RSAPrivateKey,
                reason.SSHPrivateKeyOpenSSH,
                reason.SSHPrivateKeyDSA,
                reason.SSHPrivateKeyEC,
                reason.PGPPrivateKeyBlock:
                parser = &PEMParser{}
            case reason.Entropy,
                reason.SlackToken,
                reason.GitHub,
                reason.AWSAPIKey,
                reason.GoogleOauth,
                reason.GenericSecret,
                reason.GCPServiceAccount,
                reason.TwilioApiKey,
                reason.SlackWebhook,
                reason.FacebookOauth,
                reason.HerokuAPIKey,
                reason.GenericAPIKey,
                reason.TwitterOauth:
                parser = &NotImplementedParser{}
            default:
                return errors.Errorv("commit reason not implemented in secret parser", finding.Reason.String())
            }
            if err != nil {
                return errors.WithMessage(err, "unable to parse and write secret")
            }

            var parsedSecrets []*parsedSecret
            parsedSecrets, err = parser.Parse(finding, findingString)
            if err != nil {
                return err
            }
            if len(parsedSecrets) == 0 {
                parsedSecrets = []*parsedSecret{{
                    Value:    "",
                    Decision: decision.NotImplementedWithinParser{}.New(),
                }}
            }

            for _, parsedSecret := range parsedSecrets {
                err = p.WriteSecret(parsedSecret, findingString)
                if err != nil {
                    return err
                }
            }
        }
    }

    return
}

func (p *Parser) WriteSecret(parsedSecret *parsedSecret, findingString *database.FindingString) (err error) {
    var secretID = ""
    if parsedSecret.Value != "" {
        secretID = database.CreateHashID(parsedSecret.Value)
        secret := &database.Secret{
            ID:    secretID,
            Value: parsedSecret.Value,
        }
        if err = p.db.Write(database.SecretTable, secret.ID, secret); err != nil {
            return
        }
    }

    findingStringSecret := &database.Decision{
        ID:              database.CreateHashID(fmt.Sprintf("%s-%s", findingString.ID, secretID)),
        FindingStringID: findingString.ID,
        SecretID:        secretID,
        DecisionValue:   parsedSecret.Decision.Value(),
    }
    if err = p.db.Write(database.DecisionTable, findingStringSecret.ID, findingStringSecret); err != nil {
        return
    }

    return
}
