package source_provider

//go:generate sh -c "go run github.com/gdm85/go-genums SourceProvider value string source_provider_values.go > source_provider.go"

const (
    valueLocal  = "local"
    valueGitHub = "github"
)
