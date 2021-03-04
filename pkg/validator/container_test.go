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
	"github.com/fairwindsops/polaris/pkg/kube"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var resourceConfMinimal = `---
checks:
  cpuRequestsMissing: warning
  memoryRequestsMissing: warning
  cpuLimitsMissing: danger
  memoryLimitsMissing: danger
`

var resourceConfExemptions = `---
checks:
  cpuRequestsMissing: warning
  memoryRequestsMissing: warning
  cpuLimitsMissing: danger
  memoryLimitsMissing: danger
exemptions:
  - rules:
    - cpuRequestsMissing
    - memoryRequestsMissing
    - cpuLimitsMissing
    - memoryLimitsMissing
    controllerNames:
    - foo
`

func getEmptyWorkload(t *testing.T, name string) kube.GenericResource {
	workload, err := kube.NewGenericResourceFromPod(corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}, nil)
	assert.NoError(t, err)
	return workload
}

func testValidate(t *testing.T, container *corev1.Container, resourceConf *string, controllerName string, expectedDangers []ResultMessage, expectedWarnings []ResultMessage, expectedSuccesses []ResultMessage) {
	testValidateWithWorkload(t, container, resourceConf, getEmptyWorkload(t, controllerName), expectedDangers, expectedWarnings, expectedSuccesses)
}

func testValidateWithWorkload(t *testing.T, container *corev1.Container, resourceConf *string, workload kube.GenericResource, expectedDangers []ResultMessage, expectedWarnings []ResultMessage, expectedSuccesses []ResultMessage) {
	parsedConf, err := conf.Parse([]byte(*resourceConf))
	assert.NoError(t, err, "Expected no error when parsing config")

	var results ResultSet
	results, err = applyContainerSchemaChecks(&parsedConf, workload, container, false)
	if err != nil {
		panic(err)
	}
	summary := results.GetSummary()

	assert.Equal(t, uint(len(expectedWarnings)), summary.Warnings)
	assert.ElementsMatch(t, expectedWarnings, results.GetWarnings())

	assert.Equal(t, uint(len(expectedDangers)), summary.Dangers)
	assert.ElementsMatch(t, expectedDangers, results.GetDangers())

	assert.Equal(t, uint(len(expectedSuccesses)), summary.Successes)
	assert.ElementsMatch(t, expectedSuccesses, results.GetSuccesses())
}

func TestValidateResourcesEmptyConfig(t *testing.T) {
	container := &corev1.Container{
		Name: "Empty",
	}

	results, err := applyContainerSchemaChecks(&conf.Configuration{}, getEmptyWorkload(t, ""), container, false)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, uint(0), results.GetSummary().Dangers)
}

func TestValidateResourcesEmptyContainer(t *testing.T) {
	container := corev1.Container{
		Name: "Empty",
	}

	expectedWarnings := []ResultMessage{
		{
			ID:       "cpuRequestsMissing",
			Success:  false,
			Severity: "warning",
			Message:  "CPU requests should be set",
			Category: "Efficiency",
		},
		{
			ID:       "memoryRequestsMissing",
			Success:  false,
			Severity: "warning",
			Message:  "Memory requests should be set",
			Category: "Efficiency",
		},
	}

	expectedDangers := []ResultMessage{
		{
			ID:       "cpuLimitsMissing",
			Success:  false,
			Severity: "danger",
			Message:  "CPU limits should be set",
			Category: "Efficiency",
		},
		{
			ID:       "memoryLimitsMissing",
			Success:  false,
			Severity: "danger",
			Message:  "Memory limits should be set",
			Category: "Efficiency",
		},
	}

	expectedSuccesses := []ResultMessage{}

	testValidate(t, &container, &resourceConfMinimal, "foo", expectedDangers, expectedWarnings, expectedSuccesses)
}

func TestValidateHealthChecks(t *testing.T) {

	// Test setup.
	p1 := make(map[string]conf.Severity)
	p2 := map[string]conf.Severity{
		"readinessProbeMissing": conf.SeverityIgnore,
		"livenessProbeMissing":  conf.SeverityIgnore,
	}
	p3 := map[string]conf.Severity{
		"readinessProbeMissing": conf.SeverityDanger,
		"livenessProbeMissing":  conf.SeverityWarning,
	}

	probe := corev1.Probe{}
	emptyContainer := &corev1.Container{Name: ""}
	goodContainer := &corev1.Container{
		Name:           "",
		LivenessProbe:  &probe,
		ReadinessProbe: &probe,
	}

	l := ResultMessage{ID: "livenessProbeMissing", Success: false, Severity: "warning", Message: "Liveness probe should be configured", Category: "Reliability"}
	r := ResultMessage{ID: "readinessProbeMissing", Success: false, Severity: "danger", Message: "Readiness probe should be configured", Category: "Reliability"}
	f1 := []ResultMessage{}
	f2 := []ResultMessage{r}
	w1 := []ResultMessage{l}

	var testCases = []struct {
		name      string
		probes    map[string]conf.Severity
		container *corev1.Container
		isInit    bool
		dangers   *[]ResultMessage
		warnings  *[]ResultMessage
	}{
		{name: "probes not configured", probes: p1, container: emptyContainer, dangers: &f1},
		{name: "probes not required", probes: p2, container: emptyContainer, dangers: &f1},
		{name: "probes required & configured", probes: p3, container: goodContainer, dangers: &f1},
		{name: "probes required, not configured, but init", probes: p3, container: emptyContainer, isInit: true, dangers: &f1},
		{name: "probes required & not configured", probes: p3, container: emptyContainer, dangers: &f2, warnings: &w1},
		{name: "probes configured, but not required", probes: p2, container: goodContainer, dangers: &f1},
	}

	for idx, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			controller := getEmptyWorkload(t, "")
			results, err := applyContainerSchemaChecks(&conf.Configuration{Checks: tt.probes}, controller, tt.container, tt.isInit)
			if err != nil {
				panic(err)
			}
			message := fmt.Sprintf("test case %d", idx)

			if tt.warnings != nil {
				warnings := results.GetWarnings()
				assert.Len(t, warnings, len(*tt.warnings), message)
				assert.ElementsMatch(t, warnings, *tt.warnings, message)
			}

			if tt.dangers != nil {
				dangers := results.GetDangers()
				assert.Len(t, dangers, len(*tt.dangers), message)
				assert.ElementsMatch(t, dangers, *tt.dangers, message)
			}
		})
	}
}

func TestValidateImage(t *testing.T) {
	emptyConf := make(map[string]conf.Severity)
	standardConf := map[string]conf.Severity{
		"tagNotSpecified":     conf.SeverityDanger,
		"pullPolicyNotAlways": conf.SeverityIgnore,
	}
	strongConf := map[string]conf.Severity{
		"tagNotSpecified":     conf.SeverityDanger,
		"pullPolicyNotAlways": conf.SeverityDanger,
	}

	emptyContainer := &corev1.Container{}
	badContainer := &corev1.Container{Image: "test"}
	lessBadContainer := &corev1.Container{Image: "test:latest", ImagePullPolicy: ""}
	goodContainer := &corev1.Container{Image: "test:0.1.0", ImagePullPolicy: "Always"}

	var testCases = []struct {
		name      string
		image     map[string]conf.Severity
		container *corev1.Container
		expected  []ResultMessage
	}{
		{
			name:      "emptyConf + emptyCV",
			image:     emptyConf,
			container: emptyContainer,
			expected:  []ResultMessage{},
		},
		{
			name:      "standardConf + emptyCV",
			image:     standardConf,
			container: emptyContainer,
			expected: []ResultMessage{{
				ID:       "tagNotSpecified",
				Message:  "Image tag should be specified",
				Success:  false,
				Severity: "danger",
				Category: "Reliability",
			}},
		},
		{
			name:      "standardConf + badCV",
			image:     standardConf,
			container: badContainer,
			expected: []ResultMessage{{
				ID:       "tagNotSpecified",
				Message:  "Image tag should be specified",
				Success:  false,
				Severity: "danger",
				Category: "Reliability",
			}},
		},
		{
			name:      "standardConf + lessBadCV",
			image:     standardConf,
			container: lessBadContainer,
			expected: []ResultMessage{{
				ID:       "tagNotSpecified",
				Message:  "Image tag should be specified",
				Success:  false,
				Severity: "danger",
				Category: "Reliability",
			}},
		},
		{
			name:      "strongConf + badCV",
			image:     strongConf,
			container: badContainer,
			expected: []ResultMessage{{
				ID:       "pullPolicyNotAlways",
				Message:  "Image pull policy should be \"Always\"",
				Success:  false,
				Severity: "danger",
				Category: "Reliability",
			}, {
				ID:       "tagNotSpecified",
				Message:  "Image tag should be specified",
				Success:  false,
				Severity: "danger",
				Category: "Reliability",
			}},
		},
		{
			name:      "strongConf + goodCV",
			image:     strongConf,
			container: goodContainer,
			expected:  []ResultMessage{},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			controller := getEmptyWorkload(t, "")
			results, err := applyContainerSchemaChecks(&conf.Configuration{Checks: tt.image}, controller, tt.container, false)
			if err != nil {
				panic(err)
			}
			dangers := results.GetDangers()
			assert.Len(t, dangers, len(tt.expected))
			assert.ElementsMatch(t, dangers, tt.expected)
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
		"hostPortSet": conf.SeverityDanger,
	}

	emptyContainer := &corev1.Container{Name: ""}
	badContainer := &corev1.Container{
		Ports: []corev1.ContainerPort{{
			ContainerPort: 3000,
			HostPort:      443,
		}},
	}
	goodContainer := &corev1.Container{
		Ports: []corev1.ContainerPort{{
			ContainerPort: 3000,
		}},
	}

	var testCases = []struct {
		name            string
		networkConf     map[string]conf.Severity
		container       *corev1.Container
		expectedResults []ResultMessage
	}{
		{
			name:            "empty ports + empty validation config",
			networkConf:     emptyConf,
			container:       emptyContainer,
			expectedResults: []ResultMessage{},
		},
		{
			name:        "empty ports + standard validation config",
			networkConf: standardConf,
			container:   emptyContainer,
			expectedResults: []ResultMessage{{
				ID:       "hostPortSet",
				Message:  "Host port is not configured",
				Success:  true,
				Severity: "warning",
				Category: "Security",
			}},
		},
		{
			name:        "empty ports + strong validation config",
			networkConf: standardConf,
			container:   emptyContainer,
			expectedResults: []ResultMessage{{
				ID:       "hostPortSet",
				Message:  "Host port is not configured",
				Success:  true,
				Severity: "warning",
				Category: "Security",
			}},
		},
		{
			name:            "host ports + empty validation config",
			networkConf:     emptyConf,
			container:       badContainer,
			expectedResults: []ResultMessage{},
		},
		{
			name:        "host ports + standard validation config",
			networkConf: standardConf,
			container:   badContainer,
			expectedResults: []ResultMessage{{
				ID:       "hostPortSet",
				Message:  "Host port should not be configured",
				Success:  false,
				Severity: "warning",
				Category: "Security",
			}},
		},
		{
			name:        "no host ports + standard validation config",
			networkConf: standardConf,
			container:   goodContainer,
			expectedResults: []ResultMessage{{
				ID:       "hostPortSet",
				Message:  "Host port is not configured",
				Success:  true,
				Severity: "warning",
				Category: "Security",
			}},
		},
		{
			name:        "host ports + strong validation config",
			networkConf: strongConf,
			container:   badContainer,
			expectedResults: []ResultMessage{{
				ID:       "hostPortSet",
				Message:  "Host port should not be configured",
				Success:  false,
				Severity: "danger",
				Category: "Security",
			}},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			controller := getEmptyWorkload(t, "")
			results, err := applyContainerSchemaChecks(&conf.Configuration{Checks: tt.networkConf}, controller, tt.container, false)
			if err != nil {
				panic(err)
			}
			messages := []ResultMessage{}
			for _, msg := range results {
				messages = append(messages, msg)
			}
			assert.Len(t, messages, len(tt.expectedResults))
			assert.ElementsMatch(t, messages, tt.expectedResults)
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
		"runAsPrivileged":            conf.SeverityDanger,
		"notReadOnlyRootFilesystem":  conf.SeverityWarning,
		"privilegeEscalationAllowed": conf.SeverityDanger,
		"dangerousCapabilities":      conf.SeverityDanger,
		"insecureCapabilities":       conf.SeverityWarning,
	}
	strongConf := map[string]conf.Severity{
		"runAsRootAllowed":           conf.SeverityDanger,
		"runAsPrivileged":            conf.SeverityDanger,
		"notReadOnlyRootFilesystem":  conf.SeverityDanger,
		"privilegeEscalationAllowed": conf.SeverityDanger,
		"dangerousCapabilities":      conf.SeverityDanger,
		"insecureCapabilities":       conf.SeverityDanger,
	}

	emptyContainer := &corev1.Container{Name: ""}
	badContainer := &corev1.Container{
		Name: "",
		SecurityContext: &corev1.SecurityContext{
			RunAsNonRoot:             &falseVar,
			ReadOnlyRootFilesystem:   &falseVar,
			Privileged:               &trueVar,
			AllowPrivilegeEscalation: &trueVar,
			Capabilities: &corev1.Capabilities{
				Add: []corev1.Capability{"AUDIT_WRITE", "SYS_ADMIN", "NET_ADMIN"},
			},
		},
	}
	emptyPodSpec := &corev1.PodSpec{}
	goodPodSpec := &corev1.PodSpec{
		SecurityContext: &corev1.PodSecurityContext{
			RunAsNonRoot: &trueVar,
		},
	}
	badPodSpec := &corev1.PodSpec{
		SecurityContext: &corev1.PodSecurityContext{
			RunAsNonRoot: &falseVar,
		},
	}
	inheritContainer := &corev1.Container{
		Name: "",
		SecurityContext: &corev1.SecurityContext{
			RunAsNonRoot:             nil, // this will use the default from the podspec
			ReadOnlyRootFilesystem:   &trueVar,
			Privileged:               &falseVar,
			AllowPrivilegeEscalation: &falseVar,
			Capabilities: &corev1.Capabilities{
				Drop: []corev1.Capability{"ALL"},
			},
		},
	}
	goodContainer := &corev1.Container{
		Name: "",
		SecurityContext: &corev1.SecurityContext{
			RunAsNonRoot:             &trueVar,
			ReadOnlyRootFilesystem:   &trueVar,
			Privileged:               &falseVar,
			AllowPrivilegeEscalation: &falseVar,
			Capabilities: &corev1.Capabilities{
				Drop: []corev1.Capability{"NET_BIND_SERVICE", "FOWNER"},
			},
		},
	}
	strongContainer := &corev1.Container{
		Name: "",
		SecurityContext: &corev1.SecurityContext{
			RunAsNonRoot:             &trueVar,
			ReadOnlyRootFilesystem:   &trueVar,
			Privileged:               &falseVar,
			AllowPrivilegeEscalation: &falseVar,
			Capabilities: &corev1.Capabilities{
				Drop: []corev1.Capability{"ALL"},
			},
		},
	}

	var testCases = []struct {
		name            string
		securityConf    map[string]conf.Severity
		container       *corev1.Container
		pod             *corev1.PodSpec
		expectedResults []ResultMessage
	}{
		{
			name:            "empty security context + empty validation config",
			securityConf:    emptyConf,
			container:       emptyContainer,
			pod:             emptyPodSpec,
			expectedResults: []ResultMessage{},
		},
		{
			name:         "empty security context + standard validation config",
			securityConf: standardConf,
			container:    emptyContainer,
			pod:          emptyPodSpec,
			expectedResults: []ResultMessage{{
				ID:       "runAsRootAllowed",
				Message:  "Should not be allowed to run as root",
				Success:  false,
				Severity: "warning",
				Category: "Security",
			}, {
				ID:       "notReadOnlyRootFilesystem",
				Message:  "Filesystem should be read only",
				Success:  false,
				Severity: "warning",
				Category: "Security",
			}, {
				ID:       "runAsPrivileged",
				Message:  "Not running as privileged",
				Success:  true,
				Severity: "danger",
				Category: "Security",
			}, {
				ID:       "privilegeEscalationAllowed",
				Message:  "Privilege escalation should not be allowed",
				Success:  false,
				Severity: "danger",
				Category: "Security",
			}, {
				ID:       "insecureCapabilities",
				Message:  "Container should not have insecure capabilities",
				Success:  false,
				Severity: "warning",
				Category: "Security",
			}, {
				ID:       "dangerousCapabilities",
				Message:  "Container does not have any dangerous capabilities",
				Success:  true,
				Severity: "danger",
				Category: "Security",
			}},
		},
		{
			name:         "bad security context + standard validation config",
			securityConf: standardConf,
			container:    badContainer,
			pod:          emptyPodSpec,
			expectedResults: []ResultMessage{{
				ID:       "dangerousCapabilities",
				Message:  "Container should not have dangerous capabilities",
				Success:  false,
				Severity: "danger",
				Category: "Security",
			}, {
				ID:       "privilegeEscalationAllowed",
				Message:  "Privilege escalation should not be allowed",
				Success:  false,
				Severity: "danger",
				Category: "Security",
			}, {
				ID:       "runAsPrivileged",
				Message:  "Should not be running as privileged",
				Success:  false,
				Severity: "danger",
				Category: "Security",
			}, {
				ID:       "insecureCapabilities",
				Message:  "Container should not have insecure capabilities",
				Success:  false,
				Severity: "warning",
				Category: "Security",
			}, {
				ID:       "runAsRootAllowed",
				Message:  "Should not be allowed to run as root",
				Success:  false,
				Severity: "warning",
				Category: "Security",
			}, {
				ID:       "notReadOnlyRootFilesystem",
				Message:  "Filesystem should be read only",
				Success:  false,
				Severity: "warning",
				Category: "Security",
			}},
		},
		{
			name:         "bad security context + standard validation config with good settings in podspec",
			securityConf: standardConf,
			container:    badContainer,
			pod:          goodPodSpec,
			expectedResults: []ResultMessage{{
				ID:       "dangerousCapabilities",
				Message:  "Container should not have dangerous capabilities",
				Success:  false,
				Severity: "danger",
				Category: "Security",
			}, {
				ID:       "privilegeEscalationAllowed",
				Message:  "Privilege escalation should not be allowed",
				Success:  false,
				Severity: "danger",
				Category: "Security",
			}, {
				ID:       "runAsPrivileged",
				Message:  "Should not be running as privileged",
				Success:  false,
				Severity: "danger",
				Category: "Security",
			}, {
				ID:       "insecureCapabilities",
				Message:  "Container should not have insecure capabilities",
				Success:  false,
				Severity: "warning",
				Category: "Security",
			}, {
				ID:       "runAsRootAllowed",
				Message:  "Should not be allowed to run as root",
				Success:  false,
				Severity: "warning",
				Category: "Security",
			}, {
				ID:       "notReadOnlyRootFilesystem",
				Message:  "Filesystem should be read only",
				Success:  false,
				Severity: "warning",
				Category: "Security",
			}},
		},
		{
			name:         "bad security context + standard validation config from default set in podspec",
			securityConf: standardConf,
			container:    badContainer,
			pod:          badPodSpec,
			expectedResults: []ResultMessage{{
				ID:       "dangerousCapabilities",
				Message:  "Container should not have dangerous capabilities",
				Success:  false,
				Severity: "danger",
				Category: "Security",
			}, {
				ID:       "insecureCapabilities",
				Message:  "Container should not have insecure capabilities",
				Success:  false,
				Severity: "warning",
				Category: "Security",
			}, {
				ID:       "privilegeEscalationAllowed",
				Message:  "Privilege escalation should not be allowed",
				Success:  false,
				Severity: "danger",
				Category: "Security",
			}, {
				ID:       "runAsPrivileged",
				Message:  "Should not be running as privileged",
				Success:  false,
				Severity: "danger",
				Category: "Security",
			}, {
				ID:       "runAsRootAllowed",
				Message:  "Should not be allowed to run as root",
				Success:  false,
				Severity: "warning",
				Category: "Security",
			}, {
				ID:       "notReadOnlyRootFilesystem",
				Message:  "Filesystem should be read only",
				Success:  false,
				Severity: "warning",
				Category: "Security",
			}},
		},
		{
			name:         "good security context + standard validation config",
			securityConf: standardConf,
			container:    goodContainer,
			pod:          emptyPodSpec,
			expectedResults: []ResultMessage{{
				ID:       "runAsRootAllowed",
				Message:  "Is not allowed to run as root",
				Success:  true,
				Severity: "warning",
				Category: "Security",
			}, {
				ID:       "notReadOnlyRootFilesystem",
				Message:  "Filesystem is read only",
				Success:  true,
				Severity: "warning",
				Category: "Security",
			}, {
				ID:       "runAsPrivileged",
				Message:  "Not running as privileged",
				Success:  true,
				Severity: "danger",
				Category: "Security",
			}, {
				ID:       "privilegeEscalationAllowed",
				Message:  "Privilege escalation not allowed",
				Success:  true,
				Severity: "danger",
				Category: "Security",
			}, {
				ID:       "dangerousCapabilities",
				Message:  "Container does not have any dangerous capabilities",
				Success:  true,
				Severity: "danger",
				Category: "Security",
			}, {
				ID:       "insecureCapabilities",
				Message:  "Container should not have insecure capabilities",
				Success:  false,
				Severity: "warning",
				Category: "Security",
			}},
		},
		{
			name:         "good security context + strong validation config",
			securityConf: strongConf,
			container:    goodContainer,
			pod:          emptyPodSpec,
			expectedResults: []ResultMessage{{
				ID:       "dangerousCapabilities",
				Message:  "Container does not have any dangerous capabilities",
				Success:  true,
				Severity: "danger",
				Category: "Security",
			}, {
				ID:       "insecureCapabilities",
				Message:  "Container should not have insecure capabilities",
				Success:  false,
				Severity: "danger",
				Category: "Security",
			}, {
				ID:       "runAsRootAllowed",
				Message:  "Is not allowed to run as root",
				Success:  true,
				Severity: "danger",
				Category: "Security",
			}, {
				ID:       "notReadOnlyRootFilesystem",
				Message:  "Filesystem is read only",
				Success:  true,
				Severity: "danger",
				Category: "Security",
			}, {
				ID:       "runAsPrivileged",
				Message:  "Not running as privileged",
				Success:  true,
				Severity: "danger",
				Category: "Security",
			}, {
				ID:       "privilegeEscalationAllowed",
				Message:  "Privilege escalation not allowed",
				Success:  true,
				Severity: "danger",
				Category: "Security",
			}},
		},
		{
			name:         "strong security context + strong validation config",
			securityConf: strongConf,
			container:    strongContainer,
			pod:          emptyPodSpec,
			expectedResults: []ResultMessage{{
				ID:       "runAsRootAllowed",
				Message:  "Is not allowed to run as root",
				Success:  true,
				Severity: "danger",
				Category: "Security",
			}, {
				ID:       "notReadOnlyRootFilesystem",
				Message:  "Filesystem is read only",
				Success:  true,
				Severity: "danger",
				Category: "Security",
			}, {
				ID:       "runAsPrivileged",
				Message:  "Not running as privileged",
				Success:  true,
				Severity: "danger",
				Category: "Security",
			}, {
				ID:       "privilegeEscalationAllowed",
				Message:  "Privilege escalation not allowed",
				Success:  true,
				Severity: "danger",
				Category: "Security",
			}, {
				ID:       "dangerousCapabilities",
				Message:  "Container does not have any dangerous capabilities",
				Success:  true,
				Severity: "danger",
				Category: "Security",
			}, {
				ID:       "insecureCapabilities",
				Message:  "Container does not have any insecure capabilities",
				Success:  true,
				Severity: "danger",
				Category: "Security",
			}},
		},
		{
			name:         "strong security context + strong validation config via podspec default",
			securityConf: strongConf,
			container:    inheritContainer,
			pod:          goodPodSpec,
			expectedResults: []ResultMessage{{
				ID:       "runAsRootAllowed",
				Message:  "Is not allowed to run as root",
				Success:  true,
				Severity: "danger",
				Category: "Security",
			}, {
				ID:       "notReadOnlyRootFilesystem",
				Message:  "Filesystem is read only",
				Success:  true,
				Severity: "danger",
				Category: "Security",
			}, {
				ID:       "runAsPrivileged",
				Message:  "Not running as privileged",
				Success:  true,
				Severity: "danger",
				Category: "Security",
			}, {
				ID:       "privilegeEscalationAllowed",
				Message:  "Privilege escalation not allowed",
				Success:  true,
				Severity: "danger",
				Category: "Security",
			}, {
				ID:       "dangerousCapabilities",
				Message:  "Container does not have any dangerous capabilities",
				Success:  true,
				Severity: "danger",
				Category: "Security",
			}, {
				ID:       "insecureCapabilities",
				Message:  "Container does not have any insecure capabilities",
				Success:  true,
				Severity: "danger",
				Category: "Security",
			}},
		},
		{
			name:         "strong security context + strong validation config with bad setting in podspec default",
			securityConf: strongConf,
			container:    strongContainer,
			pod:          badPodSpec,
			expectedResults: []ResultMessage{{
				ID:       "runAsRootAllowed",
				Message:  "Is not allowed to run as root",
				Success:  true,
				Severity: "danger",
				Category: "Security",
			}, {
				ID:       "notReadOnlyRootFilesystem",
				Message:  "Filesystem is read only",
				Success:  true,
				Severity: "danger",
				Category: "Security",
			}, {
				ID:       "runAsPrivileged",
				Message:  "Not running as privileged",
				Success:  true,
				Severity: "danger",
				Category: "Security",
			}, {
				ID:       "privilegeEscalationAllowed",
				Message:  "Privilege escalation not allowed",
				Success:  true,
				Severity: "danger",
				Category: "Security",
			}, {
				ID:       "dangerousCapabilities",
				Message:  "Container does not have any dangerous capabilities",
				Success:  true,
				Severity: "danger",
				Category: "Security",
			}, {
				ID:       "insecureCapabilities",
				Message:  "Container does not have any insecure capabilities",
				Success:  true,
				Severity: "danger",
				Category: "Security",
			}},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			workload, err := kube.NewGenericResourceFromPod(corev1.Pod{Spec: *tt.pod}, nil)
			assert.NoError(t, err)
			results, err := applyContainerSchemaChecks(&conf.Configuration{Checks: tt.securityConf}, workload, tt.container, false)
			if err != nil {
				panic(err)
			}
			messages := []ResultMessage{}
			for _, msg := range results {
				messages = append(messages, msg)
			}
			assert.Len(t, messages, len(tt.expectedResults))
			assert.ElementsMatch(t, tt.expectedResults, messages)
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

	goodContainer := &corev1.Container{
		SecurityContext: &corev1.SecurityContext{
			RunAsNonRoot: &trueVar,
		},
	}
	badContainer := &corev1.Container{
		SecurityContext: &corev1.SecurityContext{
			RunAsNonRoot: &falseVar,
		},
	}
	inheritContainer := &corev1.Container{
		SecurityContext: &corev1.SecurityContext{
			RunAsNonRoot: nil,
		},
	}
	runAsUserContainer := &corev1.Container{
		SecurityContext: &corev1.SecurityContext{
			RunAsUser: &nonRootUser,
		},
	}
	runAsUser0Container := &corev1.Container{
		SecurityContext: &corev1.SecurityContext{
			RunAsUser: &rootUser,
		},
	}
	badPod := &corev1.PodSpec{
		SecurityContext: &corev1.PodSecurityContext{
			RunAsNonRoot: &falseVar,
		},
	}
	runAsUserPod := &corev1.PodSpec{
		SecurityContext: &corev1.PodSecurityContext{
			RunAsUser: &nonRootUser,
		},
	}
	emptyPod := &corev1.PodSpec{}

	testCases := []struct {
		name      string
		container *corev1.Container
		pod       *corev1.PodSpec
		message   ResultMessage
	}{
		{
			name:      "pod=false,container=nil",
			container: inheritContainer,
			pod:       badPod,
			message: ResultMessage{
				ID:       "runAsRootAllowed",
				Message:  "Should not be allowed to run as root",
				Success:  false,
				Severity: "warning",
				Category: "Security",
			},
		},
		{
			name:      "pod=false,container=true",
			container: goodContainer,
			pod:       badPod,
			message: ResultMessage{
				ID:       "runAsRootAllowed",
				Message:  "Is not allowed to run as root",
				Success:  true,
				Severity: "warning",
				Category: "Security",
			},
		},
		{
			name:      "pod=nil,container=runAsUser",
			container: runAsUserContainer,
			pod:       emptyPod,
			message: ResultMessage{
				ID:       "runAsRootAllowed",
				Message:  "Is not allowed to run as root",
				Success:  true,
				Severity: "warning",
				Category: "Security",
			},
		},
		{
			name:      "pod=runAsUser,container=nil",
			container: inheritContainer,
			pod:       runAsUserPod,
			message: ResultMessage{
				ID:       "runAsRootAllowed",
				Message:  "Is not allowed to run as root",
				Success:  true,
				Severity: "warning",
				Category: "Security",
			},
		},
		{
			name:      "pod=runAsUser,container=runAsUser0",
			container: runAsUser0Container,
			pod:       runAsUserPod,
			message: ResultMessage{
				ID:       "runAsRootAllowed",
				Message:  "Should not be allowed to run as root",
				Success:  false,
				Severity: "warning",
				Category: "Security",
			},
		},
		{
			name:      "pod=runAsUser,container=false",
			pod:       runAsUserPod,
			container: badContainer,
			message: ResultMessage{
				ID:       "runAsRootAllowed",
				Message:  "Should not be allowed to run as root",
				Success:  false,
				Severity: "warning",
				Category: "Security",
			},
		},
	}
	for idx, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			workload, err := kube.NewGenericResourceFromPod(corev1.Pod{Spec: *tt.pod}, nil)
			assert.NoError(t, err)
			results, err := applyContainerSchemaChecks(&config, workload, tt.container, false)
			if err != nil {
				panic(err)
			}
			messages := []ResultMessage{}
			for _, msg := range results {
				messages = append(messages, msg)
			}
			assert.Len(t, messages, 1)
			if len(messages) > 0 {
				assert.Equal(t, tt.message, messages[0], fmt.Sprintf("Test case %d failed", idx))
			}
		})
	}
}

func TestValidateResourcesExemption(t *testing.T) {
	container := corev1.Container{
		Name: "Empty",
	}

	expectedWarnings := []ResultMessage{}
	expectedDangers := []ResultMessage{}
	expectedSuccesses := []ResultMessage{}

	testValidate(t, &container, &resourceConfExemptions, "foo", expectedDangers, expectedWarnings, expectedSuccesses)

	expectedWarnings = []ResultMessage{
		{
			ID:       "cpuRequestsMissing",
			Success:  false,
			Severity: "warning",
			Message:  "CPU requests should be set",
			Category: "Efficiency",
		},
		{
			ID:       "memoryRequestsMissing",
			Success:  false,
			Severity: "warning",
			Message:  "Memory requests should be set",
			Category: "Efficiency",
		},
	}

	expectedDangers = []ResultMessage{
		{
			ID:       "cpuLimitsMissing",
			Success:  false,
			Severity: "danger",
			Message:  "CPU limits should be set",
			Category: "Efficiency",
		},
		{
			ID:       "memoryLimitsMissing",
			Success:  false,
			Severity: "danger",
			Message:  "Memory limits should be set",
			Category: "Efficiency",
		},
	}

	disallowExemptionsConf := resourceConfExemptions + "\ndisallowExemptions: true"

	testValidate(t, &container, &disallowExemptionsConf, "foo", expectedDangers, expectedWarnings, expectedSuccesses)
}

func TestValidateResourcesEmptyContainerCPURequestsExempt(t *testing.T) {
	container := corev1.Container{
		Name: "Empty",
	}

	expectedWarnings := []ResultMessage{
		{
			ID:       "memoryRequestsMissing",
			Success:  false,
			Severity: "warning",
			Message:  "Memory requests should be set",
			Category: "Efficiency",
		},
	}

	expectedDangers := []ResultMessage{
		{
			ID:       "cpuLimitsMissing",
			Success:  false,
			Severity: "danger",
			Message:  "CPU limits should be set",
			Category: "Efficiency",
		},
		{
			ID:       "memoryLimitsMissing",
			Success:  false,
			Severity: "danger",
			Message:  "Memory limits should be set",
			Category: "Efficiency",
		},
	}

	expectedSuccesses := []ResultMessage{}

	workload, err := kube.NewGenericResourceFromPod(corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "foo",
			Annotations: map[string]string{
				"polaris.fairwinds.com/cpuRequestsMissing-exempt":    "true",   // Exempt this controller from cpuRequestsMissing
				"polaris.fairwinds.com/memoryRequestsMissing-exempt": "truthy", // Don't actually exempt this controller from memoryRequestsMissing
			},
		},
	}, nil)
	assert.NoError(t, err)
	testValidateWithWorkload(t, &container, &resourceConfMinimal, workload, expectedDangers, expectedWarnings, expectedSuccesses)
}
