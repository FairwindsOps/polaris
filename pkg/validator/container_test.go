// Copyright 2019 FairwindsOps Inc
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

	conf "github.com/fairwindsops/polaris/pkg/config"
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

var resourceConfExemptions = `---
resources:
  cpuRequestsMissing: warning
  memoryRequestsMissing: warning
  cpuLimitsMissing: error
  memoryLimitsMissing: error
exemptions:
  - rules:
    - cpuRequestsMissing
    - memoryRequestsMissing
    - cpuLimitsMissing
    - memoryLimitsMissing
    controllerNames:
    - foo
`

var resourceConfRangeExemptions = `---
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
exemptions:
  - rules:
    - cpuRequestRanges
    - memoryRequestRanges
    - cpuLimitRanges
    - memoryLimitRanges
    controllerNames:
    - foo
`

func TestValidateResourcesEmptyConfig(t *testing.T) {
	container := corev1.Container{
		Name: "Empty",
	}

	cv := ContainerValidation{
		Container:          &container,
		ResourceValidation: &ResourceValidation{},
	}

	cv.validateResources(&conf.Configuration{}, "")
	assert.Len(t, cv.Errors, 0)
}

func TestValidateResourcesEmptyContainer(t *testing.T) {
	container := corev1.Container{
		Name: "Empty",
	}

	expectedWarnings := []*ResultMessage{
		{
			ID:       "cpuRequestsMissing",
			Type:     "warning",
			Message:  "CPU requests should be set",
			Category: "Resources",
		},
		{
			ID:       "memoryRequestsMissing",
			Type:     "warning",
			Message:  "Memory requests should be set",
			Category: "Resources",
		},
	}

	expectedErrors := []*ResultMessage{
		{
			ID:       "cpuLimitsMissing",
			Type:     "error",
			Message:  "CPU limits should be set",
			Category: "Resources",
		},
		{
			ID:       "memoryLimitsMissing",
			Type:     "error",
			Message:  "Memory limits should be set",
			Category: "Resources",
		},
	}

	testValidateResources(t, &container, &resourceConf2, "foo", &expectedErrors, &expectedWarnings)
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
			ID:       "cpuRequestRanges",
			Type:     "warning",
			Message:  "CPU requests should be higher than 200m",
			Category: "Resources",
		},
		{
			ID:       "cpuLimitRanges",
			Type:     "warning",
			Message:  "CPU limits should be higher than 300m",
			Category: "Resources",
		},
	}

	expectedErrors := []*ResultMessage{
		{
			ID:       "memoryRequestRanges",
			Type:     "error",
			Message:  "Memory requests should be higher than 100M",
			Category: "Resources",
		},
		{
			ID:       "memoryLimitRanges",
			Type:     "error",
			Message:  "Memory limits should be higher than 200M",
			Category: "Resources",
		},
	}

	testValidateResources(t, &container, &resourceConf1, "foo", &expectedErrors, &expectedWarnings)
}

func TestValidateResourcesInit(t *testing.T) {
	cvEmpty := ContainerValidation{
		Container:          &corev1.Container{},
		ResourceValidation: &ResourceValidation{},
	}
	cvInit := ContainerValidation{
		Container:          &corev1.Container{},
		ResourceValidation: &ResourceValidation{},
		IsInitContainer:    true,
	}

	parsedConf, err := conf.Parse([]byte(resourceConf1))
	assert.NoError(t, err, "Expected no error when parsing config")

	cvEmpty.validateResources(&parsedConf, "")
	assert.Len(t, cvEmpty.Errors, 4)

	cvInit.validateResources(&parsedConf, "")
	assert.Len(t, cvInit.Errors, 0)
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

	testValidateResources(t, &container, &resourceConf1, "foo", &[]*ResultMessage{}, &[]*ResultMessage{})
}

func testValidateResources(t *testing.T, container *corev1.Container, resourceConf *string, controllerName string, expectedErrors *[]*ResultMessage, expectedWarnings *[]*ResultMessage) {
	cv := ContainerValidation{
		Container:          container,
		ResourceValidation: &ResourceValidation{},
	}

	parsedConf, err := conf.Parse([]byte(*resourceConf))
	assert.NoError(t, err, "Expected no error when parsing config")

	cv.validateResources(&parsedConf, controllerName)
	assert.Len(t, cv.Warnings, len(*expectedWarnings))
	assert.ElementsMatch(t, *expectedWarnings, cv.Warnings)

	assert.Len(t, cv.Errors, len(*expectedErrors))
	assert.ElementsMatch(t, *expectedErrors, cv.Errors)
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
	emptyCV := ContainerValidation{
		Container:          &corev1.Container{Name: ""},
		ResourceValidation: &ResourceValidation{},
	}
	emptyCVInit := ContainerValidation{
		Container:          &corev1.Container{Name: ""},
		ResourceValidation: &ResourceValidation{},
		IsInitContainer:    true,
	}
	goodCV := ContainerValidation{
		Container: &corev1.Container{
			Name:           "",
			LivenessProbe:  &probe,
			ReadinessProbe: &probe,
		},
		ResourceValidation: &ResourceValidation{},
	}

	l := &ResultMessage{ID: "livenessProbeMissing", Type: "warning", Message: "Liveness probe should be configured", Category: "Health Checks"}
	r := &ResultMessage{ID: "readinessProbeMissing", Type: "error", Message: "Readiness probe should be configured", Category: "Health Checks"}
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
		{name: "probes not configured", probes: p1, cv: emptyCV, errors: &f1},
		{name: "probes not required", probes: p2, cv: emptyCV, errors: &f1},
		{name: "probes required & configured", probes: p3, cv: goodCV, errors: &f1},
		{name: "probes required, not configured, but init", probes: p3, cv: emptyCVInit, errors: &f1},
		{name: "probes required & not configured", probes: p3, cv: emptyCV, errors: &f2, warnings: &w1},
		{name: "probes configured, but not required", probes: p2, cv: goodCV, errors: &f1},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			tt.cv.validateHealthChecks(&conf.Configuration{HealthChecks: tt.probes}, "")

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
	emptyConf := conf.Images{}
	standardConf := conf.Images{
		TagNotSpecified:     conf.SeverityError,
		PullPolicyNotAlways: conf.SeverityIgnore,
	}
	strongConf := conf.Images{
		TagNotSpecified:     conf.SeverityError,
		PullPolicyNotAlways: conf.SeverityError,
	}

	emptyCV := ContainerValidation{
		Container:          &corev1.Container{},
		ResourceValidation: &ResourceValidation{},
	}
	badCV := ContainerValidation{
		Container:          &corev1.Container{Image: "test"},
		ResourceValidation: &ResourceValidation{},
	}
	lessBadCV := ContainerValidation{
		Container:          &corev1.Container{Image: "test:latest", ImagePullPolicy: ""},
		ResourceValidation: &ResourceValidation{},
	}
	goodCV := ContainerValidation{
		Container:          &corev1.Container{Image: "test:0.1.0", ImagePullPolicy: "Always"},
		ResourceValidation: &ResourceValidation{},
	}

	var testCases = []struct {
		name     string
		image    conf.Images
		cv       ContainerValidation
		expected []*ResultMessage
	}{
		{
			name:     "emptyConf + emptyCV",
			image:    emptyConf,
			cv:       emptyCV,
			expected: []*ResultMessage{},
		},
		{
			name:  "standardConf + emptyCV",
			image: standardConf,
			cv:    emptyCV,
			expected: []*ResultMessage{{
				ID:       "tagNotSpecified",
				Message:  "Image tag should be specified",
				Type:     "error",
				Category: "Images",
			}},
		},
		{
			name:  "standardConf + badCV",
			image: standardConf,
			cv:    badCV,
			expected: []*ResultMessage{{
				ID:       "tagNotSpecified",
				Message:  "Image tag should be specified",
				Type:     "error",
				Category: "Images",
			}},
		},
		{
			name:  "standardConf + lessBadCV",
			image: standardConf,
			cv:    lessBadCV,
			expected: []*ResultMessage{{
				ID:       "tagNotSpecified",
				Message:  "Image tag should be specified",
				Type:     "error",
				Category: "Images",
			}},
		},
		{
			name:  "strongConf + badCV",
			image: strongConf,
			cv:    badCV,
			expected: []*ResultMessage{{
				ID:       "pullPolicyNotAlways",
				Message:  "Image pull policy should be \"Always\"",
				Type:     "error",
				Category: "Images",
			}, {
				ID:       "tagNotSpecified",
				Message:  "Image tag should be specified",
				Type:     "error",
				Category: "Images",
			}},
		},
		{
			name:     "strongConf + goodCV",
			image:    strongConf,
			cv:       goodCV,
			expected: []*ResultMessage{},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			tt.cv = resetCV(tt.cv)
			tt.cv.validateImage(&conf.Configuration{Images: tt.image}, "")
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
				ID:       "hostPortSet",
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
				ID:       "hostPortSet",
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
				ID:       "hostPortSet",
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
				ID:       "hostPortSet",
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
				ID:       "hostPortSet",
				Message:  "Host port should not be configured",
				Type:     "error",
				Category: "Networking",
			}},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			tt.cv = resetCV(tt.cv)
			tt.cv.validateNetworking(&conf.Configuration{Networking: tt.networkConf}, "")
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

	badCVWithGoodPodSpec := ContainerValidation{
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
		parentPodSpec: corev1.PodSpec{
			SecurityContext: &corev1.PodSecurityContext{
				RunAsNonRoot: &trueVar,
			},
		},
	}

	badCVWithBadPodSpec := ContainerValidation{
		Container: &corev1.Container{Name: "", SecurityContext: &corev1.SecurityContext{
			RunAsNonRoot:             nil, // this will use the default from the podspec
			ReadOnlyRootFilesystem:   &falseVar,
			Privileged:               &trueVar,
			AllowPrivilegeEscalation: &trueVar,
			Capabilities: &corev1.Capabilities{
				Add: []corev1.Capability{"AUDIT_CONTROL", "SYS_ADMIN", "NET_ADMIN"},
			},
		}},
		ResourceValidation: &ResourceValidation{},
		parentPodSpec: corev1.PodSpec{
			SecurityContext: &corev1.PodSecurityContext{
				RunAsNonRoot: &falseVar,
			},
		},
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

	strongCVWithPodSpecSecurityContext := ContainerValidation{
		Container: &corev1.Container{Name: "", SecurityContext: &corev1.SecurityContext{
			RunAsNonRoot:             nil, // not set but overridden via podSpec
			ReadOnlyRootFilesystem:   &trueVar,
			Privileged:               &falseVar,
			AllowPrivilegeEscalation: &falseVar,
			Capabilities: &corev1.Capabilities{
				Drop: []corev1.Capability{"ALL"},
			},
		}},
		ResourceValidation: &ResourceValidation{},
		parentPodSpec: corev1.PodSpec{
			SecurityContext: &corev1.PodSecurityContext{
				RunAsNonRoot: &trueVar,
			},
		},
	}

	strongCVWithBadPodSpecSecurityContext := ContainerValidation{
		Container: &corev1.Container{Name: "", SecurityContext: &corev1.SecurityContext{
			RunAsNonRoot:             &trueVar, // will override the bad setting in PodSpec
			ReadOnlyRootFilesystem:   &trueVar,
			Privileged:               &falseVar,
			AllowPrivilegeEscalation: &falseVar,
			Capabilities: &corev1.Capabilities{
				Drop: []corev1.Capability{"ALL"},
			},
		}},
		ResourceValidation: &ResourceValidation{},
		parentPodSpec: corev1.PodSpec{
			SecurityContext: &corev1.PodSecurityContext{
				RunAsNonRoot: &falseVar, // is overridden at container level with RunAsNonRoot:true
			},
		},
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
				ID:       "runAsRootAllowed",
				Message:  "Should not be allowed to run as root",
				Type:     "warning",
				Category: "Security",
			}, {
				ID:       "notReadOnlyRootFileSystem",
				Message:  "Filesystem should be read only",
				Type:     "warning",
				Category: "Security",
			}, {
				ID:       "runAsPrivileged",
				Message:  "Not running as privileged",
				Type:     "success",
				Category: "Security",
			}, {
				ID:       "privilegeEscalationAllowed",
				Message:  "Privilege escalation not allowed",
				Type:     "success",
				Category: "Security",
			}, {
				ID:       "capabilitiesAdded",
				Message:  "Disallowed security capabilities have not been added",
				Type:     "success",
				Category: "Security",
			}},
		},
		{
			name:         "bad security context + standard validation config",
			securityConf: standardConf,
			cv:           badCV,
			expectedMessages: []*ResultMessage{{
				ID:       "capabilitiesAdded",
				Message:  "The following security capabilities should not be added: SYS_ADMIN, NET_ADMIN",
				Type:     "error",
				Category: "Security",
			}, {
				ID:       "privilegeEscalationAllowed",
				Message:  "Privilege escalation should not be allowed",
				Type:     "error",
				Category: "Security",
			}, {
				ID:       "runAsPrivileged",
				Message:  "Should not be running as privileged",
				Type:     "error",
				Category: "Security",
			}, {
				ID:       "capabilitiesAdded",
				Message:  "The following security capabilities should not be added: AUDIT_CONTROL, SYS_ADMIN, NET_ADMIN",
				Type:     "warning",
				Category: "Security",
			}, {
				ID:       "runAsRootAllowed",
				Message:  "Should not be allowed to run as root",
				Type:     "warning",
				Category: "Security",
			}, {
				ID:       "notReadOnlyRootFileSystem",
				Message:  "Filesystem should be read only",
				Type:     "warning",
				Category: "Security",
			}},
		},
		{
			name:         "bad security context + standard validation config with good settings in podspec",
			securityConf: standardConf,
			cv:           badCVWithGoodPodSpec,
			expectedMessages: []*ResultMessage{{
				ID:       "capabilitiesAdded",
				Message:  "The following security capabilities should not be added: SYS_ADMIN, NET_ADMIN",
				Type:     "error",
				Category: "Security",
			}, {
				ID:       "privilegeEscalationAllowed",
				Message:  "Privilege escalation should not be allowed",
				Type:     "error",
				Category: "Security",
			}, {
				ID:       "runAsPrivileged",
				Message:  "Should not be running as privileged",
				Type:     "error",
				Category: "Security",
			}, {
				ID:       "capabilitiesAdded",
				Message:  "The following security capabilities should not be added: AUDIT_CONTROL, SYS_ADMIN, NET_ADMIN",
				Type:     "warning",
				Category: "Security",
			}, {
				ID:       "runAsRootAllowed",
				Message:  "Should not be allowed to run as root",
				Type:     "warning",
				Category: "Security",
			}, {
				ID:       "notReadOnlyRootFileSystem",
				Message:  "Filesystem should be read only",
				Type:     "warning",
				Category: "Security",
			}},
		},
		{
			name:         "bad security context + standard validation config from default set in podspec",
			securityConf: standardConf,
			cv:           badCVWithBadPodSpec,
			expectedMessages: []*ResultMessage{{
				ID:       "capabilitiesAdded",
				Message:  "The following security capabilities should not be added: SYS_ADMIN, NET_ADMIN",
				Type:     "error",
				Category: "Security",
			}, {
				ID:       "privilegeEscalationAllowed",
				Message:  "Privilege escalation should not be allowed",
				Type:     "error",
				Category: "Security",
			}, {
				ID:       "runAsPrivileged",
				Message:  "Should not be running as privileged",
				Type:     "error",
				Category: "Security",
			}, {
				ID:       "capabilitiesAdded",
				Message:  "The following security capabilities should not be added: AUDIT_CONTROL, SYS_ADMIN, NET_ADMIN",
				Type:     "warning",
				Category: "Security",
			}, {
				ID:       "runAsRootAllowed",
				Message:  "Should not be allowed to run as root",
				Type:     "warning",
				Category: "Security",
			}, {
				ID:       "notReadOnlyRootFileSystem",
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
				ID:       "runAsRootAllowed",
				Message:  "Is not allowed to run as root",
				Type:     "success",
				Category: "Security",
			}, {
				ID:       "notReadOnlyRootFileSystem",
				Message:  "Filesystem is read only",
				Type:     "success",
				Category: "Security",
			}, {
				ID:       "runAsPrivileged",
				Message:  "Not running as privileged",
				Type:     "success",
				Category: "Security",
			}, {
				ID:       "privilegeEscalationAllowed",
				Message:  "Privilege escalation not allowed",
				Type:     "success",
				Category: "Security",
			}, {
				ID:       "capabilitiesAdded",
				Message:  "Disallowed security capabilities have not been added",
				Type:     "success",
				Category: "Security",
			}},
		},
		{
			name:         "good security context + strong validation config",
			securityConf: strongConf,
			cv:           goodCV,
			expectedMessages: []*ResultMessage{{
				ID:       "capabilitiesNotDropped",
				Message:  "The following security capabilities should be dropped: DAC_OVERRIDE, SYS_CHROOT",
				Type:     "error",
				Category: "Security",
			}, {
				ID:       "capabilitiesAdded",
				Message:  "Disallowed security capabilities have not been added",
				Type:     "success",
				Category: "Security",
			}, {
				ID:       "runAsRootAllowed",
				Message:  "Is not allowed to run as root",
				Type:     "success",
				Category: "Security",
			}, {
				ID:       "notReadOnlyRootFileSystem",
				Message:  "Filesystem is read only",
				Type:     "success",
				Category: "Security",
			}, {
				ID:       "runAsPrivileged",
				Message:  "Not running as privileged",
				Type:     "success",
				Category: "Security",
			}, {
				ID:       "privilegeEscalationAllowed",
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
				ID:       "runAsRootAllowed",
				Message:  "Is not allowed to run as root",
				Type:     "success",
				Category: "Security",
			}, {
				ID:       "notReadOnlyRootFileSystem",
				Message:  "Filesystem is read only",
				Type:     "success",
				Category: "Security",
			}, {
				ID:       "runAsPrivileged",
				Message:  "Not running as privileged",
				Type:     "success",
				Category: "Security",
			}, {
				ID:       "privilegeEscalationAllowed",
				Message:  "Privilege escalation not allowed",
				Type:     "success",
				Category: "Security",
			}, {
				ID:       "capabilitiesAdded",
				Message:  "Disallowed security capabilities have not been added",
				Type:     "success",
				Category: "Security",
			}, {
				ID:       "capabilitiesDropped",
				Message:  "All disallowed security capabilities have been dropped",
				Type:     "success",
				Category: "Security",
			}},
		},
		{
			name:         "strong security context + strong validation config via podspec default",
			securityConf: strongConf,
			cv:           strongCVWithPodSpecSecurityContext,
			expectedMessages: []*ResultMessage{{
				ID:       "runAsRootAllowed",
				Message:  "Is not allowed to run as root",
				Type:     "success",
				Category: "Security",
			}, {
				ID:       "notReadOnlyRootFileSystem",
				Message:  "Filesystem is read only",
				Type:     "success",
				Category: "Security",
			}, {
				ID:       "runAsPrivileged",
				Message:  "Not running as privileged",
				Type:     "success",
				Category: "Security",
			}, {
				ID:       "privilegeEscalationAllowed",
				Message:  "Privilege escalation not allowed",
				Type:     "success",
				Category: "Security",
			}, {
				ID:       "capabilitiesAdded",
				Message:  "Disallowed security capabilities have not been added",
				Type:     "success",
				Category: "Security",
			}, {
				ID:       "capabilitiesDropped",
				Message:  "All disallowed security capabilities have been dropped",
				Type:     "success",
				Category: "Security",
			}},
		},
		{
			name:         "strong security context + strong validation config with bad setting in podspec default",
			securityConf: strongConf,
			cv:           strongCVWithBadPodSpecSecurityContext,
			expectedMessages: []*ResultMessage{{
				ID:       "runAsRootAllowed",
				Message:  "Is not allowed to run as root",
				Type:     "success",
				Category: "Security",
			}, {
				ID:       "notReadOnlyRootFileSystem",
				Message:  "Filesystem is read only",
				Type:     "success",
				Category: "Security",
			}, {
				ID:       "runAsPrivileged",
				Message:  "Not running as privileged",
				Type:     "success",
				Category: "Security",
			}, {
				ID:       "privilegeEscalationAllowed",
				Message:  "Privilege escalation not allowed",
				Type:     "success",
				Category: "Security",
			}, {
				ID:       "capabilitiesAdded",
				Message:  "Disallowed security capabilities have not been added",
				Type:     "success",
				Category: "Security",
			}, {
				ID:       "capabilitiesDropped",
				Message:  "All disallowed security capabilities have been dropped",
				Type:     "success",
				Category: "Security",
			}},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			tt.cv = resetCV(tt.cv)
			tt.cv.validateSecurity(&conf.Configuration{Security: tt.securityConf}, "")
			assert.Len(t, tt.cv.messages(), len(tt.expectedMessages))
			assert.ElementsMatch(t, tt.cv.messages(), tt.expectedMessages)
		})
	}
}

func TestValidateResourcesExemption(t *testing.T) {
	container := corev1.Container{
		Name: "Empty",
	}

	expectedWarnings := []*ResultMessage{}
	expectedErrors := []*ResultMessage{}

	testValidateResources(t, &container, &resourceConfExemptions, "foo", &expectedErrors, &expectedWarnings)

	expectedWarnings = []*ResultMessage{
		{
			ID:       "cpuRequestsMissing",
			Type:     "warning",
			Message:  "CPU requests should be set",
			Category: "Resources",
		},
		{
			ID:       "memoryRequestsMissing",
			Type:     "warning",
			Message:  "Memory requests should be set",
			Category: "Resources",
		},
	}

	expectedErrors = []*ResultMessage{
		{
			ID:       "cpuLimitsMissing",
			Type:     "error",
			Message:  "CPU limits should be set",
			Category: "Resources",
		},
		{
			ID:       "memoryLimitsMissing",
			Type:     "error",
			Message:  "Memory limits should be set",
			Category: "Resources",
		},
	}

	disallowExemptionsConf := resourceConfExemptions + "\ndisallowExemptions: true"

	testValidateResources(t, &container, &disallowExemptionsConf, "foo", &expectedErrors, &expectedWarnings)
}

func TestValidateResourceRangeExemption(t *testing.T) {
	container := corev1.Container{
		Name: "Empty",
	}

	expectedWarnings := []*ResultMessage{}
	expectedErrors := []*ResultMessage{}

	testValidateResources(t, &container, &resourceConfRangeExemptions, "foo", &expectedErrors, &expectedWarnings)
}

func resetCV(cv ContainerValidation) ContainerValidation {
	cv.Errors = []*ResultMessage{}
	cv.Successes = []*ResultMessage{}
	cv.Warnings = []*ResultMessage{}
	return cv
}
