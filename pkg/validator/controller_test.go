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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	conf "github.com/fairwindsops/polaris/pkg/config"
	"github.com/fairwindsops/polaris/pkg/kube"
	controller "github.com/fairwindsops/polaris/pkg/validator/controllers"
	"github.com/fairwindsops/polaris/test"
)

func TestValidateController(t *testing.T) {
	c := conf.Configuration{
		Checks: map[string]conf.Severity{
			"hostIPCSet": conf.SeverityError,
			"hostPIDSet": conf.SeverityError,
		},
	}
	deployment := controller.NewDeploymentController(test.MockDeploy())
	expectedSum := CountSummary{
		Successes: uint(2),
		Warnings:  uint(0),
		Errors:    uint(0),
	}

	expectedMessages := ResultSet{
		"hostIPCSet": {ID: "hostIPCSet", Message: "Host IPC is not configured", Type: "success", Severity: "error", Category: "Security"},
		"hostPIDSet": {ID: "hostPIDSet", Message: "Host PID is not configured", Type: "success", Severity: "error", Category: "Security"},
	}

	actualResult := ValidateController(&c, deployment)

	assert.Equal(t, "Deployments", actualResult.Type)
	assert.Equal(t, 1, len(actualResult.PodResult.ContainerResults), "should be equal")
	assert.EqualValues(t, expectedSum, actualResult.GetSummary())
	assert.EqualValues(t, expectedMessages, actualResult.PodResult.Messages)
}

func TestSkipHealthChecks(t *testing.T) {
	c := conf.Configuration{
		Checks: map[string]conf.Severity{
			"readinessProbeMissing": conf.SeverityError,
			"livenessProbeMissing":  conf.SeverityWarning,
		},
		ControllersToScan: []conf.SupportedController{
			conf.Deployments,
			conf.StatefulSets,
			conf.DaemonSets,
			conf.Jobs,
			conf.CronJobs,
			conf.ReplicationControllers,
		},
	}
	deploymentBase := test.MockDeploy()
	deploymentBase.Spec.Template.Spec.InitContainers = []corev1.Container{test.MockContainer("test")}
	deployment := controller.NewDeploymentController(deploymentBase)
	expectedSum := CountSummary{
		Successes: uint(0),
		Warnings:  uint(1),
		Errors:    uint(1),
	}
	expectedMessages := ResultSet{
		"readinessProbeMissing": {ID: "readinessProbeMissing", Message: "Readiness probe should be configured", Type: "failure", Severity: "error", Category: "Health Checks"},
		"livenessProbeMissing":  {ID: "livenessProbeMissing", Message: "Liveness probe should be configured", Type: "failure", Severity: "warning", Category: "Health Checks"},
	}
	actualResult := ValidateController(&c, deployment)
	assert.Equal(t, "Deployments", actualResult.Type)
	assert.Equal(t, 2, len(actualResult.PodResult.ContainerResults), "should be equal")
	assert.EqualValues(t, expectedSum, actualResult.GetSummary())
	assert.EqualValues(t, ResultSet{}, actualResult.PodResult.ContainerResults[0].Messages)
	assert.EqualValues(t, expectedMessages, actualResult.PodResult.ContainerResults[1].Messages)

	job := controller.NewJobController(test.MockJob())
	expectedSum = CountSummary{
		Successes: uint(0),
		Warnings:  uint(0),
		Errors:    uint(0),
	}
	expectedMessages = ResultSet{}
	actualResult = ValidateController(&c, job)
	assert.Equal(t, "Jobs", actualResult.Type)
	assert.Equal(t, 1, len(actualResult.PodResult.ContainerResults), "should be equal")
	assert.EqualValues(t, expectedSum, actualResult.GetSummary())
	assert.EqualValues(t, expectedMessages, actualResult.PodResult.ContainerResults[0].Messages)

	cronjob := controller.NewCronJobController(test.MockCronJob())
	expectedSum = CountSummary{
		Successes: uint(0),
		Warnings:  uint(0),
		Errors:    uint(0),
	}
	expectedMessages = ResultSet{}
	actualResult = ValidateController(&c, cronjob)
	assert.Equal(t, "CronJobs", actualResult.Type)
	assert.Equal(t, 1, len(actualResult.PodResult.ContainerResults), "should be equal")
	assert.EqualValues(t, expectedSum, actualResult.GetSummary())
	assert.EqualValues(t, expectedMessages, actualResult.PodResult.ContainerResults[0].Messages)
}

func TestControllerExemptions(t *testing.T) {
	c := conf.Configuration{
		Checks: map[string]conf.Severity{
			"readinessProbeMissing": conf.SeverityError,
			"livenessProbeMissing":  conf.SeverityWarning,
		},
		ControllersToScan: []conf.SupportedController{
			conf.Deployments,
		},
	}
	resources := &kube.ResourceProvider{
		Deployments: []appsv1.Deployment{test.MockDeploy()},
	}

	expectedSum := CountSummary{
		Successes: uint(0),
		Warnings:  uint(1),
		Errors:    uint(1),
	}
	actualResults := ValidateControllers(&c, resources)
	assert.Equal(t, 1, len(actualResults))
	assert.Equal(t, "Deployments", actualResults[0].Type)
	assert.EqualValues(t, expectedSum, actualResults[0].GetSummary())

	resources.Deployments[0].ObjectMeta.Annotations = map[string]string{
		exemptionAnnotationKey: "true",
	}
	actualResults = ValidateControllers(&c, resources)
	assert.Equal(t, 0, len(actualResults))
}
