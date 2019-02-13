package validator

import (
	"testing"

	conf "github.com/reactiveops/fairwinds/pkg/config"
	"github.com/reactiveops/fairwinds/test"
	"github.com/stretchr/testify/assert"
)

func TestValidatePod(t *testing.T) {
	r := conf.ResourceRequire{Require: true}
	c := conf.Configuration{
		HostNetworking: conf.HostNetworking{
			HostAlias:   r,
			HostIPC:     r,
			HostNetwork: r,
			HostPID:     r,
			HostPort:    r,
		},
	}

	k8s := test.SetupTestAPI()
	k8s = test.SetupAddDeploys(k8s, "test")
	pod := test.MockPod()

	expectedSum := ResultSummary{
		Successes: uint(9),
		Warnings:  uint(0),
		Failures:  uint(0),
	}

	expectedMssgs := []ResultMessage{
		ResultMessage{Message: "Host alias should not be configured", Type: "success"},
		ResultMessage{Message: "Host IPC should not be configured", Type: "success"},
		ResultMessage{Message: "Host PID should not be configured", Type: "success"},
		ResultMessage{Message: "Host network sould not be configured", Type: "success"},
	}

	actualRR := ValidatePod(c, &pod.Spec)

	assert.Equal(t, actualRR.Type, "Pod", "should be equal")
	assert.Equal(t, len(actualRR.ContainerResults), 0, "should be equal")
	assert.EqualValues(t, actualRR.Summary, &expectedSum)
	assert.EqualValues(t, actualRR.PodResults[0].Messages, expectedMssgs)
}
