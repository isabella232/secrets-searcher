package processor_type

//go:generate sh -c "go run github.com/gdm85/go-genums ProcessorType value string processor_type_values.go > processor_type.go"

const (
    valueURL     = "url"
    valueRegex   = "regex"
    valuePEM     = "pem"
    valueEntropy = "entropy"
)
