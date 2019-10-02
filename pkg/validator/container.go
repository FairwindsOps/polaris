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
	"strings"

	conf "github.com/fairwindsops/polaris/pkg/config"
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
func ValidateContainer(container *corev1.Container, parentPodResult *PodResult, cnConf *conf.Configuration, isInit bool) ContainerResult {
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

	cv.validateResources(&cnConf.Resources)
	cv.validateHealthChecks(&cnConf.HealthChecks)
	cv.validateImage(&cnConf.Images)
	cv.validateNetworking(&cnConf.Networking)
	cv.validateSecurity(&cnConf.Security)

	cRes := ContainerResult{
		Name:     container.Name,
		Messages: cv.messages(),
		Summary:  cv.summary(),
	}

	return cRes
}

func (cv *ContainerValidation) validateResources(resConf *conf.Resources) {
	// Only validate resources for primary containers. Although it can
	// be helpful to set these in certain cases, it usually isn't
	if cv.IsInitContainer {
		return
	}

	category := messages.CategoryResources
	res := cv.Container.Resources

	if resConf.CPURequestsMissing.IsActionable() && res.Requests.Cpu().MilliValue() == 0 {
		id := getIDFromField(*resConf, "CPURequestsMissing")
		cv.addFailure(messages.CPURequestsFailure, resConf.CPURequestsMissing, category, id)
	} else {
		id := getIDFromField(*resConf, "CPURequestRanges")
		cv.validateResourceRange(id, messages.CPURequestsLabel, &resConf.CPURequestRanges, res.Requests.Cpu())
	}

	if resConf.CPULimitsMissing.IsActionable() && res.Limits.Cpu().MilliValue() == 0 {
		id := getIDFromField(*resConf, "CPULimitsMissing")
		cv.addFailure(messages.CPULimitsFailure, resConf.CPULimitsMissing, category, id)
	} else {
		id := getIDFromField(*resConf, "CPULimitRanges")
		cv.validateResourceRange(id, messages.CPULimitsLabel, &resConf.CPULimitRanges, res.Requests.Cpu())
	}

	if resConf.MemoryRequestsMissing.IsActionable() && res.Requests.Memory().MilliValue() == 0 {
		id := getIDFromField(*resConf, "MemoryRequestsMissing")
		cv.addFailure(messages.MemoryRequestsFailure, resConf.MemoryRequestsMissing, category, id)
	} else {
		id := getIDFromField(*resConf, "MemoryRequestRanges")
		cv.validateResourceRange(id, messages.MemoryRequestsLabel, &resConf.MemoryRequestRanges, res.Requests.Memory())
	}

	if resConf.MemoryLimitsMissing.IsActionable() && res.Limits.Memory().MilliValue() == 0 {
		id := getIDFromField(*resConf, "MemoryLimitsMissing")
		cv.addFailure(messages.MemoryLimitsFailure, resConf.MemoryLimitsMissing, category, id)
	} else {
		id := getIDFromField(*resConf, "MemoryLimitRanges")
		cv.validateResourceRange(id, messages.MemoryLimitsLabel, &resConf.MemoryLimitRanges, res.Limits.Memory())
	}
}

func (cv *ContainerValidation) validateResourceRange(id, resourceName string, rangeConf *conf.ResourceRanges, res *resource.Quantity) {
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
	} else {
		if warnAbove != nil || warnBelow != nil || errorAbove != nil || errorBelow != nil {
			cv.addSuccess(fmt.Sprintf(messages.ResourceAmountSuccess, resourceName), category, id)
		} else {
			cv.addSuccess(fmt.Sprintf(messages.ResourcePresentSuccess, resourceName), category, id)
		}
	}
}

func (cv *ContainerValidation) validateHealthChecks(conf *conf.HealthChecks) {
	category := messages.CategoryHealthChecks

	// Don't validate readiness probes on init containers
	if !cv.IsInitContainer && conf.ReadinessProbeMissing.IsActionable() {
		id := getIDFromField(*conf, "ReadinessProbeMissing")
		if cv.Container.ReadinessProbe == nil {
			cv.addFailure(messages.ReadinessProbeFailure, conf.ReadinessProbeMissing, category, id)
		} else {
			cv.addSuccess(messages.ReadinessProbeSuccess, category, id)
		}
	}

	if conf.LivenessProbeMissing.IsActionable() {
		id := getIDFromField(*conf, "LivenessProbeMissing")
		if cv.Container.LivenessProbe == nil {
			cv.addFailure(messages.LivenessProbeFailure, conf.LivenessProbeMissing, category, id)
		} else {
			cv.addSuccess(messages.LivenessProbeSuccess, category, id)
		}
	}
}

func (cv *ContainerValidation) validateImage(imageConf *conf.Images) {
	category := messages.CategoryImages
	if imageConf.PullPolicyNotAlways.IsActionable() {
		id := getIDFromField(*imageConf, "PullPolicyNotAlways")
		if cv.Container.ImagePullPolicy != corev1.PullAlways {
			cv.addFailure(messages.ImagePullPolicyFailure, imageConf.PullPolicyNotAlways, category, id)
		} else {
			cv.addSuccess(messages.ImagePullPolicySuccess, category, id)
		}
	}

	if imageConf.TagNotSpecified.IsActionable() {
		id := getIDFromField(*imageConf, "TagNotSpecified")
		img := strings.Split(cv.Container.Image, ":")
		if len(img) == 1 || img[1] == "latest" {
			cv.addFailure(messages.ImageTagFailure, imageConf.TagNotSpecified, category, id)
		} else {
			cv.addSuccess(messages.ImageTagSuccess, category, id)
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

		id := getIDFromField(*networkConf, "HostPortSet")
		if hostPortSet {
			cv.addFailure(messages.HostPortFailure, networkConf.HostPortSet, category, id)
		} else {
			cv.addSuccess(messages.HostPortSuccess, category, id)
		}
	}
}

func (cv *ContainerValidation) validateSecurity(securityConf *conf.Security) {
	category := messages.CategorySecurity
	securityContext := cv.Container.SecurityContext
	podSecurityContext := cv.parentPodSpec.SecurityContext

	// Support an empty container security context
	if securityContext == nil {
		securityContext = &corev1.SecurityContext{}
	}

	// Support an empty pod security context
	if podSecurityContext == nil {
		podSecurityContext = &corev1.PodSecurityContext{}
	}

	if securityConf.RunAsRootAllowed.IsActionable() {
		id := getIDFromField(*securityConf, "RunAsRootAllowed")
		if getBoolValue(securityContext.RunAsNonRoot) {
			// Check if the container is explicitly set to True (pass)
			cv.addSuccess(messages.RunAsRootSuccess, category, id)
		} else if securityContext.RunAsNonRoot == nil {
			// Check if the value in the container spec if nil (thus defaulting to the podspec)
			// Check if the container value is not set
			if getBoolValue(podSecurityContext.RunAsNonRoot) {
				// if the pod spec default for containers is true, then pass
				cv.addSuccess(messages.RunAsRootSuccess, category, id)
			} else {
				// else fail as RunAsNonRoot defaults to false
				cv.addFailure(messages.RunAsRootFailure, securityConf.RunAsRootAllowed, category, id)
			}
		} else {
			cv.addFailure(messages.RunAsRootFailure, securityConf.RunAsRootAllowed, category, id)
		}
	}

	if securityConf.RunAsPrivileged.IsActionable() {
		id := getIDFromField(*securityConf, "RunAsPrivileged")
		if getBoolValue(securityContext.Privileged) {
			cv.addFailure(messages.RunAsPrivilegedFailure, securityConf.RunAsPrivileged, category, id)
		} else {
			cv.addSuccess(messages.RunAsPrivilegedSuccess, category, id)
		}
	}

	if securityConf.NotReadOnlyRootFileSystem.IsActionable() {
		id := getIDFromField(*securityConf, "NotReadOnlyRootFileSystem")
		if getBoolValue(securityContext.ReadOnlyRootFilesystem) {
			cv.addSuccess(messages.ReadOnlyFilesystemSuccess, category, id)
		} else {
			cv.addFailure(messages.ReadOnlyFilesystemFailure, securityConf.NotReadOnlyRootFileSystem, category, id)
		}
	}

	if securityConf.PrivilegeEscalationAllowed.IsActionable() {
		id := getIDFromField(*securityConf, "PrivilegeEscalationAllowed")
		if getBoolValue(securityContext.AllowPrivilegeEscalation) {
			cv.addFailure(messages.PrivilegeEscalationFailure, securityConf.PrivilegeEscalationAllowed, category, id)
		} else {
			cv.addSuccess(messages.PrivilegeEscalationSuccess, category, id)
		}
	}

	cv.validateCapabilities(&securityConf.Capabilities.Warning, &securityConf.Capabilities.Error)
}

func (cv *ContainerValidation) validateCapabilities(warningLists *conf.SecurityCapabilityLists, errorLists *conf.SecurityCapabilityLists) {
	category := messages.CategorySecurity
	capabilities := &corev1.Capabilities{}
	if cv.Container.SecurityContext != nil && cv.Container.SecurityContext.Capabilities != nil {
		capabilities = cv.Container.SecurityContext.Capabilities
	}
	allLists := []*conf.SecurityCapabilityLists{warningLists, errorLists}

	addID := "capabilitiesAdded"
	hasAddFailure := false
	hasAddCheck := false
	for _, confLists := range allLists {
		if len(confLists.IfAnyAdded) == 0 && len(confLists.IfAnyAddedBeyond) == 0 {
			continue
		}
		hasAddCheck = true
		var severity conf.Severity
		if confLists == warningLists {
			severity = conf.SeverityWarning
		} else {
			severity = conf.SeverityError
		}
		badAdds := make([]corev1.Capability, 0)
		if len(confLists.IfAnyAdded) > 0 {
			intersectAdds := capIntersection(capabilities.Add, confLists.IfAnyAdded)
			badAdds = append(badAdds, intersectAdds...)
		}
		if len(confLists.IfAnyAddedBeyond) > 0 {
			differentAdds := capDifference(capabilities.Add, confLists.IfAnyAddedBeyond)
			differentAdds = capDifference(differentAdds, badAdds)
			badAdds = append(badAdds, differentAdds...)
		}
		if capContains(capabilities.Add, "ALL") && !capContains(badAdds, "ALL") {
			badAdds = append(badAdds, "ALL")
		}
		if len(badAdds) > 0 {
			hasAddFailure = true
			capsString := commaSeparatedCapabilities(badAdds)
			cv.addFailure(fmt.Sprintf(messages.SecurityCapabilitiesAddedFailure, capsString), severity, category, addID)
		}
	}
	if hasAddCheck && !hasAddFailure {
		cv.addSuccess(messages.SecurityCapabilitiesAddedSuccess, category, addID)
	}

	dropID := "capabilitiesDropped"
	hasDropCheck := false
	hasDropFailure := false
	for _, confLists := range allLists {
		if len(confLists.IfAnyNotDropped) == 0 {
			continue
		}
		hasDropCheck = true
		var severity conf.Severity
		if confLists == warningLists {
			severity = conf.SeverityWarning
		} else {
			severity = conf.SeverityError
		}
		missingDrops := capDifference(confLists.IfAnyNotDropped, capabilities.Drop)
		id := "capabilitiesNotDropped"
		if len(missingDrops) > 0 && !capContains(capabilities.Drop, "ALL") {
			hasDropFailure = true
			capsString := commaSeparatedCapabilities(missingDrops)
			cv.addFailure(fmt.Sprintf(messages.SecurityCapabilitiesNotDroppedFailure, capsString), severity, category, id)
		}
	}
	if hasDropCheck && !hasDropFailure {
		cv.addSuccess(messages.SecurityCapabilitiesNotDroppedSuccess, category, dropID)
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

// getBoolValue returns false if nil or returns the value of the bool pointer
func getBoolValue(val *bool) bool {
	if val == nil {
		return false
	}

	return *val
}
