package logg

import (
	"io"
)

type (

	// Main logger interface
	Logg interface {
		WithField(key string, value interface{}) Logg
		WithFields(fields Fields) Logg
		WithError(err error) Logg
		WithPrefix(prefix string) Logg
		AddPrefixPath(prefix string) Logg
		Data() Fields
		Level() Level
		Output() io.Writer
		Spawn() Logg
		SetOutput(output io.Writer)
		Debugf(format string, args ...interface{})
		Infof(format string, args ...interface{})
		Warnf(format string, args ...interface{})
		Errorf(format string, args ...interface{})
		Debug(args ...interface{})
		Info(args ...interface{})
		Warn(args ...interface{})
		Error(args ...interface{})
		Tracef(format string, args ...interface{})
		Trace(args ...interface{})
		WithLazyField(key string, fnc func() interface{}) (result Logg)
	}

	Fields map[string]interface{}
)
