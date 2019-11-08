package config

import (
	"reflect"
	"strings"
)

// IsActionable determines whether a check is actionable given the current configuration
func (conf *Configuration) IsActionable(subConf interface{}, ruleName, controllerName string) bool {
	ruleID := GetIDFromField(subConf, ruleName)
	subConfRef := reflect.ValueOf(subConf)
	severity, ok := reflect.Indirect(subConfRef).FieldByName(ruleName).Interface().(Severity)
	if ok && !severity.IsActionable() {
		return false
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
