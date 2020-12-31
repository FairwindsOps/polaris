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
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/fairwindsops/polaris/pkg/config"
	"github.com/fairwindsops/polaris/pkg/kube"
	validator "github.com/fairwindsops/polaris/pkg/validator"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
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

// InjectDecoder injects the decoder.
func (v *Validator) InjectDecoder(d *admission.Decoder) error {
	logrus.Info("Injecting decoder")
	v.decoder = d
	return nil
}

var _ admission.Handler = &Validator{}

// NewWebhook creates a validating admission webhook for the apiType.
func NewWebhook(mgr manager.Manager, validator Validator) {
	path := "/validate"

	mgr.GetWebhookServer().Register(path, &webhook.Admission{Handler: &validator})
}

// GetObjectFromRawRequest returns the pod object and the controller's object from the raw json bytes.
func GetObjectFromRawRequest(raw []byte) (corev1.Pod, interface{}, error) {
	pod := corev1.Pod{}
	var originalObject interface{}

	decoded := map[string]interface{}{}
	err := json.Unmarshal(raw, &decoded)
	if err != nil {
		return pod, originalObject, err
	}
	podMap := kube.GetPodSpec(decoded)
	if podMap == nil {
		return pod, originalObject, errors.New("Object does not contain pods")
	}
	encoded, err := json.Marshal(podMap)
	if err != nil {
		return pod, originalObject, err
	}
	err = json.Unmarshal(encoded, &pod.Spec)
	if err != nil {
		return pod, originalObject, err
	}
	originalObject = decoded
	return pod, originalObject, err
}

func (v *Validator) handleInternal(req admission.Request) (*validator.PodResult, error) {
	pod := corev1.Pod{}
	var originalObject interface{}
	var err error
	if req.AdmissionRequest.Kind.Kind == "Pod" {
		err := v.decoder.Decode(req, &pod)
		if err != nil {
			return nil, err
		}
		if len(pod.ObjectMeta.OwnerReferences) > 0 {
			logrus.Infof("Allowing owned pod %s/%s to pass through webhook", pod.ObjectMeta.Namespace, pod.ObjectMeta.Name)
			return nil, nil
		}
		originalObject = pod
	} else {
		pod, originalObject, err = GetObjectFromRawRequest(req.Object.Raw)
	}
	controller, err := kube.NewGenericWorkloadFromPod(pod, originalObject)
	if err != nil {
		return nil, err
	}
	controller.Kind = req.AdmissionRequest.Kind.Kind
	var controllerResult validator.Result
	controllerResult, err = validator.ValidateController(&v.Config, controller)
	if err != nil {
		return nil, err
	}
	return controllerResult.PodResult, nil
}

// Handle for Validator to run validation checks.
func (v *Validator) Handle(ctx context.Context, req admission.Request) admission.Response {
	logrus.Info("Starting request")
	podResult, err := v.handleInternal(req)
	if err != nil {
		logrus.Errorf("Error validating request: %v", err)
		return admission.Errored(http.StatusBadRequest, err)
	}
	allowed := true
	reason := ""
	if podResult != nil {
		numDangers := podResult.GetSummary().Dangers
		if numDangers > 0 {
			allowed = false
			reason = getFailureReason(*podResult)
		}
		logrus.Infof("%d validation errors found when validating %s", numDangers, podResult.Name)
	}
	return admission.ValidationResponse(allowed, reason)
}

func getFailureReason(podResult validator.PodResult) string {
	reason := "\nPolaris prevented this deployment due to configuration problems:\n"

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

	return reason
}
