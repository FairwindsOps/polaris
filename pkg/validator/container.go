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
	"github.com/reactiveops/fairwinds/pkg/validator/messages"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// ContainerValidation tracks validation failures associated with a Container.
type ContainerValidation struct {
	*ResourceValidation
	Container *corev1.Container
}

// ValidateContainer validates that each pod conforms to the Fairwinds config, returns a ResourceResult.
func ValidateContainer(cnConf *conf.Configuration, container *corev1.Container) ResourceResult {
	cv := ContainerValidation{
		Container: container,
		ResourceValidation: &ResourceValidation{
			Summary: &ResultSummary{},
		},
	}

	cv.validateResources(&cnConf.Resources)
	cv.validateHealthChecks(&cnConf.HealthChecks)
	cv.validateImage(&cnConf.Images)
	cv.validateNetworking(&cnConf.Networking)
	cv.validateSecurity(&cnConf.Security)

	cRes := ContainerResult{
		Name:     container.Name,
		Messages: cv.messages(),
	}

	rr := ResourceResult{
		Name:             container.Name,
		Type:             "Container",
		Summary:          cv.Summary,
		ContainerResults: []ContainerResult{cRes},
	}

	return rr
}

func (cv *ContainerValidation) validateResources(resConf *conf.Resources) {
	category := messages.CategoryResources
	res := cv.Container.Resources

	if resConf.CPURequestsMissing.IsActionable() && res.Requests.Cpu().MilliValue() == 0 {
		cv.addFailure(messages.CPURequestsFailure, resConf.CPURequestsMissing, category)
	} else {
		cv.validateResourceRange(messages.CPURequestsLabel, &resConf.CPURequestRanges, res.Requests.Cpu())
	}

	if resConf.CPULimitsMissing.IsActionable() && res.Limits.Cpu().MilliValue() == 0 {
		cv.addFailure(messages.CPULimitsFailure, resConf.CPULimitsMissing, category)
	} else {
		cv.validateResourceRange(messages.CPULimitsLabel, &resConf.CPULimitRanges, res.Requests.Cpu())
	}

	if resConf.MemoryRequestsMissing.IsActionable() && res.Requests.Memory().MilliValue() == 0 {
		cv.addFailure(messages.MemoryRequestsFailure, resConf.MemoryRequestsMissing, category)
	} else {
		cv.validateResourceRange(messages.MemoryRequestsLabel, &resConf.MemoryRequestRanges, res.Requests.Memory())
	}

	if resConf.MemoryLimitsMissing.IsActionable() && res.Limits.Memory().MilliValue() == 0 {
		cv.addFailure(messages.MemoryLimitsFailure, resConf.MemoryLimitsMissing, category)
	} else {
		cv.validateResourceRange(messages.MemoryLimitsLabel, &resConf.MemoryLimitRanges, res.Limits.Memory())
	}
}

func (cv *ContainerValidation) validateResourceRange(resourceName string, rangeConf *conf.ResourceRanges, res *resource.Quantity) {
	warnAbove := rangeConf.Warning.Above
	warnBelow := rangeConf.Warning.Below
	errorAbove := rangeConf.Error.Above
	errorBelow := rangeConf.Error.Below
	category := messages.CategoryResources

	if errorAbove != nil && errorAbove.MilliValue() < res.MilliValue() {
		cv.addError(fmt.Sprintf(messages.ResourceAmountTooHighFailure, resourceName, errorAbove.String()), category)
	} else if warnAbove != nil && warnAbove.MilliValue() < res.MilliValue() {
		cv.addWarning(fmt.Sprintf(messages.ResourceAmountTooHighFailure, resourceName, warnAbove.String()), category)
	} else if errorBelow != nil && errorBelow.MilliValue() > res.MilliValue() {
		cv.addError(fmt.Sprintf(messages.ResourceAmountTooLowFailure, resourceName, errorBelow.String()), category)
	} else if warnBelow != nil && warnBelow.MilliValue() > res.MilliValue() {
		cv.addWarning(fmt.Sprintf(messages.ResourceAmountTooLowFailure, resourceName, warnBelow.String()), category)
	} else {
		cv.addSuccess(fmt.Sprintf(messages.ResourceAmountSuccess, resourceName), category)
	}
}

func (cv *ContainerValidation) validateHealthChecks(conf *conf.HealthChecks) {
	category := messages.CategoryHealthChecks
	if conf.ReadinessProbeMissing.IsActionable() {
		if cv.Container.ReadinessProbe == nil {
			cv.addFailure(messages.ReadinessProbeFailure, conf.ReadinessProbeMissing, category)
		} else {
			cv.addSuccess(messages.ReadinessProbeSuccess, category)
		}
	}

	if conf.LivenessProbeMissing.IsActionable() {
		if cv.Container.LivenessProbe == nil {
			cv.addFailure(messages.LivenessProbeFailure, conf.LivenessProbeMissing, category)
		} else {
			cv.addSuccess(messages.LivenessProbeSuccess, category)
		}
	}
}

func (cv *ContainerValidation) validateImage(imageConf *conf.Images) {
	category := messages.CategoryImages
	if imageConf.TagNotSpecified.IsActionable() {
		img := strings.Split(cv.Container.Image, ":")
		if len(img) == 1 || img[1] == "latest" {
			cv.addFailure(messages.ImageTagFailure, imageConf.TagNotSpecified, category)
		} else {
			cv.addSuccess(messages.ImageTagSuccess, category)
		}
	}
}

func (cv *ContainerValidation) validateNetworking(networkConf *conf.Networking) {
	category := messages.CategoryNetworking
	if networkConf.HostPortSet.IsActionable() {
		hostPortSet := false
		for _, port := range cv.Container.Ports {
			if port.HostPort != 0 {
				hostPortSet = true
				break
			}
		}

		if hostPortSet {
			cv.addFailure(messages.HostPortFailure, networkConf.HostPortSet, category)
		} else {
			cv.addSuccess(messages.HostPortSuccess, category)
		}
	}
}

func (cv *ContainerValidation) validateSecurity(securityConf *conf.Security) {
	category := messages.CategorySecurity
	securityContext := cv.Container.SecurityContext
	if securityContext == nil {
		securityContext = &corev1.SecurityContext{}
	}

	if securityConf.RunAsRootAllowed.IsActionable() {
		if securityContext.RunAsNonRoot == (*bool)(nil) || !*securityContext.RunAsNonRoot {
			cv.addFailure(messages.RunAsRootFailure, securityConf.RunAsRootAllowed, category)
		} else {
			cv.addSuccess(messages.RunAsRootSuccess, category)
		}
	}

	if securityConf.RunAsPrivileged.IsActionable() {
		if securityContext.Privileged == (*bool)(nil) || !*securityContext.Privileged {
			cv.addSuccess(messages.RunAsPrivilegedSuccess, category)
		} else {
			cv.addFailure(messages.RunAsPrivilegedFailure, securityConf.RunAsPrivileged, category)
		}
	}

	if securityConf.NotReadOnlyRootFileSystem.IsActionable() {
		if securityContext.ReadOnlyRootFilesystem == (*bool)(nil) || !*securityContext.ReadOnlyRootFilesystem {
			cv.addFailure(messages.ReadOnlyFilesystemFailure, securityConf.NotReadOnlyRootFileSystem, category)
		} else {
			cv.addSuccess(messages.ReadOnlyFilesystemSuccess, category)
		}
	}

	if securityConf.PrivilegeEscalationAllowed.IsActionable() {
		if securityContext.AllowPrivilegeEscalation == (*bool)(nil) || !*securityContext.AllowPrivilegeEscalation {
			cv.addSuccess(messages.PrivilegeEscalationSuccess, category)
		} else {
			cv.addFailure(messages.PrivilegeEscalationFailure, securityConf.PrivilegeEscalationAllowed, category)
		}
	}

	hasSecurityError :=
		!cv.validateCapabilities(securityConf.Capabilities.Error, conf.SeverityError)
	hasSecurityWarning :=
		!cv.validateCapabilities(securityConf.Capabilities.Warning, conf.SeverityWarning)
	hasSecurityCheck := func(confLists conf.SecurityCapabilityLists) bool {
		return len(confLists.IfAnyAdded) > 0 ||
			len(confLists.IfAnyAddedBeyond) > 0 ||
			len(confLists.IfAnyNotDropped) > 0
	}
	if !hasSecurityError && !hasSecurityWarning &&
		(hasSecurityCheck(securityConf.Capabilities.Error) ||
			hasSecurityCheck(securityConf.Capabilities.Warning)) {
		cv.addSuccess(messages.SecurityCapabilitiesSuccess, category)
	}
}

func (cv *ContainerValidation) validateCapabilities(confLists conf.SecurityCapabilityLists, severity conf.Severity) bool {
	category := messages.CategorySecurity
	capabilities := &corev1.Capabilities{}
	if cv.Container.SecurityContext != nil && cv.Container.SecurityContext.Capabilities != nil {
		capabilities = cv.Container.SecurityContext.Capabilities
	}

	everythingOK := true
	if len(confLists.IfAnyAdded) > 0 {
		intersectAdds := capIntersection(capabilities.Add, confLists.IfAnyAdded)
		if len(intersectAdds) > 0 {
			capsString := commaSeparatedCapabilities(intersectAdds)
			cv.addFailure(fmt.Sprintf(messages.SecurityCapabilitiesAddedFailure, capsString), severity, category)
			everythingOK = false
		} else if capContains(capabilities.Add, "ALL") {
			cv.addFailure(fmt.Sprintf(messages.SecurityCapabilitiesAddedFailure, "ALL"), severity, category)
			everythingOK = false
		}
	}

	if len(confLists.IfAnyAddedBeyond) > 0 {
		differentAdds := capDifference(capabilities.Add, confLists.IfAnyAddedBeyond)
		if len(differentAdds) > 0 {
			capsString := commaSeparatedCapabilities(differentAdds)
			cv.addFailure(fmt.Sprintf(messages.SecurityCapabilitiesAddedFailure, capsString), severity, category)
			everythingOK = false
		} else if capContains(capabilities.Add, "ALL") {
			cv.addFailure(fmt.Sprintf(messages.SecurityCapabilitiesAddedFailure, "ALL"), severity, category)
			everythingOK = false
		}
	}

	if len(confLists.IfAnyNotDropped) > 0 {
		missingDrops := capDifference(confLists.IfAnyNotDropped, capabilities.Drop)
		if len(missingDrops) > 0 && !capContains(capabilities.Drop, "ALL") {
			capsString := commaSeparatedCapabilities(missingDrops)
			cv.addFailure(fmt.Sprintf(messages.SecurityCapabilitiesNotDroppedFailure, capsString), severity, category)
			everythingOK = false
		}
	}

	return everythingOK
}

func commaSeparatedCapabilities(caps []corev1.Capability) string {
	capsString := ""
	for _, cap := range caps {
		capsString = fmt.Sprintf("%s, %s", capsString, cap)
	}
	return capsString[2:]
}

func capIntersection(a, b []corev1.Capability) []corev1.Capability {
	result := []corev1.Capability{}
	hash := map[corev1.Capability]bool{}

	for _, s := range a {
		hash[s] = true
	}

	for _, s := range b {
		if hash[s] {
			result = append(result, s)
		}
	}

	return result
}

func capDifference(b, a []corev1.Capability) []corev1.Capability {
	result := []corev1.Capability{}
	hash := map[corev1.Capability]bool{}

	for _, s := range a {
		hash[s] = true
	}

	for _, s := range b {
		if !hash[s] {
			result = append(result, s)
		}
	}

	return result
}

func capContains(list []corev1.Capability, val corev1.Capability) bool {
	for _, s := range list {
		if s == val {
			return true
		}
	}

	return false
}
