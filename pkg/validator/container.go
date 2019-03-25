// Copyright 2019 ReactiveOps
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

// ContainerValidation tracks validation failures associated with a Container.
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

func validateContainer(conf conf.Configuration, container corev1.Container) ResourceResult {
	cv := ContainerValidation{
		Container: container,
		Summary:   ResultSummary{},
	}

	cv.validateResources(&conf.Resources)
	cv.validateHealthChecks(&conf.HealthChecks)
	cv.validateImage(&conf.Images)
	cv.validateNetworking(&conf.Networking)

	cRes := ContainerResult{
		Name:     container.Name,
		Messages: cv.messages(),
	}

	rr := ResourceResult{
		Name:             container.Name,
		Type:             "Container",
		Summary:          &cv.Summary,
		ContainerResults: []ContainerResult{cRes},
	}

	return rr
}

func (cv *ContainerValidation) addMessage(message string, severity conf.Severity) {
	if severity == conf.SeverityError {
		cv.addFailure(message)
	} else if severity == conf.SeverityWarning {
		cv.addWarning(message)
	}
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

func (cv *ContainerValidation) validateResources(resourcesConf *conf.Resources) {
	actualRes := cv.Container.Resources
	cv.validateResource("CPU Requests", &resourcesConf.CPURequests, actualRes.Requests.Cpu())
	cv.validateResource("Memory Requests", &resourcesConf.MemoryRequests, actualRes.Requests.Memory())
	cv.validateResource("CPU Limits", &resourcesConf.CPULimits, actualRes.Limits.Cpu())
	cv.validateResource("Memory Limits", &resourcesConf.MemoryLimits, actualRes.Limits.Memory())
}

func (cv *ContainerValidation) validateResource(resourceName string, resourceConf *conf.Resource, actual *resource.Quantity) {
	if resourceConf.Absent.IsActionable() && actual.MilliValue() == 0 {
		msg := fmt.Sprintf("%s are not set", resourceName)
		if resourceConf.Absent == conf.SeverityError {
			cv.addFailure(msg)
		} else {
			cv.addWarning(msg)
		}
	} else {
		warnAbove := resourceConf.Warning.Above
		warnBelow := resourceConf.Warning.Below
		errorAbove := resourceConf.Error.Above
		errorBelow := resourceConf.Error.Below

		if errorAbove != nil && errorAbove.MilliValue() < actual.MilliValue() {
			cv.addFailure(fmt.Sprintf("%s are too high", resourceName))
		} else if warnAbove != nil && warnAbove.MilliValue() < actual.MilliValue() {
			cv.addWarning(fmt.Sprintf("%s are too high", resourceName))
		} else if errorBelow != nil && errorBelow.MilliValue() > actual.MilliValue() {
			cv.addFailure(fmt.Sprintf("%s are too low", resourceName))
		} else if warnBelow != nil && warnBelow.MilliValue() > actual.MilliValue() {
			cv.addWarning(fmt.Sprintf("%s are too low", resourceName))
		} else {
			cv.addSuccess(fmt.Sprintf("%s are within the expected range", resourceName))
		}
	}
}

func (cv *ContainerValidation) validateHealthChecks(conf *conf.HealthChecks) {
	if conf.ReadinessProbeMissing.IsActionable() {
		if cv.Container.ReadinessProbe == nil {
			cv.addMessage("Readiness probe needs to be configured", conf.ReadinessProbeMissing)
		} else {
			cv.addSuccess("Readiness probe configured")
		}
	}

	if conf.LivenessProbeMissing.IsActionable() {
		if cv.Container.LivenessProbe == nil {
			cv.addMessage("Liveness probe needs to be configured", conf.LivenessProbeMissing)
		} else {
			cv.addSuccess("Liveness probe configured")
		}
	}
}

func (cv *ContainerValidation) validateImage(imageConf *conf.Images) {
	if imageConf.TagNotSpecified.IsActionable() {
		img := strings.Split(cv.Container.Image, ":")
		if len(img) == 1 || img[1] == "latest" {
			cv.addMessage("Image tag should be specified", imageConf.TagNotSpecified)
		} else {
			cv.addSuccess("Image tag specified")
		}
	}
}

func (cv *ContainerValidation) validateNetworking(networkConf *conf.Networking) {
	if networkConf.HostPortSet.IsActionable() {
		hostPortSet := false
		for _, port := range cv.Container.Ports {
			if port.HostPort != 0 {
				hostPortSet = true
				break
			}
		}

		if hostPortSet {
			cv.addMessage("Host port is configured, but it shouldn't be", networkConf.HostAliasSet)
		} else {
			cv.addSuccess("Host port is not configured")
		}
	}
}
