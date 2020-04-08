package processor_type

//go:generate sh -c "go-genums ProcessorType value string processor_type_values.go > processor_type.go"

const (
    valueRegex   = "regex"
    valuePEM     = "pem"
    valueEntropy = "entropy"
)
