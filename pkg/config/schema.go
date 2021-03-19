package config

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/qri-io/jsonpointer"
	"github.com/qri-io/jsonschema"
	"github.com/thoas/go-funk"
	"gopkg.in/yaml.v3"
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
	ID             string                 `yaml:"id"`
	Category       string                 `yaml:"category"`
	SuccessMessage string                 `yaml:"successMessage"`
	FailureMessage string                 `yaml:"failureMessage"`
	Controllers    includeExcludeList     `yaml:"controllers"`
	Containers     includeExcludeList     `yaml:"containers"`
	Target         TargetKind             `yaml:"target"`
	SchemaTarget   TargetKind             `yaml:"schemaTarget"`
	Schema         map[string]interface{} `yaml:"schema"`
	SchemaFoo      jsonschema.Schema      `yaml:""`
	JSONSchema     string                 `yaml:"jsonSchema"`
}

type resourceMinimum string
type resourceMaximum string

func ParseCheck(rawBytes []byte) (SchemaCheck, error) {
	reader := bytes.NewReader(rawBytes)
	check := SchemaCheck{}
	d := k8sYaml.NewYAMLOrJSONDecoder(reader, 4096)
	for {
		if err := d.Decode(&check); err != nil {
			if err == io.EOF {
				return check, nil
			}
			return check, fmt.Errorf("Decoding schema check failed: %v", err)
		}
	}
}

func init() {
	jsonschema.RegisterKeyword("resourceMinimum", newResourceMinimum)
	jsonschema.RegisterKeyword("resourceMaximum", newResourceMaximum)
	jsonschema.LoadDraft2019_09()
}

type includeExcludeList struct {
	Include []string `yaml:"include"`
	Exclude []string `yaml:"exclude"`
}

func newResourceMinimum() jsonschema.Keyword {
	return new(resourceMinimum)
}

func newResourceMaximum() jsonschema.Keyword {
	return new(resourceMaximum)
}

func (min *resourceMinimum) Register(uri string, registry *jsonschema.SchemaRegistry) {}
func (min *resourceMinimum) Resolve(pointer jsonpointer.Pointer, uri string) *jsonschema.Schema {
	return nil
}
func (min *resourceMinimum) Validate(propPath string, data interface{}, errs *[]jsonschema.KeyError) {}

// Validate checks that a specified quanitity is not less than the minimum
func (min *resourceMinimum) ValidateKeyword(ctx context.Context, currentState *jsonschema.ValidationState, data interface{}) {
	err := validateRange(string(*min), data, true)
	if err != nil {
		currentState.AddError(data, err.Error())
	}
}

func (min *resourceMaximum) Register(uri string, registry *jsonschema.SchemaRegistry) {}
func (min *resourceMaximum) Resolve(pointer jsonpointer.Pointer, uri string) *jsonschema.Schema {
	return nil
}
func (min *resourceMaximum) Validate(propPath string, data interface{}, errs *[]jsonschema.KeyError) {}

// Validate checks that a specified quanitity is not less than the minimum
func (min *resourceMaximum) ValidateKeyword(ctx context.Context, currentState *jsonschema.ValidationState, data interface{}) {
	err := validateRange(string(*min), data, false)
	if err != nil {
		currentState.AddError(data, err.Error())
	}
}

func parseQuantity(i interface{}) (resource.Quantity, error) {
	resStr, ok := i.(string)
	if !ok {
		return resource.Quantity{}, fmt.Errorf("Resource quantity %v is not a string", i)
	}
	q, err := resource.ParseQuantity(resStr)
	if err != nil {
		return resource.Quantity{}, fmt.Errorf("Could not parse resource quantity: %s", resStr)
	}
	return q, nil
}

func validateRange(limit interface{}, data interface{}, isMinimum bool) error {
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
			return fmt.Errorf("quantity %v is > %v", actualQuantity, limitQuantity)
		}
	} else {
		if cmp == -1 {
			return fmt.Errorf("quantity %v is < %v", actualQuantity, limitQuantity)
		}
	}
	return nil
}

// Initialize sets up the schema
func (check *SchemaCheck) Initialize(id string) error {
	check.ID = id
	if check.JSONSchema != "" {
		if err := json.Unmarshal([]byte(check.JSONSchema), &check.SchemaFoo); err != nil {
			return err
		}
	} else {
		jsonBytes, err := json.Marshal(check.Schema)
		if err != nil {
			return err
		}
		err = json.Unmarshal(jsonBytes, &check.SchemaFoo)
		if err != nil {
			return err
		}
	}
	return nil
}

func (check SchemaCheck) TemplateForResource(res interface{}) (*SchemaCheck, error) {
	yamlBytes, err := yaml.Marshal(check)
	if err != nil {
		return nil, err
	}

	tmpl := template.New(check.ID)
	tmpl, err = tmpl.Parse(string(yamlBytes))
	if err != nil {
		return nil, err
	}

	w := bytes.Buffer{}
	err = tmpl.Execute(&w, res)
	if err != nil {
		return nil, err
	}

	newCheck, err := ParseCheck(w.Bytes())
	if err != nil {
		return nil, err
	}
	err = newCheck.Initialize(check.ID)
	if err != nil {
		return nil, err
	}
	return &newCheck, nil
}

// CheckPod checks a pod spec against the schema
func (check SchemaCheck) CheckPod(pod *corev1.PodSpec) (bool, []jsonschema.KeyError, error) {
	return check.CheckObject(pod)
}

// CheckController checks a controler's spec against the schema
func (check SchemaCheck) CheckController(bytes []byte) (bool, []jsonschema.KeyError, error) {
	errs, err := check.SchemaFoo.ValidateBytes(context.TODO(), bytes)
	return len(errs) == 0, errs, err
}

// CheckContainer checks a container spec against the schema
func (check SchemaCheck) CheckContainer(container *corev1.Container) (bool, []jsonschema.KeyError, error) {
	return check.CheckObject(container)
}

// CheckObject checks arbitrary data against the schema
func (check SchemaCheck) CheckObject(obj interface{}) (bool, []jsonschema.KeyError, error) {
	bytes, err := json.Marshal(obj)
	if err != nil {
		return false, nil, err
	}
	errs, err := check.SchemaFoo.ValidateBytes(context.TODO(), bytes)
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
