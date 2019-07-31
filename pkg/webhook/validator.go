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
	"os"

	conf "github.com/fairwindsops/polaris/pkg/config"
	validator "github.com/fairwindsops/polaris/pkg/validator"
	"github.com/fairwindsops/polaris/pkg/validator/controllers"
	"github.com/sirupsen/logrus"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission/builder"
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

// NewWebhook creates a validating admission webhook for the apiType.
func NewWebhook(name string, mgr manager.Manager, validator Validator, apiType runtime.Object) *admission.Webhook {
	name = fmt.Sprintf("%s.k8s.io", name)
	path := fmt.Sprintf("/validating-%s", name)

	webhook, err := builder.NewWebhookBuilder().
		Name(name).
		Validating().
		Path(path).
		Operations(admissionregistrationv1beta1.Create, admissionregistrationv1beta1.Update).
		WithManager(mgr).
		ForType(apiType).
		Handlers(&validator).
		Build()
	if err != nil {
		logrus.Errorf("Error building webhook: %v", err)
		os.Exit(1)
	}

	return webhook
}

// Handle for Validator to run validation checks.
func (v *Validator) Handle(ctx context.Context, req types.Request) types.Response {
	var err error
	var podResult validator.PodResult

	if req.AdmissionRequest.Kind.Kind == "Pod" {
		pod := corev1.Pod{}
		err = v.decoder.Decode(req, &pod)
		podResult = validator.ValidatePod(v.Config, &pod.Spec)
	} else {
		var controller controllers.Interface
		if yes := v.Config.CheckIfKindIsConfiguredForValidation(req.AdmissionRequest.Kind.Kind); !yes {
			logrus.Warnf("Skipping, kind (%s) isn't something we are configured to scan", req.AdmissionRequest.Kind.Kind)
			return admission.ValidationResponse(true, fmt.Sprintf("Skipping: (%s) isn't something we're configured to scan.", req.AdmissionRequest.Kind.Kind))
		}

		// We should never hit this case unless something is misconfiured in CheckIfKindIsConfiguredForValidation
		controllerType, err := conf.GetSupportedControllerFromString(req.AdmissionRequest.Kind.Kind)
		if err != nil {
			msg := fmt.Errorf("Unexpected error occurred. Expected Kind to be a supported type (%s)", req.AdmissionRequest.Kind.Kind)
			logrus.Error(msg)
			return admission.ErrorResponse(http.StatusInternalServerError, err)
		}

		// For each type, perform the scan
		// TODO: This isn't really that elegant due to the decoder and NewXXXController setup :( could use love
		switch controllerType {
		case conf.Deployments:
			deploy := appsv1.Deployment{}
			err = v.decoder.Decode(req, &deploy)
			controller = controllers.NewDeploymentController(deploy)
		case conf.StatefulSets:
			statefulSet := appsv1.StatefulSet{}
			err = v.decoder.Decode(req, &statefulSet)
			controller = controllers.NewStatefulSetController(statefulSet)
		case conf.DaemonSets:
			daemonSet := appsv1.DaemonSet{}
			err = v.decoder.Decode(req, &daemonSet)
			controller = controllers.NewDaemonSetController(daemonSet)
		case conf.Jobs:
			job := batchv1.Job{}
			err = v.decoder.Decode(req, &job)
			controller = controllers.NewJobController(job)
		case conf.CronJobs:
			cronJob := batchv1beta1.CronJob{}
			err = v.decoder.Decode(req, &cronJob)
			controller = controllers.NewCronJobController(cronJob)
		case conf.ReplicationControllers:
			replicationController := corev1.ReplicationController{}
			err = v.decoder.Decode(req, &replicationController)
			controller = controllers.NewReplicationControllerController(replicationController)
		}
		controllerResult := validator.ValidateController(v.Config, controller)
		podResult = controllerResult.PodResult
	}

	if err != nil {
		logrus.Errorf("Error validating request: %v", err)
		return admission.ErrorResponse(http.StatusBadRequest, err)
	}

	allowed := true
	reason := ""
	if podResult.Summary.Totals.Errors > 0 {
		allowed = false
		reason = getFailureReason(podResult)
	}
	logrus.Infof("%d validation errors found when validating %s", podResult.Summary.Totals.Errors, podResult.Name)
	return admission.ValidationResponse(allowed, reason)
}

func getFailureReason(podResult validator.PodResult) string {
	reason := "\nPolaris prevented this deployment due to configuration problems:\n"

	for _, message := range podResult.Messages {
		if message.Type == validator.MessageTypeError {
			reason += fmt.Sprintf("- Pod: %s\n", message.Message)
		}
	}

	for _, containerResult := range podResult.ContainerResults {
		for _, message := range containerResult.Messages {
			if message.Type == validator.MessageTypeError {
				reason += fmt.Sprintf("- Container %s: %s\n", containerResult.Name, message.Message)
			}
		}
	}

	return reason
}
