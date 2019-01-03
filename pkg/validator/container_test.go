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

func TestValidateHealthChecks(t *testing.T) {

	// Test setup.
	p1 := conf.Probes{}
	p2 := conf.Probes{
		Readiness: conf.ResourceRequire{Require: false},
		Liveness:  conf.ResourceRequire{Require: false},
	}
	p3 := conf.Probes{
		Readiness: conf.ResourceRequire{Require: true},
		Liveness:  conf.ResourceRequire{Require: true},
	}

	probe := corev1.Probe{}
	cv1 := ContainerValidation{Container: corev1.Container{Name: ""}}
	cv2 := ContainerValidation{Container: corev1.Container{Name: "", LivenessProbe: &probe, ReadinessProbe: &probe}}

	l := types.Failure{Name: "liveness", Expected: "probe needs to be configured", Actual: "nil"}
	r := types.Failure{Name: "readiness", Expected: "probe needs to be configured", Actual: "nil"}
	f1 := []types.Failure{}
	f2 := []types.Failure{r, l}

	var testCases = []struct {
		name     string
		probes   conf.Probes
		cv       ContainerValidation
		expected []types.Failure
	}{
		{name: "probes not configured", probes: p1, cv: cv1, expected: f1},
		{name: "probes not required", probes: p2, cv: cv1, expected: f1},
		{name: "probes required & configured", probes: p3, cv: cv2, expected: f1},
		{name: "probes required & not configured", probes: p3, cv: cv1, expected: f2},
		{name: "probes configured, but not required", probes: p2, cv: cv2, expected: f1},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			tt.cv.validateHealthChecks(tt.probes)
			assert.Len(t, tt.cv.Failures, len(tt.expected))
			assert.ElementsMatch(t, tt.cv.Failures, tt.expected)
		})
	}
}

func TestValidateImage(t *testing.T) {

	// Test setup.
	i1 := conf.Images{}
	i2 := conf.Images{TagRequired: false}
	i3 := conf.Images{TagRequired: true}

	cv1 := ContainerValidation{Container: corev1.Container{Name: ""}}
	cv2 := ContainerValidation{Container: corev1.Container{Name: "", Image: "test:tag"}}
	cv3 := ContainerValidation{Container: corev1.Container{Name: "", Image: "test:latest"}}
	cv4 := ContainerValidation{Container: corev1.Container{Name: "", Image: "test"}}

	f := types.Failure{Name: "Image Tag", Expected: "not latest", Actual: "latest"}
	f1 := []types.Failure{}
	f2 := []types.Failure{f}

	var testCases = []struct {
		name     string
		image    conf.Images
		cv       ContainerValidation
		expected []types.Failure
	}{
		{name: "image not configured", image: i1, cv: cv1, expected: f1},
		{name: "image not required	", image: i2, cv: cv1, expected: f1},
		{name: "image tag required and configured", image: i3, cv: cv2, expected: f1},
		{name: "image tag required, but not configured", image: i3, cv: cv1, expected: f2},
		{name: "image tag required, but is latest", image: i3, cv: cv3, expected: f2},
		{name: "image tag required, but is empty", image: i3, cv: cv4, expected: f2},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			tt.cv.validateImage(tt.image)
			assert.Len(t, tt.cv.Failures, len(tt.expected))
			assert.ElementsMatch(t, tt.cv.Failures, tt.expected)
		})
	}
}
