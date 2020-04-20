package diff_operation

// *** generated with go-genums ***

// DiffOperationEnum is the the enum interface that can be used
type DiffOperationEnum interface {
	String() string
	Value() diffOperationValue
	uniqueDiffOperationMethod()
}

// diffOperationEnumBase is the internal, non-exported type
type diffOperationEnumBase struct{ value diffOperationValue }

// Value() returns the enum value
func (eb diffOperationEnumBase) Value() diffOperationValue { return eb.value }

// String() returns the enum name as you use it in Go code,
// needs to be overriden by inheriting types
func (eb diffOperationEnumBase) String() string { return "" }

// Equal is the enum type for 'valueEqual' value
type Equal struct{ diffOperationEnumBase }

// New is the constructor for a brand new DiffOperationEnum with value 'valueEqual'
func (Equal) New() DiffOperationEnum { return Equal{diffOperationEnumBase{valueEqual}} }

// String returns always "Equal" for this enum type
func (Equal) String() string { return "Equal" }

// uniqueDiffOperationMethod() guarantees that the enum interface cannot be mis-assigned with others defined with an otherwise identical signature
func (Equal) uniqueDiffOperationMethod() {}

// Delete is the enum type for 'valueDelete' value
type Delete struct{ diffOperationEnumBase }

// New is the constructor for a brand new DiffOperationEnum with value 'valueDelete'
func (Delete) New() DiffOperationEnum { return Delete{diffOperationEnumBase{valueDelete}} }

// String returns always "Delete" for this enum type
func (Delete) String() string { return "Delete" }

// uniqueDiffOperationMethod() guarantees that the enum interface cannot be mis-assigned with others defined with an otherwise identical signature
func (Delete) uniqueDiffOperationMethod() {}

// Add is the enum type for 'valueAdd' value
type Add struct{ diffOperationEnumBase }

// New is the constructor for a brand new DiffOperationEnum with value 'valueAdd'
func (Add) New() DiffOperationEnum { return Add{diffOperationEnumBase{valueAdd}} }

// String returns always "Add" for this enum type
func (Add) String() string { return "Add" }

// uniqueDiffOperationMethod() guarantees that the enum interface cannot be mis-assigned with others defined with an otherwise identical signature
func (Add) uniqueDiffOperationMethod() {}

var internalDiffOperationEnumValues = []DiffOperationEnum{
	Equal{}.New(),
	Delete{}.New(),
	Add{}.New(),
}

// DiffOperationEnumValues will return a slice of all allowed enum value types
func DiffOperationEnumValues() []DiffOperationEnum { return internalDiffOperationEnumValues[:] }

// NewDiffOperationFromValue will generate a valid enum from a value, or return nil in case of invalid value
func NewDiffOperationFromValue(v diffOperationValue) (result DiffOperationEnum) {
	switch v {
	case valueEqual:
		result = Equal{}.New()
	case valueDelete:
		result = Delete{}.New()
	case valueAdd:
		result = Add{}.New()
	}
	return
}

// MustGetDiffOperationFromValue is the same as NewDiffOperationFromValue, but will panic in case of conversion failure
func MustGetDiffOperationFromValue(v diffOperationValue) DiffOperationEnum {
	result := NewDiffOperationFromValue(v)
	if result == nil {
		panic("invalid DiffOperationEnum value cast")
	}
	return result
}
