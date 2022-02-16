package source

import "strings"

//go:generate stringer -type Provider

type Provider int

const (
	Local Provider = iota
	Github
)

func Providers() []Provider {
	return []Provider{
		Local,
		Github,
	}
}

func (i Provider) Value() string {
	return strings.ToLower(i.String())
}

func ValidProviderValues() (result []string) {
	providers := Providers()
	result = make([]string, len(providers))
	for i := range providers {
		result[i] = providers[i].Value()
	}
	return
}
