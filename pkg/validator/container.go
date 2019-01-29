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
	"strings"

	conf "github.com/reactiveops/fairwinds/pkg/config"
	"github.com/reactiveops/fairwinds/pkg/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// ContainerValidation tracks validation failures associated with a Container
type ContainerValidation struct {
	Container corev1.Container
	Failures  []types.Failure
	Successes []ResultMessage
}

func validateContainer(conf conf.Configuration, container corev1.Container) ContainerValidation {
	cv := ContainerValidation{
		Container: container,
	}

	cv.validateResources(conf.Resources)
	cv.validateHealthChecks(conf.HealthChecks)
	cv.validateImage(conf.Images)

	return cv
}

func (cv *ContainerValidation) addFailure(name, expected, actual string) {
	cv.Failures = append(cv.Failures, types.Failure{
		Name:     name,
		Expected: expected,
		Actual:   actual,
	})
}

func (cv *ContainerValidation) addSuccess(message string) {
	cv.Successes = append(cv.Successes, ResultMessage{
		Message: message,
		Type:    "success",
	})
}

func (cv *ContainerValidation) validateResources(conf conf.RequestsAndLimits) {
	actualRes := cv.Container.Resources
	cv.withinRange("requests.cpu", conf.Requests["cpu"], actualRes.Requests.Cpu())
	cv.withinRange("requests.memory", conf.Requests["memory"], actualRes.Requests.Memory())
	cv.withinRange("limits.cpu", conf.Limits["cpu"], actualRes.Limits.Cpu())
	cv.withinRange("limits.memory", conf.Limits["memory"], actualRes.Limits.Memory())
}

func (cv *ContainerValidation) withinRange(resourceName string, expectedRange conf.ResourceMinMax, actual *resource.Quantity) {
	expectedMin := expectedRange.Min
	expectedMax := expectedRange.Max
	if expectedMin != nil && expectedMin.MilliValue() > actual.MilliValue() {
		cv.addFailure(resourceName, expectedMin.String(), actual.String())
	} else if expectedMax != nil && expectedMax.MilliValue() < actual.MilliValue() {
		cv.addFailure(resourceName, expectedMax.String(), actual.String())
	} else {
		cv.addSuccess(fmt.Sprintf("Resource %s within expected range", resourceName))
	}
}

func (cv *ContainerValidation) validateHealthChecks(conf conf.Probes) {
	if conf.Readiness.Require {
		if cv.Container.ReadinessProbe == nil {
			cv.addFailure("readiness", "probe needs to be configured", "nil")
		} else {
			cv.addSuccess("Readiness probe configured")
		}
	}

	if conf.Liveness.Require {
		if cv.Container.LivenessProbe == nil {
			cv.addFailure("liveness", "probe needs to be configured", "nil")
		} else {
			cv.addSuccess("Liveness probe configured")
		}
	}
}

func (cv *ContainerValidation) validateImage(conf conf.Images) {
	if conf.TagRequired {
		img := strings.Split(cv.Container.Image, ":")
		if len(img) == 1 || img[1] == "latest" {
			cv.addFailure("Image Tag", "not latest", "latest")
		} else {
			cv.addSuccess("Image tag specified")
		}
	}
}

// func hostPort(conf conf.ResourceRequestsAndLimits, c corev1.Container, results types.ContainerResults) types.ContainerResults {
// 	for _, port := range c.Ports {
// 		if port.HostPort != 0 {
// 			results.AddFailure("Host port", "placeholder", "placeholder")
// 		}
// 	}
// 	return results
// }
