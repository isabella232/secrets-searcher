package secret

import (
    "github.com/pantheon-systems/search-secrets/pkg/database"
    "github.com/pantheon-systems/search-secrets/pkg/database/enum/decision"
)

type (
    NotImplementedParser struct{}
)

func (p *NotImplementedParser) Parse(*database.Finding, *database.FindingString) (parsedSecrets []*parsedSecret, err error) {
    parsedSecrets = []*parsedSecret{{Value: "", Decision: decision.ParserNotImplemented{}.New()}}
    return
}
