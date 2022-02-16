package builtin

import (
	va "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pantheon-systems/secrets-searcher/pkg/app/config"
)

//
// Processor

func ProcessorConfig(procName ProcessorName) (result *config.ProcessorConfig) {
	configs := processorDefinitions()
	procToFind := procName.String()

	for _, cfg := range configs {
		name := cfg.Name
		if name != procToFind {
			continue
		}

		result = &*cfg // Copy
		prepareConfig(result)
		return
	}

	panic("undefined builtin processor: " + procToFind)

	return
}

func ProcessorConfigs() (result []*config.ProcessorConfig) {
	coreConfigs := processorDefinitions()
	result = make([]*config.ProcessorConfig, len(coreConfigs))
	for i, cfg := range coreConfigs {
		copied := &*cfg // Copy
		prepareConfig(copied)
		result[i] = copied
	}

	return
}

//
// Target

func TargetConfig(targetName TargetName) (result *config.TargetConfig) {
	configs := targetDefinitions()

	for name, cfg := range configs {
		if name != targetName {
			continue
		}

		result = &*cfg // Copy
		result.Name = targetName.String()
		prepareConfig(result)
		return
	}

	panic("undefined builtin target: " + targetName.String())

	return
}

func TargetConfigs() (result []*config.TargetConfig) {
	targetNames := TargetNames() // in order
	result = make([]*config.TargetConfig, len(targetNames))
	for i, name := range targetNames {
		result[i] = TargetConfig(name)
	}

	return
}

//
// Internal

func prepareConfig(cfg interface{}) {

	// Apply defaults
	if validatable, ok := cfg.(config.SetsDefaults); ok {
		validatable.SetDefaults()
	}

	// Validate
	if validatable, ok := cfg.(va.Validatable); ok {
		if err := validatable.Validate(); err != nil {
			panic("validation error: " + err.Error())
		}
	}
}
