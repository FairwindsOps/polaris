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
	"k8s.io/apimachinery/pkg/api/resource"
	"testing"

	conf "github.com/reactiveops/fairwinds/pkg/config"
	types "github.com/reactiveops/fairwinds/pkg/types"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

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

	resourceConf := conf.RequestsAndLimits{
		Requests: conf.ResourceList{
			"cpu": conf.ResourceMinMax{
				Min: "100m",
				Max: "1",
			},
			"memory": conf.ResourceMinMax{
				Min: "100Mi",
				Max: "3Gi",
			},
		},
		Limits: conf.ResourceList{
			"cpu": conf.ResourceMinMax{
				Min: "150m",
				Max: "2",
			},
			"memory": conf.ResourceMinMax{
				Min: "150Mi",
				Max: "4Gi",
			},
		},
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

	testValidateResources(t, &container, &resourceConf, &expectedFailures)
}

func TestValidateResourcesPartiallyValid(t *testing.T) {
	cpuRequest := resource.Quantity{}
	cpuRequest.SetMilli(100)

	cpuLimit := resource.Quantity{}
	cpuLimit.SetMilli(200)

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

	resourceConf := conf.RequestsAndLimits{
		Requests: conf.ResourceList{
			"cpu": conf.ResourceMinMax{
				Min: "100m",
				Max: "1",
			},
			"memory": conf.ResourceMinMax{
				Min: "100Mi",
				Max: "3Gi",
			},
		},
		Limits: conf.ResourceList{
			"cpu": conf.ResourceMinMax{
				Min: "150m",
				Max: "2",
			},
			"memory": conf.ResourceMinMax{
				Min: "150Mi",
				Max: "4Gi",
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

	testValidateResources(t, &container, &resourceConf, &expectedFailures)
}

func TestValidateResourcesFullyValid(t *testing.T) {
	cpuRequest := resource.Quantity{}
	cpuRequest.SetMilli(100)

	cpuLimit := resource.Quantity{}
	cpuLimit.SetMilli(200)

	memoryRequest := resource.Quantity{}
	memoryRequest.SetScaled(105, resource.Mega)

	memoryLimit := resource.Quantity{}
	memoryLimit.SetScaled(200, resource.Mega)

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

	resourceConf := conf.RequestsAndLimits{
		Requests: conf.ResourceList{
			"cpu": conf.ResourceMinMax{
				Min: "100m",
				Max: "1",
			},
			"memory": conf.ResourceMinMax{
				Min: "100Mi",
				Max: "3Gi",
			},
		},
		Limits: conf.ResourceList{
			"cpu": conf.ResourceMinMax{
				Min: "150m",
				Max: "2",
			},
			"memory": conf.ResourceMinMax{
				Min: "150Mi",
				Max: "4Gi",
			},
		},
	}

	testValidateResources(t, &container, &resourceConf, &[]types.Failure{})
}

func testValidateResources(t *testing.T, container *corev1.Container, conf *conf.RequestsAndLimits, expectedFailures *[]types.Failure) {
	cv := ContainerValidation{
		Container: *container,
	}

	cv.validateResources(*conf)
	assert.Len(t, cv.Failures, len(*expectedFailures))
	assert.ElementsMatch(t, cv.Failures, *expectedFailures)
}
