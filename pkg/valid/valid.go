package valid

import (
	"context"
	"fmt"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"time"

	va "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pantheon-systems/secrets-searcher/pkg/manip"

	"os"

	"gopkg.in/asaskevich/govalidator.v9"
)

//
// Errors

var (
	ErrPath               = va.NewError("valid_is_path", "must be a valid file path")
	ErrExtension          = va.NewError("valid_is_extension", "must be formatted like \".gz\" or \".tar.gz\"")
	ErrEmptyString        = va.NewError("valid_is_empty_string", "must be empty")
	ErrExistingFile       = va.NewError("valid_is_existing_file", "file does not exist")
	ErrExistingDir        = va.NewError("valid_is_existing_dir", "directory does not exist")
	ErrZero               = va.NewError("valid_is_zero", "must be empty")
	ErrBeforeTimeParam    = va.NewError("valid_before_time_param", "must not come before {{.param}}")
	ErrPathNotWithinParam = va.NewError("valid_not_within_dir", "must not be within {{.param}}")
	ErrRegexpPattern      = va.NewError("valid_regex", "must be a valid regular expression")
)

//
// Rules

// ParamRule

func NewParamRule(param *manip.Param, rule validatesWithParam) ParamRule {
	return ParamRule{param: param, rule: rule}
}

type (
	ParamRule struct {
		param *manip.Param
		rule  validatesWithParam
		ruleC validatesWithParamAndContext
	}
	validatesWithParam interface {
		Validate(param *manip.Param, value interface{}) error
	}
	validatesWithParamAndContext interface {
		ValidateWithContext(ctx context.Context, param *manip.Param, value interface{}) error
	}
)

func (r ParamRule) Validate(value interface{}) error {
	ctx := context.Background()
	if r.rule == nil {
		return r.ruleC.ValidateWithContext(ctx, r.param, value)
	}
	return r.rule.Validate(r.param, value)
}

func (r ParamRule) ValidateWithContext(ctx context.Context, value interface{}) error {
	if r.ruleC == nil {
		return r.rule.Validate(r.param, value)
	}
	return r.ruleC.ValidateWithContext(ctx, r.param, value)
}

func (r ParamRule) Param() *manip.Param {
	return r.param
}

func rules(rules []ParamRule) (result []va.Rule) {
	result = make([]va.Rule, len(rules))
	for i, rule := range rules {
		result[i] = rule
	}
	return
}

// WhenBothNotZero

func WhenBothNotZero(rules ...ParamRule) WhenBothNotZeroRule {
	return WhenBothNotZeroRule{rules: rules}
}

type WhenBothNotZeroRule struct{ rules []ParamRule }

func (r WhenBothNotZeroRule) Validate(value interface{}) error {
	return r.ValidateWithContext(context.Background(), value)
}

func (r WhenBothNotZeroRule) ValidateWithContext(ctx context.Context, value interface{}) error {
	paramValue := r.rules[0].Param().LeafFieldValue()
	if !IsZero(value) && !IsZero(paramValue) {
		return va.ValidateWithContext(ctx, value, rules(r.rules)...)
	}

	return nil
}

// BeforeTime

func BeforeTime(param *manip.Param) ParamRule {
	return NewParamRule(param, IsBeforeTimeParamRule{})
}

type IsBeforeTimeParamRule struct{}

func (r IsBeforeTimeParamRule) Validate(param *manip.Param, value interface{}) (err error) {
	if !value.(*time.Time).Before(*param.LeafFieldValue().(*time.Time)) {
		return setDotParam(ErrBeforeTimeParam, r)
	}
	return
}

// PathNotWithinParam

func PathNotWithinParam(param *manip.Param) IsPathNotWithinParamRule {
	return IsPathNotWithinParamRule{param}
}

type IsPathNotWithinParamRule struct{ param *manip.Param }

func (r IsPathNotWithinParamRule) Validate(value interface{}) (err error) {
	valueAbs, _ := filepath.Abs(value.(string))
	paramAbs, _ := filepath.Abs(*(r.param.LeafFieldValue().(*string)))
	if strings.HasPrefix(valueAbs, paramAbs) {
		return setDotParam(ErrPathNotWithinParam, r)
	}
	return
}

func (r IsPathNotWithinParamRule) Param() *manip.Param {
	return r.param
}

// Extension

var Extension = va.NewStringRuleWithError(func(s string) bool {
	return s == "" || (strings.HasPrefix(s, ".") && !strings.HasSuffix(s, "."))
}, ErrExtension)

// EmptyString

var EmptyString = va.NewStringRuleWithError(func(s string) bool { return s == "" }, ErrEmptyString)

// Path

var Path = va.NewStringRuleWithError(func(value string) (result bool) {
	result, _ = govalidator.IsFilePath(value)
	return
}, ErrPath)

// ExistingFile

var ExistingFile = va.NewStringRuleWithError(func(value string) bool {
	fileInfo, err := os.Stat(value)
	return !os.IsNotExist(err) && fileInfo.Mode().IsRegular()
}, ErrExistingFile)

// ExistingDir

var ExistingDir = va.NewStringRuleWithError(func(value string) bool {
	fileInfo, err := os.Stat(value)
	return !os.IsNotExist(err) && fileInfo.Mode().IsDir()
}, ErrExistingDir)

// Zero

var Zero = newRuleWithError(IsZero, ErrZero)

// RegexpPattern

var RegexpPattern = va.By(func(value interface{}) (err error) {
	input := value.(string)
	if input == "" {
		return
	}
	if _, err := regexp.Compile(input); err != nil {
		vaErr := ErrRegexpPattern
		msg := fmt.Sprintf(`%s (rule: %s; error: %s)`, vaErr.Message(), input, err.Error())
		return vaErr.SetMessage(msg)
	}
	return
})

var RegexpTmpl = va.By(func(value interface{}) (err error) {
	input := value.(string)
	if input == "" {
		return
	}
	replacer := regexp.MustCompile(`{{[^}]+}}`)
	pattern := replacer.ReplaceAllString(input, `TMPL_REF_PLACEHOLDER`)

	return RegexpPattern.Validate(pattern)
})

//
// Validators

func IsZero(value interface{}) bool {
	return value == reflect.Zero(reflect.TypeOf(value)).Interface()
}

//
// Helpers

func setDotParam(err va.Error, obj interface{}) (result va.Error) {
	if err == nil {
		return
	}
	return err.SetParams(map[string]interface{}{".": obj})
}

func newRuleWithError(f func(value interface{}) bool, err va.Error) va.Rule {
	return va.By(func(value interface{}) error {
		if !f(value) {
			return err
		}
		return nil
	})
}
