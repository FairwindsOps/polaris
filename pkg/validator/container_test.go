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
  cpuRequestRanges:
    error:
      below: 100m
      above: 1
    warning:
      below: 200m
      above: 800m
  memoryRequestRanges:
    error:
      below: 100M
      above: 3G
    warning:
      below: 200M
      above: 2G
  cpuLimitRanges:
    error:
      below: 100m
      above: 2
    warning:
      below: 300m
      above: 1800m
  memoryLimitRanges:
    error:
      below: 200M
      above: 6G
    warning:
      below: 300M
      above: 4G
`

var resourceConf2 = `---
resources:
  cpuRequestsMissing: warning
  memoryRequestsMissing: warning
  cpuLimitsMissing: error
  memoryLimitsMissing: error
`

func TestValidateResourcesEmptyConfig(t *testing.T) {
	container := corev1.Container{
		Name: "Empty",
	}

	cv := ContainerValidation{
		Container:          &container,
		ResourceValidation: &ResourceValidation{},
	}

	expected := conf.Resources{}

	cv.validateResources(&expected)
	assert.Len(t, cv.Errors, 0)
}

func TestValidateResourcesEmptyContainer(t *testing.T) {
	container := corev1.Container{
		Name: "Empty",
	}

	expectedWarnings := []*ResultMessage{
		{
			Type:     "warning",
			Message:  "CPU requests should be set",
			Category: "Resources",
		},
		{
			Type:     "warning",
			Message:  "Memory requests should be set",
			Category: "Resources",
		},
	}

	expectedErrors := []*ResultMessage{
		{
			Type:     "error",
			Message:  "CPU limits should be set",
			Category: "Resources",
		},
		{
			Type:     "error",
			Message:  "Memory limits should be set",
			Category: "Resources",
		},
	}

	testValidateResources(t, &container, &resourceConf2, &expectedErrors, &expectedWarnings)
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

	expectedWarnings := []*ResultMessage{
		{
			Type:     "warning",
			Message:  "CPU requests should be higher than 200m",
			Category: "Resources",
		},
		{
			Type:     "warning",
			Message:  "CPU limits should be higher than 300m",
			Category: "Resources",
		},
	}

	expectedErrors := []*ResultMessage{
		{
			Type:     "error",
			Message:  "Memory requests should be higher than 100M",
			Category: "Resources",
		},
		{
			Type:     "error",
			Message:  "Memory limits should be higher than 200M",
			Category: "Resources",
		},
	}

	testValidateResources(t, &container, &resourceConf1, &expectedErrors, &expectedWarnings)
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

	testValidateResources(t, &container, &resourceConf1, &[]*ResultMessage{}, &[]*ResultMessage{})
}

func testValidateResources(t *testing.T, container *corev1.Container, resourceConf *string, expectedErrors *[]*ResultMessage, expectedWarnings *[]*ResultMessage) {
	cv := ContainerValidation{
		Container:          container,
		ResourceValidation: &ResourceValidation{},
	}

	parsedConf, err := conf.Parse([]byte(*resourceConf))
	assert.NoError(t, err, "Expected no error when parsing config")

	cv.validateResources(&parsedConf.Resources)
	assert.Len(t, cv.Warnings, len(*expectedWarnings))
	assert.ElementsMatch(t, cv.Warnings, *expectedWarnings)

	assert.Len(t, cv.Errors, len(*expectedErrors))
	assert.ElementsMatch(t, cv.Errors, *expectedErrors)
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
	cv1 := ContainerValidation{
		Container:          &corev1.Container{Name: ""},
		ResourceValidation: &ResourceValidation{},
	}
	cv2 := ContainerValidation{
		Container: &corev1.Container{
			Name:           "",
			LivenessProbe:  &probe,
			ReadinessProbe: &probe,
		},
		ResourceValidation: &ResourceValidation{},
	}

	l := &ResultMessage{Type: "warning", Message: "Liveness probe should be configured", Category: "Health Checks"}
	r := &ResultMessage{Type: "error", Message: "Readiness probe should be configured", Category: "Health Checks"}
	f1 := []*ResultMessage{}
	f2 := []*ResultMessage{r}
	w1 := []*ResultMessage{l}

	var testCases = []struct {
		name     string
		probes   conf.HealthChecks
		cv       ContainerValidation
		errors   *[]*ResultMessage
		warnings *[]*ResultMessage
	}{
		{name: "probes not configured", probes: p1, cv: cv1, errors: &f1},
		{name: "probes not required", probes: p2, cv: cv1, errors: &f1},
		{name: "probes required & configured", probes: p3, cv: cv2, errors: &f1},
		{name: "probes required & not configured", probes: p3, cv: cv1, errors: &f2, warnings: &w1},
		{name: "probes configured, but not required", probes: p2, cv: cv2, errors: &f1},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			tt.cv.validateHealthChecks(&tt.probes)

			if tt.warnings != nil {
				assert.Len(t, tt.cv.Warnings, len(*tt.warnings))
				assert.ElementsMatch(t, tt.cv.Warnings, *tt.warnings)
			}

			assert.Len(t, tt.cv.Errors, len(*tt.errors))
			assert.ElementsMatch(t, tt.cv.Errors, *tt.errors)
		})
	}
}

func TestValidateImage(t *testing.T) {

	// Test setup.
	i1 := conf.Images{}
	i2 := conf.Images{TagNotSpecified: conf.SeverityIgnore}
	i3 := conf.Images{TagNotSpecified: conf.SeverityError}

	cv1 := ContainerValidation{
		Container:          &corev1.Container{Name: ""},
		ResourceValidation: &ResourceValidation{},
	}

	cv2 := ContainerValidation{
		Container:          &corev1.Container{Name: "", Image: "test:tag"},
		ResourceValidation: &ResourceValidation{},
	}

	cv3 := ContainerValidation{
		Container:          &corev1.Container{Name: "", Image: "test:latest"},
		ResourceValidation: &ResourceValidation{},
	}

	cv4 := ContainerValidation{
		Container:          &corev1.Container{Name: "", Image: "test"},
		ResourceValidation: &ResourceValidation{},
	}

	f := &ResultMessage{Message: "Image tag should be specified", Type: "error", Category: "Images"}
	f1 := []*ResultMessage{}
	f2 := []*ResultMessage{f}

	var testCases = []struct {
		name     string
		image    conf.Images
		cv       ContainerValidation
		expected []*ResultMessage
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
			assert.Len(t, tt.cv.Errors, len(tt.expected))
			assert.ElementsMatch(t, tt.cv.Errors, tt.expected)
		})
	}
}

func TestValidateNetworking(t *testing.T) {
	// Test setup.
	emptyConf := conf.Networking{}
	standardConf := conf.Networking{
		HostPortSet: conf.SeverityWarning,
	}
	strongConf := conf.Networking{
		HostPortSet: conf.SeverityError,
	}

	emptyCV := ContainerValidation{
		Container:          &corev1.Container{Name: ""},
		ResourceValidation: &ResourceValidation{},
	}

	badCV := ContainerValidation{
		Container: &corev1.Container{
			Ports: []corev1.ContainerPort{{
				ContainerPort: 3000,
				HostPort:      443,
			}},
		},
		ResourceValidation: &ResourceValidation{},
	}

	goodCV := ContainerValidation{
		Container: &corev1.Container{
			Ports: []corev1.ContainerPort{{
				ContainerPort: 3000,
			}},
		},
		ResourceValidation: &ResourceValidation{},
	}

	var testCases = []struct {
		name             string
		networkConf      conf.Networking
		cv               ContainerValidation
		expectedMessages []*ResultMessage
	}{
		{
			name:             "empty ports + empty validation config",
			networkConf:      emptyConf,
			cv:               emptyCV,
			expectedMessages: []*ResultMessage{},
		},
		{
			name:        "empty ports + standard validation config",
			networkConf: standardConf,
			cv:          emptyCV,
			expectedMessages: []*ResultMessage{{
				Message:  "Host port is not configured",
				Type:     "success",
				Category: "Networking",
			}},
		},
		{
			name:        "empty ports + strong validation config",
			networkConf: standardConf,
			cv:          emptyCV,
			expectedMessages: []*ResultMessage{{
				Message:  "Host port is not configured",
				Type:     "success",
				Category: "Networking",
			}},
		},
		{
			name:             "host ports + empty validation config",
			networkConf:      emptyConf,
			cv:               badCV,
			expectedMessages: []*ResultMessage{},
		},
		{
			name:        "host ports + standard validation config",
			networkConf: standardConf,
			cv:          badCV,
			expectedMessages: []*ResultMessage{{
				Message:  "Host port should not be configured",
				Type:     "warning",
				Category: "Networking",
			}},
		},
		{
			name:        "no host ports + standard validation config",
			networkConf: standardConf,
			cv:          goodCV,
			expectedMessages: []*ResultMessage{{
				Message:  "Host port is not configured",
				Type:     "success",
				Category: "Networking",
			}},
		},
		{
			name:        "host ports + strong validation config",
			networkConf: strongConf,
			cv:          badCV,
			expectedMessages: []*ResultMessage{{
				Message:  "Host port should not be configured",
				Type:     "error",
				Category: "Networking",
			}},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			tt.cv = resetCV(tt.cv)
			tt.cv.validateNetworking(&tt.networkConf)
			assert.Len(t, tt.cv.messages(), len(tt.expectedMessages))
			assert.ElementsMatch(t, tt.cv.messages(), tt.expectedMessages)
		})
	}
}

func TestValidateSecurity(t *testing.T) {
	trueVar := true
	falseVar := false

	// Test setup.
	emptyConf := conf.Security{}
	standardConf := conf.Security{
		RunAsRootAllowed:           conf.SeverityWarning,
		RunAsPrivileged:            conf.SeverityError,
		NotReadOnlyRootFileSystem:  conf.SeverityWarning,
		PrivilegeEscalationAllowed: conf.SeverityError,
		Capabilities: conf.SecurityCapabilities{
			Error: conf.SecurityCapabilityLists{
				IfAnyAdded: []corev1.Capability{"ALL", "SYS_ADMIN", "NET_ADMIN"},
			},
			Warning: conf.SecurityCapabilityLists{
				IfAnyAddedBeyond: []corev1.Capability{"NONE"},
			},
		},
	}
	strongConf := conf.Security{
		RunAsRootAllowed:           conf.SeverityError,
		RunAsPrivileged:            conf.SeverityError,
		NotReadOnlyRootFileSystem:  conf.SeverityError,
		PrivilegeEscalationAllowed: conf.SeverityError,
		Capabilities: conf.SecurityCapabilities{
			Error: conf.SecurityCapabilityLists{
				IfAnyAdded:      []corev1.Capability{"ALL", "SYS_ADMIN", "NET_ADMIN"},
				IfAnyNotDropped: []corev1.Capability{"NET_BIND_SERVICE", "DAC_OVERRIDE", "SYS_CHROOT"},
			},
			Warning: conf.SecurityCapabilityLists{
				IfAnyAddedBeyond: []corev1.Capability{"NONE"},
			},
		},
	}

	emptyCV := ContainerValidation{
		Container:          &corev1.Container{Name: ""},
		ResourceValidation: &ResourceValidation{},
	}

	badCV := ContainerValidation{
		Container: &corev1.Container{Name: "", SecurityContext: &corev1.SecurityContext{
			RunAsNonRoot:             &falseVar,
			ReadOnlyRootFilesystem:   &falseVar,
			Privileged:               &trueVar,
			AllowPrivilegeEscalation: &trueVar,
			Capabilities: &corev1.Capabilities{
				Add: []corev1.Capability{"AUDIT_CONTROL", "SYS_ADMIN", "NET_ADMIN"},
			},
		}},
		ResourceValidation: &ResourceValidation{},
	}

	goodCV := ContainerValidation{
		Container: &corev1.Container{Name: "", SecurityContext: &corev1.SecurityContext{
			RunAsNonRoot:             &trueVar,
			ReadOnlyRootFilesystem:   &trueVar,
			Privileged:               &falseVar,
			AllowPrivilegeEscalation: &falseVar,
			Capabilities: &corev1.Capabilities{
				Drop: []corev1.Capability{"NET_BIND_SERVICE", "FOWNER"},
			},
		}},
		ResourceValidation: &ResourceValidation{},
	}

	strongCV := ContainerValidation{
		Container: &corev1.Container{Name: "", SecurityContext: &corev1.SecurityContext{
			RunAsNonRoot:             &trueVar,
			ReadOnlyRootFilesystem:   &trueVar,
			Privileged:               &falseVar,
			AllowPrivilegeEscalation: &falseVar,
			Capabilities: &corev1.Capabilities{
				Drop: []corev1.Capability{"ALL"},
			},
		}},
		ResourceValidation: &ResourceValidation{},
	}

	var testCases = []struct {
		name             string
		securityConf     conf.Security
		cv               ContainerValidation
		expectedMessages []*ResultMessage
	}{
		{
			name:             "empty security context + empty validation config",
			securityConf:     emptyConf,
			cv:               emptyCV,
			expectedMessages: []*ResultMessage{},
		},
		{
			name:         "empty security context + standard validation config",
			securityConf: standardConf,
			cv:           emptyCV,
			expectedMessages: []*ResultMessage{{
				Message:  "Should not be running as root",
				Type:     "warning",
				Category: "Security",
			}, {
				Message:  "Filesystem should be read only",
				Type:     "warning",
				Category: "Security",
			}, {
				Message:  "Not running as privileged",
				Type:     "success",
				Category: "Security",
			}, {
				Message:  "Privilege escalation not allowed",
				Type:     "success",
				Category: "Security",
			}, {
				Message:  "Security capabilities are within the configured limits",
				Type:     "success",
				Category: "Security",
			}},
		},
		{
			name:         "bad security context + standard validation config",
			securityConf: standardConf,
			cv:           badCV,
			expectedMessages: []*ResultMessage{{
				Message:  "The following security capabilities should not be added: SYS_ADMIN, NET_ADMIN",
				Type:     "error",
				Category: "Security",
			}, {
				Message:  "Privilege escalation should not be allowed",
				Type:     "error",
				Category: "Security",
			}, {
				Message:  "Should not be running as privileged",
				Type:     "error",
				Category: "Security",
			}, {
				Message:  "The following security capabilities should not be added: AUDIT_CONTROL, SYS_ADMIN, NET_ADMIN",
				Type:     "warning",
				Category: "Security",
			}, {
				Message:  "Should not be running as root",
				Type:     "warning",
				Category: "Security",
			}, {
				Message:  "Filesystem should be read only",
				Type:     "warning",
				Category: "Security",
			}},
		},
		{
			name:         "good security context + standard validation config",
			securityConf: standardConf,
			cv:           goodCV,
			expectedMessages: []*ResultMessage{{
				Message:  "Not running as root",
				Type:     "success",
				Category: "Security",
			}, {
				Message:  "Filesystem is read only",
				Type:     "success",
				Category: "Security",
			}, {
				Message:  "Not running as privileged",
				Type:     "success",
				Category: "Security",
			}, {
				Message:  "Privilege escalation not allowed",
				Type:     "success",
				Category: "Security",
			}, {
				Message:  "Security capabilities are within the configured limits",
				Type:     "success",
				Category: "Security",
			}},
		},
		{
			name:         "good security context + strong validation config",
			securityConf: strongConf,
			cv:           goodCV,
			expectedMessages: []*ResultMessage{{
				Message:  "The following security capabilities should be dropped: DAC_OVERRIDE, SYS_CHROOT",
				Type:     "error",
				Category: "Security",
			}, {
				Message:  "Not running as root",
				Type:     "success",
				Category: "Security",
			}, {
				Message:  "Filesystem is read only",
				Type:     "success",
				Category: "Security",
			}, {
				Message:  "Not running as privileged",
				Type:     "success",
				Category: "Security",
			}, {
				Message:  "Privilege escalation not allowed",
				Type:     "success",
				Category: "Security",
			}},
		},
		{
			name:         "strong security context + strong validation config",
			securityConf: strongConf,
			cv:           strongCV,
			expectedMessages: []*ResultMessage{{
				Message:  "Not running as root",
				Type:     "success",
				Category: "Security",
			}, {
				Message:  "Filesystem is read only",
				Type:     "success",
				Category: "Security",
			}, {
				Message:  "Not running as privileged",
				Type:     "success",
				Category: "Security",
			}, {
				Message:  "Privilege escalation not allowed",
				Type:     "success",
				Category: "Security",
			}, {
				Message:  "Security capabilities are within the configured limits",
				Type:     "success",
				Category: "Security",
			}},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			tt.cv = resetCV(tt.cv)
			tt.cv.validateSecurity(&tt.securityConf)
			assert.Len(t, tt.cv.messages(), len(tt.expectedMessages))
			assert.ElementsMatch(t, tt.cv.messages(), tt.expectedMessages)
		})
	}
}

func resetCV(cv ContainerValidation) ContainerValidation {
	cv.Errors = []*ResultMessage{}
	cv.Successes = []*ResultMessage{}
	cv.Warnings = []*ResultMessage{}
	return cv
}
