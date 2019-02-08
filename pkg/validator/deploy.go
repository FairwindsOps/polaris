package validator

import (
	"context"
	"net/http"

	conf "github.com/reactiveops/fairwinds/pkg/config"
	"github.com/reactiveops/fairwinds/pkg/kube"
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

	if results.Summary.Failures > 0 {
		// TODO: Decide what message we want to return here.
		allowed, reason = false, "failed validation checks, view details on dashbaord."
	}
	return admission.ValidationResponse(allowed, reason)
}

// ValidateDeploy validates a single deployment, returns a ResourceResult.
func ValidateDeploy(conf conf.Configuration, deploy *appsv1.Deployment) ResourceResult {
	pod := deploy.Spec.Template.Spec
	resResult := ValidatePod(conf, &pod)
	resResult.Name = deploy.Name
	resResult.Type = "Deployment"
	return resResult
}

// ValidateDeploys validates that each deployment conforms to the Fairwinds config,
// returns a list of ResourceResults organized by namespace.
func ValidateDeploys(config conf.Configuration, k8sAPI *kube.API) (NamespacedResults, error) {
	nsResults := NamespacedResults{}
	deploys, err := k8sAPI.GetDeploys()
	if err != nil {
		return nsResults, err
	}

	for _, deploy := range deploys.Items {
		resResult := ValidateDeploy(config, &deploy)
		nsResults = addResult(resResult, nsResults, deploy.Namespace)
	}

	return nsResults, nil
}

func addResult(resResult ResourceResult, nsResults NamespacedResults, nsName string) NamespacedResults {
	nsResult := &NamespacedResult{}

	// If there is already data stored for this namespace name,
	// then append to the ResourceResults to the existing data.
	switch nsResults[nsName] {
	case nil:
		nsResult = &NamespacedResult{
			Summary: &ResultSummary{},
			Results: []ResourceResult{},
		}
		nsResults[nsName] = nsResult
	default:
		nsResult = nsResults[nsName]
	}

	nsResult.Results = append(nsResult.Results, resResult)

	// Aggregate all resource results summary counts to get a namespace wide count.
	nsResult.Summary.Successes += resResult.Summary.Successes
	nsResult.Summary.Warnings += resResult.Summary.Warnings
	nsResult.Summary.Failures += resResult.Summary.Successes
	return nsResults
}
