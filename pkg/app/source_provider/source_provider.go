package source_provider

// *** generated with go-genums ***

// SourceProviderEnum is the the enum interface that can be used
type SourceProviderEnum interface {
    String() string
    Value() string
    uniqueSourceProviderMethod()
}

// sourceProviderEnumBase is the internal, non-exported type
type sourceProviderEnumBase struct{ value string }

// Value() returns the enum value
func (eb sourceProviderEnumBase) Value() string { return eb.value }

// String() returns the enum name as you use it in Go code,
// needs to be overriden by inheriting types
func (eb sourceProviderEnumBase) String() string { return "" }

// Local is the enum type for 'valueLocal' value
type Local struct{ sourceProviderEnumBase }

// New is the constructor for a brand new SourceProviderEnum with value 'valueLocal'
func (Local) New() SourceProviderEnum { return Local{sourceProviderEnumBase{valueLocal}} }

// String returns always "Local" for this enum type
func (Local) String() string { return "Local" }

// uniqueSourceProviderMethod() guarantees that the enum interface cannot be mis-assigned with others defined with an otherwise identical signature
func (Local) uniqueSourceProviderMethod() {}

// GitHub is the enum type for 'valueGitHub' value
type GitHub struct{ sourceProviderEnumBase }

// New is the constructor for a brand new SourceProviderEnum with value 'valueGitHub'
func (GitHub) New() SourceProviderEnum { return GitHub{sourceProviderEnumBase{valueGitHub}} }

// String returns always "GitHub" for this enum type
func (GitHub) String() string { return "GitHub" }

// uniqueSourceProviderMethod() guarantees that the enum interface cannot be mis-assigned with others defined with an otherwise identical signature
func (GitHub) uniqueSourceProviderMethod() {}

var internalSourceProviderEnumValues = []SourceProviderEnum{
    Local{}.New(),
    GitHub{}.New(),
}

// SourceProviderEnumValues will return a slice of all allowed enum value types
func SourceProviderEnumValues() []SourceProviderEnum { return internalSourceProviderEnumValues[:] }

// NewSourceProviderFromValue will generate a valid enum from a value, or return nil in case of invalid value
func NewSourceProviderFromValue(v string) (result SourceProviderEnum) {
    switch v {
    case valueLocal:
        result = Local{}.New()
    case valueGitHub:
        result = GitHub{}.New()
    }
    return
}

// MustGetSourceProviderFromValue is the same as NewSourceProviderFromValue, but will panic in case of conversion failure
func MustGetSourceProviderFromValue(v string) SourceProviderEnum {
    result := NewSourceProviderFromValue(v)
    if result == nil {
        panic("invalid SourceProviderEnum value cast")
    }
    return result
}
