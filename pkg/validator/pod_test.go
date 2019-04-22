// Copyright 2019 ReactiveOps
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

	conf "github.com/reactiveops/fairwinds/pkg/config"
	"github.com/reactiveops/fairwinds/test"
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
	k8s = test.SetupAddDeploys(k8s, "test")
	pod := test.MockPod()

	expectedSum := ResultSummary{
		Successes: uint(8),
		Warnings:  uint(0),
		Errors:    uint(0),
	}

	expectedMessages := []*ResultMessage{
		{Message: "Host IPC is not configured", Type: "success", Category: "Security"},
		{Message: "Host PID is not configured", Type: "success", Category: "Security"},
		{Message: "Host network is not configured", Type: "success", Category: "Networking"},
	}

	actualRR := ValidatePod(c, &pod.Spec)

	assert.Equal(t, actualRR.Type, "Pod", "should be equal")
	assert.Equal(t, len(actualRR.ContainerResults), 0, "should be equal")
	assert.EqualValues(t, actualRR.Summary, &expectedSum)
	assert.EqualValues(t, actualRR.PodResults[0].Messages, expectedMessages)
}
