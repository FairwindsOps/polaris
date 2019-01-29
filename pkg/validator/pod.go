// Copyright 2018 ReactiveOps
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
	conf "github.com/reactiveops/fairwinds/pkg/config"
	corev1 "k8s.io/api/core/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("Fairwinds Validator")

// ValidatePods validates that each pod conforms to the Fairwinds config.
func ValidatePods(conf conf.Configuration, pod *corev1.PodSpec) Results {
	results := Results{}
	for _, container := range pod.InitContainers {
		results.InitContainerValidations = append(
			results.InitContainerValidations,
			validateContainer(conf, container),
		)
	}

	for _, container := range pod.Containers {
		results.ContainerValidations = append(
			results.ContainerValidations,
			validateContainer(conf, container),
		)
	}

	return results
}
