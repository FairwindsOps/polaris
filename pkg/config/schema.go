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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/qri-io/jsonschema"
	"github.com/thoas/go-funk"
	"gomodules.xyz/jsonpatch/v2"
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

// MutationComment is the comments added to a mutated file
type MutationComment struct {
	Find    string `yaml:"find" json:"find"`
	Comment string `yaml:"comment" json:"comment"`
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
	Validator               jsonschema.RootSchema             `yaml:"-" json:"-"`
	AdditionalSchemas       map[string]map[string]interface{} `yaml:"additionalSchemas" json:"additionalSchemas"`
	AdditionalSchemaStrings map[string]string                 `yaml:"additionalSchemaStrings" json:"additionalSchemaStrings"`
	AdditionalValidators    map[string]jsonschema.RootSchema  `yaml:"-" json:"-"`
	Mutations               []jsonpatch.Operation             `yaml:"mutations" json:"mutations"`
	Comments                []MutationComment                 `yaml:"comments" json:"comments"`
}

type resourceMinimum string
type resourceMaximum string

func unmarshalYAMLOrJSON(raw []byte, dest interface{}) error {
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
	err := unmarshalYAMLOrJSON(rawBytes, &check)
	if err != nil {
		return check, err
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
		tmpl := template.New(newCheck.ID)
		tmpl, err := tmpl.Parse(tmplString)
		if err != nil {
			return nil, err
		}
		w := bytes.Buffer{}
		err = tmpl.Execute(&w, res)
		if err != nil {
			return nil, err
		}

		if kind == "" {
			newCheck.SchemaString = w.String()
		} else {
			newCheck.AdditionalSchemaStrings[kind] = w.String()
		}
	}

	newCheck.AdditionalValidators = map[string]jsonschema.RootSchema{}
	for kind, schemaStr := range newCheck.AdditionalSchemaStrings {
		val := jsonschema.RootSchema{}
		err := unmarshalYAMLOrJSON([]byte(schemaStr), &val)
		if err != nil {
			return nil, err
		}
		newCheck.AdditionalValidators[kind] = val
	}
	err := unmarshalYAMLOrJSON([]byte(newCheck.SchemaString), &newCheck.Validator)
	if err != nil {
		return nil, err
	}
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

// CheckAdditionalObjects looks for an object that passes the specified additional schema
func (check SchemaCheck) CheckAdditionalObjects(groupkind string, objects []interface{}) (bool, error) {
	val, ok := check.AdditionalValidators[groupkind]
	if !ok {
		return false, errors.New("No validator found for " + groupkind)
	}
	for _, obj := range objects {
		bytes, err := json.Marshal(obj)
		if err != nil {
			return false, err
		}
		errs, err := val.ValidateBytes(bytes)
		if err != nil {
			return false, err
		}
		if len(errs) == 0 {
			return true, nil
		}
	}
	return false, nil
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
