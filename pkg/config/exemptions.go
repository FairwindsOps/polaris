package config

import (
	"reflect"
	"strings"
)

// IsActionable determines whether a check is actionable given the current configuration
func (conf *Configuration) IsActionable(subConf interface{}, ruleName, controllerName string) bool {
	ruleID := GetIDFromField(subConf, ruleName)
	subConfRef := reflect.ValueOf(subConf)
	fieldVal := reflect.Indirect(subConfRef).FieldByName(ruleName).Interface()
	if severity, ok := fieldVal.(Severity); ok && !severity.IsActionable() {
		return false
	}
	if ranges, ok := fieldVal.(ResourceRanges); ok {
		if ranges.Warning.Above == nil && ranges.Warning.Below == nil &&
			ranges.Error.Above == nil && ranges.Error.Below == nil {
			return false
		}
	}
	if conf.DisallowExemptions {
		return true
	}
	for _, example := range conf.Exemptions {
		for _, rule := range example.Rules {
			if rule != ruleID {
				continue
			}
			for _, controller := range example.ControllerNames {
				if strings.HasPrefix(controllerName, controller) {
					return false
				}
			}
		}
	}
	return true
}
