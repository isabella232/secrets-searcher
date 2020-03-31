package processor_type

//go:generate sh -c "go-genums ProcessorType value string pkg/database/enum/processor_type/processor_type_values.go > pkg/database/enum/processor_type/processor_type.go"/

const (
    valueRegex   = "regex"
    valuePEM     = "pem"
    valueEntropy = "entropy"
)
