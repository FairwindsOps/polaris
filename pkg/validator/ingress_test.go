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
	// "github.com/stretchr/testify/assert"
	// conf "github.com/fairwindsops/polaris/pkg/config"
	// "github.com/fairwindsops/polaris/pkg/kube"
	// "github.com/fairwindsops/polaris/test"
)

func TestValidateIngress(t *testing.T) {
	// c := conf.Configuration{
	// Checks: map[string]conf.Severity{
	// "hostIPCSet": conf.SeverityDanger,
	// "hostPIDSet": conf.SeverityDanger,
	// },
	// }
	// deployment, err := kube.NewGenericWorkloadFromPod(test.MockPod(), nil)
	// assert.NoError(t, err)
	// deployment.Kind = "Deployment"
	// expectedSum := CountSummary{
	// Successes: uint(2),
	// Warnings:  uint(0),
	// Dangers:   uint(0),
	// }

	// expectedResults := ResultSet{
	// "hostIPCSet": {ID: "hostIPCSet", Message: "Host IPC is not configured", Success: true, Severity: "danger", Category: "Security"},
	// "hostPIDSet": {ID: "hostPIDSet", Message: "Host PID is not configured", Success: true, Severity: "danger", Category: "Security"},
	// }

	// var actualResult Result
	// actualResult, err = ValidateController(&c, deployment)
	// if err != nil {
	// panic(err)
	// }

	// assert.Equal(t, "Deployment", actualResult.Kind)
	// assert.Equal(t, 1, len(actualResult.PodResult.ContainerResults), "should be equal")
	// assert.EqualValues(t, expectedSum, actualResult.GetSummary())
	// assert.EqualValues(t, expectedResults, actualResult.PodResult.Results)
}
