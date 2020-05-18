// Code generated by "stringer -type ProcessorType"; DO NOT EDIT.

package search

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Regex-0]
	_ = x[PEM-1]
	_ = x[Setter-2]
	_ = x[Entropy-3]
}

const _ProcessorType_name = "RegexPEMSetterEntropy"

var _ProcessorType_index = [...]uint8{0, 5, 8, 14, 21}

func (i ProcessorType) String() string {
	if i < 0 || i >= ProcessorType(len(_ProcessorType_index)-1) {
		return "ProcessorType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _ProcessorType_name[_ProcessorType_index[i]:_ProcessorType_index[i+1]]
}
