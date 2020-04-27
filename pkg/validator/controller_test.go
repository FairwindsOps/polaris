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
	corev1 "k8s.io/api/core/v1"

	conf "github.com/fairwindsops/polaris/pkg/config"
	"github.com/fairwindsops/polaris/pkg/kube"
	"github.com/fairwindsops/polaris/test"
)

func TestValidateController(t *testing.T) {
	c := conf.Configuration{
		Checks: map[string]conf.Severity{
			"hostIPCSet": conf.SeverityError,
			"hostPIDSet": conf.SeverityError,
		},
	}
	deployment := kube.NewGenericWorkload(test.MockPod(), nil, nil)
	deployment.Kind = "Deployment"
	expectedSum := CountSummary{
		Successes: uint(2),
		Warnings:  uint(0),
		Errors:    uint(0),
	}

	expectedResults := ResultSet{
		"hostIPCSet": {ID: "hostIPCSet", Message: "Host IPC is not configured", Success: true, Severity: "error", Category: "Security"},
		"hostPIDSet": {ID: "hostPIDSet", Message: "Host PID is not configured", Success: true, Severity: "error", Category: "Security"},
	}

	actualResult, err := ValidateController(&c, deployment)
	if err != nil {
		panic(err)
	}

	assert.Equal(t, "Deployment", actualResult.Kind)
	assert.Equal(t, 1, len(actualResult.PodResult.ContainerResults), "should be equal")
	assert.EqualValues(t, expectedSum, actualResult.GetSummary())
	assert.EqualValues(t, expectedResults, actualResult.PodResult.Results)
}

func TestSkipHealthChecks(t *testing.T) {
	c := conf.Configuration{
		Checks: map[string]conf.Severity{
			"readinessProbeMissing": conf.SeverityError,
			"livenessProbeMissing":  conf.SeverityWarning,
		},
	}
	pod := test.MockPod()
	pod.Spec.InitContainers = []corev1.Container{test.MockContainer("test")}
	deployment := kube.NewGenericWorkload(pod, nil, nil)
	deployment.Kind = "Deployment"
	expectedSum := CountSummary{
		Successes: uint(0),
		Warnings:  uint(1),
		Errors:    uint(1),
	}
	expectedResults := ResultSet{
		"readinessProbeMissing": {ID: "readinessProbeMissing", Message: "Readiness probe should be configured", Success: false, Severity: "error", Category: "Health Checks"},
		"livenessProbeMissing":  {ID: "livenessProbeMissing", Message: "Liveness probe should be configured", Success: false, Severity: "warning", Category: "Health Checks"},
	}
	actualResult, err := ValidateController(&c, deployment)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, "Deployment", actualResult.Kind)
	assert.Equal(t, 2, len(actualResult.PodResult.ContainerResults), "should be equal")
	assert.EqualValues(t, expectedSum, actualResult.GetSummary())
	assert.EqualValues(t, ResultSet{}, actualResult.PodResult.ContainerResults[0].Results)
	assert.EqualValues(t, expectedResults, actualResult.PodResult.ContainerResults[1].Results)

	job := kube.NewGenericWorkload(test.MockPod(), nil, nil)
	job.Kind = "Job"
	expectedSum = CountSummary{
		Successes: uint(0),
		Warnings:  uint(0),
		Errors:    uint(0),
	}
	expectedResults = ResultSet{}
	actualResult, err = ValidateController(&c, job)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, "Job", actualResult.Kind)
	assert.Equal(t, 1, len(actualResult.PodResult.ContainerResults), "should be equal")
	assert.EqualValues(t, expectedSum, actualResult.GetSummary())
	assert.EqualValues(t, expectedResults, actualResult.PodResult.ContainerResults[0].Results)

	cronjob := kube.NewGenericWorkload(test.MockPod(), nil, nil)
	cronjob.Kind = "CronJob"
	expectedSum = CountSummary{
		Successes: uint(0),
		Warnings:  uint(0),
		Errors:    uint(0),
	}
	expectedResults = ResultSet{}
	actualResult, err = ValidateController(&c, cronjob)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, "CronJob", actualResult.Kind)
	assert.Equal(t, 1, len(actualResult.PodResult.ContainerResults), "should be equal")
	assert.EqualValues(t, expectedSum, actualResult.GetSummary())
	assert.EqualValues(t, expectedResults, actualResult.PodResult.ContainerResults[0].Results)
}

func TestControllerExemptions(t *testing.T) {
	c := conf.Configuration{
		Checks: map[string]conf.Severity{
			"readinessProbeMissing": conf.SeverityError,
			"livenessProbeMissing":  conf.SeverityWarning,
		},
	}
	pod := test.MockPod()
	workload := kube.NewGenericWorkload(pod, nil, nil)
	workload.Kind = "Deployment"
	resources := &kube.ResourceProvider{
		Controllers: []kube.GenericWorkload{workload},
	}

	expectedSum := CountSummary{
		Successes: uint(0),
		Warnings:  uint(1),
		Errors:    uint(1),
	}
	actualResults, err := ValidateControllers(&c, resources)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, 1, len(actualResults))
	assert.Equal(t, "Deployment", actualResults[0].Kind)
	assert.EqualValues(t, expectedSum, actualResults[0].GetSummary())

	resources.Controllers[0].ObjectMeta.SetAnnotations(map[string]string{
		exemptionAnnotationKey: "true",
	})
	actualResults, err = ValidateControllers(&c, resources)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, 0, len(actualResults))
}
