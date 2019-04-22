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
	"github.com/reactiveops/fairwinds/pkg/validator/messages"
	corev1 "k8s.io/api/core/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("Fairwinds Validator")

// PodValidation tracks validation failures associated with a Pod.
type PodValidation struct {
	*ResourceValidation
	Pod *corev1.PodSpec
}

// ValidatePod validates that each pod conforms to the Fairwinds config, returns a ResourceResult.
func ValidatePod(podConf conf.Configuration, pod *corev1.PodSpec) ResourceResult {
	pv := PodValidation{
		Pod: pod,
		ResourceValidation: &ResourceValidation{
			Summary: &ResultSummary{},
		},
	}

	pv.validateSecurity(&podConf.Security)
	pv.validateNetworking(&podConf.Networking)

	pRes := PodResult{
		Messages:         pv.messages(),
		ContainerResults: []ContainerResult{},
	}

	pv.validateContainers(pod.InitContainers, &pRes, &podConf)
	pv.validateContainers(pod.Containers, &pRes, &podConf)

	rr := ResourceResult{
		Type:       "Pod",
		Summary:    pv.Summary,
		PodResults: []PodResult{pRes},
	}

	return rr
}

func (pv *PodValidation) validateContainers(containers []corev1.Container, pRes *PodResult, podConf *conf.Configuration) {
	for _, container := range containers {
		ctrRR := ValidateContainer(podConf, &container)
		pv.Summary.Successes += ctrRR.Summary.Successes
		pv.Summary.Warnings += ctrRR.Summary.Warnings
		pv.Summary.Errors += ctrRR.Summary.Errors
		pRes.ContainerResults = append(
			pRes.ContainerResults,
			ctrRR.ContainerResults[0],
		)
	}
}

func (pv *PodValidation) validateSecurity(securityConf *conf.Security) {
	category := messages.CategorySecurity

	if securityConf.HostIPCSet.IsActionable() {
		if pv.Pod.HostIPC {
			pv.addFailure(messages.HostIPCFailure, securityConf.HostIPCSet, category)
		} else {
			pv.addSuccess(messages.HostIPCSuccess, category)
		}
	}

	if securityConf.HostPIDSet.IsActionable() {
		if pv.Pod.HostPID {
			pv.addFailure(messages.HostPIDFailure, securityConf.HostPIDSet, category)
		} else {
			pv.addSuccess(messages.HostPIDSuccess, category)
		}
	}
}

func (pv *PodValidation) validateNetworking(networkConf *conf.Networking) {
	category := messages.CategoryNetworking

	if networkConf.HostNetworkSet.IsActionable() {
		if pv.Pod.HostNetwork {
			pv.addFailure(messages.HostNetworkFailure, networkConf.HostNetworkSet, category)
		} else {
			pv.addSuccess(messages.HostNetworkSuccess, category)
		}
	}
}
