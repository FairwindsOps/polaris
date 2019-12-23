package config

import (
	"reflect"
	"strings"
)

// IsActionable determines whether a check is actionable given the current configuration
func (conf Configuration) IsActionable(subConf interface{}, ruleName, controllerName string) bool {
	if subConfStr, ok := subConf.(string); ok {
		subConf = conf.GetCategoryConfig(subConfStr)
	}
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

// GetCategoryConfig returns the configuration for a particular category name
func (conf Configuration) GetCategoryConfig(category string) interface{} {
	if category == "Networking" {
		return conf.Networking
	} else if category == "Security" {
		return conf.Security
	} else if category == "Health Checks" {
		return conf.HealthChecks
	} else if category == "Resources" {
		return conf.Resources
	} else if category == "Images" {
		return conf.Images
	}
	return nil
}

// GetSeverity returns the severity configured for a particular check
func (conf Configuration) GetSeverity(category string, name string) Severity {
	subConf := conf.GetCategoryConfig(category)
	subConfRef := reflect.ValueOf(subConf)
	fieldVal := reflect.Indirect(subConfRef).FieldByName(name).Interface()
	if severity, ok := fieldVal.(Severity); ok {
		return severity
	}
	// TODO: don't panic
	panic("Unknown severity: " + category + "/" + name)
}
