package config

import (
	"context"
	"encoding/json"
	"fmt"

	jptr "github.com/qri-io/jsonpointer"
	"github.com/qri-io/jsonschema"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// TargetKind represents the part of the config to be validated
type TargetKind string

const (
	// TargetContainer points to the container spec
	TargetContainer TargetKind = "Container"
	// TargetPod points to the pod spec
	TargetPod TargetKind = "Pod"
	// TargetController points to the controller's spec
	TargetController TargetKind = "Controller"
)

// SchemaCheck is a Polaris check that runs using JSON Schema
type SchemaCheck struct {
	ID             string             `yaml:"id"`
	Category       string             `yaml:"category"`
	SuccessMessage string             `yaml:"successMessage"`
	FailureMessage string             `yaml:"failureMessage"`
	Controllers    includeExcludeList `yaml:"controllers"`
	Containers     includeExcludeList `yaml:"containers"`
	Target         TargetKind         `yaml:"target"`
	SchemaTarget   TargetKind         `yaml:"schemaTarget"`
	Schema         jsonschema.Schema  `yaml:"schema"`
	JSONSchema     string             `yaml:"jsonSchema"`
}

type resourceMinimum string
type resourceMaximum string

func init() {
	jsonschema.RegisterKeyword("resourceMinimum", newResourceMinimum)
	jsonschema.RegisterKeyword("resourceMaximum", newResourceMaximum)
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

// Register implements jsonschema.Keyword
func (min *resourceMinimum) Register(uri string, registry *jsonschema.SchemaRegistry) {
}

// Register implements jsonschema.Keyword
func (max *resourceMaximum) Register(uri string, registry *jsonschema.SchemaRegistry) {
}

// Resolve implements jsonschema.Keyword
func (min *resourceMinimum) Resolve(pointer jptr.Pointer, uri string) *jsonschema.Schema {
	return nil
}

// Resolve implements jsonschema.Keyword
func (max *resourceMaximum) Resolve(pointer jptr.Pointer, uri string) *jsonschema.Schema {
	return nil
}

// ValidateKeyword checks that a specified quanitity is not less than the minimum
func (min *resourceMinimum) ValidateKeyword(ctx context.Context, currentState *jsonschema.ValidationState, data interface{}) {
	errorMessage := validateRange(string(*min), data, true)
	if errorMessage != "" {
		currentState.AddError(data, errorMessage)
	}
}

// Validate checks that a specified quanitity is not greater than the maximum
func (max *resourceMaximum) ValidateKeyword(ctx context.Context, currentState *jsonschema.ValidationState, data interface{}) {
	errorMessage := validateRange(string(*max), data, false)
	if errorMessage != "" {
		currentState.AddError(data, errorMessage)
	}
}

func parseQuantity(i interface{}) (resource.Quantity, *[]jsonschema.KeyError) {
	resStr, ok := i.(string)
	if !ok {
		return resource.Quantity{}, &[]jsonschema.KeyError{
			{Message: fmt.Sprintf("Resource quantity %v is not a string", i)},
		}
	}
	q, err := resource.ParseQuantity(resStr)
	if err != nil {
		return resource.Quantity{}, &[]jsonschema.KeyError{
			{Message: fmt.Sprintf("Could not parse resource quantity: %s", resStr)},
		}
	}
	return q, nil
}

func validateRange(limit interface{}, data interface{}, isMinimum bool) string {
	limitQuantity, err := parseQuantity(limit)
	if err != nil {
		return fmt.Sprintf("%v", err)
	}
	actualQuantity, err := parseQuantity(data)
	if err != nil {
		return fmt.Sprintf("%v", err)
	}
	cmp := limitQuantity.Cmp(actualQuantity)
	if isMinimum {
		if cmp == 1 {
			return fmt.Sprintf("quantity %v is > %v", actualQuantity, limitQuantity)
		}
	} else {
		if cmp == -1 {
			return fmt.Sprintf("quantity %v is < %v", actualQuantity, limitQuantity)
		}
	}
	return ""
}

// Initialize sets up the schema
func (check *SchemaCheck) Initialize(id string) error {
	check.ID = id
	if check.JSONSchema != "" {
		if err := json.Unmarshal([]byte(check.JSONSchema), &check.Schema); err != nil {
			return err
		}
	}
	return nil
}

// CheckPod checks a pod spec against the schema
func (check SchemaCheck) CheckPod(ctx context.Context, pod *corev1.PodSpec) (bool, error) {
	return check.CheckObject(ctx, pod)
}

// CheckController checks a controler's spec against the schema
func (check SchemaCheck) CheckController(ctx context.Context, bytes []byte) (bool, error) {
	errs, err := check.Schema.ValidateBytes(ctx, bytes)
	return len(errs) == 0, err
}

// CheckContainer checks a container spec against the schema
func (check SchemaCheck) CheckContainer(ctx context.Context, container *corev1.Container) (bool, error) {
	return check.CheckObject(ctx, container)
}

// CheckObject checks arbitrary data against the schema
func (check SchemaCheck) CheckObject(ctx context.Context, obj interface{}) (bool, error) {
	bytes, err := json.Marshal(obj)
	if err != nil {
		return false, err
	}
	errs, err := check.Schema.ValidateBytes(ctx, bytes)
	return len(errs) == 0, err
}

// IsActionable decides if this check applies to a particular target
func (check SchemaCheck) IsActionable(target TargetKind, controllerType string, isInit bool) bool {
	if check.Target != target {
		return false
	}
	isIncluded := len(check.Controllers.Include) == 0
	for _, inclusion := range check.Controllers.Include {
		if inclusion == controllerType {
			isIncluded = true
			break
		}
	}
	if !isIncluded {
		return false
	}
	for _, exclusion := range check.Controllers.Exclude {
		if exclusion == controllerType {
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
