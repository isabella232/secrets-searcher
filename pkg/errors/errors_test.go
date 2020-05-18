package errors_test

import (
	"fmt"
	"testing"

	"github.com/pantheon-systems/search-secrets/pkg/dev"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/pantheon-systems/search-secrets/pkg/errors"
)

const (
	errMsg     = "errMsg"
	panicMsg   = "panicMsg"
	prependMsg = "prependMsg"
)

func TestErrors(t *testing.T) {
	dev.RunningTests = true
	RegisterFailHandler(Fail)
	RunSpecs(t, "Error Library Test Suite")
}

var _ = Describe("Errors", func() {

	DescribeTable("Value-add functions",
		func(message string, args []interface{}, expected string) {

			// Fire
			response := errors.Errorv(message, args[0], args[1:]...)

			Expect(response).To(Not(BeNil()))
			Expect(response.Error()).To(Equal(expected))
		},
		Entry("string value",
			errMsg, []interface{}{"bar"}, errMsg+" (bar)"),
		Entry("nil value",
			errMsg, []interface{}{nil}, errMsg+" ([nil])"),
		Entry("empty string",
			errMsg, []interface{}{""}, errMsg+" ([empty string])"),
		Entry("multiple string values",
			errMsg, []interface{}{"bar", "bar2"}, errMsg+" (bar; bar2)"),
		Entry("logrus style Fields",
			errMsg, []interface{}{map[string]interface{}{"foo": "bar", "foo2": "bar2"}}, errMsg+" (foo=bar foo2=bar2)"),
		Entry("struct with data",
			errMsg, []interface{}{testStruct{foo: "bar", baz: "uuhhh"}}, errMsg+" ({foo:bar baz:uuhhh})"),
	)

	Describe("Categorizing panic handlers", func() {

		Context("If a method panics", func() {

			It("we should be able to catch the panic", func() {
				response := func() (response interface{}) {

					// Fire
					defer errors.CatchPanicValueDo(func(recovered interface{}) {
						response = recovered
					})
					panic(panicMsg)

					return
				}()

				Expect(response).To(Equal(panicMsg))
			})

			It("we should be able to catch the panic as an error", func() {
				response := func() (response error) {

					// Fire
					defer errors.CatchPanicDo(func(err error) {
						response = err
					})
					panic(panicMsg)

					return
				}()

				Expect(response).To(Not(BeNil()))
				Expect(response.Error()).To(Equal("panic caught: " + panicMsg))
			})

			It("we should be able to catch the panic and set an error var with message", func() {
				response := func() (response error) {

					// Fire
					defer errors.CatchPanicSetErr(&response, prependMsg)
					panic(panicMsg)

					return
				}()

				Expect(response).To(Not(BeNil()))
				Expect(response.Error()).To(Equal(prependMsg + ": panic caught: " + panicMsg))
			})
		})
	})

	Describe("Categorizing stacktrace building", func() {

		Context("If an error is provided that directly contains a stacktrace", func() {

			It("it should be returned with StackTrace()", func() {
				err := errors.New("") // Stacktrace recorded

				// Fire
				stacktrace := errors.StackTrace(err)

				Expect(stacktrace).To(Not(BeEmpty()))
			})
		})

		Context("If an error is provided whose root error contains a stacktrace", func() {

			It("it should be returned with StackTrace()", func() {
				err0 := errors.New("")               // Stacktrace recorded
				err1 := errors.WithMessage(err0, "") // No stacktrace
				err := errors.WithMessage(err1, "")  // No stacktrace

				// Fire
				stacktrace := errors.StackTrace(err)

				Expect(stacktrace).To(Not(BeEmpty()))
			})
		})

		Context("If an error is provided whose ancestor error contains a stacktrace", func() {

			It("it should be returned with StackTrace()", func() {
				err0 := &simpleError{msg: ""}       // No stacktrace
				err1 := errors.Wrap(err0, "")       // Stacktrace recorded
				err := errors.WithMessage(err1, "") // No stacktrace

				// Fire
				stacktrace := errors.StackTrace(err)

				Expect(stacktrace).To(Not(BeEmpty()))
			})
		})

		Context("If an error is provided whose ancestor error contains a caught panic", func() {

			It("a stacktrace should be returned with StackTrace()", func() {
				var err error
				func() {
					defer errors.CatchPanicSetErr(&err, "") // Stacktrace recorded
					panic(panicMsg)
				}()

				// Fire
				stacktrace := errors.StackTrace(err)

				Expect(stacktrace).To(Not(BeEmpty()))
			})
		})

		Context("If an error has nested stacktraces", func() {

			It("the innermost stacktrace should be returned by StackTrace()", func() {
				err0 := errors.New("")               // Stacktrace recorded
				err1 := errors.WithMessage(err0, "") // No stacktrace
				err2 := errors.WithMessage(err1, "") // No stacktrace
				err := errors.Wrap(err2, "")         // Stacktrace recorded again

				// Fire
				stacktrace := errors.StackTrace(err)

				Expect(stacktrace).To(Not(BeEmpty()))
				firstFunctionName := fmt.Sprintf("%+v", stacktrace[0])
				Expect(firstFunctionName).To(ContainSubstring("errors.New"))
			})
		})

		Context("If an error is provided that contains no stacktrace", func() {

			It("nothing should be returned with StackTrace()", func() {
				err := &simpleError{msg: ""} // No stacktrace

				// Fire
				stacktrace := errors.StackTrace(err)

				Expect(stacktrace).To(BeEmpty())
			})
		})
	})
})

type simpleError struct {
	msg string
}

func (f *simpleError) Error() string { return f.msg }

type testStruct struct {
	foo string
	baz string
}

func (ts testStruct) HopefullyHiddenFunction() string { return "hi" }
