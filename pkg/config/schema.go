// Copyright 2022 FairwindsOps, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/qri-io/jsonpointer"
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
	// TargetPodSpec points to the pod spec
	TargetPodSpec TargetKind = "PodSpec"
	// TargetPodTemplate points to the pod template
	TargetPodTemplate TargetKind = "PodTemplate"
)

// HandledTargets is a list of target names that are explicitly handled
var HandledTargets = []TargetKind{
	TargetController,
	TargetContainer,
	TargetPodSpec,
	TargetPodTemplate,
}

// Mutation defines how to change a YAML file, in the style of JSON Patch
type Mutation struct {
	Path    string
	Op      string
	Value   interface{}
	Comment string
}

// SchemaCheck is a Polaris check that runs using JSON Schema
type SchemaCheck struct {
	ID                      string                            `yaml:"id" json:"id"`
	Category                string                            `yaml:"category" json:"category"`
	SuccessMessage          string                            `yaml:"successMessage" json:"successMessage"`
	FailureMessage          string                            `yaml:"failureMessage" json:"failureMessage"`
	Controllers             includeExcludeList                `yaml:"controllers" json:"controllers"`
	Containers              includeExcludeList                `yaml:"containers" json:"containers"`
	Target                  TargetKind                        `yaml:"target" json:"target"`
	SchemaTarget            TargetKind                        `yaml:"schemaTarget" json:"schemaTarget"`
	Schema                  map[string]interface{}            `yaml:"schema" json:"schema"`
	SchemaString            string                            `yaml:"schemaString" json:"schemaString"`
	Validator               *jsonschema.Schema                `yaml:"-" json:"-"`
	AdditionalSchemas       map[string]map[string]interface{} `yaml:"additionalSchemas" json:"additionalSchemas"`
	AdditionalSchemaStrings map[string]string                 `yaml:"additionalSchemaStrings" json:"additionalSchemaStrings"`
	AdditionalValidators    map[string]*jsonschema.Schema     `yaml:"-" json:"-"`
	Mutations               []Mutation                        `yaml:"mutations" json:"mutations"`
}

type resourceMinimum string
type resourceMaximum string

// UnmarshalYAMLOrJSON is a helper function to unmarshal data in an arbitrary format
func UnmarshalYAMLOrJSON(raw []byte, dest interface{}) error {
	reader := bytes.NewReader(raw)
	d := k8sYaml.NewYAMLOrJSONDecoder(reader, 4096)
	for {
		if err := d.Decode(dest); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("Decoding schema check failed: %v", err)
		}
	}
	return nil
}

// ParseCheck parses a check from a byte array
func ParseCheck(id string, rawBytes []byte) (SchemaCheck, error) {
	check := SchemaCheck{}
	err := UnmarshalYAMLOrJSON(rawBytes, &check)
	if err != nil {
		return check, err
	}
	check.Initialize(id)
	return check, nil
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

// ValidateKeyword checks that a specified quantity is not less than the minimum
func (min resourceMinimum) ValidateKeyword(ctx context.Context, currentState *jsonschema.ValidationState, data interface{}) {
	err := validateRange(currentState.InstanceLocation.String(), string(min), data, true)
	if len(err) > 0 {
		currentState.AddSubErrors(err...)
	}
}

// Register implements jsonschema.Keyword interface
func (min resourceMinimum) Register(uri string, registry *jsonschema.SchemaRegistry) {
	// No registration needed for custom validators
}

// Resolve implements jsonschema.Keyword interface
func (min resourceMinimum) Resolve(pointer jsonpointer.Pointer, uri string) *jsonschema.Schema {
	return nil
}

// ValidateKeyword checks that a specified quantity is not greater than the maximum
func (max resourceMaximum) ValidateKeyword(ctx context.Context, currentState *jsonschema.ValidationState, data interface{}) {
	err := validateRange(currentState.InstanceLocation.String(), string(max), data, false)
	if len(err) > 0 {
		currentState.AddSubErrors(err...)
	}
}

// Register implements jsonschema.Keyword interface
func (max resourceMaximum) Register(uri string, registry *jsonschema.SchemaRegistry) {
	// No registration needed for custom validators
}

// Resolve implements jsonschema.Keyword interface
func (max resourceMaximum) Resolve(pointer jsonpointer.Pointer, uri string) *jsonschema.Schema {
	return nil
}

func parseQuantity(i interface{}) (resource.Quantity, []jsonschema.KeyError) {
	if resNum, ok := i.(float64); ok {
		i = fmt.Sprintf("%f", resNum)
	}
	resStr, ok := i.(string)
	if !ok {
		return resource.Quantity{}, []jsonschema.KeyError{
			{Message: fmt.Sprintf("Resource quantity %v is not a string", i)},
		}
	}
	q, err := resource.ParseQuantity(resStr)
	if err != nil {
		return resource.Quantity{}, []jsonschema.KeyError{
			{Message: fmt.Sprintf("Could not parse resource quantity: %s", resStr)},
		}
	}
	return q, nil
}

func validateRange(path string, limit interface{}, data interface{}, isMinimum bool) []jsonschema.KeyError {
	limitQuantity, err := parseQuantity(limit)
	if len(err) > 0 {
		return err
	}
	actualQuantity, err := parseQuantity(data)
	if len(err) > 0 {
		return err
	}
	cmp := limitQuantity.Cmp(actualQuantity)
	if isMinimum {
		if cmp == 1 {
			return []jsonschema.KeyError{
				{Message: fmt.Sprintf("%s quantity %v is > %v", path, actualQuantity, limitQuantity)},
			}
		}
	} else {
		if cmp == -1 {
			return []jsonschema.KeyError{
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
	if check.AdditionalSchemaStrings == nil {
		check.AdditionalSchemaStrings = make(map[string]string)
	}
	for kind, schema := range check.AdditionalSchemas {
		jsonBytes, err := json.Marshal(schema)
		if err != nil {
			return err
		}
		check.AdditionalSchemaStrings[kind] = string(jsonBytes)
	}
	check.Schema = map[string]interface{}{}
	check.AdditionalSchemas = map[string]map[string]interface{}{}
	return nil
}

// TemplateForResource fills out a check's templated fields given a particular resource
func (check SchemaCheck) TemplateForResource(res interface{}) (*SchemaCheck, error) {
	newCheck := check // Make a copy of the check, since we're going to modify the schema

	templateStrings := map[string]string{
		"": newCheck.SchemaString,
	}
	for kind, schema := range newCheck.AdditionalSchemaStrings {
		templateStrings[kind] = schema
	}
	newCheck.SchemaString = ""
	newCheck.AdditionalSchemaStrings = map[string]string{}

	for kind, tmplString := range templateStrings {
		tmpl := template.New(newCheck.ID).Funcs(template.FuncMap{
			"hasPrefix": strings.HasPrefix,
			"hasSuffix": strings.HasSuffix,
		})
		tmpl, err := tmpl.Parse(tmplString)
		if err != nil {
			return nil, err
		}
		w := bytes.Buffer{}
		err = tmpl.Execute(&w, res)
		if err != nil {
			return nil, err
		}
		templated := w.String()
		if strings.TrimSpace(templated) == "" {
			continue
		}

		if kind == "" {
			newCheck.SchemaString = templated
		} else {
			newCheck.AdditionalSchemaStrings[kind] = templated
		}
	}

	newCheck.AdditionalValidators = map[string]*jsonschema.Schema{}
	for kind, schemaStr := range newCheck.AdditionalSchemaStrings {
		// Replace draft-07 with draft 2019-09 for compatibility
		//schemaStr = strings.ReplaceAll(schemaStr, `"$schema": "http://json-schema.org/draft/2019-09/schema"`, `"$schema": "http://json-schema.org/draft/2019-09/schema"`)
		val := jsonschema.Must(schemaStr)
		val.Register("", jsonschema.GetSchemaRegistry())
		newCheck.AdditionalValidators[kind] = val
	}
	// Replace draft-07 with draft 2019-09 for compatibility
	//newCheck.SchemaString = strings.ReplaceAll(newCheck.SchemaString, `"$schema": "http://json-schema.org/draft/2019-09/schema"`, `"$schema": "http://json-schema.org/draft/2019-09/schema"`)
	fmt.Println("x======", newCheck.SchemaString)
	newCheck.Validator = jsonschema.Must(newCheck.SchemaString)
	newCheck.Validator.Register("", jsonschema.GetSchemaRegistry())
	return &newCheck, nil
}

// CheckPodSpec checks a pod spec against the schema
func (check SchemaCheck) CheckPodSpec(pod *corev1.PodSpec) (bool, []jsonschema.KeyError, error) {
	return check.CheckObject(pod)
}

// CheckPodTemplate checks a pod template against the schema
func (check SchemaCheck) CheckPodTemplate(podTemplate interface{}) (bool, []jsonschema.KeyError, error) {
	return check.CheckObject(podTemplate)
}

// CheckController checks a controler's spec against the schema
func (check SchemaCheck) CheckController(bytes []byte) (bool, []jsonschema.KeyError, error) {
	var data interface{}
	err := json.Unmarshal(bytes, &data)
	if err != nil {
		return false, nil, err
	}
	validationState := check.Validator.Validate(context.Background(), data)
	return validationState.IsValid(), *validationState.Errs, nil
}

// CheckContainer checks a container spec against the schema
func (check SchemaCheck) CheckContainer(container *corev1.Container) (bool, []jsonschema.KeyError, error) {
	return check.CheckObject(container)
}

// CheckObject checks arbitrary data against the schema
func (check SchemaCheck) CheckObject(obj interface{}) (bool, []jsonschema.KeyError, error) {
	validationState := check.Validator.Validate(context.Background(), obj)
	return validationState.IsValid(), *validationState.Errs, nil
}

// CheckAdditionalObjects looks for an object that passes the specified additional schema
func (check SchemaCheck) CheckAdditionalObjects(groupkind string, objects []interface{}) (bool, error) {
	val, ok := check.AdditionalValidators[groupkind]
	if !ok {
		return false, errors.New("No validator found for " + groupkind)
	}
	for _, obj := range objects {
		validationState := val.Validate(context.Background(), obj)
		if validationState.IsValid() {
			return true, nil
		}
	}
	return false, nil
}

// IsActionable decides if this check applies to a particular target
func (check SchemaCheck) IsActionable(target TargetKind, kind string, isInit bool) bool {
	if funk.Contains(HandledTargets, target) {
		if check.Target == TargetPodTemplate && target == TargetPodSpec {
			// A target=PodSpec and check.Target=PodTemplate is expected
			// because applyPodSchemaChecks() explicitly sets check.Target
			return true
		}
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
