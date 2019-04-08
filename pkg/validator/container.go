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
		cv.addFailure("CPU Requests are not set", resConf.CPURequestsMissing)
	} else {
		cv.validateResourceRange("CPU Requests", &resConf.CPURequestRanges, res.Requests.Cpu())
	}

	if resConf.CPULimitsMissing.IsActionable() && res.Limits.Cpu().MilliValue() == 0 {
		cv.addFailure("CPU Limits are not set", resConf.CPULimitsMissing)
	} else {
		cv.validateResourceRange("CPU Limits", &resConf.CPULimitRanges, res.Requests.Cpu())
	}

	if resConf.MemoryRequestsMissing.IsActionable() && res.Requests.Memory().MilliValue() == 0 {
		cv.addFailure("Memory Requests are not set", resConf.MemoryRequestsMissing)
	} else {
		cv.validateResourceRange("Memory Requests", &resConf.MemoryRequestRanges, res.Requests.Memory())
	}

	if resConf.MemoryLimitsMissing.IsActionable() && res.Limits.Memory().MilliValue() == 0 {
		cv.addFailure("Memory Limits are not set", resConf.MemoryLimitsMissing)
	} else {
		cv.validateResourceRange("Memory Limits", &resConf.MemoryLimitRanges, res.Limits.Memory())
	}
}

func (cv *ContainerValidation) validateResourceRange(resourceName string, rangeConf *conf.ResourceRanges, res *resource.Quantity) {
	warnAbove := rangeConf.Warning.Above
	warnBelow := rangeConf.Warning.Below
	errorAbove := rangeConf.Error.Above
	errorBelow := rangeConf.Error.Below

	if errorAbove != nil && errorAbove.MilliValue() < res.MilliValue() {
		cv.addError(fmt.Sprintf("%s are too high", resourceName))
	} else if warnAbove != nil && warnAbove.MilliValue() < res.MilliValue() {
		cv.addWarning(fmt.Sprintf("%s are too high", resourceName))
	} else if errorBelow != nil && errorBelow.MilliValue() > res.MilliValue() {
		cv.addError(fmt.Sprintf("%s are too low", resourceName))
	} else if warnBelow != nil && warnBelow.MilliValue() > res.MilliValue() {
		cv.addWarning(fmt.Sprintf("%s are too low", resourceName))
	} else {
		cv.addSuccess(fmt.Sprintf("%s are within the expected range", resourceName))
	}
}

func (cv *ContainerValidation) validateHealthChecks(conf *conf.HealthChecks) {
	if conf.ReadinessProbeMissing.IsActionable() {
		if cv.Container.ReadinessProbe == nil {
			cv.addFailure("Readiness probe needs to be configured", conf.ReadinessProbeMissing)
		} else {
			cv.addSuccess("Readiness probe configured")
		}
	}

	if conf.LivenessProbeMissing.IsActionable() {
		if cv.Container.LivenessProbe == nil {
			cv.addFailure("Liveness probe needs to be configured", conf.LivenessProbeMissing)
		} else {
			cv.addSuccess("Liveness probe configured")
		}
	}
}

func (cv *ContainerValidation) validateImage(imageConf *conf.Images) {
	if imageConf.TagNotSpecified.IsActionable() {
		img := strings.Split(cv.Container.Image, ":")
		if len(img) == 1 || img[1] == "latest" {
			cv.addFailure("Image tag should be specified", imageConf.TagNotSpecified)
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
			cv.addFailure("Host port is configured, but it shouldn't be", networkConf.HostAliasSet)
		} else {
			cv.addSuccess("Host port is not configured")
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
			cv.addFailure("Container is allowed to run as root", securityConf.RunAsRootAllowed)
		} else {
			cv.addSuccess("Container is not allowed to run as root")
		}
	}

	if securityConf.RunAsPrivileged.IsActionable() {
		if securityContext.Privileged == (*bool)(nil) || !*securityContext.Privileged {
			cv.addSuccess("Container is not running as privileged")
		} else {
			cv.addFailure("Container is running as privileged", securityConf.RunAsPrivileged)
		}
	}

	if securityConf.NotReadOnlyRootFileSystem.IsActionable() {
		if securityContext.ReadOnlyRootFilesystem == (*bool)(nil) || !*securityContext.ReadOnlyRootFilesystem {
			cv.addFailure("Container is not running with a read only filesystem", securityConf.NotReadOnlyRootFileSystem)
		} else {
			cv.addSuccess("Container is running with a read only filesystem")
		}
	}

	if securityConf.PrivilegeEscalationAllowed.IsActionable() {
		if securityContext.AllowPrivilegeEscalation == (*bool)(nil) || !*securityContext.AllowPrivilegeEscalation {
			cv.addSuccess("Container does not allow privilege escalation")
		} else {
			cv.addFailure("Container allows privilege escalation", securityConf.PrivilegeEscalationAllowed)
		}
	}

	cv.validateCapabilities(securityConf.Capabilities.Error, conf.SeverityError)
	cv.validateCapabilities(securityConf.Capabilities.Warning, conf.SeverityWarning)
}

func (cv *ContainerValidation) validateCapabilities(confLists conf.SecurityCapabilityLists, severity conf.Severity) {
	capabilities := &corev1.Capabilities{}
	if cv.Container.SecurityContext != nil && cv.Container.SecurityContext.Capabilities != nil {
		capabilities = cv.Container.SecurityContext.Capabilities
	}

	if len(confLists.IfAnyAdded) > 0 {
		intersectAdds := capIntersection(capabilities.Add, confLists.IfAnyAdded)
		if len(intersectAdds) > 0 {
			capsString := commaSeparatedCapabilities(intersectAdds)
			cv.addFailure(fmt.Sprintf("Security capabilities added from %v list: %v", severity, capsString), severity)
		} else if capContains(capabilities.Add, "ALL") {
			cv.addFailure(fmt.Sprintf("All security capabilities added, violating %v list", severity), severity)
		} else {
			cv.addSuccess(fmt.Sprintf("No security capabilities added from %v list", severity))
		}
	}

	if len(confLists.IfAnyAddedBeyond) > 0 {
		differentAdds := capDifference(capabilities.Add, confLists.IfAnyAddedBeyond)
		if len(differentAdds) > 0 {
			capsString := commaSeparatedCapabilities(differentAdds)
			cv.addFailure(fmt.Sprintf("Security capabilities added beyond %v list: %v", severity, capsString), severity)
		} else if capContains(capabilities.Add, "ALL") {
			cv.addFailure(fmt.Sprintf("All security capabilities added, going beyond %v list", severity), severity)
		} else {
			cv.addSuccess(fmt.Sprintf("No security capabilities added beyond %v list", severity))
		}
	}

	if len(confLists.IfAnyNotDropped) > 0 {
		missingDrops := capDifference(confLists.IfAnyNotDropped, capabilities.Drop)

		if len(missingDrops) > 0 && !capContains(capabilities.Drop, "ALL") {
			capsString := commaSeparatedCapabilities(missingDrops)
			cv.addFailure(fmt.Sprintf("Security capabilities not dropped from %v list: %v", severity, capsString), severity)
		} else {
			cv.addSuccess(fmt.Sprintf("Required security capabilities dropped from %v list", severity))
		}
	}
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
