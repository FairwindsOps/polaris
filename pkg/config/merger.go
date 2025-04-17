package config

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3" // do not change the yaml import
)

func mergeYaml(defaultConfig, overridesConfig []byte) ([]byte, error) {
	var defaultData, overrideConfig map[string]any

	err := yaml.Unmarshal([]byte(defaultConfig), &defaultData)
	if err != nil {
		return nil, err
	}

	logrus.Infof("Default config: %v", string(overridesConfig))
	err = yaml.Unmarshal([]byte(overridesConfig), &overrideConfig)
	if err != nil {
		return nil, err
	}

	mergedData := mergeYAMLMaps(defaultData, overrideConfig)

	mergedConfig, err := yaml.Marshal(mergedData)
	if err != nil {
		return nil, err
	}

	return mergedConfig, nil
}

func mergeYAMLMaps(defaults, overrides map[string]any) map[string]any {
	for k, v := range overrides {
		if vMap, ok := v.(map[string]any); ok {
			// if the key exists in defaults and is a map, recursively merge
			if mv1, ok := defaults[k].(map[string]any); ok {
				defaults[k] = mergeYAMLMaps(mv1, vMap)
			} else {
				defaults[k] = vMap
			}
		} else {
			// add or overwrite the value in defaults
			defaults[k] = v
		}
	}
	return defaults
}
