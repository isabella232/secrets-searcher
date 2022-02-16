package logg

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"sync"

	"github.com/pantheon-systems/secrets-searcher/pkg/manip"

	lr "github.com/sirupsen/logrus"
)

const prefixFieldName = "prefix"

type (
	// Logrus logg implementation
	LogrusLogg struct {
		ent         *lr.Entry
		runFncs     map[string]func() interface{}
		runOnceFncs map[string]func() interface{}
		mu          sync.Mutex
	}
)

func NewLogrusLogg(logrusInput interface{}) *LogrusLogg {
	return &LogrusLogg{
		ent:         getEntry(logrusInput),
		runFncs:     make(map[string]func() interface{}),
		runOnceFncs: make(map[string]func() interface{}),
		mu:          sync.Mutex{},
	}
}

func (l *LogrusLogg) Output() io.Writer {
	return l.ent.Logger.Out
}

func (l *LogrusLogg) SetOutput(output io.Writer) {
	l.ent.Logger.SetOutput(output)
}

//
// Fiddly wrappers

func (l *LogrusLogg) Data() (result Fields) {
	return getMap(l.ent.Data)
}

func (l *LogrusLogg) Level() Level {
	return NewLevelFromValue(l.ent.Logger.Level.String())
}

func (l *LogrusLogg) WithPrefix(prefix string) Logg {
	return l.spawnMe(l.ent.WithField(prefixFieldName, prefix))
}

func (l *LogrusLogg) Spawn() Logg {
	return l.spawnMe(l.spawnEntry())
}

func (l *LogrusLogg) AddPrefixPath(prefix string) Logg {
	var pieces []string

	// Existing prefix?
	prevPrefix, ok := l.ent.Data[prefixFieldName]
	if ok {
		pieces = append(pieces, prevPrefix.(string))
	}

	// Add new prefix
	pieces = append(pieces, prefix)

	// Build
	newPrefix := strings.Join(pieces, "/")

	return l.spawnMe(l.spawnEntry().WithField(prefixFieldName, newPrefix))
}

func (l *LogrusLogg) WithError(err error) Logg {
	return l.spawnMe(l.spawnEntry().WithError(err))
}

func (l *LogrusLogg) WithField(key string, value interface{}) (result Logg) {
	if fnc, ok := getFnc(value); ok {
		newLog := l.spawnMe(l.spawnEntry())
		newLog.runFncs[key] = fnc
		return newLog
	}

	entry := l.spawnEntry()
	field := entry.WithField(key, value)
	return l.spawnMe(field)
}

func (l *LogrusLogg) WithFields(fields Fields) (result Logg) {
	l.mu.Lock()
	defer l.mu.Unlock()

	runFncs := make(map[string]func() interface{})
	newFields := make(map[string]interface{})
	for key := range fields {
		if fnc, ok := getFnc(fields[key]); ok {
			runFncs[key] = fnc
			continue
		}

		if strVal, ok := fields[key].(string); ok {
			newFields[key] = manip.MakeOneLine(strVal, `\n`)
			continue
		}

		newFields[key] = fields[key]
	}

	newLog := l.spawnMe(l.spawnEntry().WithFields(newFields))
	for key, fnc := range runFncs {
		newLog.setRunFnc(key, fnc)
	}

	return newLog
}

func (l *LogrusLogg) WithLazyField(key string, fnc func() interface{}) (result Logg) {
	newLog := l.spawnMe(l.spawnEntry())
	newLog.setRunOnceFnc(key, fnc)

	return newLog
}

func (l *LogrusLogg) Tracef(format string, args ...interface{}) {
	l.logf(lr.TraceLevel, format, args)
}

func (l *LogrusLogg) Debugf(format string, args ...interface{}) {
	l.logf(lr.DebugLevel, format, args)
}

func (l *LogrusLogg) Infof(format string, args ...interface{}) {
	l.logf(lr.InfoLevel, format, args)
}

func (l *LogrusLogg) Warnf(format string, args ...interface{}) {
	l.logf(lr.WarnLevel, format, args)
}

func (l *LogrusLogg) Errorf(format string, args ...interface{}) {
	l.logf(lr.ErrorLevel, format, args)
}

func (l *LogrusLogg) Trace(args ...interface{}) {
	l.log(lr.TraceLevel, args)
}

func (l *LogrusLogg) Debug(args ...interface{}) {
	l.log(lr.DebugLevel, args)
}

func (l *LogrusLogg) Info(args ...interface{}) {
	l.log(lr.InfoLevel, args)
}

func (l *LogrusLogg) Warn(args ...interface{}) {
	l.log(lr.WarnLevel, args)
}

func (l *LogrusLogg) Error(args ...interface{}) {
	l.log(lr.ErrorLevel, args)
}

func getEntry(logrusInput interface{}) (result *lr.Entry) {
	switch logrusHmm := logrusInput.(type) {
	case *lr.Logger:
		result = lr.NewEntry(logrusHmm)
	case lr.Logger:
		result = lr.NewEntry(&logrusHmm)
	case lr.Entry:
		result = &logrusHmm
	case *lr.Entry:
		result = logrusHmm
	case *LogrusLogg:
		result = logrusHmm.ent
	case LogrusLogg:
		result = logrusHmm.ent
	default:
		panic("invalid object")
	}
	return
}

func (l *LogrusLogg) logf(lrLevel lr.Level, format string, args []interface{}) {
	if !l.ent.Logger.IsLevelEnabled(lrLevel) {
		return
	}
	args = []interface{}{fmt.Sprintf(format, args...)}
	l.log(lrLevel, args)
}

func (l *LogrusLogg) log(lrLevel lr.Level, args []interface{}) {
	if !l.ent.Logger.IsLevelEnabled(lrLevel) {
		return
	}

	// Run functions in args
	for i := range args {
		if fnc, ok := getFnc(args[i]); ok {
			args[i] = fnc()
		}
	}

	// Fields to pass to logrus entry
	fncResults := make(map[string]interface{})

	// Run the RunOnce functions and remove them from the struct
	for key := range l.runOnceFncs {
		fncResults[key] = l.callRunOnceFunc(key)
	}
	// Run the Run functions
	for key := range l.runFncs {
		fncResults[key] = l.callRunFunc(key)
	}

	ent := l.ent
	if len(fncResults) > 0 {
		ent = ent.WithFields(fncResults)
	}
	ent.Log(lrLevel, fmt.Sprint(args...))
}

func (l *LogrusLogg) spawnMe(entry *lr.Entry) *LogrusLogg {
	return &LogrusLogg{
		ent:         entry,
		runFncs:     l.runFncs,
		runOnceFncs: l.runOnceFncs,
	}
}

func (l *LogrusLogg) spawnEntry() *lr.Entry {
	dataCopy := make(lr.Fields, len(l.ent.Data))
	for k, v := range l.ent.Data {
		dataCopy[k] = v
	}
	return &lr.Entry{
		Logger:  l.ent.Logger,
		Data:    dataCopy,
		Time:    l.ent.Time,
		Context: l.ent.Context,
	}
}

func (l *LogrusLogg) setRunFnc(key string, value func() interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.runFncs[key] = value
}

func (l *LogrusLogg) setRunOnceFnc(key string, value func() interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.runOnceFncs[key] = value
}

func (l *LogrusLogg) callRunOnceFunc(key string) (result interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	result = l.runOnceFncs[key]()
	delete(l.runOnceFncs, key)

	return
}

func (l *LogrusLogg) callRunFunc(key string) (result interface{}) {
	return l.runFncs[key]()
}

func getMap(lgFields map[string]interface{}) (result map[string]interface{}) {
	result = make(map[string]interface{}, len(lgFields))
	for i, value := range lgFields {
		result[i] = value
	}
	return
}

func validateFnc(fnc interface{}) error {
	fncType := reflect.TypeOf(fnc)
	if fncType == nil {
		return errors.New("no function object type")
	}
	if fncType.Kind() != reflect.Func {
		return errors.New("function object not a function")
	}
	if fncType.NumIn() > 0 {
		return errors.New("function object cannot accept args")
	}
	if fncType.NumOut() != 1 {
		return errors.New("function object must have a single return value")
	}
	return nil
}

func getFnc(fnc interface{}) (result func() interface{}, ok bool) {
	if err := validateFnc(fnc); err != nil {
		ok = false
		return
	}

	ok = true
	result = func() (result interface{}) {
		fncVal := reflect.ValueOf(fnc)
		resultVals := fncVal.Call(nil)
		result = resultVals[0].Interface()
		return
	}

	return
}
