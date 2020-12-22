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

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var confContainerTest = `
checks:
  multipleReplicasForDeployment: warning
  priorityClassNotSet: warning
  pullPolicyNotAlways: warning
exemptions:
  - namespace: prometheus
    rules:
      - multipleReplicasForDeployment
  - controllerNames: 
      - controller2
    rules:
      - multipleReplicasForDeployment
  - namespace: kube-system
    controllerNames:
      - controller3
    rules:
      - multipleReplicasForDeployment
  - containerNames:
      - container41
      - container42
    rules:
      - multipleReplicasForDeployment
  - namespace: kube-system
    containerNames:
      - container51
      - container52
    rules:
      - multipleReplicasForDeployment
  - controllerNames:
      - controller6
    containerNames:
      - container61
      - container62
    rules:
      - multipleReplicasForDeployment
  - namespace: kube-system
    controllerNames:
      - controller7
    containerNames:
      - container71
      - container72
    rules:
      - multipleReplicasForDeployment
      - priorityClassNotSet
  - namespace: polaris
`

func TestNamespaceExemptionForSpecifiedRules(t *testing.T) {
	parsedConf, err := Parse([]byte(confContainerTest))
	assert.NoError(t, err)

	actionable := parsedConf.IsActionable("multipleReplicasForDeployment", "prometheus", "", "")
	assert.False(t, actionable)

	actionable = parsedConf.IsActionable("multipleReplicasForDeployment", "prometheus", "controller1", "container11")
	assert.False(t, actionable)

	actionable = parsedConf.IsActionable("multipleReplicasForDeployment", "prometheus", "", "container11")
	assert.False(t, actionable)

	actionable = parsedConf.IsActionable("multipleReplicasForDeployment", "prometheus", "controller1", "")
	assert.False(t, actionable)

	actionable = parsedConf.IsActionable("pullPolicyNotAlways", "prometheus", "controller1", "")
	assert.True(t, actionable)

	actionable = parsedConf.IsActionable("multipleReplicasForDeployment", "kube-system", "", "")
	assert.True(t, actionable)
}

func TestNamespaceExemptionForAllRules(t *testing.T) {
	parsedConf, err := Parse([]byte(confContainerTest))
	assert.NoError(t, err)

	actionable := parsedConf.IsActionable("multipleReplicasForDeployment", "polaris", "", "")
	assert.False(t, actionable)

	actionable = parsedConf.IsActionable("multipleReplicasForDeployment", "polaris", "controller1", "container11")
	assert.False(t, actionable)

	actionable = parsedConf.IsActionable("multipleReplicasForDeployment", "polaris", "", "container11")
	assert.False(t, actionable)

	actionable = parsedConf.IsActionable("multipleReplicasForDeployment", "polaris", "controller1", "")
	assert.False(t, actionable)

	actionable = parsedConf.IsActionable("pullPolicyNotAlways", "polaris", "controller1", "")
	assert.False(t, actionable)
}

func TestControllerExemption(t *testing.T) {
	parsedConf, err := Parse([]byte(confContainerTest))
	assert.NoError(t, err)

	actionable := parsedConf.IsActionable("multipleReplicasForDeployment", "", "controller2", "")
	assert.False(t, actionable)

	actionable = parsedConf.IsActionable("multipleReplicasForDeployment", "", "controller2", "container21")
	assert.False(t, actionable)

	actionable = parsedConf.IsActionable("multipleReplicasForDeployment", "prometheus", "controller2", "container21")
	assert.False(t, actionable)

	actionable = parsedConf.IsActionable("multipleReplicasForDeployment", "prometheus", "controller2", "")
	assert.False(t, actionable)

	actionable = parsedConf.IsActionable("multipleReplicasForDeployment", "", "controller3", "")
	assert.True(t, actionable)

	actionable = parsedConf.IsActionable("multipleReplicasForDeployment", "kube-system", "controller3", "")
	assert.False(t, actionable)

	actionable = parsedConf.IsActionable("multipleReplicasForDeployment", "kube-system", "controller3", "container31")
	assert.False(t, actionable)

	actionable = parsedConf.IsActionable("multipleReplicasForDeployment", "kube-system", "controller4", "")
	assert.True(t, actionable)
}

func TestOnlyContainerExemption(t *testing.T) {
	parsedConf, err := Parse([]byte(confContainerTest))
	assert.NoError(t, err)

	actionable := parsedConf.IsActionable("multipleReplicasForDeployment", "", "", "container41")
	assert.False(t, actionable)

	actionable = parsedConf.IsActionable("multipleReplicasForDeployment", "", "", "container42")
	assert.False(t, actionable)

	actionable = parsedConf.IsActionable("multipleReplicasForDeployment", "", "controller4", "container41")
	assert.False(t, actionable)

	actionable = parsedConf.IsActionable("multipleReplicasForDeployment", "kube-system", "", "container41")
	assert.False(t, actionable)

	actionable = parsedConf.IsActionable("multipleReplicasForDeployment", "kube-system", "controller4", "container41")
	assert.False(t, actionable)

	actionable = parsedConf.IsActionable("multipleReplicasForDeployment", "", "", "container51")
	assert.True(t, actionable)
}

func TestNamespaceAndContainerExemption(t *testing.T) {
	parsedConf, err := Parse([]byte(confContainerTest))
	assert.NoError(t, err)

	actionable := parsedConf.IsActionable("multipleReplicasForDeployment", "kube-system", "", "container51")
	assert.False(t, actionable)

	actionable = parsedConf.IsActionable("priorityClassNotSet", "kube-system", "", "container51")
	assert.True(t, actionable)

	actionable = parsedConf.IsActionable("multipleReplicasForDeployment", "kube-system", "controller5", "container51")
	assert.False(t, actionable)

	actionable = parsedConf.IsActionable("multipleReplicasForDeployment", "kube-system", "controller5", "")
	assert.True(t, actionable)

	actionable = parsedConf.IsActionable("multipleReplicasForDeployment", "insights-agent", "", "container51")
	assert.True(t, actionable)

	actionable = parsedConf.IsActionable("multipleReplicasForDeployment", "", "", "container51")
	assert.True(t, actionable)

	actionable = parsedConf.IsActionable("multipleReplicasForDeployment", "", "controller5", "container51")
	assert.True(t, actionable)
}

func TestControllerAndContainerExemption(t *testing.T) {
	parsedConf, err := Parse([]byte(confContainerTest))
	assert.NoError(t, err)

	actionable := parsedConf.IsActionable("multipleReplicasForDeployment", "", "controller6", "container61")
	assert.False(t, actionable)

	actionable = parsedConf.IsActionable("priorityClassNotSet", "", "controller6", "container61")
	assert.True(t, actionable)

	actionable = parsedConf.IsActionable("multipleReplicasForDeployment", "kube-system", "controller6", "container61")
	assert.False(t, actionable)

	actionable = parsedConf.IsActionable("multipleReplicasForDeployment", "kube-system", "controller6", "")
	assert.True(t, actionable)

	actionable = parsedConf.IsActionable("multipleReplicasForDeployment", "", "controller7", "container61")
	assert.True(t, actionable)

	actionable = parsedConf.IsActionable("multipleReplicasForDeployment", "", "", "container61")
	assert.True(t, actionable)

	actionable = parsedConf.IsActionable("multipleReplicasForDeployment", "kube-system", "", "container61")
	assert.True(t, actionable)
}

func TestContainerExemption(t *testing.T) {
	parsedConf, err := Parse([]byte(confContainerTest))
	assert.NoError(t, err)

	actionable := parsedConf.IsActionable("multipleReplicasForDeployment", "", "", "container71")
	assert.True(t, actionable)

	actionable = parsedConf.IsActionable("multipleReplicasForDeployment", "kube-system", "", "container71")
	assert.True(t, actionable)

	actionable = parsedConf.IsActionable("multipleReplicasForDeployment", "", "controller7", "container71")
	assert.True(t, actionable)

	actionable = parsedConf.IsActionable("multipleReplicasForDeployment", "kube-system", "controller7", "")
	assert.True(t, actionable)

	actionable = parsedConf.IsActionable("multipleReplicasForDeployment", "kube-system", "controller7", "container71")
	assert.False(t, actionable)

	actionable = parsedConf.IsActionable("multipleReplicasForDeployment", "insights-agent", "controller7", "container71")
	assert.True(t, actionable)

	actionable = parsedConf.IsActionable("multipleReplicasForDeployment", "kube-system", "controller6", "container71")
	assert.True(t, actionable)

	actionable = parsedConf.IsActionable("multipleReplicasForDeployment", "kube-system", "controller7", "container61")
	assert.True(t, actionable)

	actionable = parsedConf.IsActionable("priorityClassNotSet", "kube-system", "controller7", "container71")
	assert.False(t, actionable)

	actionable = parsedConf.IsActionable("pullPolicyNotAlways", "kube-system", "controller8", "container71")
	assert.True(t, actionable)
}
