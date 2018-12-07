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
	"context"
	"net/http"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission/types"
)

var log = logf.Log.WithName("Fairwinds Validator")

// PodValidator validates Pods
type PodValidator struct {
	client  client.Client
	decoder types.Decoder
}

// Implement admission.Handler so the controller can handle admission request.
var _ admission.Handler = &PodValidator{}

// Handle for PodValidator admits a pod if validation passes.
func (v *PodValidator) Handle(ctx context.Context, req types.Request) types.Response {
	pod := &corev1.Pod{}

	err := v.decoder.Decode(req, pod)
	if err != nil {
		return admission.ErrorResponse(http.StatusBadRequest, err)
	}

	allowed, reason, err := v.validatePods(ctx, pod)
	if err != nil {
		return admission.ErrorResponse(http.StatusInternalServerError, err)
	}
	return admission.ValidationResponse(allowed, reason)
}

func (v *PodValidator) validatePods(ctx context.Context, pod *corev1.Pod) (bool, string, error) {
	for _, container := range pod.Spec.InitContainers {
		if container.Resources.Requests.Cpu().IsZero() {
			return false, "CPU resource request not Set", nil
		}
		if container.Resources.Requests.Memory().IsZero() {
			return false, "Memory resource request not Set", nil
		}
		if container.Resources.Limits.Cpu().IsZero() {
			return false, "CPU resource limit not Set", nil
		}
		if container.Resources.Limits.Memory().IsZero() {
			return false, "Memory resource limit not Set", nil
		}
	}

	for _, container := range pod.Spec.Containers {
		log.Info("validating container", "container", container.Resources, "memoryLimit", container.Resources.Limits.Memory().Value(), "isZero", container.Resources.Limits.Memory().IsZero())
		if container.Resources.Requests.Cpu().IsZero() {
			return false, "CPU resource request not Set", nil
		}
		if container.Resources.Requests.Memory().IsZero() {
			return false, "Memory resource request not Set", nil
		}
		if container.Resources.Limits.Cpu().IsZero() {
			return false, "CPU resource limit not Set", nil
		}
		if container.Resources.Limits.Memory().IsZero() {
			return false, "Memory resource limit not Set", nil
		}
	}

	return true, "", nil
}

// PodValidator implements inject.Client.
// A client will be automatically injected.
var _ inject.Client = &PodValidator{}

// InjectClient injects the client.
func (v *PodValidator) InjectClient(c client.Client) error {
	v.client = c
	return nil
}

// PodValidator implements inject.Decoder.
// A decoder will be automatically injected.
var _ inject.Decoder = &PodValidator{}

// InjectDecoder injects the decoder.
func (v *PodValidator) InjectDecoder(d types.Decoder) error {
	v.decoder = d
	return nil
}
