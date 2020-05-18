package config

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/pantheon-systems/search-secrets/pkg/errors"

	"github.com/pantheon-systems/search-secrets/pkg/app/vars"
	"github.com/pantheon-systems/search-secrets/pkg/manip"
)

func ParseEnvVars(appCfg *AppConfig, envVars []string) (err error) {
	if envVars == nil {
		return
	}

	// Parse env into map
	valuesByKey := envVarsToMap(envVars)

	// Parse struct fields into map of params by env var name
	paramsByEnvVarName := getParamMap(appCfg)

	// For each param with env var name defined, look for a value and set the new value passed
	for paramEnvVarName, param := range paramsByEnvVarName {

		// Look for the env var value passed for this field
		envVarValue, ok := valuesByKey[paramEnvVarName]
		if !ok {
			continue
		}

		// Set value in field
		if err = param.SetLeafFieldValueFromString(envVarValue); err != nil {
			return errors.WithMessage(err, "unable to set field")
		}
	}

	return
}

func getParamMap(cfg *AppConfig) (result map[string]*manip.Param) {
	result = map[string]*manip.Param{}

	// For each param in struct and in nested structs, grab a param
	structParams := manip.NewStructParams(cfg, vars.ConfigParamTag, func(structElemVal reflect.Value) (result bool) {
		for i := structElemVal.NumField() - 1; i >= 0; i-- {
			structField := structElemVal.Type().Field(i)
			if _, ok := structField.Tag.Lookup(vars.EnvParamTag); ok {
				return true
			}
		}
		return
	})

	for _, param := range structParams.Params {
		varName := getParamEnvVarName(param)
		if varName != "" {
			result[varName] = param
		}
	}

	return
}

func getParamEnvVarName(param *manip.Param) (result string) {

	// Get `env:"VALUE"` from struct field
	envTagValue := param.LeafStructField().Tag.Get(vars.EnvParamTag)
	if envTagValue == "" {
		return
	}

	// If it's a bool string we'll assume it means to include it or not
	if boolValue, err := strconv.ParseBool(envTagValue); err == nil {
		// Since it's set to "false", next field
		if !boolValue {
			return
		}

		// It's a true string, so we build the env var name from the field param
		result = envVarNameForParam(param)
		return
	}

	// It's not a bool string so it must be the env var name
	result = strings.ToUpper(envTagValue)
	return
}

func envVarNameForParam(param *manip.Param) (result string) {
	result = vars.EnvVarPrefix + param.PathName()
	result = strings.NewReplacer(".", "_", "-", "_").Replace(result)
	result = strings.ToUpper(result)
	return
}

func envVarsToMap(envVars []string) (result map[string]string) {
	result = make(map[string]string, len(envVars))
	for _, envVar := range envVars {
		kv := strings.SplitN(envVar, "=", 2)
		result[kv[0]] = kv[1]
	}
	return
}
