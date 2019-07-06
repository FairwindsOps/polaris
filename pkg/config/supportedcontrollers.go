package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

const (
	// Unsupported is the default enum for non-defined controller types
	Unsupported SupportedController = iota
	// Deployments are a supported controller for scanning pod specs
	Deployments
	// StatefulSets are a supported controller for scanning pod specs
	StatefulSets
	// DaemonSets are a supported controller for scanning pod specs
	DaemonSets
	// Jobs are a supported controller for scanning pod specs
	Jobs
	// CronJobs are a supported controller for scanning pod specs
	CronJobs
)

// ControllerStrings are strongly ordered to match the SupportedController enum
var ControllerStrings = []string{
	"Unsupported",
	"Deployments",
	"StatefulSets",
	"DaemonSets",
	"Jobs",
	"CronJobs",
}

// stringLookupForSupportedControllers is the list of lowercase singular and plural strings for string to enum lookup
var stringLookupForSupportedControllers = map[string]SupportedController{
	"deployment":   Deployments,
	"deployments":  Deployments,
	"statefulset":  StatefulSets,
	"statefulsets": StatefulSets,
	"daemonset":    DaemonSets,
	"daemonsets":   DaemonSets,
	"job":          Jobs,
	"jobs":         Jobs,
	"cronjob":      CronJobs,
	"cronjobs":     CronJobs,
}

// SupportedController is a constant item of a controller that is supported for scanning pod specs
type SupportedController int

// String returns the string name for a given SupportedController enum
func (s SupportedController) String() string {
	return ControllerStrings[s]
}

// MarshalJSON manages writing the enum into json data or error on unsupported value
func (s SupportedController) MarshalJSON() ([]byte, error) {
	if s == Unsupported {
		return []byte{}, fmt.Errorf("Unsupported is not a valid Supported Controller")
	}
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(s.String())
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

// UnmarshalJSON handles reading json data into enum
func (s *SupportedController) UnmarshalJSON(b []byte) error {
	var j string
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}

	*s, err = GetSupportedControllerFromString(j)
	if err != nil {
		return err
	}
	return nil
}

// GetSupportedControllerFromString fuzzy matches a string with a SupportedController Enum
func GetSupportedControllerFromString(str string) (SupportedController, error) {
	lowerStr := strings.ToLower(str)
	if stringLookupForSupportedControllers[lowerStr] == Unsupported {
		return 0, fmt.Errorf("Value ('%v') in configuration was not found in Supported Controllers: (%v)", str, strings.Join(ControllerStrings, ","))
	}
	return stringLookupForSupportedControllers[lowerStr], nil
}

// CheckIfKindIsConfiguredForValidation takes a kind (in string format) and checks if Polaris is configured to scan this type of controller
func (c Configuration) CheckIfKindIsConfiguredForValidation(kind string) bool {
	controller, err := GetSupportedControllerFromString(kind)
	// if no errors then we found the kind in supported controller types
	if err == nil {
		// see if the kind exists in the controllers to scan config
		for _, controllerToScan := range c.ControllersToScan {
			if controller == controllerToScan {
				return true
			}
		}
	}
	return false
}
