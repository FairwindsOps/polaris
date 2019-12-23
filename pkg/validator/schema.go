package validator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	packr "github.com/gobuffalo/packr/v2"
	"github.com/qri-io/jsonschema"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/fairwindsops/polaris/pkg/config"
	controller "github.com/fairwindsops/polaris/pkg/validator/controllers"
)

type IncludeExcludeList struct {
	Include []string `yaml:"include"`
	Exclude []string `yaml:"exclude"`
}

type Target string

const (
	TargetContainer Target = "Container"
	TargetPod       Target = "Pod"
)

type SchemaCheck struct {
	Name           string                `yaml:"name"`
	ID             string                `yaml:"id"`
	Category       string                `yaml:"category"`
	SuccessMessage string                `yaml:"success_message"`
	FailureMessage string                `yaml:"failure_message"`
	Controllers    IncludeExcludeList    `yaml:"controllers"`
	Containers     IncludeExcludeList    `yaml:"containers"`
	Target         Target                `yaml:"target"`
	Schema         jsonschema.RootSchema `yaml:"schema"`
}

var (
	schemaBox = (*packr.Box)(nil)
	checks    = map[Target][]SchemaCheck{
		TargetContainer: []SchemaCheck{},
		TargetPod:       []SchemaCheck{},
	}
	// We explicitly set the order to avoid thrash in the
	// tests as we migrate toward JSON schema
	checkOrder = []string{
		// Pod checks
		"hostIPC",
		"hostPID",
		"hostNetwork",
		// Container checks
		"readinessProbe",
		"livenessProbe",
		"pullPolicyNotAlways",
		"tagNotSpecified",
		"hostPortSet",
	}
)

func init() {
	schemaBox = packr.New("Schemas", "../../checks")
	for _, file := range checkOrder {
		contents, err := schemaBox.Find(file + ".yaml")
		if err != nil {
			panic(err)
		}
		check, err := parseCheck(contents)
		if err != nil {
			panic(err)
		}
		checks[check.Target] = append(checks[check.Target], check)
	}
}

func parseCheck(rawBytes []byte) (SchemaCheck, error) {
	reader := bytes.NewReader(rawBytes)
	check := SchemaCheck{}
	d := yaml.NewYAMLOrJSONDecoder(reader, 4096)
	for {
		if err := d.Decode(&check); err != nil {
			if err == io.EOF {
				return check, nil
			}
			return check, fmt.Errorf("Decoding schema check failed: %v", err)
		}
	}
}

func (check SchemaCheck) check(controller controller.Interface) (bool, error) {
	pod := controller.GetPodSpec()
	if check.Target == TargetPod {
		return check.checkPod(pod)
	} else if check.Target == TargetContainer {
		for _, container := range pod.Containers {
			bytes, err := json.Marshal(container)
			if err != nil {
				return false, err
			}
			errors, err := check.Schema.ValidateBytes(bytes)
			if err != nil || len(errors) > 0 {
				return false, err
			}
		}
		// TODO: initcontainers
	}
	return true, nil
}

func (check SchemaCheck) checkPod(pod *corev1.PodSpec) (bool, error) {
	bytes, err := json.Marshal(pod)
	if err != nil {
		return false, err
	}
	errors, err := check.Schema.ValidateBytes(bytes)
	return len(errors) == 0, err
}

func (check SchemaCheck) checkContainer(container *corev1.Container) (bool, error) {
	bytes, err := json.Marshal(container)
	if err != nil {
		return false, err
	}
	errors, err := check.Schema.ValidateBytes(bytes)
	return len(errors) == 0, err
}

func (check SchemaCheck) isActionable(target Target, controllerType config.SupportedController, isInit bool) bool {
	if check.Target != target {
		return false
	}
	isIncluded := len(check.Controllers.Include) == 0
	for _, inclusion := range check.Controllers.Include {
		if config.GetSupportedControllerFromString(inclusion) == controllerType {
			isIncluded = true
			break
		}
	}
	if !isIncluded {
		return false
	}
	for _, exclusion := range check.Controllers.Exclude {
		if config.GetSupportedControllerFromString(exclusion) == controllerType {
			return false
		}
	}
	if check.Target == TargetContainer {
		isIncluded := len(check.Containers.Include) == 0
		for _, inclusion := range check.Containers.Include {
			if (inclusion == "initContainer" && isInit) || (inclusion == "container" && !isInit) {
				isIncluded = true
				break
			}
		}
		if !isIncluded {
			return false
		}
		for _, exclusion := range check.Containers.Exclude {
			if (exclusion == "initContainer" && isInit) || (exclusion == "container" && !isInit) {
				return false
			}
		}
	}
	return true
}

func applyPodSchemaChecks(conf *config.Configuration, pod *corev1.PodSpec, controllerName string, controllerType config.SupportedController, pv *PodValidation) error {
	for _, check := range checks[TargetPod] {
		if !conf.IsActionable(check.Category, check.Name, controllerName) {
			continue
		}
		if !check.isActionable(TargetPod, controllerType, false) {
			continue
		}
		severity := conf.GetSeverity(check.Category, check.Name)
		passes, err := check.checkPod(pod)
		if err != nil {
			return err
		}
		if passes {
			pv.addSuccess(check.SuccessMessage, check.Category, check.ID)
		} else {
			pv.addFailure(check.FailureMessage, severity, check.Category, check.ID)
		}
	}
	return nil
}

func applyContainerSchemaChecks(conf *config.Configuration, container *corev1.Container, controllerName string, controllerType config.SupportedController, isInit bool, cv *ContainerValidation) error {
	for _, check := range checks[TargetContainer] {
		if !conf.IsActionable(check.Category, check.Name, controllerName) {
			continue
		}
		if !check.isActionable(TargetContainer, controllerType, isInit) {
			continue
		}
		severity := conf.GetSeverity(check.Category, check.Name)
		passes, err := check.checkContainer(container)
		if err != nil {
			return err
		}
		if passes {
			cv.addSuccess(check.SuccessMessage, check.Category, check.ID)
		} else {
			cv.addFailure(check.FailureMessage, severity, check.Category, check.ID)
		}
	}
	return nil
}
