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
	mssgs := []ResultMessage{}
	mssgs = append(mssgs, pv.Failures...)
	mssgs = append(mssgs, pv.Warnings...)
	mssgs = append(mssgs, pv.Successes...)
	return mssgs
}

// ValidatePod validates that each pod conforms to the Fairwinds config, returns a ResourceResult.
func ValidatePod(conf conf.Configuration, pod *corev1.PodSpec) ResourceResult {
	pv := PodValidation{
		Pod:     *pod,
		Summary: ResultSummary{},
	}

	pv.validateHostNetwork(conf.HostNetworking)

	pRes := PodResult{
		Messages:         pv.messages(),
		ContainerResults: []ContainerResult{},
	}

	// Add container resource results to the pod resource results.
	for _, container := range pod.InitContainers {
		ctrRR := validateContainer(conf, container)
		pv.Summary.Successes += ctrRR.Summary.Successes
		pv.Summary.Warnings += ctrRR.Summary.Warnings
		pv.Summary.Failures += ctrRR.Summary.Failures
		pRes.ContainerResults = append(
			pRes.ContainerResults,
			ctrRR.ContainerResults[0],
		)
	}

	for _, container := range pod.Containers {
		ctrRR := validateContainer(conf, container)
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

func (pv *PodValidation) validateHostNetwork(conf conf.HostNetworking) {
	pv.hostAlias(conf)
	pv.hostIPC(conf)
	pv.hostPID(conf)
	pv.hostNetwork(conf)
}

func (pv *PodValidation) hostAlias(conf conf.HostNetworking) {
	if conf.HostAlias.Require {
		for _, alias := range pv.Pod.HostAliases {
			if alias.IP != "" && len(alias.Hostnames) == 0 {
				pv.addFailure("Host alias should is configured, but it shouldn't be")
				return
			}
		}
		pv.addSuccess("Host alias is not configured")
	}
}

func (pv *PodValidation) hostIPC(conf conf.HostNetworking) {
	if conf.HostIPC.Require {
		if pv.Pod.HostIPC {
			pv.addFailure("Host IPC is configured, but it shouldn't be")
			return
		}
		pv.addSuccess("Host IPC is not configured")
	}
}

func (pv *PodValidation) hostPID(conf conf.HostNetworking) {
	if conf.HostPID.Require {
		if pv.Pod.HostPID {
			pv.addFailure("Host PID is configured, but it shouldn't be")
			return
		}
		pv.addSuccess("Host PID is not configured")
	}
}

func (pv *PodValidation) hostNetwork(conf conf.HostNetworking) {
	if conf.HostNetwork.Require {
		if pv.Pod.HostNetwork {
			pv.addFailure("Host network is configured, but it shouldn't be")
			return
		}
		pv.addSuccess("Host network is not configured")
	}
}
