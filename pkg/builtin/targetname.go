package builtin

//go:generate stringer -type TargetName

type TargetName int

const (
	Passwords TargetName = iota
	APIKeysAndTokens
)

// This is the order in which the targets are run.
func TargetNames() []TargetName {
	return []TargetName{
		Passwords,
		APIKeysAndTokens,
	}
}
