package processor_type

// *** generated with go-genums ***

// ProcessorTypeEnum is the the enum interface that can be used
type ProcessorTypeEnum interface {
	String() string
	Value() string
	uniqueProcessorTypeMethod()
}

// processorTypeEnumBase is the internal, non-exported type
type processorTypeEnumBase struct{ value string }

// Value() returns the enum value
func (eb processorTypeEnumBase) Value() string { return eb.value }

// String() returns the enum name as you use it in Go code,
// needs to be overriden by inheriting types
func (eb processorTypeEnumBase) String() string { return "" }

// Regex is the enum type for 'valueRegex' value
type Regex struct{ processorTypeEnumBase }

// New is the constructor for a brand new ProcessorTypeEnum with value 'valueRegex'
func (Regex) New() ProcessorTypeEnum { return Regex{processorTypeEnumBase{valueRegex}} }

// String returns always "Regex" for this enum type
func (Regex) String() string { return "Regex" }

// uniqueProcessorTypeMethod() guarantees that the enum interface cannot be mis-assigned with others defined with an otherwise identical signature
func (Regex) uniqueProcessorTypeMethod() {}

// PEM is the enum type for 'valuePEM' value
type PEM struct{ processorTypeEnumBase }

// New is the constructor for a brand new ProcessorTypeEnum with value 'valuePEM'
func (PEM) New() ProcessorTypeEnum { return PEM{processorTypeEnumBase{valuePEM}} }

// String returns always "PEM" for this enum type
func (PEM) String() string { return "PEM" }

// uniqueProcessorTypeMethod() guarantees that the enum interface cannot be mis-assigned with others defined with an otherwise identical signature
func (PEM) uniqueProcessorTypeMethod() {}

// Entropy is the enum type for 'valueEntropy' value
type Entropy struct{ processorTypeEnumBase }

// New is the constructor for a brand new ProcessorTypeEnum with value 'valueEntropy'
func (Entropy) New() ProcessorTypeEnum { return Entropy{processorTypeEnumBase{valueEntropy}} }

// String returns always "Entropy" for this enum type
func (Entropy) String() string { return "Entropy" }

// uniqueProcessorTypeMethod() guarantees that the enum interface cannot be mis-assigned with others defined with an otherwise identical signature
func (Entropy) uniqueProcessorTypeMethod() {}

var internalProcessorTypeEnumValues = []ProcessorTypeEnum{
	Regex{}.New(),
	PEM{}.New(),
	Entropy{}.New(),
}

// ProcessorTypeEnumValues will return a slice of all allowed enum value types
func ProcessorTypeEnumValues() []ProcessorTypeEnum { return internalProcessorTypeEnumValues[:] }

// NewProcessorTypeFromValue will generate a valid enum from a value, or return nil in case of invalid value
func NewProcessorTypeFromValue(v string) (result ProcessorTypeEnum) {
	switch v {
	case valueRegex:
		result = Regex{}.New()
	case valuePEM:
		result = PEM{}.New()
	case valueEntropy:
		result = Entropy{}.New()
	}
	return
}

// MustGetProcessorTypeFromValue is the same as NewProcessorTypeFromValue, but will panic in case of conversion failure
func MustGetProcessorTypeFromValue(v string) ProcessorTypeEnum {
	result := NewProcessorTypeFromValue(v)
	if result == nil {
		panic("invalid ProcessorTypeEnum value cast")
	}
	return result
}
