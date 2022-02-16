package build

import (
	"regexp"

	"github.com/pantheon-systems/secrets-searcher/pkg/builtin"

	"github.com/pantheon-systems/secrets-searcher/pkg/app/config"
	"github.com/pantheon-systems/secrets-searcher/pkg/errors"
	"github.com/pantheon-systems/secrets-searcher/pkg/logg"
	"github.com/pantheon-systems/secrets-searcher/pkg/manip"
	"github.com/pantheon-systems/secrets-searcher/pkg/search"
	"github.com/pantheon-systems/secrets-searcher/pkg/search/contract"
	"github.com/pantheon-systems/secrets-searcher/pkg/search/processor/entropy"
	"github.com/pantheon-systems/secrets-searcher/pkg/search/processor/pem"
	"github.com/pantheon-systems/secrets-searcher/pkg/search/processor/regex"
	"github.com/pantheon-systems/secrets-searcher/pkg/search/processor/setter"
)

func Procs(searchCfg *config.SearchConfig, targets *search.TargetSet, processorsLog logg.Logg) (result []contract.ProcessorI, err error) {
	result = []contract.ProcessorI{}

	// Custom procs names for filter
	customProcNames := make([]string, len(searchCfg.ProcessorConfigs))
	for i, procConfig := range searchCfg.ProcessorConfigs {
		var proc contract.ProcessorI
		if proc, err = Proc(procConfig, targets, processorsLog); err != nil {
			err = errors.New("unable to create custom processesor: " + procConfig.Name)
			return
		}

		result = append(result, proc)
		customProcNames[i] = procConfig.Name
	}

	// Core procs are run after custom procs
	coreProcConfigs := builtin.ProcessorConfigs()
	processorFilter := manip.StringFilter(searchCfg.IncludeProcessors, searchCfg.ExcludeProcessors)
	for _, procConfig := range coreProcConfigs {
		if processorFilter.Includes(procConfig.Name) && !manip.SliceContains(customProcNames, procConfig.Name) {
			var proc contract.ProcessorI
			if proc, err = Proc(procConfig, targets, processorsLog); err != nil {
				err = errors.New("unable to create builtin processesor: " + procConfig.Name)
				return
			}
			result = append(result, proc)
		}
	}

	if len(result) == 0 {
		err = errors.New("no processors are configured")
		return
	}

	return
}

func Proc(procCfg *config.ProcessorConfig, targets *search.TargetSet, processorsLog logg.Logg) (result contract.ProcessorI, err error) {
	processorLog := processorsLog.AddPrefixPath(procCfg.GetName())

	switch procCfg.Processor {
	case search.Regex.String():
		result = ProcRegexWrapped(procCfg.Name, &procCfg.RegexProcessorConfig, processorLog)
	case search.PEM.String():
		result = ProcPEM(procCfg.Name, &procCfg.PEMProcessorConfig, processorLog)
	case search.Setter.String():
		result, err = ProcSetterWrapped(procCfg.Name, &procCfg.SetterProcessorConfig, targets, processorLog)
	case search.Entropy.String():
		result = ProcEntropy(procCfg.Name, &procCfg.EntropyProcessorConfig, processorLog)
	default:
		err = errors.Errorv("unknown processor", procCfg.Processor)
		return
	}
	return
}

//
// Regex processor

func ProcRegexWrapped(name string, regexProcCfg *config.RegexProcessorConfig, processorLog logg.Logg) (result contract.ProcessorI) {
	lineProc := ProcRegex(name, regexProcCfg, processorLog)
	result = search.NewLineProcessorWrapper(lineProc, processorLog)
	return
}

func ProcRegex(name string, regexProcCfg *config.RegexProcessorConfig, processorLog logg.Logg) (result *regex.Processor) {
	codeWhitelist := CodeWhitelist(regexProcCfg.WhitelistCodeMatch, processorLog)
	re := regexp.MustCompile(regexProcCfg.RegexString)
	result = regex.NewProcessor(name, re, codeWhitelist, processorLog)

	return
}

//
// PEM processor

func ProcPEM(name string, pemProcCfg *config.PEMProcessorConfig, processorLog logg.Logg) (result contract.ProcessorI) {
	result = pem.NewProcessor(name, pemProcCfg.PEMType, processorLog)

	return
}

//
// Setter processor

func ProcSetterWrapped(name string, setterProcCfg *config.SetterProcessorConfig, targets *search.TargetSet, processorLog logg.Logg) (result contract.ProcessorI, err error) {
	var proc contract.LineProcessorI
	proc, err = ProcSetter(name, setterProcCfg, targets, processorLog)
	result = LineProcessor(proc, processorLog)

	return
}

func ProcSetter(name string, setterProcCfg *config.SetterProcessorConfig, targets *search.TargetSet, processorLog logg.Logg) (result *setter.Processor, err error) {
	rule := ProcSetterRule(name+"Rule", setterProcCfg, targets)
	rules := []*setter.Rule{rule} // FIXME There's only ever one
	result = setter.NewProcessor(name, rules, targets, processorLog)
	return
}

func ProcSetterRule(name string, setterProcCfg *config.SetterProcessorConfig, targets *search.TargetSet) (result *setter.Rule) {
	fileExtFilter := manip.NewStringRegexpFilter(setterProcCfg.FileExts, nil)

	return setter.NewRule(
		name,
		targets,
		fileExtFilter,
		setterProcCfg.MainTmpl,
		setterProcCfg.KeyTmpls,
		setterProcCfg.KeyChars,
		setterProcCfg.Operator,
		setterProcCfg.ValTmpls,
		setterProcCfg.NoWhitespace,
		setterProcCfg.NotValChars,
	)
}

//
// Entropy processor

func ProcEntropy(name string, entropyProcCfg *config.EntropyProcessorConfig, processorLog logg.Logg) (result contract.ProcessorI) {
	codeWhitelist := CodeWhitelist(entropyProcCfg.WhitelistCodeMatch, processorLog)
	return entropy.NewProcessor(
		name,
		entropyProcCfg.Charset,
		entropyProcCfg.WordLengthThreshold,
		entropyProcCfg.Threshold,
		codeWhitelist,
		entropyProcCfg.SkipPEMs,
		processorLog,
	)
}

//
// Helpers

func CodeWhitelist(whitelistCodeMatch []string, baseLog logg.Logg) *search.CodeWhitelist {
	codeWhitelistLog := baseLog.AddPrefixPath("code-whitelist")
	return search.NewCodeWhitelist(whitelistCodeMatch, codeWhitelistLog)
}

func LineProcessor(proc contract.LineProcessorI, baseLog logg.Logg) *search.LineProcessorWrapper {
	lineProcessorLog := baseLog.AddPrefixPath("line-processor")
	return search.NewLineProcessorWrapper(proc, lineProcessorLog)
}
