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
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
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
			"hostPathSet":    conf.SeverityWarning,
			"procMount":      conf.SeverityWarning,
			"hostProcess":    conf.SeverityWarning,
		},
	}

	p := test.MockPod()
	deployment, err := kube.NewGenericResourceFromPod(p, nil)
	assert.NoError(t, err)
	expectedSum := CountSummary{
		Successes: uint(7),
		Warnings:  uint(0),
		Dangers:   uint(0),
	}

	expectedResults := ResultSet{
		"hostIPCSet":     {ID: "hostIPCSet", Message: "Host IPC is not configured", Success: true, Severity: "danger", Category: "Security"},
		"hostNetworkSet": {ID: "hostNetworkSet", Message: "Host network is not configured", Success: true, Severity: "warning", Category: "Security"},
		"hostPIDSet":     {ID: "hostPIDSet", Message: "Host PID is not configured", Success: true, Severity: "danger", Category: "Security"},
		"hostPathSet":    {ID: "hostPathSet", Message: "HostPath volumes are not configured", Success: true, Severity: "warning", Category: "Security"},
		"procMount":      {ID: "procMount", Message: "The default /proc masks are set up to reduce attack surface, and should be required", Success: true, Severity: "warning", Category: "Security"},
		"hostProcess":    {ID: "hostProcess", Message: "Privileged access to the host check is valid", Success: true, Severity: "warning", Category: "Security"},
	}

	actualPodResult, err := applyControllerSchemaChecks(context.Background(), &c, nil, deployment)
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
			"hostPathSet":    conf.SeverityWarning,
			"procMount":      conf.SeverityWarning,
			"hostProcess":    conf.SeverityWarning,
		},
	}

	p := test.MockPod()
	p.Spec.HostIPC = true
	p.Spec.Volumes = append(p.Spec.Volumes, v1.Volume{
		Name: "hostpath",
		VolumeSource: v1.VolumeSource{
			HostPath: &v1.HostPathVolumeSource{
				Path: "/var/run/docker.sock",
			},
		},
	})
	procMount := v1.UnmaskedProcMount
	p.Spec.Containers[0].SecurityContext = &v1.SecurityContext{
		ProcMount: &procMount,
	}
	hostProcess := true
	p.Spec.Containers[0].SecurityContext.WindowsOptions = &v1.WindowsSecurityContextOptions{
		HostProcess: &hostProcess,
	}

	workload, err := kube.NewGenericResourceFromPod(p, nil)
	assert.NoError(t, err)
	expectedSum := CountSummary{
		Successes: uint(3),
		Warnings:  uint(3),
		Dangers:   uint(1),
	}
	expectedResults := ResultSet{
		"hostIPCSet":     {ID: "hostIPCSet", Message: "Host IPC should not be configured", Success: false, Severity: "danger", Category: "Security"},
		"hostNetworkSet": {ID: "hostNetworkSet", Message: "Host network is not configured", Success: true, Severity: "warning", Category: "Security"},
		"hostPIDSet":     {ID: "hostPIDSet", Message: "Host PID is not configured", Success: true, Severity: "danger", Category: "Security"},
		"hostPathSet":    {ID: "hostPathSet", Message: "HostPath volumes must be forbidden", Success: false, Severity: "warning", Category: "Security"},
		"procMount":      {ID: "procMount", Message: "Proc mount must not be changed from the default", Success: false, Severity: "warning", Category: "Security"},
		"hostProcess":    {ID: "hostProcess", Message: "Privileged access to the host is disallowed", Success: false, Severity: "warning", Category: "Security"},
	}

	actualPodResult, err := applyControllerSchemaChecks(context.Background(), &c, nil, workload)
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

	actualPodResult, err := applyControllerSchemaChecks(context.Background(), &c, nil, workload)
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

	actualPodResult, err := applyControllerSchemaChecks(context.Background(), &c, nil, workload)
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

	actualPodResult, err := applyControllerSchemaChecks(context.Background(), &c, nil, workload)
	if err != nil {
		panic(err)
	}

	assert.Equal(t, 1, len(actualPodResult.PodResult.ContainerResults), "should be equal")
	assert.EqualValues(t, expectedSum, actualPodResult.GetSummary())
	assert.EqualValues(t, expectedResults, actualPodResult.PodResult.Results)
}
