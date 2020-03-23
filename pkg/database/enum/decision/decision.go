package decision

// *** generated with go-genums ***

// DecisionEnum is the the enum interface that can be used
type DecisionEnum interface {
	String() string
	Value() string
	uniqueDecisionMethod()
}

// decisionEnumBase is the internal, non-exported type
type decisionEnumBase struct{ value string }

// Value() returns the enum value
func (eb decisionEnumBase) Value() string { return eb.value }

// String() returns the enum name as you use it in Go code,
// needs to be overriden by inheriting types
func (eb decisionEnumBase) String() string { return "" }

// IgnoreExampleCode is the enum type for 'valueIgnoreExampleCode' value
type IgnoreExampleCode struct{ decisionEnumBase }

// New is the constructor for a brand new DecisionEnum with value 'valueIgnoreExampleCode'
func (IgnoreExampleCode) New() DecisionEnum { return IgnoreExampleCode{decisionEnumBase{valueIgnoreExampleCode}} }

// String returns always "IgnoreExampleCode" for this enum type
func (IgnoreExampleCode) String() string { return "IgnoreExampleCode" }

// uniqueDecisionMethod() guarantees that the enum interface cannot be mis-assigned with others defined with an otherwise identical signature
func (IgnoreExampleCode) uniqueDecisionMethod() {}

// IgnoreTemplateVars is the enum type for 'valueIgnoreTemplateVars' value
type IgnoreTemplateVars struct{ decisionEnumBase }

// New is the constructor for a brand new DecisionEnum with value 'valueIgnoreTemplateVars'
func (IgnoreTemplateVars) New() DecisionEnum { return IgnoreTemplateVars{decisionEnumBase{valueIgnoreTemplateVars}} }

// String returns always "IgnoreTemplateVars" for this enum type
func (IgnoreTemplateVars) String() string { return "IgnoreTemplateVars" }

// uniqueDecisionMethod() guarantees that the enum interface cannot be mis-assigned with others defined with an otherwise identical signature
func (IgnoreTemplateVars) uniqueDecisionMethod() {}

// NeedsInvestigation is the enum type for 'valueNeedsInvestigation' value
type NeedsInvestigation struct{ decisionEnumBase }

// New is the constructor for a brand new DecisionEnum with value 'valueNeedsInvestigation'
func (NeedsInvestigation) New() DecisionEnum { return NeedsInvestigation{decisionEnumBase{valueNeedsInvestigation}} }

// String returns always "NeedsInvestigation" for this enum type
func (NeedsInvestigation) String() string { return "NeedsInvestigation" }

// uniqueDecisionMethod() guarantees that the enum interface cannot be mis-assigned with others defined with an otherwise identical signature
func (NeedsInvestigation) uniqueDecisionMethod() {}

// DoNotKnowYet is the enum type for 'valueDoNotKnowYet' value
type DoNotKnowYet struct{ decisionEnumBase }

// New is the constructor for a brand new DecisionEnum with value 'valueDoNotKnowYet'
func (DoNotKnowYet) New() DecisionEnum { return DoNotKnowYet{decisionEnumBase{valueDoNotKnowYet}} }

// String returns always "DoNotKnowYet" for this enum type
func (DoNotKnowYet) String() string { return "DoNotKnowYet" }

// uniqueDecisionMethod() guarantees that the enum interface cannot be mis-assigned with others defined with an otherwise identical signature
func (DoNotKnowYet) uniqueDecisionMethod() {}

// NotImplemented is the enum type for 'valueNotImplemented' value
type NotImplemented struct{ decisionEnumBase }

// New is the constructor for a brand new DecisionEnum with value 'valueNotImplemented'
func (NotImplemented) New() DecisionEnum { return NotImplemented{decisionEnumBase{valueNotImplemented}} }

// String returns always "NotImplemented" for this enum type
func (NotImplemented) String() string { return "NotImplemented" }

// uniqueDecisionMethod() guarantees that the enum interface cannot be mis-assigned with others defined with an otherwise identical signature
func (NotImplemented) uniqueDecisionMethod() {}

// IgnoreTestFiles is the enum type for 'valueIgnoreTestFiles' value
type IgnoreTestFiles struct{ decisionEnumBase }

// New is the constructor for a brand new DecisionEnum with value 'valueIgnoreTestFiles'
func (IgnoreTestFiles) New() DecisionEnum { return IgnoreTestFiles{decisionEnumBase{valueIgnoreTestFiles}} }

// String returns always "IgnoreTestFiles" for this enum type
func (IgnoreTestFiles) String() string { return "IgnoreTestFiles" }

// uniqueDecisionMethod() guarantees that the enum interface cannot be mis-assigned with others defined with an otherwise identical signature
func (IgnoreTestFiles) uniqueDecisionMethod() {}

var internalDecisionEnumValues = []DecisionEnum{
	IgnoreExampleCode{}.New(),
	IgnoreTemplateVars{}.New(),
	NeedsInvestigation{}.New(),
	DoNotKnowYet{}.New(),
	NotImplemented{}.New(),
	IgnoreTestFiles{}.New(),
}

// DecisionEnumValues will return a slice of all allowed enum value types
func DecisionEnumValues() []DecisionEnum { return internalDecisionEnumValues[:] }

// NewDecisionFromValue will generate a valid enum from a value, or return nil in case of invalid value
func NewDecisionFromValue(v string) (result DecisionEnum) {
	switch v {
	case valueIgnoreExampleCode:
		result = IgnoreExampleCode{}.New()
	case valueIgnoreTemplateVars:
		result = IgnoreTemplateVars{}.New()
	case valueNeedsInvestigation:
		result = NeedsInvestigation{}.New()
	case valueDoNotKnowYet:
		result = DoNotKnowYet{}.New()
	case valueNotImplemented:
		result = NotImplemented{}.New()
	case valueIgnoreTestFiles:
		result = IgnoreTestFiles{}.New()
	}
	return
}

// MustGetDecisionFromValue is the same as NewDecisionFromValue, but will panic in case of conversion failure
func MustGetDecisionFromValue(v string) DecisionEnum {
	result := NewDecisionFromValue(v)
	if result == nil {
		panic("invalid DecisionEnum value cast")
	}
	return result
}
