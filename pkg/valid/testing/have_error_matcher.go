package testing

import (
	"fmt"
	"strings"

	va "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/onsi/gomega/types"
	"github.com/pantheon-systems/search-secrets/pkg/errors"
	"github.com/pantheon-systems/search-secrets/pkg/manip"
)

func HaveError(param *manip.Param) types.GomegaMatcher {
	return &haveErrorMatcher{
		param: param,
	}
}

type haveErrorMatcher struct {
	param  *manip.Param
	result va.Error
}

func (m *haveErrorMatcher) Match(inputErr interface{}) (result bool, err error) {
	if inputErr == nil {
		result = false
		return
	}
	errs, ok := inputErr.(va.Errors)
	if !ok {
		err = errors.New("need a validation.Errors object")
		return
	}

	m.result = findErrorForParam(m.param, errs)

	result = m.result != nil
	return
}

func (m *haveErrorMatcher) FailureMessage(_ interface{}) (message string) {
	var sb strings.Builder
	sb.WriteString("Expected an error for field\n")
	fmt.Fprintf(&sb, "\t%s\n", m.param.PathName())
	sb.WriteString("but there is no error\n")
	return sb.String()
}

func (m *haveErrorMatcher) NegatedFailureMessage(_ interface{}) (message string) {
	var sb strings.Builder
	sb.WriteString("Expected field\n")
	fmt.Fprintf(&sb, "\t%s\n", m.param.PathName())
	sb.WriteString("to not have an error, but it had an error of type\n")
	fmt.Fprintf(&sb, "\t%s\n", m.result.Code())
	return sb.String()
}
