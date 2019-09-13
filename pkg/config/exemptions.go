package config

import (
	"reflect"
)

func IsActionable(conf *Configuration, subConf interface{}, ruleName, controllerName string) bool {
	ruleID := GetIDFromField(subConf, ruleName)
	subConfRef := reflect.ValueOf(subConf)
	severity, ok := reflect.Indirect(subConfRef).FieldByName(ruleName).Interface().(Severity)
	if ok && !severity.IsActionable() {
		return false
	}
	for _, example := range conf.Exemptions {
		for _, rule := range example.Rules {
			if rule != ruleID {
				continue
			}
			for _, controller := range example.ControllerNames {
				if controller == controllerName {
					return false
				}
			}
		}
	}
	return true
}
