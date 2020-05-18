package testing

import (
	"fmt"
	"strings"

	va "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/onsi/gomega/types"
	"github.com/pantheon-systems/search-secrets/pkg/manip"
)

func HaveErrorType(param *manip.Param, expectErrObj va.Error) types.GomegaMatcher {
	return HaveErrorCode(param, expectErrObj.Code())
}

func HaveErrorCode(param *manip.Param, expectErrCode string) types.GomegaMatcher {
	return &haveErrorTypeMatcher{
		param:         param,
		expectErrCode: expectErrCode,
	}
}

type (
	haveErrorTypeMatcher struct {
		param         *manip.Param
		expectErrCode string
		*matchResult
	}
	matchResult struct {
		errObj va.Error
	}
)

func (m *haveErrorTypeMatcher) Match(inputErr interface{}) (result bool, err error) {
	if inputErr == nil {
		m.matchResult = &matchResult{}
		result = false
		return
	}
	errs, ok := inputErr.(va.Errors)
	if !ok {
		panic("need a validation.Errors object")
	}

	m.matchResult = &matchResult{findErrorForParam(m.param, errs)}
	result = m.matchResult.errObj != nil && m.matchResult.errObj.Code() == m.expectErrCode
	return
}

func (m *haveErrorTypeMatcher) FailureMessage(_ interface{}) (message string) {
	var sb strings.Builder
	sb.WriteString("Expected field\n")
	fmt.Fprintf(&sb, "\t%s\n", m.param.PathName())
	sb.WriteString("to have an error of type\n")
	fmt.Fprintf(&sb, "\t%s\n", m.expectErrCode)
	if m.matchResult.errObj == nil {
		sb.WriteString("but no error exists")
	} else {
		sb.WriteString("but instead we found an error with code\n")
		fmt.Fprintf(&sb, "\t%s\n", m.matchResult.errObj.Code())
	}
	return sb.String()
}

func (m *haveErrorTypeMatcher) NegatedFailureMessage(_ interface{}) (message string) {
	var sb strings.Builder
	sb.WriteString("Expected field\n")
	fmt.Fprintf(&sb, "\t%s\n", m.param.PathName())
	sb.WriteString("to not have an error of type\n")
	fmt.Fprintf(&sb, "\t%s\n", m.expectErrCode)
	sb.WriteString("but it did")
	return sb.String()
}
