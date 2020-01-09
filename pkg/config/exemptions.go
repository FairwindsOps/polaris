package config

import (
	"strings"
)

// IsActionable determines whether a check is actionable given the current configuration
func (conf Configuration) IsActionable(ruleID, controllerName string) bool {
	if severity, ok := conf.Checks[ruleID]; !ok || !severity.IsActionable() {
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
