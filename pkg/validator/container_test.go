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
	"fmt"
	"testing"

	conf "github.com/fairwindsops/polaris/pkg/config"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

var resourceConfMinimal = `---
checks:
  cpuRequestsMissing: warning
  memoryRequestsMissing: warning
  cpuLimitsMissing: error
  memoryLimitsMissing: error
`

var resourceConfExemptions = `---
checks:
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

var resourceConfRanges = `
checks:
  memoryRequestsRange: error
  memoryLimitsRange: warning
customChecks:
  memoryLimitsRange:
    containers:
      exclude:
      - initContainer
    successMessage: Memory limits are within the required range
    failureMessage: Memory limits should be within the required range
    category: Resources
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
    category: Resources
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

func testValidateResources(t *testing.T, container *corev1.Container, resourceConf *string, controllerName string, expectedErrors []*ResultMessage, expectedWarnings []*ResultMessage, expectedSuccesses []*ResultMessage) {
	cv := ContainerValidation{
		Container:          container,
		ResourceValidation: &ResourceValidation{},
	}

	parsedConf, err := conf.Parse([]byte(*resourceConf))
	assert.NoError(t, err, "Expected no error when parsing config")

	err = applyContainerSchemaChecks(&parsedConf, controllerName, conf.Deployments, &cv)
	if err != nil {
		panic(err)
	}

	assert.Len(t, cv.Warnings, len(expectedWarnings))
	assert.ElementsMatch(t, expectedWarnings, cv.Warnings)

	assert.Len(t, cv.Errors, len(expectedErrors))
	assert.ElementsMatch(t, expectedErrors, cv.Errors)

	assert.Len(t, cv.Successes, len(expectedSuccesses))
	assert.ElementsMatch(t, expectedSuccesses, cv.Successes)
}

func TestValidateResourcesEmptyConfig(t *testing.T) {
	container := corev1.Container{
		Name: "Empty",
	}

	cv := ContainerValidation{
		Container:          &container,
		ResourceValidation: &ResourceValidation{},
	}

	err := applyContainerSchemaChecks(&conf.Configuration{}, "", conf.Deployments, &cv)
	if err != nil {
		panic(err)
	}
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

	expectedSuccesses := []*ResultMessage{}

	testValidateResources(t, &container, &resourceConfMinimal, "foo", expectedErrors, expectedWarnings, expectedSuccesses)
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
			ID:       "memoryLimitsRange",
			Type:     "warning",
			Message:  "Memory limits should be within the required range",
			Category: "Resources",
		},
	}

	expectedErrors := []*ResultMessage{
		{
			ID:       "memoryRequestsRange",
			Type:     "error",
			Message:  "Memory requests should be within the required range",
			Category: "Resources",
		},
	}

	expectedSuccesses := []*ResultMessage{}

	testValidateResources(t, &container, &resourceConfRanges, "foo", expectedErrors, expectedWarnings, expectedSuccesses)
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

	parsedConf, err := conf.Parse([]byte(resourceConfRanges))
	assert.NoError(t, err, "Expected no error when parsing config")

	err = applyContainerSchemaChecks(&parsedConf, "", conf.Deployments, &cvEmpty)
	if err != nil {
		panic(err)
	}
	assert.Len(t, cvEmpty.Errors, 1)
	assert.Len(t, cvEmpty.Warnings, 1)

	err = applyContainerSchemaChecks(&parsedConf, "", conf.Deployments, &cvInit)
	if err != nil {
		panic(err)
	}
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

	expectedSuccesses := []*ResultMessage{
		{
			ID:       "memoryRequestsRange",
			Type:     "success",
			Message:  "Memory requests are within the required range",
			Category: "Resources",
		},
		{
			ID:       "memoryLimitsRange",
			Type:     "success",
			Message:  "Memory limits are within the required range",
			Category: "Resources",
		},
	}

	testValidateResources(t, &container, &resourceConfRanges, "foo", []*ResultMessage{}, []*ResultMessage{}, expectedSuccesses)

	expectedSuccesses = []*ResultMessage{
		{
			ID:       "cpuRequestsMissing",
			Type:     "success",
			Message:  "CPU requests are set",
			Category: "Resources",
		},
		{
			ID:       "memoryRequestsMissing",
			Type:     "success",
			Message:  "Memory requests are set",
			Category: "Resources",
		},
		{
			ID:       "cpuLimitsMissing",
			Type:     "success",
			Message:  "CPU limits are set",
			Category: "Resources",
		},
		{
			ID:       "memoryLimitsMissing",
			Type:     "success",
			Message:  "Memory limits are set",
			Category: "Resources",
		},
	}

	testValidateResources(t, &container, &resourceConfMinimal, "foo", []*ResultMessage{}, []*ResultMessage{}, expectedSuccesses)
}

func TestValidateHealthChecks(t *testing.T) {

	// Test setup.
	p1 := make(map[string]conf.Severity)
	p2 := map[string]conf.Severity{
		"readinessProbeMissing": conf.SeverityIgnore,
		"livenessProbeMissing":  conf.SeverityIgnore,
	}
	p3 := map[string]conf.Severity{
		"readinessProbeMissing": conf.SeverityError,
		"livenessProbeMissing":  conf.SeverityWarning,
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
		probes   map[string]conf.Severity
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

	for idx, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			err := applyContainerSchemaChecks(&conf.Configuration{Checks: tt.probes}, "", conf.Deployments, &tt.cv)
			if err != nil {
				panic(err)
			}
			message := fmt.Sprintf("test case %d", idx)

			if tt.warnings != nil {
				assert.Len(t, tt.cv.Warnings, len(*tt.warnings), message)
				assert.ElementsMatch(t, tt.cv.Warnings, *tt.warnings, message)
			}

			assert.Len(t, tt.cv.Errors, len(*tt.errors), message)
			assert.ElementsMatch(t, tt.cv.Errors, *tt.errors, message)
		})
	}
}

func TestValidateImage(t *testing.T) {
	emptyConf := make(map[string]conf.Severity)
	standardConf := map[string]conf.Severity{
		"tagNotSpecified":     conf.SeverityError,
		"pullPolicyNotAlways": conf.SeverityIgnore,
	}
	strongConf := map[string]conf.Severity{
		"tagNotSpecified":     conf.SeverityError,
		"pullPolicyNotAlways": conf.SeverityError,
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
		image    map[string]conf.Severity
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
			err := applyContainerSchemaChecks(&conf.Configuration{Checks: tt.image}, "", conf.Deployments, &tt.cv)
			if err != nil {
				panic(err)
			}
			assert.Len(t, tt.cv.Errors, len(tt.expected))
			assert.ElementsMatch(t, tt.cv.Errors, tt.expected)
		})
	}
}

func TestValidateNetworking(t *testing.T) {
	// Test setup.
	emptyConf := make(map[string]conf.Severity)
	standardConf := map[string]conf.Severity{
		"hostPortSet": conf.SeverityWarning,
	}
	strongConf := map[string]conf.Severity{
		"hostPortSet": conf.SeverityError,
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
		networkConf      map[string]conf.Severity
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
			err := applyContainerSchemaChecks(&conf.Configuration{Checks: tt.networkConf}, "", conf.Deployments, &tt.cv)
			if err != nil {
				panic(err)
			}
			assert.Len(t, tt.cv.messages(), len(tt.expectedMessages))
			assert.ElementsMatch(t, tt.cv.messages(), tt.expectedMessages)
		})
	}
}

func TestValidateSecurity(t *testing.T) {
	trueVar := true
	falseVar := false

	// Test setup.
	emptyConf := map[string]conf.Severity{}
	standardConf := map[string]conf.Severity{
		"runAsRootAllowed":           conf.SeverityWarning,
		"runAsPrivileged":            conf.SeverityError,
		"notReadOnlyRootFileSystem":  conf.SeverityWarning,
		"privilegeEscalationAllowed": conf.SeverityError,
		"dangerousCapabilities":      conf.SeverityError,
		"insecureCapabilities":       conf.SeverityWarning,
	}
	strongConf := map[string]conf.Severity{
		"runAsRootAllowed":           conf.SeverityError,
		"runAsPrivileged":            conf.SeverityError,
		"notReadOnlyRootFileSystem":  conf.SeverityError,
		"privilegeEscalationAllowed": conf.SeverityError,
		"dangerousCapabilities":      conf.SeverityError,
		"insecureCapabilities":       conf.SeverityError,
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
				Add: []corev1.Capability{"AUDIT_WRITE", "SYS_ADMIN", "NET_ADMIN"},
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
				Add: []corev1.Capability{"AUDIT_WRITE", "SYS_ADMIN", "NET_ADMIN"},
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
				Add: []corev1.Capability{"AUDIT_WRITE", "SYS_ADMIN", "NET_ADMIN"},
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
		securityConf     map[string]conf.Severity
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
				ID:       "insecureCapabilities",
				Message:  "Container does not have any insecure capabilities",
				Type:     "success",
				Category: "Security",
			}, {
				ID:       "dangerousCapabilities",
				Message:  "Container does not have any dangerous capabilities",
				Type:     "success",
				Category: "Security",
			}},
		},
		{
			name:         "bad security context + standard validation config",
			securityConf: standardConf,
			cv:           badCV,
			expectedMessages: []*ResultMessage{{
				ID:       "dangerousCapabilities",
				Message:  "Container should not have dangerous capabilities",
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
				ID:       "insecureCapabilities",
				Message:  "Container should not have insecure capabilities",
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
				ID:       "dangerousCapabilities",
				Message:  "Container should not have dangerous capabilities",
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
				ID:       "insecureCapabilities",
				Message:  "Container should not have insecure capabilities",
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
				ID:       "dangerousCapabilities",
				Message:  "Container should not have dangerous capabilities",
				Type:     "error",
				Category: "Security",
			}, {
				ID:       "insecureCapabilities",
				Message:  "Container should not have insecure capabilities",
				Type:     "warning",
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
				ID:       "dangerousCapabilities",
				Message:  "Container does not have any dangerous capabilities",
				Type:     "success",
				Category: "Security",
			}, {
				ID:       "insecureCapabilities",
				Message:  "Container does not have any insecure capabilities",
				Type:     "success",
				Category: "Security",
			}},
		},
		{
			name:         "good security context + strong validation config",
			securityConf: strongConf,
			cv:           goodCV,
			expectedMessages: []*ResultMessage{{
				ID:       "dangerousCapabilities",
				Message:  "Container does not have any dangerous capabilities",
				Type:     "success",
				Category: "Security",
			}, {
				ID:       "insecureCapabilities",
				Message:  "Container does not have any insecure capabilities",
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
				ID:       "dangerousCapabilities",
				Message:  "Container does not have any dangerous capabilities",
				Type:     "success",
				Category: "Security",
			}, {
				ID:       "insecureCapabilities",
				Message:  "Container does not have any insecure capabilities",
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
				ID:       "dangerousCapabilities",
				Message:  "Container does not have any dangerous capabilities",
				Type:     "success",
				Category: "Security",
			}, {
				ID:       "insecureCapabilities",
				Message:  "Container does not have any insecure capabilities",
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
				ID:       "dangerousCapabilities",
				Message:  "Container does not have any dangerous capabilities",
				Type:     "success",
				Category: "Security",
			}, {
				ID:       "insecureCapabilities",
				Message:  "Container does not have any insecure capabilities",
				Type:     "success",
				Category: "Security",
			}},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			tt.cv = resetCV(tt.cv)
			err := applyContainerSchemaChecks(&conf.Configuration{Checks: tt.securityConf}, "", conf.Deployments, &tt.cv)
			if err != nil {
				panic(err)
			}
			assert.Len(t, tt.cv.messages(), len(tt.expectedMessages))
			assert.ElementsMatch(t, tt.expectedMessages, tt.cv.messages())
		})
	}
}

func TestValidateRunAsRoot(t *testing.T) {
	falseVar := false
	trueVar := true
	nonRootUser := int64(1000)
	rootUser := int64(0)
	config := conf.Configuration{
		Checks: map[string]conf.Severity{
			"runAsRootAllowed": conf.SeverityWarning,
		},
	}
	testCases := []struct {
		name    string
		cv      ContainerValidation
		message ResultMessage
	}{
		{
			name: "pod=false,container=nil",
			cv: ContainerValidation{
				ResourceValidation: &ResourceValidation{},
				Container: &corev1.Container{Name: "", SecurityContext: &corev1.SecurityContext{
					RunAsNonRoot: nil,
				}},
				parentPodSpec: corev1.PodSpec{
					SecurityContext: &corev1.PodSecurityContext{
						RunAsNonRoot: &falseVar,
					},
				},
			},
			message: ResultMessage{
				ID:       "runAsRootAllowed",
				Message:  "Should not be allowed to run as root",
				Type:     "warning",
				Category: "Security",
			},
		},
		{
			name: "pod=false,container=true",
			cv: ContainerValidation{
				ResourceValidation: &ResourceValidation{},
				Container: &corev1.Container{Name: "", SecurityContext: &corev1.SecurityContext{
					RunAsNonRoot: &trueVar,
				}},
				parentPodSpec: corev1.PodSpec{
					SecurityContext: &corev1.PodSecurityContext{
						RunAsNonRoot: &falseVar,
					},
				},
			},
			message: ResultMessage{
				ID:       "runAsRootAllowed",
				Message:  "Is not allowed to run as root",
				Type:     "success",
				Category: "Security",
			},
		},
		{
			name: "pod=nil,container=runAsUser",
			cv: ContainerValidation{
				ResourceValidation: &ResourceValidation{},
				Container: &corev1.Container{Name: "", SecurityContext: &corev1.SecurityContext{
					RunAsUser: &nonRootUser,
				}},
			},
			message: ResultMessage{
				ID:       "runAsRootAllowed",
				Message:  "Is not allowed to run as root",
				Type:     "success",
				Category: "Security",
			},
		},
		{
			name: "pod=runAsUser,container=nil",
			cv: ContainerValidation{
				ResourceValidation: &ResourceValidation{},
				Container:          &corev1.Container{Name: "", SecurityContext: &corev1.SecurityContext{}},
				parentPodSpec: corev1.PodSpec{
					SecurityContext: &corev1.PodSecurityContext{
						RunAsUser: &nonRootUser,
					},
				},
			},
			message: ResultMessage{
				ID:       "runAsRootAllowed",
				Message:  "Is not allowed to run as root",
				Type:     "success",
				Category: "Security",
			},
		},
		{
			name: "pod=runAsUser,container=runAsUser0",
			cv: ContainerValidation{
				ResourceValidation: &ResourceValidation{},
				Container: &corev1.Container{Name: "", SecurityContext: &corev1.SecurityContext{
					RunAsUser: &rootUser,
				}},
				parentPodSpec: corev1.PodSpec{
					SecurityContext: &corev1.PodSecurityContext{
						RunAsUser: &nonRootUser,
					},
				},
			},
			message: ResultMessage{
				ID:       "runAsRootAllowed",
				Message:  "Should not be allowed to run as root",
				Type:     "warning",
				Category: "Security",
			},
		},
		{
			name: "pod=false,container=runAsUser",
			cv: ContainerValidation{
				ResourceValidation: &ResourceValidation{},
				Container: &corev1.Container{Name: "", SecurityContext: &corev1.SecurityContext{
					RunAsNonRoot: &falseVar,
				}},
				parentPodSpec: corev1.PodSpec{
					SecurityContext: &corev1.PodSecurityContext{
						RunAsUser: &nonRootUser,
					},
				},
			},
			message: ResultMessage{
				ID:       "runAsRootAllowed",
				Message:  "Should not be allowed to run as root",
				Type:     "warning",
				Category: "Security",
			},
		},
	}
	for idx, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			err := applyContainerSchemaChecks(&config, "", conf.Deployments, &tt.cv)
			if err != nil {
				panic(err)
			}
			assert.Len(t, tt.cv.messages(), 1)
			if len(tt.cv.messages()) > 0 {
				assert.Equal(t, &tt.message, tt.cv.messages()[0], fmt.Sprintf("Test case %d failed", idx))
			}
		})
	}
}

func TestValidateResourcesExemption(t *testing.T) {
	container := corev1.Container{
		Name: "Empty",
	}

	expectedWarnings := []*ResultMessage{}
	expectedErrors := []*ResultMessage{}
	expectedSuccesses := []*ResultMessage{}

	testValidateResources(t, &container, &resourceConfExemptions, "foo", expectedErrors, expectedWarnings, expectedSuccesses)

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

	testValidateResources(t, &container, &disallowExemptionsConf, "foo", expectedErrors, expectedWarnings, expectedSuccesses)
}

/*
func TestValidateResourceRangeExemption(t *testing.T) {
	container := corev1.Container{
		Name: "Empty",
	}

	expectedWarnings := []*ResultMessage{}
	expectedErrors := []*ResultMessage{}
	expectedSuccesses := []*ResultMessage{}

	testValidateResources(t, &container, &resourceConfRangeExemptions, "foo", expectedErrors, expectedWarnings, expectedSuccesses)
}
*/

func resetCV(cv ContainerValidation) ContainerValidation {
	cv.Errors = []*ResultMessage{}
	cv.Successes = []*ResultMessage{}
	cv.Warnings = []*ResultMessage{}
	return cv
}
