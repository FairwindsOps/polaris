// Copyright 2018 ReactiveOps
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package validator

import (
	"testing"

	conf "github.com/reactiveops/fairwinds/pkg/config"
	types "github.com/reactiveops/fairwinds/pkg/types"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

var resourceConf1 = `---
resources:
  requests:
    cpu:
      min: 100m
      max: 1
    memory:
      min: 100Mi
      max: 3Gi
  limits:
    cpu:
      min: 150m
      max: 2
    memory:
      min: 150Mi
      max: 4Gi
`

func TestValidateResourcesEmptyConfig(t *testing.T) {
	container := corev1.Container{
		Name: "Empty",
	}

	cv := ContainerValidation{
		Container: container,
	}

	expected := conf.RequestsAndLimits{}

	cv.validateResources(expected)
	assert.Len(t, cv.Failures, 0)
}

func TestValidateResourcesEmptyContainer(t *testing.T) {
	container := corev1.Container{
		Name: "Empty",
	}

	expectedFailures := []types.Failure{
		{
			Name:     "requests.cpu",
			Expected: "100m",
			Actual:   "0",
		},
		{
			Name:     "requests.memory",
			Expected: "100Mi",
			Actual:   "0",
		},
		{
			Name:     "limits.cpu",
			Expected: "150m",
			Actual:   "0",
		},
		{
			Name:     "limits.memory",
			Expected: "150Mi",
			Actual:   "0",
		},
	}

	testValidateResources(t, &container, &resourceConf1, &expectedFailures)
}

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

	expectedFailures := []types.Failure{
		{
			Name:     "requests.memory",
			Expected: "100Mi",
			Actual:   "0",
		},
		{
			Name:     "limits.memory",
			Expected: "150Mi",
			Actual:   "0",
		},
	}

	testValidateResources(t, &container, &resourceConf1, &expectedFailures)
}

func TestValidateResourcesFullyValid(t *testing.T) {
	cpuRequest, err := resource.ParseQuantity("100m")
	assert.NoError(t, err, "Error parsing quantity")

	cpuLimit, err := resource.ParseQuantity("200m")
	assert.NoError(t, err, "Error parsing quantity")

	memoryRequest, err := resource.ParseQuantity("100Mi")
	assert.NoError(t, err, "Error parsing quantity")

	memoryLimit, err := resource.ParseQuantity("200Mi")
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

	testValidateResources(t, &container, &resourceConf1, &[]types.Failure{})
}

func testValidateResources(t *testing.T, container *corev1.Container, resourceConf *string, expectedFailures *[]types.Failure) {
	cv := ContainerValidation{
		Container: *container,
	}

	parsedConf, err := conf.Parse([]byte(*resourceConf))
	assert.NoError(t, err, "Expected no error when parsing config")

	cv.validateResources(parsedConf.Resources)
	assert.Len(t, cv.Failures, len(*expectedFailures))
	assert.ElementsMatch(t, cv.Failures, *expectedFailures)
}
