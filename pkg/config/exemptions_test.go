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
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var confContainerTest = `
checks:
  deploymentMissingReplicas: warning
  priorityClassNotSet: warning
  pullPolicyNotAlways: warning
exemptions:
  - namespace: prometheus
    rules:
      - deploymentMissingReplicas
  - controllerNames: 
      - controller2
    rules:
      - deploymentMissingReplicas
  - namespace: kube-system
    controllerNames:
      - controller3
    rules:
      - deploymentMissingReplicas
  - containerNames:
      - container41
      - container42
    rules:
      - deploymentMissingReplicas
  - namespace: kube-system
    containerNames:
      - container51
      - container52
    rules:
      - deploymentMissingReplicas
  - controllerNames:
      - controller6
    containerNames:
      - container61
      - container62
    rules:
      - deploymentMissingReplicas
  - namespace: kube-system
    controllerNames:
      - controller7
    containerNames:
      - container71
      - container72
    rules:
      - deploymentMissingReplicas
      - priorityClassNotSet
  - namespace: polaris
`

func createMeta(namespace, name string) metav1.Object {
	unst := unstructured.Unstructured{}
	obj, err := meta.Accessor(&unst)
	if err != nil {
		panic(err)
	}
	obj.SetName(name)
	obj.SetNamespace(namespace)
	return obj
}

func TestNamespaceExemptionForSpecifiedRules(t *testing.T) {
	parsedConf, err := Parse([]byte(confContainerTest))
	assert.NoError(t, err)

	actionable := parsedConf.IsActionable("deploymentMissingReplicas", createMeta("prometheus", ""), "")
	assert.False(t, actionable)

	actionable = parsedConf.IsActionable("deploymentMissingReplicas", createMeta("prometheus", "controller1"), "container11")
	assert.False(t, actionable)

	actionable = parsedConf.IsActionable("deploymentMissingReplicas", createMeta("prometheus", ""), "container11")
	assert.False(t, actionable)

	actionable = parsedConf.IsActionable("deploymentMissingReplicas", createMeta("prometheus", "controller1"), "")
	assert.False(t, actionable)

	actionable = parsedConf.IsActionable("pullPolicyNotAlways", createMeta("prometheus", "controller1"), "")
	assert.True(t, actionable)

	actionable = parsedConf.IsActionable("deploymentMissingReplicas", createMeta("kube-system", ""), "")
	assert.True(t, actionable)
}

func TestNamespaceExemptionForAllRules(t *testing.T) {
	parsedConf, err := Parse([]byte(confContainerTest))
	assert.NoError(t, err)

	actionable := parsedConf.IsActionable("deploymentMissingReplicas", createMeta("polaris", ""), "")
	assert.False(t, actionable)

	actionable = parsedConf.IsActionable("deploymentMissingReplicas", createMeta("polaris", "controller1"), "container11")
	assert.False(t, actionable)

	actionable = parsedConf.IsActionable("deploymentMissingReplicas", createMeta("polaris", ""), "container11")
	assert.False(t, actionable)

	actionable = parsedConf.IsActionable("deploymentMissingReplicas", createMeta("polaris", "controller1"), "")
	assert.False(t, actionable)

	actionable = parsedConf.IsActionable("pullPolicyNotAlways", createMeta("polaris", "controller1"), "")
	assert.False(t, actionable)
}

func TestControllerExemption(t *testing.T) {
	parsedConf, err := Parse([]byte(confContainerTest))
	assert.NoError(t, err)

	actionable := parsedConf.IsActionable("deploymentMissingReplicas", createMeta("", "controller2"), "")
	assert.False(t, actionable)

	actionable = parsedConf.IsActionable("deploymentMissingReplicas", createMeta("", "controller2"), "container21")
	assert.False(t, actionable)

	actionable = parsedConf.IsActionable("deploymentMissingReplicas", createMeta("prometheus", "controller2"), "container21")
	assert.False(t, actionable)

	actionable = parsedConf.IsActionable("deploymentMissingReplicas", createMeta("prometheus", "controller2"), "")
	assert.False(t, actionable)

	actionable = parsedConf.IsActionable("deploymentMissingReplicas", createMeta("", "controller3"), "")
	assert.True(t, actionable)

	actionable = parsedConf.IsActionable("deploymentMissingReplicas", createMeta("kube-system", "controller3"), "")
	assert.False(t, actionable)

	actionable = parsedConf.IsActionable("deploymentMissingReplicas", createMeta("kube-system", "controller3"), "container31")
	assert.False(t, actionable)

	actionable = parsedConf.IsActionable("deploymentMissingReplicas", createMeta("kube-system", "controller4"), "")
	assert.True(t, actionable)
}

func TestOnlyContainerExemption(t *testing.T) {
	parsedConf, err := Parse([]byte(confContainerTest))
	assert.NoError(t, err)

	actionable := parsedConf.IsActionable("deploymentMissingReplicas", createMeta("", ""), "container41")
	assert.False(t, actionable)

	actionable = parsedConf.IsActionable("deploymentMissingReplicas", createMeta("", ""), "container42")
	assert.False(t, actionable)

	actionable = parsedConf.IsActionable("deploymentMissingReplicas", createMeta("", "controller4"), "container41")
	assert.False(t, actionable)

	actionable = parsedConf.IsActionable("deploymentMissingReplicas", createMeta("kube-system", ""), "container41")
	assert.False(t, actionable)

	actionable = parsedConf.IsActionable("deploymentMissingReplicas", createMeta("kube-system", "controller4"), "container41")
	assert.False(t, actionable)

	actionable = parsedConf.IsActionable("deploymentMissingReplicas", createMeta("", ""), "container51")
	assert.True(t, actionable)
}

func TestNamespaceAndContainerExemption(t *testing.T) {
	parsedConf, err := Parse([]byte(confContainerTest))
	assert.NoError(t, err)

	actionable := parsedConf.IsActionable("deploymentMissingReplicas", createMeta("kube-system", ""), "container51")
	assert.False(t, actionable)

	actionable = parsedConf.IsActionable("priorityClassNotSet", createMeta("kube-system", ""), "container51")
	assert.True(t, actionable)

	actionable = parsedConf.IsActionable("deploymentMissingReplicas", createMeta("kube-system", "controller5"), "container51")
	assert.False(t, actionable)

	actionable = parsedConf.IsActionable("deploymentMissingReplicas", createMeta("kube-system", "controller5"), "")
	assert.True(t, actionable)

	actionable = parsedConf.IsActionable("deploymentMissingReplicas", createMeta("insights-agent", ""), "container51")
	assert.True(t, actionable)

	actionable = parsedConf.IsActionable("deploymentMissingReplicas", createMeta("", ""), "container51")
	assert.True(t, actionable)

	actionable = parsedConf.IsActionable("deploymentMissingReplicas", createMeta("", "controller5"), "container51")
	assert.True(t, actionable)
}

func TestControllerAndContainerExemption(t *testing.T) {
	parsedConf, err := Parse([]byte(confContainerTest))
	assert.NoError(t, err)

	actionable := parsedConf.IsActionable("deploymentMissingReplicas", createMeta("", "controller6"), "container61")
	assert.False(t, actionable)

	actionable = parsedConf.IsActionable("priorityClassNotSet", createMeta("", "controller6"), "container61")
	assert.True(t, actionable)

	actionable = parsedConf.IsActionable("deploymentMissingReplicas", createMeta("kube-system", "controller6"), "container61")
	assert.False(t, actionable)

	actionable = parsedConf.IsActionable("deploymentMissingReplicas", createMeta("kube-system", "controller6"), "")
	assert.True(t, actionable)

	actionable = parsedConf.IsActionable("deploymentMissingReplicas", createMeta("", "controller7"), "container61")
	assert.True(t, actionable)

	actionable = parsedConf.IsActionable("deploymentMissingReplicas", createMeta("", ""), "container61")
	assert.True(t, actionable)

	actionable = parsedConf.IsActionable("deploymentMissingReplicas", createMeta("kube-system", ""), "container61")
	assert.True(t, actionable)
}

func TestContainerExemption(t *testing.T) {
	parsedConf, err := Parse([]byte(confContainerTest))
	assert.NoError(t, err)

	actionable := parsedConf.IsActionable("deploymentMissingReplicas", createMeta("", ""), "container71")
	assert.True(t, actionable)

	actionable = parsedConf.IsActionable("deploymentMissingReplicas", createMeta("kube-system", ""), "container71")
	assert.True(t, actionable)

	actionable = parsedConf.IsActionable("deploymentMissingReplicas", createMeta("", "controller7"), "container71")
	assert.True(t, actionable)

	actionable = parsedConf.IsActionable("deploymentMissingReplicas", createMeta("kube-system", "controller7"), "")
	assert.True(t, actionable)

	actionable = parsedConf.IsActionable("deploymentMissingReplicas", createMeta("kube-system", "controller7"), "container71")
	assert.False(t, actionable)

	actionable = parsedConf.IsActionable("deploymentMissingReplicas", createMeta("insights-agent", "controller7"), "container71")
	assert.True(t, actionable)

	actionable = parsedConf.IsActionable("deploymentMissingReplicas", createMeta("kube-system", "controller6"), "container71")
	assert.True(t, actionable)

	actionable = parsedConf.IsActionable("deploymentMissingReplicas", createMeta("kube-system", "controller7"), "container61")
	assert.True(t, actionable)

	actionable = parsedConf.IsActionable("priorityClassNotSet", createMeta("kube-system", "controller7"), "container71")
	assert.False(t, actionable)

	actionable = parsedConf.IsActionable("pullPolicyNotAlways", createMeta("kube-system", "controller8"), "container71")
	assert.True(t, actionable)
}
