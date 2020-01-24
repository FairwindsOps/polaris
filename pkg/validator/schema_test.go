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
  foo: error
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
			Category: "Resources",
		},
	}

	expectedErrors := []ResultMessage{
		{
			ID:       "memoryRequestsRange",
			Success:  false,
			Severity: "error",
			Message:  "Memory requests should be within the required range",
			Category: "Resources",
		},
	}

	expectedSuccesses := []ResultMessage{}

	testValidate(t, &container, &resourceConfRanges, "foo", expectedErrors, expectedWarnings, expectedSuccesses)
}

func TestValidateResourcesInit(t *testing.T) {
	emptyContainer := &corev1.Container{}
	controller := getEmptyController("")

	parsedConf, err := conf.Parse([]byte(resourceConfRanges))
	assert.NoError(t, err, "Expected no error when parsing config")

	results, err := applyContainerSchemaChecks(&parsedConf, controller, emptyContainer, false)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, uint(1), results.GetSummary().Errors)
	assert.Equal(t, uint(1), results.GetSummary().Warnings)

	results, err = applyContainerSchemaChecks(&parsedConf, controller, emptyContainer, true)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, uint(0), results.GetSummary().Errors)
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
			Severity: "error",
			Message:  "Memory requests are within the required range",
			Category: "Resources",
		},
		{
			ID:       "memoryLimitsRange",
			Success:  true,
			Severity: "warning",
			Message:  "Memory limits are within the required range",
			Category: "Resources",
		},
	}

	testValidate(t, &container, &resourceConfRanges, "foo", []ResultMessage{}, []ResultMessage{}, expectedSuccesses)

	expectedSuccesses = []ResultMessage{
		{
			ID:       "cpuRequestsMissing",
			Success:  true,
			Severity: "warning",
			Message:  "CPU requests are set",
			Category: "Resources",
		},
		{
			ID:       "memoryRequestsMissing",
			Success:  true,
			Severity: "warning",
			Message:  "Memory requests are set",
			Category: "Resources",
		},
		{
			ID:       "cpuLimitsMissing",
			Success:  true,
			Severity: "error",
			Message:  "CPU limits are set",
			Category: "Resources",
		},
		{
			ID:       "memoryLimitsMissing",
			Success:  true,
			Severity: "error",
			Message:  "Memory limits are set",
			Category: "Resources",
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
	expectedErrors := []ResultMessage{}
	expectedSuccesses := []ResultMessage{}
	testValidate(t, &container, &customCheckExemptions, "exempt", expectedErrors, expectedWarnings, expectedSuccesses)

	expectedErrors = []ResultMessage{
		{
			ID:       "foo",
			Success:  false,
			Severity: "error",
			Message:  "fail!",
			Category: "Security",
		},
	}
	testValidate(t, &container, &customCheckExemptions, "notexempt", expectedErrors, expectedWarnings, expectedSuccesses)
}

func TestGetExemptKey(t *testing.T) {
	keyMap := map[string]string {
		"hostIPCSet": "polaris.fairwinds.com/host-ipc-set-exempt",
		"hostPIDSet": "polaris.fairwinds.com/host-pid-set-exempt",
		"hostNetworkSet": "polaris.fairwinds.com/host-network-set-exempt",
		"memoryLimitsMissing": "polaris.fairwinds.com/memory-limits-missing-exempt",
		"memoryRequestsMissing": "polaris.fairwinds.com/memory-requests-missing-exempt",
		"cpuLimitsMissing": "polaris.fairwinds.com/cpu-limits-missing-exempt",
		"cpuRequestsMissing": "polaris.fairwinds.com/cpu-requests-missing-exempt",
		"readinessProbeMissing": "polaris.fairwinds.com/readiness-probe-missing-exempt",
		"livenessProbeMissing": "polaris.fairwinds.com/liveness-probe-missing-exempt",
		"pullPolicyNotAlways": "polaris.fairwinds.com/pull-policy-not-always-exempt",
		"tagNotSpecified": "polaris.fairwinds.com/tag-not-specified-exempt",
		"hostPortSet": "polaris.fairwinds.com/host-port-set-exempt",
		"runAsRootAllowed": "polaris.fairwinds.com/run-as-root-allowed-exempt",
		"runAsPrivileged": "polaris.fairwinds.com/run-as-privileged-exempt",
		"notReadOnlyRootFileSystem": "polaris.fairwinds.com/not-read-only-root-file-system-exempt",
		"privilegeEscalationAllowed": "polaris.fairwinds.com/privilege-escalation-allowed-exempt",
		"dangerousCapabilities": "polaris.fairwinds.com/dangerous-capabilities-exempt",
		"insecureCapabilities": "polaris.fairwinds.com/insecure-capabilities-exempt",
	}
	for id, key := range keyMap {
		exemptKey := getExemptKey(id)
		assert.Equal(t, key, exemptKey)
	}
	

}