package source_provider

//go:generate sh -c "go-genums SourceProvider value string source_provider_values.go > source_provider.go"

const (
    valueLocal  = "local"
    valueGitHub = "github"
)
