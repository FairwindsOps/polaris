package config

import (
	"strings"
)

// IsActionable determines whether a check is actionable given the current configuration
func (conf Configuration) IsActionable(ruleID, namespace, controllerName string) bool {
	if severity, ok := conf.Checks[ruleID]; !ok || !severity.IsActionable() {
		return false
	}
	if conf.DisallowExemptions {
		return true
	}

	for _, example := range conf.Exemptions {
		if example.Namespace != "" && example.Namespace != namespace {
			continue
		}

		checkIfActionable := false
		for _, rule := range example.Rules {
			if rule != ruleID {
				continue
			}
			checkIfActionable = true
			break
		}

		if len(example.Rules) == 0 || checkIfActionable {
			for _, controller := range example.ControllerNames {
				if strings.HasPrefix(controllerName, controller) {
					return false
				}
			}
		}
	}
	return true
}
