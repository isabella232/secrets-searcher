package errors

import (
    "bytes"
    "fmt"
    "regexp"
    "strings"

    "github.com/sirupsen/logrus"

    errorsOrig "github.com/pkg/errors"
)

var (
    logrusTextFormatter = logrus.TextFormatter{DisableColors: true, DisableTimestamp: true}
    logrusRegex         = regexp.MustCompile("\\s?level=[^ ]+\\s?")
)

type (
    DoWithPanicFunc         func(recovered interface{})
    DoWithErrFunc           func(err error)
    DoWithLogFunc           func(err error, log logrus.FieldLogger)
    DoWithLogAndMessageFunc func(err error, log logrus.FieldLogger, message string)
)

// Add contextual information to the end of the error string
func Errorv(message string, arg0 interface{}, args ...interface{}) error {
    return errorsOrig.New(messageWithValue(message, arg0, args...))
}

// Like Errorv(), but for WithMessage()
func WithMessagev(err error, message string, arg0 interface{}, args ...interface{}) error {
    return errorsOrig.WithMessage(err, messageWithValue(message, arg0, args...))
}

// Like Errorv(), but for Wrap()
func Wrapv(err error, message string, arg0 interface{}, args ...interface{}) error {
    return errorsOrig.Wrap(err, messageWithValue(message, arg0, args...))
}

// Wrapped
func New(message string) error {
    return errorsOrig.New(message)
}

// Wrapped
func Errorf(format string, args ...interface{}) error {
    return errorsOrig.Errorf(format, args...)
}

// Wrapped
func WithStack(err error) error {
    return errorsOrig.WithStack(err)
}

// Wrapped
func Wrap(err error, message string) error {
    return errorsOrig.Wrap(err, message)
}

// Wrapped
func Wrapf(err error, message string, args ...interface{}) error {
    return errorsOrig.Wrapf(err, message, args...)
}

// Wrapped
func WithMessage(err error, message string) error {
    return errorsOrig.WithMessage(err, message)
}

// Wrapped
func WithMessagef(err error, format string, args ...interface{}) error {
    return errorsOrig.WithMessagef(err, format, args...)
}

// Wrapped
func Cause(err error) error {
    return errorsOrig.Cause(err)
}

// Log error and return Logger object
func ErrorLogger(log logrus.FieldLogger, err error) logrus.FieldLogger {
    log = WithStacktrace(log, err)
    return log.WithError(err)
}

// Log error and return logger object
func LogErrorThenDie(log logrus.FieldLogger, err error) {
    ErrorLogger(log, err).Fatal("fatal error")
}

// Panic handling

// Catch panic and do something with it
/* Usage:
func main() {
    defer errors.CatchPanicValueDo(func(recovered interface{}) {
        fmt.Print(recovered) // "this was inevitable"
    })
    panic("this was inevitable")
}
*/
func CatchPanicValueDo(panicHandle DoWithPanicFunc) {
    if recovered := recover(); recovered != nil {
        panicHandle(recovered)
    }
}

// Catch panic, convert it to an error object, and do something with it
/* Usage:
func main() {
    defer errors.CatchPanicDo(func(err error) {
        fmt.Print(err.Error()) // "panic caught: this was inevitable"
    })
    panic("this was inevitable")
}
*/
func CatchPanicDo(doFunc DoWithErrFunc) {
    if recovered := recover(); recovered != nil {
        err := panicValueToErr(recovered)
        doFunc(err)
    }
}

// Catch panic and log it
/* Usage:
func main() {
    defer errors.CatchPanicAndLogIt(log)
    panic("this was inevitable") // logged: "panic caught: this was inevitable"
}
*/
func CatchPanicAndLogIt(log logrus.FieldLogger) {
    if recovered := recover(); recovered != nil {
        err := panicValueToErr(recovered)
        log = ErrorLogger(log, err)
        log.Error(err.Error())
    }
}

// Catch panic, convert it to an error object, and set an error pointer with it with a message
/* Usage:
func do() (err error) {
    defer errors.CatchPanicSetErr(&err, "something happened")
    panic("this was inevitable")
}
func main() {
    if err := do(); err != nil {
        fmt.Print(err.Error()) // "something happened: panic caught: this was inevitable"
    }
}
*/
func CatchPanicSetErr(err *error, message string) {
    if recovered := recover(); recovered != nil {
        *err = panicValueToErr(recovered)
        if message != "" {
            *err = WithMessage(*err, message)
        }
    }
}

// Get stacktrace from error object
func StackTraceString(err error) string {
    buf := bytes.Buffer{}
    stackTrace := StackTrace(err)

    if stackTrace != nil {
        for _, f := range stackTrace {
            buf.WriteString(fmt.Sprintf("%+v \n", f))
        }
    }

    return buf.String()
}

func StackTrace(err error) errorsOrig.StackTrace {
    var st errorsOrig.StackTrace
    for err != nil {

        // Stacktrace on this err?
        ster, ok := err.(interface{ StackTrace() errorsOrig.StackTrace })
        if ok {
            st = ster.StackTrace()
        }

        // Climb tree
        err = getInnerError(err)
    }
    return st
}

func WithStacktrace(log logrus.FieldLogger, err error) logrus.FieldLogger {
    return log.WithField("stacktrace", StackTraceString(err))
}

func messageWithValue(message string, arg0 interface{}, args ...interface{}) string {
    return fmt.Sprintf("%s (%v)", message, value(arg0, args...))
}

func value(arg0 interface{}, args ...interface{}) string {
    if len(args) == 0 {
        if arg0 == "" {
            return "[empty string]"
        }
        if arg0 == nil {
            return "[nil]"
        }

        switch v := arg0.(type) {
        case logrus.Fields:
            return fieldsString(v)
        case map[string]interface{}:
            return fieldsString(v)
        }

        return fmt.Sprintf("%+v", arg0)
    }

    values := make([]string, len(args)+1)
    values[0] = value(arg0)
    for i, arg := range args {
        values[i+1] = value(arg)
    }

    return strings.Join(values, "; ")
}

// Yeah we can just use logrus for this
func fieldsString(fields map[string]interface{}) string {
    logrusFields := logrus.Fields{}
    for key, value := range fields {
        logrusFields[key] = value
    }

    formattedFields, err := logrusTextFormatter.Format(logrus.WithFields(logrusFields))
    if err != nil {
        return "[unknown var]"
    }
    formattedFields = logrusRegex.ReplaceAll(formattedFields, []byte(""))

    return strings.TrimSpace(string(formattedFields))
}

func getInnerError(err error) error {
    cer, ok := err.(interface {
        Cause() error
    })
    if !ok {
        return nil
    }
    return cer.Cause()
}

func panicValueToErr(recovered interface{}) (result error) {
    return Errorf(fmt.Sprintf("panic caught: %v", recovered))
}
