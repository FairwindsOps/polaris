// Copyright 2019 ReactiveOps
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
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

var resourceConf1 = `---
resources:
  cpuRequests:
    error:
      below: 100m
      above: 1
    warning:
      below: 200m
      above: 800m
  memoryRequests:
    error:
      below: 100M
      above: 3G
    warning:
      below: 200M
      above: 2G
  cpuLimits:
    error:
      below: 100m
      above: 2
    warning:
      below: 300m
      above: 1800m
  memoryLimits:
    error:
      below: 200M
      above: 6G
    warning:
      below: 300M
      above: 4G
`

var resourceConf2 = `---
resources:
  cpuRequests:
    absent: warning
  memoryRequests:
    absent: warning
  cpuLimits:
    absent: error
  memoryLimits:
    absent: error
`

func TestValidateResourcesEmptyConfig(t *testing.T) {
	container := corev1.Container{
		Name: "Empty",
	}

	cv := ContainerValidation{
		Container: container,
	}

	expected := conf.Resources{}

	cv.validateResources(&expected)
	assert.Len(t, cv.Failures, 0)
}

func TestValidateResourcesEmptyContainer(t *testing.T) {
	container := corev1.Container{
		Name: "Empty",
	}

	expectedWarnings := []ResultMessage{
		{
			Type:    "warning",
			Message: "CPU Requests are not set",
		},
		{
			Type:    "warning",
			Message: "Memory Requests are not set",
		},
	}

	expectedFailures := []ResultMessage{
		{
			Type:    "failure",
			Message: "CPU Limits are not set",
		},
		{
			Type:    "failure",
			Message: "Memory Limits are not set",
		},
	}

	testValidateResources(t, &container, &resourceConf2, &expectedFailures, &expectedWarnings)
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

	expectedWarnings := []ResultMessage{
		{
			Type:    "warning",
			Message: "CPU Requests are too low",
		},
		{
			Type:    "warning",
			Message: "CPU Limits are too low",
		},
	}

	expectedFailures := []ResultMessage{
		{
			Type:    "failure",
			Message: "Memory Requests are too low",
		},
		{
			Type:    "failure",
			Message: "Memory Limits are too low",
		},
	}

	testValidateResources(t, &container, &resourceConf1, &expectedFailures, &expectedWarnings)
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

	testValidateResources(t, &container, &resourceConf1, &[]ResultMessage{}, &[]ResultMessage{})
}

func testValidateResources(t *testing.T, container *corev1.Container, resourceConf *string, expectedFailures *[]ResultMessage, expectedWarnings *[]ResultMessage) {
	cv := ContainerValidation{
		Container: *container,
	}

	parsedConf, err := conf.Parse([]byte(*resourceConf))
	assert.NoError(t, err, "Expected no error when parsing config")

	cv.validateResources(&parsedConf.Resources)
	assert.Len(t, cv.Warnings, len(*expectedWarnings))
	assert.ElementsMatch(t, cv.Warnings, *expectedWarnings)

	assert.Len(t, cv.Failures, len(*expectedFailures))
	assert.ElementsMatch(t, cv.Failures, *expectedFailures)
}

func TestValidateHealthChecks(t *testing.T) {

	// Test setup.
	p1 := conf.HealthChecks{}
	p2 := conf.HealthChecks{
		ReadinessProbeMissing: conf.SeverityIgnore,
		LivenessProbeMissing:  conf.SeverityIgnore,
	}
	p3 := conf.HealthChecks{
		ReadinessProbeMissing: conf.SeverityError,
		LivenessProbeMissing:  conf.SeverityWarning,
	}

	probe := corev1.Probe{}
	cv1 := ContainerValidation{Container: corev1.Container{Name: ""}}
	cv2 := ContainerValidation{Container: corev1.Container{Name: "", LivenessProbe: &probe, ReadinessProbe: &probe}}

	l := ResultMessage{Type: "warning", Message: "Liveness probe needs to be configured"}
	r := ResultMessage{Type: "failure", Message: "Readiness probe needs to be configured"}
	f1 := []ResultMessage{}
	f2 := []ResultMessage{r}
	w1 := []ResultMessage{l}

	var testCases = []struct {
		name     string
		probes   conf.HealthChecks
		cv       ContainerValidation
		failures *[]ResultMessage
		warnings *[]ResultMessage
	}{
		{name: "probes not configured", probes: p1, cv: cv1, failures: &f1},
		{name: "probes not required", probes: p2, cv: cv1, failures: &f1},
		{name: "probes required & configured", probes: p3, cv: cv2, failures: &f1},
		{name: "probes required & not configured", probes: p3, cv: cv1, failures: &f2, warnings: &w1},
		{name: "probes configured, but not required", probes: p2, cv: cv2, failures: &f1},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			tt.cv.validateHealthChecks(&tt.probes)

			if tt.warnings != nil {
				assert.Len(t, tt.cv.Warnings, len(*tt.warnings))
				assert.ElementsMatch(t, tt.cv.Warnings, *tt.warnings)
			}

			assert.Len(t, tt.cv.Failures, len(*tt.failures))
			assert.ElementsMatch(t, tt.cv.Failures, *tt.failures)
		})
	}
}

func TestValidateImage(t *testing.T) {

	// Test setup.
	i1 := conf.Images{}
	i2 := conf.Images{TagNotSpecified: conf.SeverityIgnore}
	i3 := conf.Images{TagNotSpecified: conf.SeverityError}

	cv1 := ContainerValidation{Container: corev1.Container{Name: ""}}
	cv2 := ContainerValidation{Container: corev1.Container{Name: "", Image: "test:tag"}}
	cv3 := ContainerValidation{Container: corev1.Container{Name: "", Image: "test:latest"}}
	cv4 := ContainerValidation{Container: corev1.Container{Name: "", Image: "test"}}

	f := ResultMessage{Message: "Image tag should be specified", Type: "failure"}
	f1 := []ResultMessage{}
	f2 := []ResultMessage{f}

	var testCases = []struct {
		name     string
		image    conf.Images
		cv       ContainerValidation
		expected []ResultMessage
	}{
		{name: "image not configured", image: i1, cv: cv1, expected: f1},
		{name: "image not required", image: i2, cv: cv1, expected: f1},
		{name: "image tag required and configured", image: i3, cv: cv2, expected: f1},
		{name: "image tag required, but not configured", image: i3, cv: cv1, expected: f2},
		{name: "image tag required, but is latest", image: i3, cv: cv3, expected: f2},
		{name: "image tag required, but is empty", image: i3, cv: cv4, expected: f2},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			tt.cv.validateImage(&tt.image)
			assert.Len(t, tt.cv.Failures, len(tt.expected))
			assert.ElementsMatch(t, tt.cv.Failures, tt.expected)
		})
	}
}
