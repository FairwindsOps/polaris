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
	"github.com/fairwindsops/polaris/pkg/kube"
	"github.com/fairwindsops/polaris/test"
	"github.com/stretchr/testify/require"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestGetTemplateInputReturnsPolarisSubKeys(t *testing.T) {
	pod := test.MockPod() // Includes a container, required by GetPodSpec
	pod.Spec.NodeName = "testNodeName"
	pod.ObjectMeta.Name = "testpod"
	genRes, err := kube.NewGenericResourceFromPod(pod, pod)
	require.NoError(t, err, "creating new generic resource from a pod")
	schemaTest := schemaTestCase{
		Target:   conf.TargetPodSpec, // ends up being set in the case of target: PodTemplate
		Resource: genRes,
	}

	templateInput, err := getTemplateInput(schemaTest)
	require.NoError(t, err, "getting template input from a generic resource")
	require.NotNil(t, templateInput)
	nodeName, ok, err := unstructured.NestedString(templateInput, "Polaris", "PodSpec", "nodeName")
	require.NoError(t, err, "getting Polaris.PodSpec.nodeName from template input")
	require.True(t, ok, "getting Polaris.PodSpec.nodeName from template input")
	require.Equal(t, "testNodeName", nodeName, "the nodeName from template output")
	podName, ok, err := unstructured.NestedString(templateInput, "Polaris", "PodTemplate", "metadata", "name")
	require.NoError(t, err, "getting Polaris.PodTemplate.metadata.name from template input")
	require.True(t, ok, "getting Polaris.PodTemplate.metadata.name from template input")
	require.Equal(t, "testpod", podName, "the pod from template input")
}
