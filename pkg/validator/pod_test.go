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

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	conf "github.com/fairwindsops/polaris/pkg/config"
	"github.com/fairwindsops/polaris/pkg/kube"
	"github.com/fairwindsops/polaris/test"
)

func TestValidatePod(t *testing.T) {
	c := conf.Configuration{
		Checks: map[string]conf.Severity{
			"hostIPCSet":     conf.SeverityDanger,
			"hostPIDSet":     conf.SeverityDanger,
			"hostNetworkSet": conf.SeverityWarning,
			"hostPortSet":    conf.SeverityDanger,
		},
	}

	p := test.MockPod()
	deployment, err := kube.NewGenericResourceFromPod(p, nil)
	assert.NoError(t, err)
	expectedSum := CountSummary{
		Successes: uint(4),
		Warnings:  uint(0),
		Dangers:   uint(0),
	}

	expectedResults := ResultSet{
		"hostIPCSet":     {ID: "hostIPCSet", Message: "Host IPC is not configured", Success: true, Severity: "danger", Category: "Security"},
		"hostNetworkSet": {ID: "hostNetworkSet", Message: "Host network is not configured", Success: true, Severity: "warning", Category: "Security"},
		"hostPIDSet":     {ID: "hostPIDSet", Message: "Host PID is not configured", Success: true, Severity: "danger", Category: "Security"},
	}

	actualPodResult, err := applyControllerSchemaChecks(&c, deployment)
	if err != nil {
		panic(err)
	}

	assert.Equal(t, 1, len(actualPodResult.PodResult.ContainerResults), "should be equal")
	assert.EqualValues(t, expectedSum, actualPodResult.GetSummary())
	assert.EqualValues(t, expectedResults, actualPodResult.PodResult.Results)
}

func TestInvalidIPCPod(t *testing.T) {
	c := conf.Configuration{
		Checks: map[string]conf.Severity{
			"hostIPCSet":     conf.SeverityDanger,
			"hostPIDSet":     conf.SeverityDanger,
			"hostNetworkSet": conf.SeverityWarning,
			"hostPortSet":    conf.SeverityDanger,
		},
	}

	p := test.MockPod()
	p.Spec.HostIPC = true
	workload, err := kube.NewGenericResourceFromPod(p, nil)
	assert.NoError(t, err)
	expectedSum := CountSummary{
		Successes: uint(3),
		Warnings:  uint(0),
		Dangers:   uint(1),
	}
	expectedResults := ResultSet{
		"hostIPCSet":     {ID: "hostIPCSet", Message: "Host IPC should not be configured", Success: false, Severity: "danger", Category: "Security"},
		"hostNetworkSet": {ID: "hostNetworkSet", Message: "Host network is not configured", Success: true, Severity: "warning", Category: "Security"},
		"hostPIDSet":     {ID: "hostPIDSet", Message: "Host PID is not configured", Success: true, Severity: "danger", Category: "Security"},
	}

	actualPodResult, err := applyControllerSchemaChecks(&c, workload)
	if err != nil {
		panic(err)
	}

	assert.Equal(t, 1, len(actualPodResult.PodResult.ContainerResults), "should be equal")
	assert.EqualValues(t, expectedSum, actualPodResult.GetSummary())
	assert.EqualValues(t, expectedResults, actualPodResult.PodResult.Results)
}

func TestInvalidNetworkPod(t *testing.T) {
	c := conf.Configuration{
		Checks: map[string]conf.Severity{
			"hostNetworkSet": conf.SeverityWarning,
			"hostPortSet":    conf.SeverityDanger,
			"hostIPCSet":     conf.SeverityDanger,
			"hostPIDSet":     conf.SeverityDanger,
		},
	}

	p := test.MockPod()
	p.Spec.HostNetwork = true
	workload, err := kube.NewGenericResourceFromPod(p, nil)
	assert.NoError(t, err)
	expectedSum := CountSummary{
		Successes: uint(3),
		Warnings:  uint(1),
		Dangers:   uint(0),
	}

	expectedResults := ResultSet{
		"hostNetworkSet": {ID: "hostNetworkSet", Message: "Host network should not be configured", Success: false, Severity: "warning", Category: "Security"},
		"hostIPCSet":     {ID: "hostIPCSet", Message: "Host IPC is not configured", Success: true, Severity: "danger", Category: "Security"},
		"hostPIDSet":     {ID: "hostPIDSet", Message: "Host PID is not configured", Success: true, Severity: "danger", Category: "Security"},
	}

	actualPodResult, err := applyControllerSchemaChecks(&c, workload)
	if err != nil {
		panic(err)
	}

	assert.Equal(t, 1, len(actualPodResult.PodResult.ContainerResults), "should be equal")
	assert.EqualValues(t, expectedSum, actualPodResult.GetSummary())
	assert.EqualValues(t, expectedResults, actualPodResult.PodResult.Results)
}

func TestInvalidPIDPod(t *testing.T) {
	c := conf.Configuration{
		Checks: map[string]conf.Severity{
			"hostIPCSet":     conf.SeverityDanger,
			"hostPIDSet":     conf.SeverityDanger,
			"hostNetworkSet": conf.SeverityWarning,
			"hostPortSet":    conf.SeverityDanger,
		},
	}

	p := test.MockPod()
	p.Spec.HostPID = true
	workload, err := kube.NewGenericResourceFromPod(p, nil)
	assert.NoError(t, err)
	expectedSum := CountSummary{
		Successes: uint(3),
		Warnings:  uint(0),
		Dangers:   uint(1),
	}

	expectedResults := ResultSet{
		"hostPIDSet":     {ID: "hostPIDSet", Message: "Host PID should not be configured", Success: false, Severity: "danger", Category: "Security"},
		"hostIPCSet":     {ID: "hostIPCSet", Message: "Host IPC is not configured", Success: true, Severity: "danger", Category: "Security"},
		"hostNetworkSet": {ID: "hostNetworkSet", Message: "Host network is not configured", Success: true, Severity: "warning", Category: "Security"},
	}

	actualPodResult, err := applyControllerSchemaChecks(&c, workload)
	if err != nil {
		panic(err)
	}

	assert.Equal(t, 1, len(actualPodResult.PodResult.ContainerResults), "should be equal")
	assert.EqualValues(t, expectedSum, actualPodResult.GetSummary())
	assert.EqualValues(t, expectedResults, actualPodResult.PodResult.Results)
}

func TestExemption(t *testing.T) {
	c := conf.Configuration{
		Checks: map[string]conf.Severity{
			"hostIPCSet":     conf.SeverityDanger,
			"hostNetworkSet": conf.SeverityWarning,
			"hostPIDSet":     conf.SeverityDanger,
			"hostPortSet":    conf.SeverityDanger,
		},
		Exemptions: []conf.Exemption{
			{
				Rules:           []string{"hostIPCSet"},
				ControllerNames: []string{"foo"},
			},
		},
	}

	p := test.MockPod()
	p.Spec.HostIPC = true
	p.ObjectMeta = metav1.ObjectMeta{
		Name: "foo",
	}
	workload, err := kube.NewGenericResourceFromPod(p, nil)
	assert.NoError(t, err)
	expectedSum := CountSummary{
		Successes: uint(3),
		Warnings:  uint(0),
		Dangers:   uint(0),
	}
	expectedResults := ResultSet{
		"hostNetworkSet": {ID: "hostNetworkSet", Message: "Host network is not configured", Success: true, Severity: "warning", Category: "Security"},
		"hostPIDSet":     {ID: "hostPIDSet", Message: "Host PID is not configured", Success: true, Severity: "danger", Category: "Security"},
	}

	actualPodResult, err := applyControllerSchemaChecks(&c, workload)
	if err != nil {
		panic(err)
	}

	assert.Equal(t, 1, len(actualPodResult.PodResult.ContainerResults), "should be equal")
	assert.EqualValues(t, expectedSum, actualPodResult.GetSummary())
	assert.EqualValues(t, expectedResults, actualPodResult.PodResult.Results)
}
