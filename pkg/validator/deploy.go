package validator

import (
	"context"
	"net/http"

	conf "github.com/reactiveops/fairwinds/pkg/config"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission/types"
)

// DeployValidator validates Pods
type DeployValidator struct {
	client  client.Client
	decoder types.Decoder
	Config  conf.Configuration
}

// PodValidator implements inject.Client.
// A client will be automatically injected.
var _ inject.Client = &DeployValidator{}

// InjectClient injects the client.
func (v *DeployValidator) InjectClient(c client.Client) error {
	v.client = c
	return nil
}

// DeployValidator implements inject.Decoder.
// A decoder will be automatically injected.
var _ inject.Decoder = &DeployValidator{}

// InjectDecoder injects the decoder.
func (v *DeployValidator) InjectDecoder(d types.Decoder) error {
	v.decoder = d
	return nil
}

// Implement admission.Handler so the controller can handle admission request.
var _ admission.Handler = &DeployValidator{}

// Handle for DeployValidator admits a pod if validation passes.
func (v *DeployValidator) Handle(ctx context.Context, req types.Request) types.Response {
	deploy := appsv1.Deployment{}

	err := v.decoder.Decode(req, &deploy)
	if err != nil {
		return admission.ErrorResponse(http.StatusBadRequest, err)
	}

	results := ValidateDeploys(v.Config, &deploy, Results{})
	allowed, reason := results.Format()

	return admission.ValidationResponse(allowed, reason)
}

// ValidateDeploys does validates that each deployment conforms to the Fairwinds config.
func ValidateDeploys(conf conf.Configuration, deploy *appsv1.Deployment, results Results) Results {
	for _, container := range deploy.Spec.Template.Spec.InitContainers {
		results.InitContainerValidations = append(
			results.InitContainerValidations,
			validateContainer(conf, container),
		)
	}

	for _, container := range deploy.Spec.Template.Spec.Containers {
		results.ContainerValidations = append(
			results.ContainerValidations,
			validateContainer(conf, container),
		)
	}

	return results
}
