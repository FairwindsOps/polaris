package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/qri-io/jsonschema"
	"github.com/thoas/go-funk"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	k8sYaml "k8s.io/apimachinery/pkg/util/yaml"
)

// TargetKind represents the part of the config to be validated
type TargetKind string

const (
	// TargetController points to the controller's spec
	TargetController TargetKind = "Controller"
	// TargetContainer points to the container spec
	TargetContainer TargetKind = "Container"
	// TargetPod points to the pod spec
	TargetPod TargetKind = "Pod"
)

// HandledTargets is a list of target names that are explicitly handled
var HandledTargets = []TargetKind{
	TargetController,
	TargetContainer,
	TargetPod,
}

// SchemaCheck is a Polaris check that runs using JSON Schema
type SchemaCheck struct {
	ID             string                 `yaml:"id" json:"id"`
	Category       string                 `yaml:"category" json:"category"`
	SuccessMessage string                 `yaml:"successMessage" json:"successMessage"`
	FailureMessage string                 `yaml:"failureMessage" json:"failureMessage"`
	Controllers    includeExcludeList     `yaml:"controllers" json:"controllers"`
	Containers     includeExcludeList     `yaml:"containers" json:"containers"`
	Target         TargetKind             `yaml:"target" json:"target"`
	SchemaTarget   TargetKind             `yaml:"schemaTarget" json:"schemaTarget"`
	Schema         map[string]interface{} `yaml:"schema" json:"schema"`
	SchemaString   string                 `yaml:"jsonSchema" json:"jsonSchema"`
	Validator      jsonschema.RootSchema  `yaml:"-"`
}

type resourceMinimum string
type resourceMaximum string

func ParseCheck(id string, rawBytes []byte) (SchemaCheck, error) {
	reader := bytes.NewReader(rawBytes)
	check := SchemaCheck{}
	d := k8sYaml.NewYAMLOrJSONDecoder(reader, 4096)
	for {
		if err := d.Decode(&check); err != nil {
			if err == io.EOF {
				break
			}
			return check, fmt.Errorf("Decoding schema check failed: %v", err)
		}
	}
	check.Initialize(id)
	return check, nil
}

func init() {
	jsonschema.RegisterValidator("resourceMinimum", newResourceMinimum)
	jsonschema.RegisterValidator("resourceMaximum", newResourceMaximum)
}

type includeExcludeList struct {
	Include []string `yaml:"include"`
	Exclude []string `yaml:"exclude"`
}

func newResourceMinimum() jsonschema.Validator {
	return new(resourceMinimum)
}

func newResourceMaximum() jsonschema.Validator {
	return new(resourceMaximum)
}

// Validate checks that a specified quanitity is not less than the minimum
func (min resourceMinimum) Validate(path string, data interface{}, errs *[]jsonschema.ValError) {
	err := validateRange(path, string(min), data, true)
	if err != nil {
		*errs = append(*errs, *err...)
	}
}

// Validate checks that a specified quanitity is not greater than the maximum
func (max resourceMaximum) Validate(path string, data interface{}, errs *[]jsonschema.ValError) {
	err := validateRange(path, string(max), data, false)
	if err != nil {
		*errs = append(*errs, *err...)
	}
}

func parseQuantity(i interface{}) (resource.Quantity, *[]jsonschema.ValError) {
	resStr, ok := i.(string)
	if !ok {
		return resource.Quantity{}, &[]jsonschema.ValError{
			{Message: fmt.Sprintf("Resource quantity %v is not a string", i)},
		}
	}
	q, err := resource.ParseQuantity(resStr)
	if err != nil {
		return resource.Quantity{}, &[]jsonschema.ValError{
			{Message: fmt.Sprintf("Could not parse resource quantity: %s", resStr)},
		}
	}
	return q, nil
}

func validateRange(path string, limit interface{}, data interface{}, isMinimum bool) *[]jsonschema.ValError {
	limitQuantity, err := parseQuantity(limit)
	if err != nil {
		return err
	}
	actualQuantity, err := parseQuantity(data)
	if err != nil {
		return err
	}
	cmp := limitQuantity.Cmp(actualQuantity)
	if isMinimum {
		if cmp == 1 {
			return &[]jsonschema.ValError{
				{Message: fmt.Sprintf("%s quantity %v is > %v", path, actualQuantity, limitQuantity)},
			}
		}
	} else {
		if cmp == -1 {
			return &[]jsonschema.ValError{
				{Message: fmt.Sprintf("%s quantity %v is < %v", path, actualQuantity, limitQuantity)},
			}
		}
	}
	return nil
}

// Initialize sets up the schema
func (check *SchemaCheck) Initialize(id string) error {
	check.ID = id
	if check.SchemaString == "" {
		jsonBytes, err := json.Marshal(check.Schema)
		if err != nil {
			return err
		}
		check.SchemaString = string(jsonBytes)
	}
	err := json.Unmarshal([]byte(check.SchemaString), &check.Validator)
	return err
}

func (check SchemaCheck) TemplateForResource(res interface{}) (*SchemaCheck, error) {
	newCheck := check
	tmpl := template.New(newCheck.ID)
	tmpl, err := tmpl.Parse(newCheck.SchemaString)
	if err != nil {
		return nil, err
	}

	w := bytes.Buffer{}
	err = tmpl.Execute(&w, res)
	if err != nil {
		return nil, err
	}

	newCheck.SchemaString = w.String()
	newCheck.Initialize(newCheck.ID)
	return &newCheck, err
}

// CheckPod checks a pod spec against the schema
func (check SchemaCheck) CheckPod(pod *corev1.PodSpec) (bool, []jsonschema.ValError, error) {
	return check.CheckObject(pod)
}

// CheckController checks a controler's spec against the schema
func (check SchemaCheck) CheckController(bytes []byte) (bool, []jsonschema.ValError, error) {
	errs, err := check.Validator.ValidateBytes(bytes)
	return len(errs) == 0, errs, err
}

// CheckContainer checks a container spec against the schema
func (check SchemaCheck) CheckContainer(container *corev1.Container) (bool, []jsonschema.ValError, error) {
	return check.CheckObject(container)
}

// CheckObject checks arbitrary data against the schema
func (check SchemaCheck) CheckObject(obj interface{}) (bool, []jsonschema.ValError, error) {
	bytes, err := json.Marshal(obj)
	if err != nil {
		return false, nil, err
	}
	errs, err := check.Validator.ValidateBytes(bytes)
	return len(errs) == 0, errs, err
}

// IsActionable decides if this check applies to a particular target
func (check SchemaCheck) IsActionable(target TargetKind, kind string, isInit bool) bool {
	if funk.Contains(HandledTargets, target) {
		if check.Target != target {
			return false
		}
	} else if string(check.Target) != kind && !strings.HasSuffix(string(check.Target), "/"+kind) {
		return false
	}
	isIncluded := len(check.Controllers.Include) == 0
	for _, inclusion := range check.Controllers.Include {
		if inclusion == kind {
			isIncluded = true
			break
		}
	}
	if !isIncluded {
		return false
	}
	for _, exclusion := range check.Controllers.Exclude {
		if exclusion == kind {
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
