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

package webhook

import (
	"context"
	"fmt"
	"net/http"

	"github.com/fairwindsops/polaris/pkg/config"
	"github.com/fairwindsops/polaris/pkg/kube"
	validator "github.com/fairwindsops/polaris/pkg/validator"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// Validator validates k8s resources.
type Validator struct {
	Client  client.Client
	decoder *admission.Decoder
	Config  config.Configuration
}

// NewValidateWebhook creates a validating admission webhook for the apiType.
func NewValidateWebhook(mgr manager.Manager, c config.Configuration) {
	path := "/validate"
	validator := Validator{
		Client:  mgr.GetClient(),
		decoder: admission.NewDecoder(runtime.NewScheme()),
		Config:  c,
	}
	mgr.GetWebhookServer().Register(path, &webhook.Admission{Handler: &validator})
}

func (v *Validator) handleInternal(req admission.Request) (*validator.Result, kube.GenericResource, error) {
	return GetValidatedResults(req.AdmissionRequest.Kind.Kind, v.decoder, req, v.Config)
}

// GetValidatedResults returns the validated results.
func GetValidatedResults(kind string, decoder *admission.Decoder, req admission.Request, config config.Configuration) (*validator.Result, kube.GenericResource, error) {
	var resource kube.GenericResource
	var err error
	if kind == "Pod" {
		if decoder == nil {
			panic("Decoder is nil!")
		}
		pod := corev1.Pod{}
		err := decoder.Decode(req, &pod)
		if err != nil {
			logrus.Errorf("Failed to decode pod: %v", err)
			return nil, resource, err
		}
		if len(pod.ObjectMeta.OwnerReferences) > 0 {
			logrus.Infof("Allowing owned pod %s/%s to pass through webhook", pod.ObjectMeta.Namespace, pod.ObjectMeta.Name)
			return nil, resource, nil
		}
		resource, err = kube.NewGenericResourceFromPod(pod, pod)
	} else {
		resource, err = kube.NewGenericResourceFromBytes(req.Object.Raw)
	}
	if err != nil {
		logrus.Errorf("Failed to create resource: %v", err)
		return nil, resource, err
	}
	resourceResult, err := validator.ApplyAllSchemaChecks(&config, nil, resource)
	if err != nil {
		return nil, resource, err
	}
	return &resourceResult, resource, nil
}

// Handle for Validator to run validation checks.
func (v *Validator) Handle(ctx context.Context, req admission.Request) admission.Response {
	logrus.Info("Starting admission request")
	result, _, err := v.handleInternal(req)
	if err != nil {
		logrus.Errorf("Error validating request: %v", err)
		return admission.Errored(http.StatusBadRequest, err)
	}
	allowed := true
	reason := ""
	if result != nil {
		numDangers := result.GetSummary().Dangers
		if numDangers > 0 {
			allowed = false
			reason = getFailureReason(*result)
		}
		logrus.Infof("%d validation errors found when validating %s", numDangers, result.Name)
	}
	return admission.ValidationResponse(allowed, reason)
}

func getFailureReason(result validator.Result) string {
	reason := "\nPolaris prevented this deployment due to configuration problems:\n"

	for _, message := range result.Results {
		if !message.Success && message.Severity == config.SeverityDanger {
			reason += fmt.Sprintf("- %s: %s\n", result.Kind, message.Message)
		}
	}

	podResult := result.PodResult
	if podResult != nil {
		for _, message := range podResult.Results {
			if !message.Success && message.Severity == config.SeverityDanger {
				reason += fmt.Sprintf("- Pod: %s\n", message.Message)
			}
		}

		for _, containerResult := range podResult.ContainerResults {
			for _, message := range containerResult.Results {
				if !message.Success && message.Severity == config.SeverityDanger {
					reason += fmt.Sprintf("- Container %s: %s\n", containerResult.Name, message.Message)
				}
			}
		}
	}

	return reason
}
