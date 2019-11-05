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
	"github.com/fairwindsops/polaris/test"
	"github.com/stretchr/testify/assert"
)

func TestValidatePod(t *testing.T) {
	c := conf.Configuration{
		Security: conf.Security{
			HostIPCSet: conf.SeverityError,
			HostPIDSet: conf.SeverityError,
		},
		Networking: conf.Networking{
			HostNetworkSet: conf.SeverityWarning,
			HostPortSet:    conf.SeverityError,
		},
	}

	k8s := test.SetupTestAPI()
	k8s = test.SetupAddControllers(k8s, "test")
	pod := test.MockPod()

	expectedSum := ResultSummary{
		Totals: CountSummary{
			Successes: uint(4),
			Warnings:  uint(0),
			Errors:    uint(0),
		},
		ByCategory: make(map[string]*CountSummary),
	}
	expectedSum.ByCategory["Networking"] = &CountSummary{
		Successes: uint(2),
		Warnings:  uint(0),
		Errors:    uint(0),
	}
	expectedSum.ByCategory["Security"] = &CountSummary{
		Successes: uint(2),
		Warnings:  uint(0),
		Errors:    uint(0),
	}

	expectedMessages := []*ResultMessage{
		{ID: "hostIPCSet", Message: "Host IPC is not configured", Type: "success", Category: "Security"},
		{ID: "hostPIDSet", Message: "Host PID is not configured", Type: "success", Category: "Security"},
		{ID: "hostNetworkSet", Message: "Host network is not configured", Type: "success", Category: "Networking"},
	}

	actualPodResult := ValidatePod(c, &pod.Spec, "", conf.Deployments)

	assert.Equal(t, 1, len(actualPodResult.ContainerResults), "should be equal")
	assert.EqualValues(t, &expectedSum, actualPodResult.Summary)
	assert.EqualValues(t, expectedMessages, actualPodResult.Messages)
}

func TestInvalidIPCPod(t *testing.T) {
	c := conf.Configuration{
		Security: conf.Security{
			HostIPCSet: conf.SeverityError,
			HostPIDSet: conf.SeverityError,
		},
		Networking: conf.Networking{
			HostNetworkSet: conf.SeverityWarning,
			HostPortSet:    conf.SeverityError,
		},
	}

	k8s := test.SetupTestAPI()
	k8s = test.SetupAddControllers(k8s, "test")
	pod := test.MockPod()
	pod.Spec.HostIPC = true

	expectedSum := ResultSummary{
		Totals: CountSummary{
			Successes: uint(3),
			Warnings:  uint(0),
			Errors:    uint(1),
		},
		ByCategory: make(map[string]*CountSummary),
	}
	expectedSum.ByCategory["Networking"] = &CountSummary{
		Successes: uint(2),
		Warnings:  uint(0),
		Errors:    uint(0),
	}
	expectedSum.ByCategory["Security"] = &CountSummary{
		Successes: uint(1),
		Warnings:  uint(0),
		Errors:    uint(1),
	}
	expectedMessages := []*ResultMessage{
		{ID: "hostIPCSet", Message: "Host IPC should not be configured", Type: "error", Category: "Security"},
		{ID: "hostPIDSet", Message: "Host PID is not configured", Type: "success", Category: "Security"},
		{ID: "hostNetworkSet", Message: "Host network is not configured", Type: "success", Category: "Networking"},
	}

	actualPodResult := ValidatePod(c, &pod.Spec, "", conf.Deployments)

	assert.Equal(t, 1, len(actualPodResult.ContainerResults), "should be equal")
	assert.EqualValues(t, &expectedSum, actualPodResult.Summary)
	assert.EqualValues(t, expectedMessages, actualPodResult.Messages)
}

func TestInvalidNeworkPod(t *testing.T) {
	c := conf.Configuration{
		Networking: conf.Networking{
			HostNetworkSet: conf.SeverityWarning,
			HostPortSet:    conf.SeverityError,
		},
		Security: conf.Security{
			HostIPCSet: conf.SeverityError,
			HostPIDSet: conf.SeverityError,
		},
	}

	k8s := test.SetupTestAPI()
	k8s = test.SetupAddControllers(k8s, "test")
	pod := test.MockPod()
	pod.Spec.HostNetwork = true

	expectedSum := ResultSummary{
		Totals: CountSummary{
			Successes: uint(3),
			Warnings:  uint(1),
			Errors:    uint(0),
		},
		ByCategory: make(map[string]*CountSummary),
	}
	expectedSum.ByCategory["Networking"] = &CountSummary{
		Successes: uint(1),
		Warnings:  uint(1),
		Errors:    uint(0),
	}

	expectedSum.ByCategory["Security"] = &CountSummary{
		Successes: uint(2),
		Warnings:  uint(0),
		Errors:    uint(0),
	}

	expectedMessages := []*ResultMessage{
		{ID: "hostNetworkSet", Message: "Host network should not be configured", Type: "warning", Category: "Networking"},
		{ID: "hostIPCSet", Message: "Host IPC is not configured", Type: "success", Category: "Security"},
		{ID: "hostPIDSet", Message: "Host PID is not configured", Type: "success", Category: "Security"},
	}

	actualPodResult := ValidatePod(c, &pod.Spec, "", conf.Deployments)

	assert.Equal(t, 1, len(actualPodResult.ContainerResults), "should be equal")
	assert.EqualValues(t, &expectedSum, actualPodResult.Summary)
	assert.EqualValues(t, expectedMessages, actualPodResult.Messages)
}

func TestInvalidPIDPod(t *testing.T) {
	c := conf.Configuration{
		Security: conf.Security{
			HostIPCSet: conf.SeverityError,
			HostPIDSet: conf.SeverityError,
		},
		Networking: conf.Networking{
			HostNetworkSet: conf.SeverityWarning,
			HostPortSet:    conf.SeverityError,
		},
	}

	k8s := test.SetupTestAPI()
	k8s = test.SetupAddControllers(k8s, "test")
	pod := test.MockPod()
	pod.Spec.HostPID = true

	expectedSum := ResultSummary{
		Totals: CountSummary{
			Successes: uint(3),
			Warnings:  uint(0),
			Errors:    uint(1),
		},
		ByCategory: make(map[string]*CountSummary),
	}
	expectedSum.ByCategory["Networking"] = &CountSummary{
		Successes: uint(2),
		Warnings:  uint(0),
		Errors:    uint(0),
	}
	expectedSum.ByCategory["Security"] = &CountSummary{
		Successes: uint(1),
		Warnings:  uint(0),
		Errors:    uint(1),
	}

	expectedMessages := []*ResultMessage{
		{ID: "hostPIDSet", Message: "Host PID should not be configured", Type: "error", Category: "Security"},
		{ID: "hostIPCSet", Message: "Host IPC is not configured", Type: "success", Category: "Security"},
		{ID: "hostNetworkSet", Message: "Host network is not configured", Type: "success", Category: "Networking"},
	}

	actualPodResult := ValidatePod(c, &pod.Spec, "", conf.Deployments)

	assert.Equal(t, 1, len(actualPodResult.ContainerResults), "should be equal")
	assert.EqualValues(t, &expectedSum, actualPodResult.Summary)
	assert.EqualValues(t, expectedMessages, actualPodResult.Messages)
}

func TestExemption(t *testing.T) {
	c := conf.Configuration{
		Security: conf.Security{
			HostIPCSet: conf.SeverityError,
			HostPIDSet: conf.SeverityError,
		},
		Networking: conf.Networking{
			HostNetworkSet: conf.SeverityWarning,
			HostPortSet:    conf.SeverityError,
		},
		Exemptions: []conf.Exemption{
			conf.Exemption{
				Rules:           []string{"hostIPCSet"},
				ControllerNames: []string{"foo"},
			},
		},
	}

	k8s := test.SetupTestAPI()
	k8s = test.SetupAddControllers(k8s, "test")
	pod := test.MockPod()
	pod.Spec.HostIPC = true

	expectedSum := ResultSummary{
		Totals: CountSummary{
			Successes: uint(3),
			Warnings:  uint(0),
			Errors:    uint(0),
		},
		ByCategory: make(map[string]*CountSummary),
	}
	expectedSum.ByCategory["Networking"] = &CountSummary{
		Successes: uint(2),
		Warnings:  uint(0),
		Errors:    uint(0),
	}
	expectedSum.ByCategory["Security"] = &CountSummary{
		Successes: uint(1),
		Warnings:  uint(0),
		Errors:    uint(0),
	}
	expectedMessages := []*ResultMessage{
		{ID: "hostPIDSet", Message: "Host PID is not configured", Type: "success", Category: "Security"},
		{ID: "hostNetworkSet", Message: "Host network is not configured", Type: "success", Category: "Networking"},
	}

	actualPodResult := ValidatePod(c, &pod.Spec, "foo", conf.Deployments)

	assert.Equal(t, 1, len(actualPodResult.ContainerResults), "should be equal")
	assert.EqualValues(t, &expectedSum, actualPodResult.Summary)
	assert.EqualValues(t, expectedMessages, actualPodResult.Messages)
}
