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
	Target         Target                `yaml:"target"`
	Schema         jsonschema.RootSchema `yaml:"schema"`
}

var (
	schemaBox = (*packr.Box)(nil)
	checks    = map[Target][]SchemaCheck{
		TargetContainer: []SchemaCheck{},
		TargetPod:       []SchemaCheck{},
	}
	checkOrder = []string{
		"hostIPC",
		"hostPID",
		"hostNetwork",
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

func applyPodSchemaChecks(conf *config.Configuration, pod *corev1.PodSpec, controllerName string, pv *PodValidation) error {
	for _, check := range checks[TargetPod] {
		if !conf.IsActionable(check.Category, check.Name, controllerName) {
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
