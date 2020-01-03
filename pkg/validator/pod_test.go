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
		Checks: map[string]conf.Severity{
			"hostIPCSet":     conf.SeverityError,
			"hostPIDSet":     conf.SeverityError,
			"hostNetworkSet": conf.SeverityWarning,
			"hostPortSet":    conf.SeverityError,
		},
	}

	k8s := test.SetupTestAPI()
	k8s = test.SetupAddControllers(k8s, "test")
	pod := test.MockPod()

	expectedSum := CountSummary{
		Successes: uint(4),
		Warnings:  uint(0),
		Errors:    uint(0),
	}

	expectedMessages := ResultSet{
		"hostIPCSet":     {ID: "hostIPCSet", Message: "Host IPC is not configured", Success: true, Severity: "error", Category: "Security"},
		"hostNetworkSet": {ID: "hostNetworkSet", Message: "Host network is not configured", Success: true, Severity: "warning", Category: "Networking"},
		"hostPIDSet":     {ID: "hostPIDSet", Message: "Host PID is not configured", Success: true, Severity: "error", Category: "Security"},
	}

	actualPodResult := ValidatePod(&c, &pod.Spec, "", conf.Deployments)

	assert.Equal(t, 1, len(actualPodResult.ContainerResults), "should be equal")
	assert.EqualValues(t, expectedSum, actualPodResult.GetSummary())
	assert.EqualValues(t, expectedMessages, actualPodResult.Messages)
}

func TestInvalidIPCPod(t *testing.T) {
	c := conf.Configuration{
		Checks: map[string]conf.Severity{
			"hostIPCSet":     conf.SeverityError,
			"hostPIDSet":     conf.SeverityError,
			"hostNetworkSet": conf.SeverityWarning,
			"hostPortSet":    conf.SeverityError,
		},
	}

	k8s := test.SetupTestAPI()
	k8s = test.SetupAddControllers(k8s, "test")
	pod := test.MockPod()
	pod.Spec.HostIPC = true

	expectedSum := CountSummary{
		Successes: uint(3),
		Warnings:  uint(0),
		Errors:    uint(1),
	}
	expectedMessages := ResultSet{
		"hostIPCSet":     {ID: "hostIPCSet", Message: "Host IPC should not be configured", Success: false, Severity: "error", Category: "Security"},
		"hostNetworkSet": {ID: "hostNetworkSet", Message: "Host network is not configured", Success: true, Severity: "warning", Category: "Networking"},
		"hostPIDSet":     {ID: "hostPIDSet", Message: "Host PID is not configured", Success: true, Severity: "error", Category: "Security"},
	}

	actualPodResult := ValidatePod(&c, &pod.Spec, "", conf.Deployments)

	assert.Equal(t, 1, len(actualPodResult.ContainerResults), "should be equal")
	assert.EqualValues(t, expectedSum, actualPodResult.GetSummary())
	assert.EqualValues(t, expectedMessages, actualPodResult.Messages)
}

func TestInvalidNeworkPod(t *testing.T) {
	c := conf.Configuration{
		Checks: map[string]conf.Severity{
			"hostNetworkSet": conf.SeverityWarning,
			"hostPortSet":    conf.SeverityError,
			"hostIPCSet":     conf.SeverityError,
			"hostPIDSet":     conf.SeverityError,
		},
	}

	k8s := test.SetupTestAPI()
	k8s = test.SetupAddControllers(k8s, "test")
	pod := test.MockPod()
	pod.Spec.HostNetwork = true

	expectedSum := CountSummary{
		Successes: uint(3),
		Warnings:  uint(1),
		Errors:    uint(0),
	}

	expectedMessages := ResultSet{
		"hostNetworkSet": {ID: "hostNetworkSet", Message: "Host network should not be configured", Success: false, Severity: "warning", Category: "Networking"},
		"hostIPCSet":     {ID: "hostIPCSet", Message: "Host IPC is not configured", Success: true, Severity: "error", Category: "Security"},
		"hostPIDSet":     {ID: "hostPIDSet", Message: "Host PID is not configured", Success: true, Severity: "error", Category: "Security"},
	}

	actualPodResult := ValidatePod(&c, &pod.Spec, "", conf.Deployments)

	assert.Equal(t, 1, len(actualPodResult.ContainerResults), "should be equal")
	assert.EqualValues(t, expectedSum, actualPodResult.GetSummary())
	assert.EqualValues(t, expectedMessages, actualPodResult.Messages)
}

func TestInvalidPIDPod(t *testing.T) {
	c := conf.Configuration{
		Checks: map[string]conf.Severity{
			"hostIPCSet":     conf.SeverityError,
			"hostPIDSet":     conf.SeverityError,
			"hostNetworkSet": conf.SeverityWarning,
			"hostPortSet":    conf.SeverityError,
		},
	}

	k8s := test.SetupTestAPI()
	k8s = test.SetupAddControllers(k8s, "test")
	pod := test.MockPod()
	pod.Spec.HostPID = true

	expectedSum := CountSummary{
		Successes: uint(3),
		Warnings:  uint(0),
		Errors:    uint(1),
	}

	expectedMessages := ResultSet{
		"hostPIDSet":     {ID: "hostPIDSet", Message: "Host PID should not be configured", Success: false, Severity: "error", Category: "Security"},
		"hostIPCSet":     {ID: "hostIPCSet", Message: "Host IPC is not configured", Success: true, Severity: "error", Category: "Security"},
		"hostNetworkSet": {ID: "hostNetworkSet", Message: "Host network is not configured", Success: true, Severity: "warning", Category: "Networking"},
	}

	actualPodResult := ValidatePod(&c, &pod.Spec, "", conf.Deployments)

	assert.Equal(t, 1, len(actualPodResult.ContainerResults), "should be equal")
	assert.EqualValues(t, expectedSum, actualPodResult.GetSummary())
	assert.EqualValues(t, expectedMessages, actualPodResult.Messages)
}

func TestExemption(t *testing.T) {
	c := conf.Configuration{
		Checks: map[string]conf.Severity{
			"hostIPCSet":     conf.SeverityError,
			"hostNetworkSet": conf.SeverityWarning,
			"hostPIDSet":     conf.SeverityError,
			"hostPortSet":    conf.SeverityError,
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

	expectedSum := CountSummary{
		Successes: uint(3),
		Warnings:  uint(0),
		Errors:    uint(0),
	}
	expectedMessages := ResultSet{
		"hostNetworkSet": {ID: "hostNetworkSet", Message: "Host network is not configured", Success: true, Severity: "warning", Category: "Networking"},
		"hostPIDSet":     {ID: "hostPIDSet", Message: "Host PID is not configured", Success: true, Severity: "error", Category: "Security"},
	}

	actualPodResult := ValidatePod(&c, &pod.Spec, "foo", conf.Deployments)

	assert.Equal(t, 1, len(actualPodResult.ContainerResults), "should be equal")
	assert.EqualValues(t, expectedSum, actualPodResult.GetSummary())
	assert.EqualValues(t, expectedMessages, actualPodResult.Messages)
}
