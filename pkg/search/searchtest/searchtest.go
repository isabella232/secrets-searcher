package searchtest

import (
	"fmt"
	"regexp"

	va "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pantheon-systems/search-secrets/pkg/app/build"
	"github.com/pantheon-systems/search-secrets/pkg/app/config"
	"github.com/pantheon-systems/search-secrets/pkg/builtin"
	"github.com/pantheon-systems/search-secrets/pkg/dev"
	gitpkg "github.com/pantheon-systems/search-secrets/pkg/git"
	"github.com/pantheon-systems/search-secrets/pkg/logg"
	"github.com/pantheon-systems/search-secrets/pkg/manip"
	"github.com/pantheon-systems/search-secrets/pkg/search"
	"github.com/pantheon-systems/search-secrets/pkg/search/contract"
	"github.com/pantheon-systems/search-secrets/pkg/search/processor/setter"
	"github.com/sirupsen/logrus"
)

//
// LineProcJobMock

type LineProcJobMock struct {
	LineRanges        []*manip.LineRange
	ContextLineRanges []*manip.LineRange
	SecretValues      []string
	Ignored           []*manip.LineRange
	SecretExtras      []*contract.ResultExtra
	FindingExtras     []*contract.ResultExtra
	Path              string
	Logger            logg.Logg
}

func (l *LineProcJobMock) Diff() (result *gitpkg.Diff) {
	panic("implement me")
}

func (l *LineProcJobMock) FileChange() (fileChange *gitpkg.FileChange) {
	commit := &gitpkg.Commit{}

	return &gitpkg.FileChange{
		Commit: commit,
		Path:   l.Path,
	}
}

func (l *LineProcJobMock) Line() (line int) {
	panic("implement me")
}

func (l *LineProcJobMock) SubmitLineResult(lineResult *contract.LineResult) {
	l.LineRanges = append(l.LineRanges, lineResult.LineRange)
	l.ContextLineRanges = append(l.ContextLineRanges, lineResult.ContextLineRange)
	l.SecretValues = append(l.SecretValues, lineResult.SecretValue)
	l.SecretExtras = append(l.SecretExtras, lineResult.SecretExtras...)
	l.FindingExtras = append(l.FindingExtras, lineResult.FindingExtras...)
}

func (l *LineProcJobMock) SubmitLineRangeIgnore(lineRange *manip.LineRange) {
	l.Ignored = append(l.Ignored, lineRange)
}

func (l *LineProcJobMock) Log(logg.Logg) (result logg.Logg) {
	return l.Logger
}

func (l *LineProcJobMock) secretExtra(key string) (result *contract.ResultExtra) {
	for _, extra := range l.SecretExtras {
		if extra.Key == key {
			return extra
		}
	}
	return
}

func (l *LineProcJobMock) findingExtra(key string) (result *contract.ResultExtra) {
	for _, extra := range l.FindingExtras {
		if extra.Key == key {
			return extra
		}
	}
	return
}

//
// From testing setter rules
// FIXME Mess

func Log() *logg.LogrusLogg {
	return logg.NewLogrusLogg(logrus.New())
}

func CoreTarget(name builtin.TargetName) (result *search.Target) {
	targetConfig := builtin.TargetConfig(name)
	result = build.Target(targetConfig)
	return
}

func CoreTargets(names ...builtin.TargetName) (result *search.TargetSet) {
	var targets []*search.Target
	for _, name := range names {
		targetConfig := builtin.TargetConfig(name)
		target := build.Target(targetConfig)
		targets = append(targets, target)
	}
	result = search.NewTargetSet(targets)
	return
}

func CustomTargets(targetConfigs []*config.TargetConfig) (result *search.TargetSet) {
	var targets []*search.Target
	for _, targetConfig := range targetConfigs {

		PrepareConfig(targetConfig)

		target := build.Target(targetConfig)
		targets = append(targets, target)
	}
	result = search.NewTargetSet(targets)
	return
}

func CoreProcessor(name builtin.ProcessorName, targets *search.TargetSet) (result contract.ProcessorI) {
	procConfig := builtin.ProcessorConfig(name)

	var err error
	result, err = build.Proc(procConfig, targets, Log())
	if err != nil {
		panic("error building processor: " + procConfig.Name)
	}

	return
}

func CustomProcessor(procConfig *config.ProcessorConfig, targets *search.TargetSet) (result contract.ProcessorI) {
	PrepareConfig(procConfig)

	var err error
	result, err = build.Proc(procConfig, targets, Log())
	if err != nil {
		panic("error building processor: " + procConfig.Name)
	}

	return
}

func CoreRule(procName builtin.ProcessorName, targets *search.TargetSet) (result *setter.Rule) {
	procConfig := builtin.ProcessorConfig(procName)
	var procSetter *setter.Processor
	var err error
	if procSetter, err = build.ProcSetter(procName.String(), &procConfig.SetterProcessorConfig, targets, Log()); err != nil {
		panic("error building processor: " + procName.String())
	}
	return procSetter.Rules[0]
}

func CustomRule(procConfig *config.ProcessorConfig, targets *search.TargetSet) (result *setter.Rule) {
	PrepareConfig(procConfig)

	return build.ProcSetterRule("rule", &procConfig.SetterProcessorConfig, targets)
}

// Helpers

func ReMatchInfo(re *regexp.Regexp, input string, ok bool) (result string) {
	if ok {
		result += fmt.Sprintf("%s\n", "MATCH SUCCESSFUL")
	} else {
		result += fmt.Sprintf("%s\n", "MATCH FAILED")
	}

	result += fmt.Sprintf("%s\n", "Input:")
	result += fmt.Sprintf("  %s\n", input)
	result += fmt.Sprintf("%s\n", "Expression:")

	bullet := ""
	result += fmt.Sprintf("%-2v%-7s%s\n", bullet, "Rule:", re.String())
	result += fmt.Sprintf("  %-7s%s\n", "Debug:", dev.RegexpTestLink(re, input))

	return
}

func RuleMatchInfo(rule *setter.Rule, keyValue, secretValue *manip.LineRangeValue, matchingRe *regexp.Regexp, ok bool, input string) (result string) {
	if ok {
		result += fmt.Sprintf("%s\n", "MATCH SUCCESSFUL")
		result += fmt.Sprintf("%s\n", "Matched key:")
		result += fmt.Sprintf("  %s\n", keyValue.Value)
		result += fmt.Sprintf("%s\n", "Matched value:")
		result += fmt.Sprintf("  %s\n", secretValue.Value)
	} else {
		result += fmt.Sprintf("%s\n", "MATCH FAILED")
	}

	result += fmt.Sprintf("%s\n", "Input:")
	result += fmt.Sprintf("  %s\n", input)
	result += fmt.Sprintf("%s\n", "Expressions:")

	bullet := ""
	match := false
	for _, re := range rule.Res() {
		if ok && !match {
			bullet = "x"
			if matchingRe == re {
				bullet = "✓"
				match = true
			}
		}
		result += fmt.Sprintf("%-2v%-7s%s\n", bullet, "Rule:", re.String())
		result += fmt.Sprintf("  %-7s%s\n", "Debug:", dev.RegexpTestLink(re, input))
	}

	return
}

func PrepareConfig(ruleConfig interface{}) {

	// Apply defaults
	if validatable, ok := ruleConfig.(config.SetsDefaults); ok {
		validatable.SetDefaults()
	}

	// Validate
	if validatable, ok := ruleConfig.(va.Validatable); ok {
		if err := validatable.Validate(); err != nil {
			panic("validation error: " + err.Error())
		}
	}
}
