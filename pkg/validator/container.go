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
	"github.com/fairwindsops/polaris/pkg/kube"

	corev1 "k8s.io/api/core/v1"
)

// ValidateContainer validates a single container from a given controller
func ValidateContainer(conf *config.Configuration, controller kube.GenericWorkload, container *corev1.Container, isInit bool) (ContainerResult, error) {
	results, err := applyContainerSchemaChecks(conf, controller, container, isInit)
	if err != nil {
		return ContainerResult{}, err
	}

	cRes := ContainerResult{
		Name:    container.Name,
		Results: results,
	}

	return cRes, nil
}

// ValidateAllContainers validates both init and regular containers
func ValidateAllContainers(conf *config.Configuration, controller kube.GenericWorkload) ([]ContainerResult, error) {
	results := []ContainerResult{}
	pod := controller.PodSpec
	for _, container := range pod.InitContainers {
		result, err := ValidateContainer(conf, controller, &container, true)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	for _, container := range pod.Containers {
		result, err := ValidateContainer(conf, controller, &container, false)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}
