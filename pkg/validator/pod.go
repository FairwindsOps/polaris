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
	conf "github.com/reactiveops/fairwinds/pkg/config"
	corev1 "k8s.io/api/core/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("Fairwinds Validator")

// PodValidation tracks validation failures associated with a Pod.
type PodValidation struct {
	Pod       corev1.PodSpec
	Summary   ResultSummary
	Failures  []ResultMessage
	Warnings  []ResultMessage
	Successes []ResultMessage
}

func (pv *PodValidation) messages() []ResultMessage {
	messages := []ResultMessage{}
	messages = append(messages, pv.Failures...)
	messages = append(messages, pv.Warnings...)
	messages = append(messages, pv.Successes...)
	return messages
}

// ValidatePod validates that each pod conforms to the Fairwinds config, returns a ResourceResult.
func ValidatePod(podConf conf.Configuration, pod *corev1.PodSpec) ResourceResult {
	pv := PodValidation{
		Pod:     *pod,
		Summary: ResultSummary{},
	}

	pv.validateNetworking(&podConf.Networking)

	pRes := PodResult{
		Messages:         pv.messages(),
		ContainerResults: []ContainerResult{},
	}

	// Add container resource results to the pod resource results.
	for _, container := range pod.InitContainers {
		ctrRR := validateContainer(podConf, container)
		pv.Summary.Successes += ctrRR.Summary.Successes
		pv.Summary.Warnings += ctrRR.Summary.Warnings
		pv.Summary.Failures += ctrRR.Summary.Failures
		pRes.ContainerResults = append(
			pRes.ContainerResults,
			ctrRR.ContainerResults[0],
		)
	}

	for _, container := range pod.Containers {
		ctrRR := validateContainer(podConf, container)
		pv.Summary.Successes += ctrRR.Summary.Successes
		pv.Summary.Warnings += ctrRR.Summary.Warnings
		pv.Summary.Failures += ctrRR.Summary.Failures
		pRes.ContainerResults = append(
			pRes.ContainerResults,
			ctrRR.ContainerResults[0],
		)
	}
	rr := ResourceResult{
		Type:       "Pod",
		Summary:    &pv.Summary,
		PodResults: []PodResult{pRes},
	}

	return rr
}

func (pv *PodValidation) addMessage(message string, severity conf.Severity) {
	if severity == conf.SeverityError {
		pv.addFailure(message)
	} else if severity == conf.SeverityWarning {
		pv.addWarning(message)
	}
}

func (pv *PodValidation) addFailure(message string) {
	pv.Summary.Failures++
	pv.Failures = append(pv.Failures, ResultMessage{
		Message: message,
		Type:    "failure",
	})
}

func (pv *PodValidation) addWarning(message string) {
	pv.Summary.Warnings++
	pv.Warnings = append(pv.Warnings, ResultMessage{
		Message: message,
		Type:    "warning",
	})
}

func (pv *PodValidation) addSuccess(message string) {
	pv.Summary.Successes++
	pv.Successes = append(pv.Successes, ResultMessage{
		Message: message,
		Type:    "success",
	})
}

func (pv *PodValidation) validateNetworking(networkConf *conf.Networking) {
	pv.validateHostAlias(networkConf)
	pv.validateHostIPC(networkConf)
	pv.validateHostPID(networkConf)
	pv.validateHostNetwork(networkConf)
}

func (pv *PodValidation) validateHostAlias(networkConf *conf.Networking) {
	if networkConf.HostAliasSet.IsActionable() {
		hostAliasSet := false
		for _, alias := range pv.Pod.HostAliases {
			if alias.IP != "" && len(alias.Hostnames) == 0 {
				hostAliasSet = true
				break
			}
		}

		if hostAliasSet {
			pv.addMessage("Host alias should is configured, but it shouldn't be", networkConf.HostAliasSet)
		} else {
			pv.addSuccess("Host alias is not configured")
		}
	}
}

func (pv *PodValidation) validateHostIPC(networkConf *conf.Networking) {
	if networkConf.HostIPCSet.IsActionable() {
		if pv.Pod.HostIPC {
			pv.addMessage("Host IPC is configured, but it shouldn't be", networkConf.HostIPCSet)
		} else {
			pv.addSuccess("Host IPC is not configured")
		}
	}
}

func (pv *PodValidation) validateHostPID(networkConf *conf.Networking) {
	if networkConf.HostPIDSet.IsActionable() {
		if pv.Pod.HostPID {
			pv.addMessage("Host PID is configured, but it shouldn't be", networkConf.HostPIDSet)
		} else {
			pv.addSuccess("Host PID is not configured")
		}
	}
}

func (pv *PodValidation) validateHostNetwork(networkConf *conf.Networking) {
	if networkConf.HostNetworkSet.IsActionable() {
		if pv.Pod.HostNetwork {
			pv.addMessage("Host network is configured, but it shouldn't be", networkConf.HostNetworkSet)
		} else {
			pv.addSuccess("Host network is not configured")
		}
	}
}
