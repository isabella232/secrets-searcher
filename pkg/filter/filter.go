package filter

import (
	"github.com/pantheon-systems/search-secrets/pkg/database/enum/reason"
	"github.com/pantheon-systems/search-secrets/pkg/structures"
)

type Filter struct {
	Repos   *structures.Index
	Reasons *structures.Index
}

func New(repos []string, reasons []string, skipEntropy bool) *Filter {

	var reasonNames []string
	if skipEntropy {
		entropy := reason.Entropy{}.New()
		for _, r := range reason.ReasonEnumValues() {
			if r != entropy {
				reasonNames = append(reasonNames, r.Value())
			}
		}
	} else {
		reasonNames = reasons
	}

	return &Filter{
		Repos:   structures.NewIndex(repos),
		Reasons: structures.NewIndex(reasonNames),
	}
}
