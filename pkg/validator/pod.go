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

// PodValidation tracks validation failures associated with a Pod.
type PodValidation struct {
	*ResourceValidation
	Pod *corev1.PodSpec
}

// ValidatePod validates that each pod conforms to the Polaris config, returns a ResourceResult.
func ValidatePod(conf config.Configuration, pod *corev1.PodSpec, controllerName string, controllerType config.SupportedController) PodResult {
	pv := PodValidation{
		Pod:                pod,
		ResourceValidation: &ResourceValidation{},
	}

	applyPodSchemaChecks(&conf, pod, controllerName, &pv)

	pRes := PodResult{
		Messages:         pv.messages(),
		ContainerResults: []ContainerResult{},
		Summary:          pv.summary(),
		podSpec:          *pod,
	}

	pv.validateContainers(pod.InitContainers, &pRes, &conf, controllerName, controllerType, true)
	pv.validateContainers(pod.Containers, &pRes, &conf, controllerName, controllerType, false)

	for _, cRes := range pRes.ContainerResults {
		pRes.Summary.appendResults(*cRes.Summary)
	}

	return pRes
}

func (pv *PodValidation) validateContainers(containers []corev1.Container, pRes *PodResult, conf *config.Configuration, controllerName string, controllerType config.SupportedController, isInit bool) {
	for _, container := range containers {
		cRes := ValidateContainer(&container, pRes, conf, controllerName, controllerType, isInit)
		pRes.ContainerResults = append(pRes.ContainerResults, cRes)
	}
}
