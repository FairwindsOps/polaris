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

	testValidate(t, &container, &resourceConfRanges, "foo", expectedErrors, expectedWarnings, expectedSuccesses)
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

	testValidate(t, &container, &resourceConfRanges, "foo", []*ResultMessage{}, []*ResultMessage{}, expectedSuccesses)

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

	testValidate(t, &container, &resourceConfMinimal, "foo", []*ResultMessage{}, []*ResultMessage{}, expectedSuccesses)
}

func TestValidateCustomCheckExemptions(t *testing.T) {
	container := corev1.Container{
		Name:  "example",
		Image: "hub.docker.com/foo",
	}

	expectedWarnings := []*ResultMessage{}
	expectedErrors := []*ResultMessage{}
	expectedSuccesses := []*ResultMessage{}
	testValidate(t, &container, &customCheckExemptions, "exempt", expectedErrors, expectedWarnings, expectedSuccesses)

	expectedErrors = []*ResultMessage{
		{
			ID:       "foo",
			Type:     "error",
			Message:  "fail!",
			Category: "Security",
		},
	}
	testValidate(t, &container, &customCheckExemptions, "notexempt", expectedErrors, expectedWarnings, expectedSuccesses)
}
