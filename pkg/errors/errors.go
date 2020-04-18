package errors

import (
    "bytes"
    "fmt"
    "os"
    "regexp"
    "strings"

    "github.com/sirupsen/logrus"

    errorsOrig "github.com/pkg/errors"
)

var (
    logrusTextFormatter = logrus.TextFormatter{DisableColors: true, DisableTimestamp: true}
    logrusRegex         = regexp.MustCompile("\\s?level=[^ ]+\\s?")
)

// Add contextual information to the end of the error string
func Errorv(message string, args ...interface{}) error {
    return errorsOrig.New(messageWithValue(message, value(args...)))
}

// Like Errorv(), but for WithMessage()
func WithMessagev(err error, message string, args ...interface{}) error {
    return errorsOrig.WithMessage(err, messageWithValue(message, args...))
}

// Like Errorv(), but for Wrap()
func Wrapv(err error, message string, args ...interface{}) error {
    return errorsOrig.Wrap(err, messageWithValue(message, args...))
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

// Log error and return logrus.Entry object
func ErrorLog(log *logrus.Logger, err error) *logrus.Entry {
    stacktrace := StackTraceString(err)
    return log.WithError(err).WithField("stacktrace", stacktrace)
}

// Log error and return logrus.Entry object
func ErrorLogForEntry(log *logrus.Entry, err error) *logrus.Entry {
    stacktrace := StackTraceString(err)
    return log.WithError(err).WithField("stacktrace", stacktrace)
}

// Log error and return logrus.Entry object
func PanicLogError(log *logrus.Logger, recovered interface{}) *logrus.Entry {
    err := PanicError(recovered)
    stacktrace := StackTraceString(err)
    return log.WithError(err).WithField("stacktrace", stacktrace)
}

// Log error and return logrus.Entry object
func PanicLogEntryError(log *logrus.Entry, recovered interface{}) *logrus.Entry {
    err := PanicError(recovered)
    stacktrace := StackTraceString(err)
    return log.WithError(err).WithField("stacktrace", stacktrace)
}

// Log error and return logrus.Entry object
func Fatal(log *logrus.Logger, err error) {
    stacktrace := StackTraceString(err)
    log.WithField("stacktrace", stacktrace).Error(err)
    os.Exit(1)
}

func PanicWithMessage(recovered interface{}, message string) error {
    return WithMessagef(PanicError(recovered), message)
}

func PanicError(recovered interface{}) error {
    return Errorf(fmt.Sprintf("panic caught: %v", recovered))
}

func LogPanicAndContinue(log *logrus.Logger) {
    if recovered := recover(); recovered != nil {
        PanicLogError(log, recovered)
    }
}

func LogEntryPanicAndContinue(log *logrus.Entry) {
    if recovered := recover(); recovered != nil {
        PanicLogEntryError(log, recovered)
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

func messageWithValue(message string, args ...interface{}) string {
    return fmt.Sprintf("%s (%v)", message, value(args...))
}

func value(args ...interface{}) string {
    if len(args) == 1 {
        var arg = args[0]
        if arg == "" {
            return "[empty string]"
        }
        if arg == nil {
            return "[nil]"
        }

        switch v := arg.(type) {
        case logrus.Fields:
            return fieldsString(v)
        case map[string]interface{}:
            return fieldsString(v)
        default:
            return fmt.Sprintf("%s", v)
        }
    }

    values := make([]string, len(args))
    for i, arg := range args {
        values[i] = value(arg)
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
