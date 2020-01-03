package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
	// ReplicationControllers are supported controllers for scanning pod specs
	ReplicationControllers
	// NakedPods are a pseudo-controller for scanning pod specs
	NakedPods
)

// ControllerStrings are strongly ordered to match the SupportedController enum
var ControllerStrings = []string{
	"Unsupported",
	"Deployment",
	"StatefulSet",
	"DaemonSet",
	"Job",
	"CronJob",
	"ReplicationController",
	"NakedPod",
}

// stringLookupForSupportedControllers is the list of lowercase singular and plural strings for string to enum lookup
var stringLookupForSupportedControllers = map[string]SupportedController{
	"deployment":             Deployments,
	"deployments":            Deployments,
	"statefulset":            StatefulSets,
	"statefulsets":           StatefulSets,
	"daemonset":              DaemonSets,
	"daemonsets":             DaemonSets,
	"job":                    Jobs,
	"jobs":                   Jobs,
	"cronjob":                CronJobs,
	"cronjobs":               CronJobs,
	"replicationcontroller":  ReplicationControllers,
	"replicationcontrollers": ReplicationControllers,
	"nakedpod":               NakedPods,
	"nakedpods":              NakedPods,
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

	*s = GetSupportedControllerFromString(j)
	if *s == Unsupported {
		return fmt.Errorf("Unsupported controller kind: %s", j)
	}
	return nil
}

// ListSupportedAPIVersions for SupportedController returns all the apimachinery object type supported
func (s SupportedController) ListSupportedAPIVersions() []runtime.Object {
	var supportedVersions []runtime.Object
	switch s {
	case Deployments:
		supportedVersions = []runtime.Object{
			&appsv1.Deployment{},
		}
	case StatefulSets:
		supportedVersions = []runtime.Object{
			&appsv1.StatefulSet{},
		}
	case DaemonSets:
		supportedVersions = []runtime.Object{
			&appsv1.DaemonSet{},
		}
	case Jobs:
		supportedVersions = []runtime.Object{
			&batchv1.Job{},
		}
	case CronJobs:
		supportedVersions = []runtime.Object{
			&batchv1beta1.CronJob{},
		}
	case ReplicationControllers:
		supportedVersions = []runtime.Object{
			&corev1.ReplicationController{},
		}
	case NakedPods:
		supportedVersions = []runtime.Object{
			&corev1.Pod{},
		}
	}
	return supportedVersions
}

// GetSupportedControllerFromString fuzzy matches a string with a SupportedController Enum
func GetSupportedControllerFromString(str string) SupportedController {
	lowerStr := strings.ToLower(str)
	controller, keyFound := stringLookupForSupportedControllers[lowerStr]
	if !keyFound {
		controller = Unsupported
	}
	return controller
}

// CheckIfKindIsConfiguredForValidation takes a kind (in string format) and checks if Polaris is configured to scan this type of controller
func (c Configuration) CheckIfKindIsConfiguredForValidation(kind string) bool {
	controller := GetSupportedControllerFromString(kind)
	if controller != Unsupported {
		for _, controllerToScan := range c.ControllersToScan {
			if controller == controllerToScan {
				return true
			}
		}
	}
	return false
}
