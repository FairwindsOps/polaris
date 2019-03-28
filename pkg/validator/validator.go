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
	"context"
	"net/http"

	conf "github.com/reactiveops/fairwinds/pkg/config"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission/types"
)

// Validator validates k8s resources.
type Validator struct {
	client  client.Client
	decoder types.Decoder
	Config  conf.Configuration
}

var _ inject.Client = &Validator{}

// InjectClient injects the client.
func (v *Validator) InjectClient(c client.Client) error {
	v.client = c
	return nil
}

var _ inject.Decoder = &Validator{}

// InjectDecoder injects the decoder.
func (v *Validator) InjectDecoder(d types.Decoder) error {
	v.decoder = d
	return nil
}

var _ admission.Handler = &Validator{}

// Handle for Validator to run validation checks.
func (v *Validator) Handle(ctx context.Context, req types.Request) types.Response {
	var err error
	var allowed bool
	var reason string
	var results ResourceResult

	switch req.AdmissionRequest.Kind.Kind {
	case "Deployment":
		deploy := appsv1.Deployment{}
		err = v.decoder.Decode(req, &deploy)
		results = ValidateDeploy(v.Config, &deploy)
	case "Pod":
		pod := corev1.Pod{}
		err = v.decoder.Decode(req, &pod)
		results = ValidatePod(v.Config, &pod.Spec)
	}
	if err != nil {
		return admission.ErrorResponse(http.StatusBadRequest, err)
	}

	if results.Summary.Errors > 0 {
		// TODO: Decide what message we want to return here.
		allowed, reason = false, "failed validation checks, view details on dashbaord."
	}
	return admission.ValidationResponse(allowed, reason)
}
