package logg

import (
	"strings"
)

//go:generate stringer -type Level

type Level int

const (
	Panic Level = iota
	Fatal
	Error
	Warning
	Info
	Debug
	Trace
)

func Levels() []Level {
	return []Level{
		Panic,
		Fatal,
		Error,
		Warning,
		Info,
		Debug,
		Trace,
	}
}

func (i Level) Value() string {
	return strings.ToLower(i.String())
}

func NewLevelFromValue(val string) Level {
	for _, e := range Levels() {
		if e.Value() == val {
			return e
		}
	}
	panic("unknown level: " + val)
}

func ValidLevelValues() (result []string) {
	levels := Levels()
	result = make([]string, len(levels))
	for i := range levels {
		result[i] = levels[i].Value()
	}
	return
}
