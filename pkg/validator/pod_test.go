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

	conf "github.com/reactiveops/polaris/pkg/config"
	"github.com/reactiveops/polaris/test"
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
			Successes: uint(8),
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
	expectedSum.ByCategory["Resources"] = &CountSummary{
		Successes: uint(4),
		Warnings:  uint(0),
		Errors:    uint(0),
	}
	expectedSum.ByCategory["Security"] = &CountSummary{
		Successes: uint(2),
		Warnings:  uint(0),
		Errors:    uint(0),
	}

	expectedMessages := []*ResultMessage{
		{Message: "Host IPC is not configured", Type: "success", Category: "Security"},
		{Message: "Host PID is not configured", Type: "success", Category: "Security"},
		{Message: "Host network is not configured", Type: "success", Category: "Networking"},
	}

	actualPodResult := ValidatePod(c, &pod.Spec)

	assert.Equal(t, len(actualPodResult.ContainerResults), 1, "should be equal")
	assert.EqualValues(t, actualPodResult.Summary, &expectedSum)
	assert.EqualValues(t, actualPodResult.Messages, expectedMessages)
}
