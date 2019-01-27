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
	deploy := appsv1.Deployment{}
	err := v.decoder.Decode(req, &deploy)
	if err == nil {
		results := ValidateDeploys(v.Config, &deploy, Results{})
		allowed, reason := results.Format()
		return admission.ValidationResponse(allowed, reason)
	}

	pod := corev1.Pod{}
	err = v.decoder.Decode(req, &pod)
	if err == nil {
		results := ValidatePods(v.Config, &pod.Spec, Results{})
		allowed, reason := results.Format()
		return admission.ValidationResponse(allowed, reason)
	}
	return admission.ErrorResponse(http.StatusBadRequest, err)
}

// ValidateDeploys validates that each deployment conforms to the Fairwinds config.
func ValidateDeploys(conf conf.Configuration, deploy *appsv1.Deployment, results Results) Results {
	pod := deploy.Spec.Template.Spec
	return ValidatePods(conf, &pod, results)
}
