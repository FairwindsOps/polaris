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
	controller "github.com/fairwindsops/polaris/pkg/validator/controllers"
	corev1 "k8s.io/api/core/v1"

	"github.com/fairwindsops/polaris/test"
	"github.com/stretchr/testify/assert"
)

func TestValidateController(t *testing.T) {
	c := conf.Configuration{
		Security: conf.Security{
			HostIPCSet: conf.SeverityError,
			HostPIDSet: conf.SeverityError,
		},
	}
	deployment := controller.NewDeploymentController(test.MockDeploy())
	expectedSum := ResultSummary{
		Totals: CountSummary{
			Successes: uint(2),
			Warnings:  uint(0),
			Errors:    uint(0),
		},
		ByCategory: make(map[string]*CountSummary),
	}
	expectedSum.ByCategory["Security"] = &CountSummary{
		Successes: uint(2),
		Warnings:  uint(0),
		Errors:    uint(0),
	}

	expectedMessages := []*ResultMessage{
		{ID: "hostIPCSet", Message: "Host IPC is not configured", Type: "success", Category: "Security"},
		{ID: "hostPIDSet", Message: "Host PID is not configured", Type: "success", Category: "Security"},
	}

	actualResult := ValidateController(c, deployment)

	assert.Equal(t, "Deployments", actualResult.Type)
	assert.Equal(t, 1, len(actualResult.PodResult.ContainerResults), "should be equal")
	assert.EqualValues(t, &expectedSum, actualResult.PodResult.Summary)
	assert.EqualValues(t, expectedMessages, actualResult.PodResult.Messages)
}

func TestSkipHealthChecks(t *testing.T) {
	c := conf.Configuration{
		HealthChecks: conf.HealthChecks{
			ReadinessProbeMissing: conf.SeverityError,
			LivenessProbeMissing:  conf.SeverityWarning,
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
	expectedSum := ResultSummary{
		Totals: CountSummary{
			Successes: uint(0),
			Warnings:  uint(1),
			Errors:    uint(1),
		},
		ByCategory: make(map[string]*CountSummary),
	}
	expectedSum.ByCategory["Health Checks"] = &CountSummary{
		Successes: uint(0),
		Warnings:  uint(1),
		Errors:    uint(1),
	}
	expectedMessages := []*ResultMessage{
		{ID: "readinessProbeMissing", Message: "Readiness probe should be configured", Type: "error", Category: "Health Checks"},
		{ID: "livenessProbeMissing", Message: "Liveness probe should be configured", Type: "warning", Category: "Health Checks"},
	}
	actualResult := ValidateController(c, deployment)
	assert.Equal(t, "Deployments", actualResult.Type)
	assert.Equal(t, 2, len(actualResult.PodResult.ContainerResults), "should be equal")
	assert.EqualValues(t, &expectedSum, actualResult.PodResult.Summary)
	assert.EqualValues(t, []*ResultMessage{}, actualResult.PodResult.ContainerResults[0].Messages)
	assert.EqualValues(t, expectedMessages, actualResult.PodResult.ContainerResults[1].Messages)

	job := controller.NewJobController(test.MockJob())
	expectedSum = ResultSummary{
		Totals: CountSummary{
			Successes: uint(0),
			Warnings:  uint(0),
			Errors:    uint(0),
		},
		ByCategory: make(map[string]*CountSummary),
	}
	expectedMessages = []*ResultMessage{}
	actualResult = ValidateController(c, job)
	assert.Equal(t, "Jobs", actualResult.Type)
	assert.Equal(t, 1, len(actualResult.PodResult.ContainerResults), "should be equal")
	assert.EqualValues(t, &expectedSum, actualResult.PodResult.Summary)
	assert.EqualValues(t, expectedMessages, actualResult.PodResult.ContainerResults[0].Messages)

	cronjob := controller.NewCronJobController(test.MockCronJob())
	expectedSum = ResultSummary{
		Totals: CountSummary{
			Successes: uint(0),
			Warnings:  uint(0),
			Errors:    uint(0),
		},
		ByCategory: make(map[string]*CountSummary),
	}
	expectedMessages = []*ResultMessage{}
	actualResult = ValidateController(c, cronjob)
	assert.Equal(t, "CronJobs", actualResult.Type)
	assert.Equal(t, 1, len(actualResult.PodResult.ContainerResults), "should be equal")
	assert.EqualValues(t, &expectedSum, actualResult.PodResult.Summary)
	assert.EqualValues(t, expectedMessages, actualResult.PodResult.ContainerResults[0].Messages)
}
