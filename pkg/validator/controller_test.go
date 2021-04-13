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
	"encoding/json"
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
			"hostIPCSet": conf.SeverityDanger,
			"hostPIDSet": conf.SeverityDanger,
		},
	}
	deployment, err := kube.NewGenericResourceFromPod(test.MockPod(), nil)
	assert.NoError(t, err)
	deployment.Kind = "Deployment"
	expectedSum := CountSummary{
		Successes: uint(2),
		Warnings:  uint(0),
		Dangers:   uint(0),
	}

	expectedResults := ResultSet{
		"hostIPCSet": {ID: "hostIPCSet", Message: "Host IPC is not configured", Success: true, Severity: "danger", Category: "Security"},
		"hostPIDSet": {ID: "hostPIDSet", Message: "Host PID is not configured", Success: true, Severity: "danger", Category: "Security"},
	}

	var actualResult Result
	actualResult, err = applyControllerSchemaChecks(&c, nil, deployment)
	if err != nil {
		panic(err)
	}

	assert.Equal(t, "Deployment", actualResult.Kind)
	assert.Equal(t, 1, len(actualResult.PodResult.ContainerResults), "should be equal")
	assert.EqualValues(t, expectedSum, actualResult.GetSummary())
	assert.EqualValues(t, expectedResults, actualResult.PodResult.Results)
}

func TestControllerLevelChecks(t *testing.T) {
	testResources := func(res *kube.ResourceProvider) {
		c := conf.Configuration{
			Checks: map[string]conf.Severity{
				"multipleReplicasForDeployment": conf.SeverityDanger,
			},
		}
		expectedResult := ResultMessage{
			ID:       "multipleReplicasForDeployment",
			Severity: "danger",
			Category: "Reliability",
		}
		for _, controller := range res.Resources["Deployment"] {
			actualResult, err := applyControllerSchemaChecks(&c, nil, controller)
			if err != nil {
				panic(err)
			}
			if controller.ObjectMeta.GetName() == "test-deployment-2" {
				expectedResult.Success = true
				expectedResult.Message = "Multiple replicas are scheduled"
			} else if controller.ObjectMeta.GetName() == "test-deployment" {
				expectedResult.Success = false
				expectedResult.Message = "Only one replica is scheduled"
			}
			expectedResults := ResultSet{
				"multipleReplicasForDeployment": expectedResult,
			}

			assert.Equal(t, "Deployment", actualResult.Kind)
			assert.Equal(t, 1, len(actualResult.Results), "should be equal")
			assert.EqualValues(t, expectedResults, actualResult.Results, controller.ObjectMeta.GetName())
		}
	}

	res, err := kube.CreateResourceProviderFromPath("../kube/test_files/test_1")
	assert.Equal(t, nil, err, "Error should be nil")
	assert.Equal(t, 11, res.Resources.GetLength())
	testResources(res)

	replicaSpec := map[string]interface{}{"replicas": 2}
	b, err := json.Marshal(replicaSpec)
	assert.NoError(t, err)
	err = json.Unmarshal(b, &replicaSpec)

	d1, p1 := test.MockDeploy("test", "test-deployment")
	d2, p2 := test.MockDeploy("test", "test-deployment-2")
	two := int32(2)
	d2.Spec.Replicas = &two
	k8s, dynamicClient := test.SetupTestAPI(&d1, &p1, &d2, &p2)
	res, err = kube.CreateResourceProviderFromAPI(context.Background(), k8s, "test", &dynamicClient, conf.Configuration{})
	assert.Equal(t, err, nil, "error should be nil")
	assert.Equal(t, 2, res.Resources.GetLength(), "Should have two controllers")
	testResources(res)
}

func TestSkipHealthChecks(t *testing.T) {
	c := conf.Configuration{
		Checks: map[string]conf.Severity{
			"readinessProbeMissing": conf.SeverityDanger,
			"livenessProbeMissing":  conf.SeverityWarning,
		},
	}
	pod := test.MockPod()
	pod.Spec.InitContainers = []corev1.Container{test.MockContainer("test")}
	deployment, err := kube.NewGenericResourceFromPod(pod, nil)
	assert.NoError(t, err)
	deployment.Kind = "Deployment"
	expectedSum := CountSummary{
		Successes: uint(0),
		Warnings:  uint(1),
		Dangers:   uint(1),
	}
	expectedResults := ResultSet{
		"readinessProbeMissing": {ID: "readinessProbeMissing", Message: "Readiness probe should be configured", Success: false, Severity: "danger", Category: "Reliability"},
		"livenessProbeMissing":  {ID: "livenessProbeMissing", Message: "Liveness probe should be configured", Success: false, Severity: "warning", Category: "Reliability"},
	}
	var actualResult Result
	actualResult, err = applyControllerSchemaChecks(&c, nil, deployment)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, "Deployment", actualResult.Kind)
	assert.Equal(t, 2, len(actualResult.PodResult.ContainerResults), "should be equal")
	assert.EqualValues(t, expectedSum, actualResult.GetSummary())
	assert.EqualValues(t, ResultSet{}, actualResult.PodResult.ContainerResults[0].Results)
	assert.EqualValues(t, expectedResults, actualResult.PodResult.ContainerResults[1].Results)

	job, err := kube.NewGenericResourceFromPod(test.MockPod(), nil)
	assert.NoError(t, err)
	job.Kind = "Job"
	expectedSum = CountSummary{
		Successes: uint(0),
		Warnings:  uint(0),
		Dangers:   uint(0),
	}
	expectedResults = ResultSet{}
	actualResult, err = applyControllerSchemaChecks(&c, nil, job)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, "Job", actualResult.Kind)
	assert.Equal(t, 1, len(actualResult.PodResult.ContainerResults), "should be equal")
	assert.EqualValues(t, expectedSum, actualResult.GetSummary())
	assert.EqualValues(t, expectedResults, actualResult.PodResult.ContainerResults[0].Results)

	cronjob, err := kube.NewGenericResourceFromPod(test.MockPod(), nil)
	assert.NoError(t, err)
	cronjob.Kind = "CronJob"
	expectedSum = CountSummary{
		Successes: uint(0),
		Warnings:  uint(0),
		Dangers:   uint(0),
	}
	expectedResults = ResultSet{}
	actualResult, err = applyControllerSchemaChecks(&c, nil, cronjob)
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
			"readinessProbeMissing": conf.SeverityDanger,
			"livenessProbeMissing":  conf.SeverityWarning,
		},
	}
	expectedSum := CountSummary{
		Successes: uint(0),
		Warnings:  uint(1),
		Dangers:   uint(1),
	}
	expectedExemptSum := CountSummary{
		Successes: uint(0),
		Warnings:  uint(0),
		Dangers:   uint(0),
	}

	pod := test.MockPod()
	pod.ObjectMeta.Namespace = "foo"
	workload, err := kube.NewGenericResourceFromPod(pod, nil)
	assert.NoError(t, err)
	workload.Kind = "Deployment"
	resources := []kube.GenericResource{workload}

	var actualResults []Result
	actualResults, err = ApplyAllSchemaChecksToAllResources(&c, nil, resources)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, 1, len(actualResults))
	assert.Equal(t, "Deployment", actualResults[0].Kind)
	assert.EqualValues(t, expectedSum, actualResults[0].GetSummary())

	c.Exemptions = []conf.Exemption{{
		Namespace: "foo",
	}}
	actualResults, err = ApplyAllSchemaChecksToAllResources(&c, nil, resources)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, 1, len(actualResults))
	assert.Equal(t, "Deployment", actualResults[0].Kind)
	assert.EqualValues(t, expectedExemptSum, actualResults[0].GetSummary())

	c.Exemptions = nil
	resources[0].ObjectMeta.SetAnnotations(map[string]string{
		exemptionAnnotationKey: "true",
	})
	actualResults, err = ApplyAllSchemaChecksToAllResources(&c, nil, resources)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, 1, len(actualResults))
	assert.Equal(t, "Deployment", actualResults[0].Kind)
	assert.EqualValues(t, expectedExemptSum, actualResults[0].GetSummary())

	c.DisallowExemptions = true
	actualResults, err = ApplyAllSchemaChecksToAllResources(&c, nil, resources)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, 1, len(actualResults))
	assert.Equal(t, "Deployment", actualResults[0].Kind)
	assert.EqualValues(t, expectedSum, actualResults[0].GetSummary())
}
