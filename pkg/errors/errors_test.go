package errors_test

import (
    "fmt"
    "strings"
    "testing"

    "github.com/sirupsen/logrus"
    "github.com/stretchr/testify/assert"

    "github.com/pantheon-systems/search-secrets/pkg/errors"
)

type simpleError struct {
    msg string
}

func (f *simpleError) Error() string { return f.msg }

func TestErrorv(t *testing.T) {
    var halfTests = []struct {
        message  string
        args     []interface{}
        expected string
    }{
        {"error string", []interface{}{"bar"}, "error string (bar)"},
        {"error string", []interface{}{nil}, "error string ([nil])"},
        {"error string", []interface{}{""}, "error string ([empty string])"},
        {"error string", []interface{}{"bar", "bar2"}, "error string (bar; bar2)"},
        {"error string", []interface{}{logrus.Fields{"foo": "bar", "foo2": "bar2"}}, "error string (foo=bar foo2=bar2)"},
    }

    for _, tt := range halfTests {

        // Fire
        err := errors.Errorv(tt.message, tt.args...)

        assert.Error(t, err)
        assert.Equal(t, tt.expected, err.Error())
    }
}

func TestStackTrace_happy(t *testing.T) {
    err0 := errors.New("Message 0")
    err1 := errors.WithMessage(err0, "Message 1")
    err2 := errors.WithMessage(err1, "Message 2")

    stacktrace := errors.StackTrace(err2)

    assert.Greater(t, len(stacktrace), 0)
}

func TestStackTrace_stackTraceNotOnCause(t *testing.T) {
    err0 := &simpleError{msg: "Resource not found (resource=binding-file-ops-malwares)."}
    err1 := errors.Wrap(err0, "unable to start pubsub subscription (subscriptionID=binding-file-ops-malwares)")
    err2 := errors.WithMessage(err1, "unable to run subscriber")

    stacktrace := errors.StackTrace(err2)

    assert.Greater(t, len(stacktrace), 0)
    assert.NotEmpty(t, stacktrace)
}

func TestStackTrace_stackTraceDoubled_returnInnermost(t *testing.T) {
    err0 := errors.New("Message 0") // Stacktrace recorded
    err1 := errors.WithMessage(err0, "Message 1")
    err2 := errors.WithMessage(err1, "Message 2")
    err3 := errors.Wrap(err2, "Message 3") // Stacktrace recorded again

    stacktrace := errors.StackTrace(err3)

    assert.Greater(t, len(stacktrace), 0)
    functionName := fmt.Sprintf("%+v", stacktrace[0])
    assert.True(t, strings.Contains(functionName, "errors.New"))
}
