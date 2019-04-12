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
	res := cv.Container.Resources

	if resConf.CPURequestsMissing.IsActionable() && res.Requests.Cpu().MilliValue() == 0 {
		cv.addFailure(messages.CPURequestsFailure, resConf.CPURequestsMissing)
	} else {
		cv.validateResourceRange(messages.CPURequestsLabel, &resConf.CPURequestRanges, res.Requests.Cpu())
	}

	if resConf.CPULimitsMissing.IsActionable() && res.Limits.Cpu().MilliValue() == 0 {
		cv.addFailure(messages.CPULimitsFailure, resConf.CPULimitsMissing)
	} else {
		cv.validateResourceRange(messages.CPULimitsLabel, &resConf.CPULimitRanges, res.Requests.Cpu())
	}

	if resConf.MemoryRequestsMissing.IsActionable() && res.Requests.Memory().MilliValue() == 0 {
		cv.addFailure(messages.MemoryRequestsFailure, resConf.MemoryRequestsMissing)
	} else {
		cv.validateResourceRange(messages.MemoryRequestsLabel, &resConf.MemoryRequestRanges, res.Requests.Memory())
	}

	if resConf.MemoryLimitsMissing.IsActionable() && res.Limits.Memory().MilliValue() == 0 {
		cv.addFailure(messages.MemoryLimitsFailure, resConf.MemoryLimitsMissing)
	} else {
		cv.validateResourceRange(messages.MemoryLimitsLabel, &resConf.MemoryLimitRanges, res.Limits.Memory())
	}
}

func (cv *ContainerValidation) validateResourceRange(resourceName string, rangeConf *conf.ResourceRanges, res *resource.Quantity) {
	warnAbove := rangeConf.Warning.Above
	warnBelow := rangeConf.Warning.Below
	errorAbove := rangeConf.Error.Above
	errorBelow := rangeConf.Error.Below

	if errorAbove != nil && errorAbove.MilliValue() < res.MilliValue() {
		cv.addError(fmt.Sprintf(messages.ResourceAmountTooHighFailure, resourceName, errorAbove.String()))
	} else if warnAbove != nil && warnAbove.MilliValue() < res.MilliValue() {
		cv.addWarning(fmt.Sprintf(messages.ResourceAmountTooHighFailure, resourceName, warnAbove.String()))
	} else if errorBelow != nil && errorBelow.MilliValue() > res.MilliValue() {
		cv.addError(fmt.Sprintf(messages.ResourceAmountTooLowFailure, resourceName, errorBelow.String()))
	} else if warnBelow != nil && warnBelow.MilliValue() > res.MilliValue() {
		cv.addWarning(fmt.Sprintf(messages.ResourceAmountTooLowFailure, resourceName, warnBelow.String()))
	} else {
		cv.addSuccess(fmt.Sprintf(messages.ResourceAmountSuccess, resourceName))
	}
}

func (cv *ContainerValidation) validateHealthChecks(conf *conf.HealthChecks) {
	if conf.ReadinessProbeMissing.IsActionable() {
		if cv.Container.ReadinessProbe == nil {
			cv.addFailure(messages.ReadinessProbeFailure, conf.ReadinessProbeMissing)
		} else {
			cv.addSuccess(messages.ReadinessProbeSuccess)
		}
	}

	if conf.LivenessProbeMissing.IsActionable() {
		if cv.Container.LivenessProbe == nil {
			cv.addFailure(messages.LivenessProbeFailure, conf.LivenessProbeMissing)
		} else {
			cv.addSuccess(messages.LivenessProbeSuccess)
		}
	}
}

func (cv *ContainerValidation) validateImage(imageConf *conf.Images) {
	if imageConf.TagNotSpecified.IsActionable() {
		img := strings.Split(cv.Container.Image, ":")
		if len(img) == 1 || img[1] == "latest" {
			cv.addFailure(messages.ImageTagFailure, imageConf.TagNotSpecified)
		} else {
			cv.addSuccess(messages.ImageTagSuccess)
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
			cv.addFailure(messages.HostPortFailure, networkConf.HostAliasSet)
		} else {
			cv.addSuccess(messages.HostPortSuccess)
		}
	}
}

func (cv *ContainerValidation) validateSecurity(securityConf *conf.Security) {
	securityContext := cv.Container.SecurityContext
	if securityContext == nil {
		securityContext = &corev1.SecurityContext{}
	}

	if securityConf.RunAsRootAllowed.IsActionable() {
		if securityContext.RunAsNonRoot == (*bool)(nil) || !*securityContext.RunAsNonRoot {
			cv.addFailure(messages.RunAsRootFailure, securityConf.RunAsRootAllowed)
		} else {
			cv.addSuccess(messages.RunAsRootSuccess)
		}
	}

	if securityConf.RunAsPrivileged.IsActionable() {
		if securityContext.Privileged == (*bool)(nil) || !*securityContext.Privileged {
			cv.addSuccess(messages.RunAsPrivilegedSuccess)
		} else {
			cv.addFailure(messages.RunAsPrivilegedFailure, securityConf.RunAsPrivileged)
		}
	}

	if securityConf.NotReadOnlyRootFileSystem.IsActionable() {
		if securityContext.ReadOnlyRootFilesystem == (*bool)(nil) || !*securityContext.ReadOnlyRootFilesystem {
			cv.addFailure(messages.ReadOnlyFilesystemFailure, securityConf.NotReadOnlyRootFileSystem)
		} else {
			cv.addSuccess(messages.ReadOnlyFilesystemSuccess)
		}
	}

	if securityConf.PrivilegeEscalationAllowed.IsActionable() {
		if securityContext.AllowPrivilegeEscalation == (*bool)(nil) || !*securityContext.AllowPrivilegeEscalation {
			cv.addSuccess(messages.PrivilegeEscalationSuccess)
		} else {
			cv.addFailure(messages.PrivilegeEscalationFailure, securityConf.PrivilegeEscalationAllowed)
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
		cv.addSuccess(messages.SecurityCapabilitiesSuccess)
	}
}

func (cv *ContainerValidation) validateCapabilities(confLists conf.SecurityCapabilityLists, severity conf.Severity) bool {
	capabilities := &corev1.Capabilities{}
	if cv.Container.SecurityContext != nil && cv.Container.SecurityContext.Capabilities != nil {
		capabilities = cv.Container.SecurityContext.Capabilities
	}

	everythingOK := true
	if len(confLists.IfAnyAdded) > 0 {
		intersectAdds := capIntersection(capabilities.Add, confLists.IfAnyAdded)
		if len(intersectAdds) > 0 {
			capsString := commaSeparatedCapabilities(intersectAdds)
			cv.addFailure(fmt.Sprintf(messages.SecurityCapabilitiesAddedFailure, capsString), severity)
			everythingOK = false
		} else if capContains(capabilities.Add, "ALL") {
			cv.addFailure(fmt.Sprintf(messages.SecurityCapabilitiesAddedFailure, "ALL"), severity)
			everythingOK = false
		}
	}

	if len(confLists.IfAnyAddedBeyond) > 0 {
		differentAdds := capDifference(capabilities.Add, confLists.IfAnyAddedBeyond)
		if len(differentAdds) > 0 {
			capsString := commaSeparatedCapabilities(differentAdds)
			cv.addFailure(fmt.Sprintf(messages.SecurityCapabilitiesAddedFailure, capsString), severity)
			everythingOK = false
		} else if capContains(capabilities.Add, "ALL") {
			cv.addFailure(fmt.Sprintf(messages.SecurityCapabilitiesAddedFailure, "ALL"), severity)
			everythingOK = false
		}
	}

	if len(confLists.IfAnyNotDropped) > 0 {
		missingDrops := capDifference(confLists.IfAnyNotDropped, capabilities.Drop)
		if len(missingDrops) > 0 && !capContains(capabilities.Drop, "ALL") {
			capsString := commaSeparatedCapabilities(missingDrops)
			cv.addFailure(fmt.Sprintf(messages.SecurityCapabilitiesNotDroppedFailure, capsString), severity)
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
