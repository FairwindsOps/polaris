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

package validator

import (
	"testing"

	conf "github.com/fairwindsops/polaris/pkg/config"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

var customCheckExemptions = `
checks:
  foo: danger
customChecks:
  foo:
    successMessage: success!
    failureMessage: fail!
    target: Container
    category: Security
    schema:
      properties:
        image:
          pattern: ^quay.io
exemptions:
- controllerNames:
  - exempt
  rules:
  - foo
`

var resourceConfRanges = `
checks:
  memoryRequestsRange: danger
  memoryLimitsRange: warning
customChecks:
  memoryLimitsRange:
    containers:
      exclude:
      - initContainer
    successMessage: Memory limits are within the required range
    failureMessage: Memory limits should be within the required range
    category: Efficiency
    target: Container
    schema:
      '$schema': http://json-schema.org/draft-07/schema
      type: object
      required:
      - resources
      properties:
        resources:
          type: object
          required:
          - limits
          properties:
            limits:
              type: object
              required:
              - memory
              properties:
                memory:
                  type: string
                  resourceMinimum: 200M
                  resourceMaximum: 6G
  memoryRequestsRange:
    successMessage: Memory requests are within the required range
    failureMessage: Memory requests should be within the required range
    category: Efficiency
    target: Container
    containers:
      exclude:
      - initContainer
    schema:
      '$schema': http://json-schema.org/draft-07/schema
      type: object
      required:
      - resources
      properties:
        resources:
          type: object
          required:
          - requests
          properties:
            requests:
              required:
              - memory
              properties:
                memory:
                  type: string
                  resourceMinimum: 200M
                  resourceMaximum: 3G
`

func TestValidateResourcesPartiallyValid(t *testing.T) {
	cpuRequest, err := resource.ParseQuantity("100m")
	assert.NoError(t, err, "Error parsing quantity")

	cpuLimit, err := resource.ParseQuantity("200m")
	assert.NoError(t, err, "Error parsing quantity")

	container := corev1.Container{
		Name: "Empty",
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				"cpu": cpuRequest,
			},
			Limits: corev1.ResourceList{
				"cpu": cpuLimit,
			},
		},
	}

	expectedWarnings := []ResultMessage{
		{
			ID:       "memoryLimitsRange",
			Success:  false,
			Severity: "warning",
			Message:  "Memory limits should be within the required range",
			Category: "Efficiency",
		},
	}

	expectedDangers := []ResultMessage{
		{
			ID:       "memoryRequestsRange",
			Success:  false,
			Severity: "danger",
			Message:  "Memory requests should be within the required range",
			Category: "Efficiency",
		},
	}

	expectedSuccesses := []ResultMessage{}

	testValidate(t, &container, &resourceConfRanges, "foo", expectedDangers, expectedWarnings, expectedSuccesses)
}

func TestValidateResourcesInit(t *testing.T) {
	emptyContainer := &corev1.Container{}
	controller := getEmptyWorkload(t, "")

	parsedConf, err := conf.Parse([]byte(resourceConfRanges))
	assert.NoError(t, err, "Expected no error when parsing config")

	var results ResultSet
	results, err = applyContainerSchemaChecks(&parsedConf, nil, controller, emptyContainer, false)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, uint(1), results.GetSummary().Dangers)
	assert.Equal(t, uint(1), results.GetSummary().Warnings)

	results, err = applyContainerSchemaChecks(&parsedConf, nil, controller, emptyContainer, true)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, uint(0), results.GetSummary().Dangers)
	assert.Equal(t, uint(0), results.GetSummary().Warnings)
}

func TestValidateResourcesFullyValid(t *testing.T) {
	cpuRequest, err := resource.ParseQuantity("300m")
	assert.NoError(t, err, "Error parsing quantity")

	cpuLimit, err := resource.ParseQuantity("400m")
	assert.NoError(t, err, "Error parsing quantity")

	memoryRequest, err := resource.ParseQuantity("400Mi")
	assert.NoError(t, err, "Error parsing quantity")

	memoryLimit, err := resource.ParseQuantity("500Mi")
	assert.NoError(t, err, "Error parsing quantity")

	container := corev1.Container{
		Name: "Empty",
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				"cpu":    cpuRequest,
				"memory": memoryRequest,
			},
			Limits: corev1.ResourceList{
				"cpu":    cpuLimit,
				"memory": memoryLimit,
			},
		},
	}

	expectedSuccesses := []ResultMessage{
		{
			ID:       "memoryRequestsRange",
			Success:  true,
			Severity: "danger",
			Message:  "Memory requests are within the required range",
			Category: "Efficiency",
		},
		{
			ID:       "memoryLimitsRange",
			Success:  true,
			Severity: "warning",
			Message:  "Memory limits are within the required range",
			Category: "Efficiency",
		},
	}

	testValidate(t, &container, &resourceConfRanges, "foo", []ResultMessage{}, []ResultMessage{}, expectedSuccesses)

	expectedSuccesses = []ResultMessage{
		{
			ID:       "cpuRequestsMissing",
			Success:  true,
			Severity: "warning",
			Message:  "CPU requests are set",
			Category: "Efficiency",
		},
		{
			ID:       "memoryRequestsMissing",
			Success:  true,
			Severity: "warning",
			Message:  "Memory requests are set",
			Category: "Efficiency",
		},
		{
			ID:       "cpuLimitsMissing",
			Success:  true,
			Severity: "danger",
			Message:  "CPU limits are set",
			Category: "Efficiency",
		},
		{
			ID:       "memoryLimitsMissing",
			Success:  true,
			Severity: "danger",
			Message:  "Memory limits are set",
			Category: "Efficiency",
		},
	}

	testValidate(t, &container, &resourceConfMinimal, "foo", []ResultMessage{}, []ResultMessage{}, expectedSuccesses)
}

func TestValidateCustomCheckExemptions(t *testing.T) {
	container := corev1.Container{
		Name:  "example",
		Image: "hub.docker.com/foo",
	}

	expectedWarnings := []ResultMessage{}
	expectedDangers := []ResultMessage{}
	expectedSuccesses := []ResultMessage{}
	testValidate(t, &container, &customCheckExemptions, "exempt", expectedDangers, expectedWarnings, expectedSuccesses)

	expectedDangers = []ResultMessage{
		{
			ID:       "foo",
			Success:  false,
			Severity: "danger",
			Message:  "fail!",
			Category: "Security",
		},
	}
	testValidate(t, &container, &customCheckExemptions, "notexempt", expectedDangers, expectedWarnings, expectedSuccesses)
}
