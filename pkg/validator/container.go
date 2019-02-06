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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// ContainerValidation tracks validation failures associated with a Container
type ContainerValidation struct {
	Container corev1.Container
	Summary   ResultSummary
	Failures  []ResultMessage
	Warnings  []ResultMessage
	Successes []ResultMessage
}

func (cv *ContainerValidation) messages() []ResultMessage {
	mssgs := []ResultMessage{}
	mssgs = append(mssgs, cv.Failures...)
	mssgs = append(mssgs, cv.Warnings...)
	mssgs = append(mssgs, cv.Successes...)
	return mssgs
}

func validateContainer(conf conf.Configuration, container corev1.Container) (ContainerResult, ResultSummary) {
	cv := ContainerValidation{
		Container: container,
		Summary:   ResultSummary{},
	}

	cv.validateResources(conf.Resources)
	cv.validateHealthChecks(conf.HealthChecks)
	cv.validateImage(conf.Images)

	cRes := ContainerResult{
		Name:     container.Name,
		Messages: cv.messages(),
	}

	return cRes, cv.Summary
}

func (cv *ContainerValidation) addFailure(message string) {
	cv.Summary.Failures++
	cv.Failures = append(cv.Failures, ResultMessage{
		Message: message,
		Type:    "failure",
	})
}

func (cv *ContainerValidation) addWarning(message string) {
	cv.Summary.Warnings++
	cv.Warnings = append(cv.Warnings, ResultMessage{
		Message: message,
		Type:    "warning",
	})
}

func (cv *ContainerValidation) addSuccess(message string) {
	cv.Summary.Successes++
	cv.Successes = append(cv.Successes, ResultMessage{
		Message: message,
		Type:    "success",
	})
}

func (cv *ContainerValidation) validateResources(conf conf.RequestsAndLimits) {
	actualRes := cv.Container.Resources
	cv.withinRange("CPU Requests", conf.Requests["cpu"], actualRes.Requests.Cpu())
	cv.withinRange("Memory Requests", conf.Requests["memory"], actualRes.Requests.Memory())
	cv.withinRange("CPU Limits", conf.Limits["cpu"], actualRes.Limits.Cpu())
	cv.withinRange("Memory Limits", conf.Limits["memory"], actualRes.Limits.Memory())
}

func (cv *ContainerValidation) withinRange(resourceName string, expectedRange conf.ResourceMinMax, actual *resource.Quantity) {
	expectedMin := expectedRange.Min
	expectedMax := expectedRange.Max
	if expectedMin != nil && expectedMin.MilliValue() > actual.MilliValue() {
		if actual.MilliValue() == 0 {
			cv.addFailure(fmt.Sprintf("%s are not set", resourceName))
		} else {
			cv.addFailure(fmt.Sprintf("%s are too low", resourceName))
		}
	} else if expectedMax != nil && expectedMax.MilliValue() < actual.MilliValue() {
		cv.addFailure(fmt.Sprintf("%s are too high", resourceName))
	} else {
		cv.addSuccess(fmt.Sprintf("%s are within the expected range", resourceName))
	}
}

func (cv *ContainerValidation) validateHealthChecks(conf conf.Probes) {
	if conf.Readiness.Require {
		if cv.Container.ReadinessProbe == nil {
			cv.addFailure("Readiness probe needs to be configured")
		} else {
			cv.addSuccess("Readiness probe configured")
		}
	}

	if conf.Liveness.Require {
		if cv.Container.LivenessProbe == nil {
			cv.addFailure("Liveness probe needs to be configured")
		} else {
			cv.addSuccess("Liveness probe configured")
		}
	}
}

func (cv *ContainerValidation) validateImage(conf conf.Images) {
	if conf.TagRequired {
		img := strings.Split(cv.Container.Image, ":")
		if len(img) == 1 || img[1] == "latest" {
			cv.addFailure("Image tag should be specified")
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
