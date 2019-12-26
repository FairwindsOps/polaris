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
	"fmt"

	"github.com/fairwindsops/polaris/pkg/config"
	"github.com/fairwindsops/polaris/pkg/validator/messages"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// ContainerValidation tracks validation failures associated with a Container.
type ContainerValidation struct {
	*ResourceValidation
	Container       *corev1.Container
	IsInitContainer bool
	parentPodSpec   corev1.PodSpec
}

// ValidateContainer validates that each pod conforms to the Polaris config, returns a ResourceResult.
// FIXME When validating a container, there are some things in a container spec
//       that can be affected by the podSpec. This means we need a copy of the
//       relevant podSpec in order to check certain aspects of a containerSpec.
//       Perhaps there is a more ideal solution instead of attaching a parent
//       podSpec to every container Validation struct...
func ValidateContainer(container *corev1.Container, parentPodResult *PodResult, conf *config.Configuration, controllerName string, controllerType config.SupportedController, isInit bool) ContainerResult {
	cv := ContainerValidation{
		Container:          container,
		ResourceValidation: &ResourceValidation{},
		IsInitContainer:    isInit,
	}

	// Support initializing
	// FIXME This is a product of pulling in the podSpec, ideally we'd never
	//       expect this be nil but our tests have conditions in which the
	//       parent podResult isn't initialized in this ContainerValidation
	//       struct.
	if parentPodResult == nil {
		// initialize a blank pod spec
		cv.parentPodSpec = corev1.PodSpec{}
	} else {
		cv.parentPodSpec = parentPodResult.podSpec
	}

	cv.validateResources(conf, controllerName)

	err := applyContainerSchemaChecks(conf, controllerName, controllerType, &cv)
	// FIXME: don't panic
	if err != nil {
		panic(err)
	}

	cRes := ContainerResult{
		Name:     container.Name,
		Messages: cv.messages(),
		Summary:  cv.summary(),
	}

	return cRes
}

func (cv *ContainerValidation) validateResources(conf *config.Configuration, controllerName string) {
	// Only validate resources for primary containers. Although it can
	// be helpful to set these in certain cases, it usually isn't
	if cv.IsInitContainer {
		return
	}

	category := messages.CategoryResources
	res := cv.Container.Resources

	missingName := "CPURequestsMissing"
	rangeName := "CPURequestRanges"
	id := config.GetIDFromField(conf.Resources, missingName)
	if conf.IsActionable(conf.Resources, missingName, controllerName) && res.Requests.Cpu().MilliValue() == 0 {
		cv.addFailure(messages.CPURequestsFailure, conf.Resources.CPURequestsMissing, category, id)
	} else if conf.IsActionable(conf.Resources, rangeName, controllerName) {
		id := config.GetIDFromField(conf.Resources, rangeName)
		cv.validateResourceRange(id, messages.CPURequestsLabel, &conf.Resources.CPURequestRanges, res.Requests.Cpu())
	} else if conf.IsActionable(conf.Resources, missingName, controllerName) {
		cv.addSuccess(fmt.Sprintf(messages.ResourcePresentSuccess, messages.CPURequestsLabel), category, id)
	}

	missingName = "CPULimitsMissing"
	rangeName = "CPULimitRanges"
	id = config.GetIDFromField(conf.Resources, missingName)
	if conf.IsActionable(conf.Resources, missingName, controllerName) && res.Limits.Cpu().MilliValue() == 0 {
		cv.addFailure(messages.CPULimitsFailure, conf.Resources.CPULimitsMissing, category, id)
	} else if conf.IsActionable(conf.Resources, rangeName, controllerName) {
		id := config.GetIDFromField(conf.Resources, rangeName)
		cv.validateResourceRange(id, messages.CPULimitsLabel, &conf.Resources.CPULimitRanges, res.Requests.Cpu())
	} else if conf.IsActionable(conf.Resources, missingName, controllerName) {
		cv.addSuccess(fmt.Sprintf(messages.ResourcePresentSuccess, messages.CPULimitsLabel), category, id)
	}

	missingName = "MemoryRequestsMissing"
	rangeName = "MemoryRequestRanges"
	id = config.GetIDFromField(conf.Resources, missingName)
	if conf.IsActionable(conf.Resources, missingName, controllerName) && res.Requests.Memory().MilliValue() == 0 {
		cv.addFailure(messages.MemoryRequestsFailure, conf.Resources.MemoryRequestsMissing, category, id)
	} else if conf.IsActionable(conf.Resources, rangeName, controllerName) {
		id := config.GetIDFromField(conf.Resources, rangeName)
		cv.validateResourceRange(id, messages.MemoryRequestsLabel, &conf.Resources.MemoryRequestRanges, res.Requests.Memory())
	} else if conf.IsActionable(conf.Resources, missingName, controllerName) {
		cv.addSuccess(fmt.Sprintf(messages.ResourcePresentSuccess, messages.MemoryRequestsLabel), category, id)
	}

	missingName = "MemoryLimitsMissing"
	rangeName = "MemoryLimitRanges"
	id = config.GetIDFromField(conf.Resources, missingName)
	if conf.IsActionable(conf.Resources, missingName, controllerName) && res.Limits.Memory().MilliValue() == 0 {
		cv.addFailure(messages.MemoryLimitsFailure, conf.Resources.MemoryLimitsMissing, category, id)
	} else if conf.IsActionable(conf.Resources, rangeName, controllerName) {
		id := config.GetIDFromField(conf.Resources, rangeName)
		cv.validateResourceRange(id, messages.MemoryLimitsLabel, &conf.Resources.MemoryLimitRanges, res.Limits.Memory())
	} else if conf.IsActionable(conf.Resources, missingName, controllerName) {
		cv.addSuccess(fmt.Sprintf(messages.ResourcePresentSuccess, messages.MemoryLimitsLabel), category, id)
	}
}

func (cv *ContainerValidation) validateResourceRange(id, resourceName string, rangeConf *config.ResourceRanges, res *resource.Quantity) {
	warnAbove := rangeConf.Warning.Above
	warnBelow := rangeConf.Warning.Below
	errorAbove := rangeConf.Error.Above
	errorBelow := rangeConf.Error.Below
	category := messages.CategoryResources

	if errorAbove != nil && errorAbove.MilliValue() < res.MilliValue() {
		cv.addError(fmt.Sprintf(messages.ResourceAmountTooHighFailure, resourceName, errorAbove.String()), category, id)
	} else if warnAbove != nil && warnAbove.MilliValue() < res.MilliValue() {
		cv.addWarning(fmt.Sprintf(messages.ResourceAmountTooHighFailure, resourceName, warnAbove.String()), category, id)
	} else if errorBelow != nil && errorBelow.MilliValue() > res.MilliValue() {
		cv.addError(fmt.Sprintf(messages.ResourceAmountTooLowFailure, resourceName, errorBelow.String()), category, id)
	} else if warnBelow != nil && warnBelow.MilliValue() > res.MilliValue() {
		cv.addWarning(fmt.Sprintf(messages.ResourceAmountTooLowFailure, resourceName, warnBelow.String()), category, id)
	} else if errorAbove != nil && warnAbove != nil && errorBelow != nil && warnBelow != nil {
		cv.addSuccess(fmt.Sprintf(messages.ResourceAmountSuccess, resourceName), category, id)
	}
}
