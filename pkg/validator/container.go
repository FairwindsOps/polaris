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
	"fmt"

	conf "github.com/reactiveops/fairwinds/pkg/config"
	"github.com/reactiveops/fairwinds/pkg/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// ContainerValidation tracks validation failures associated with a Container
type ContainerValidation struct {
	Container corev1.Container
	Failures  []types.Failure
}

func validateContainer(conf conf.Configuration, container corev1.Container) ContainerValidation {
	cv := ContainerValidation{
		Container: container,
	}

	cv.validateResources(conf.Resources)
	// cv.validateHealthChecks(conf.HealthChecks)
	// cv.validateTags(conf.Image)

	return cv
}

func (cv *ContainerValidation) addFailure(name, expected, actual string) {
	cv.Failures = append(cv.Failures, types.Failure{
		Name:     name,
		Expected: expected,
		Actual:   actual,
	})
}

func (cv *ContainerValidation) validateResources(conf conf.RequestsAndLimits) {
	actualRes := cv.Container.Resources
	cv.withinRange("requests.cpu", conf.Requests["cpu"], actualRes.Requests.Cpu())
	cv.withinRange("requests.memory", conf.Requests["memory"], actualRes.Requests.Memory())
	cv.withinRange("limits.cpu", conf.Limits["cpu"], actualRes.Limits.Cpu())
	cv.withinRange("limits.memory", conf.Limits["memory"], actualRes.Limits.Memory())
}

func (cv *ContainerValidation) withinRange(resName string, expectedRange conf.ResourceMinMax, actual *resource.Quantity) {
	expectedMin, err := resource.ParseQuantity(expectedRange.Min)
	if err != nil {
		log.Error(err, fmt.Sprintf("Error parsing min quantity for %s", resName))
	} else if expectedMin.MilliValue() > actual.MilliValue() {
		cv.addFailure(resName, expectedMin.String(), actual.String())
	}

	expectedMax, err := resource.ParseQuantity(expectedRange.Max)
	if err != nil {
		log.Error(err, fmt.Sprintf("Error parsing max quantity for %s", resName))
	} else if expectedMax.MilliValue() < actual.MilliValue() {
		cv.addFailure(resName, expectedMax.String(), actual.String())
	}
}

// func probes(conf conf.ResourceRequestsAndLimits, c corev1.Container, results types.ContainerResults) types.ContainerResults {
// 	if c.ReadinessProbe == nil {
// 		results.AddFailure("Readiness Probe", "placeholder", "placeholder")
// 	}

// 	if c.LivenessProbe == nil {
// 		results.AddFailure("Liveness Probe", "placeholder", "placeholder")
// 	}
// 	return results
// }

// func tag(conf conf.ResourceRequestsAndLimits, c corev1.Container, results types.ContainerResults) types.ContainerResults {
// 	img := strings.Split(c.Image, ":")
// 	if len(img) == 1 || img[1] == "latest" {
// 		results.AddFailure("Image Tag", "not latest", "latest")
// 	}
// 	return results
// }

// func hostPort(conf conf.ResourceRequestsAndLimits, c corev1.Container, results types.ContainerResults) types.ContainerResults {
// 	for _, port := range c.Ports {
// 		if port.HostPort != 0 {
// 			results.AddFailure("Host port", "placeholder", "placeholder")
// 		}
// 	}
// 	return results
// }
