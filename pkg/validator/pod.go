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
	"github.com/fairwindsops/polaris/pkg/config"
	corev1 "k8s.io/api/core/v1"
)

// ValidatePod validates that each pod conforms to the Polaris config, returns a ResourceResult.
func ValidatePod(conf *config.Configuration, pod *corev1.PodSpec, controllerName string, controllerKind config.SupportedController) PodResult {
	podResults, err := applyPodSchemaChecks(conf, pod, controllerName, controllerKind)
	// FIXME: don't panic
	if err != nil {
		panic(err)
	}

	pRes := PodResult{
		Results:          podResults,
		ContainerResults: []ContainerResult{},
	}

	podCopy := *pod
	podCopy.InitContainers = []corev1.Container{}
	podCopy.Containers = []corev1.Container{}

	containerResults := ValidateContainers(conf, &podCopy, pod.InitContainers, controllerName, controllerKind, true)
	pRes.ContainerResults = append(pRes.ContainerResults, containerResults...)
	containerResults = ValidateContainers(conf, &podCopy, pod.Containers, controllerName, controllerKind, false)
	pRes.ContainerResults = append(pRes.ContainerResults, containerResults...)

	return pRes
}
