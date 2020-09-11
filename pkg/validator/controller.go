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

package validator

import (
	"context"
	"strings"

	"github.com/sirupsen/logrus"

	conf "github.com/fairwindsops/polaris/pkg/config"
	"github.com/fairwindsops/polaris/pkg/kube"
)

const exemptionAnnotationKey = "polaris.fairwinds.com/exempt"

// ValidateController validates a single controller, returns a ControllerResult.
func ValidateController(ctx context.Context, conf *conf.Configuration, controller kube.GenericWorkload) (ControllerResult, error) {
	podResult, err := ValidatePod(ctx, conf, controller)
	if err != nil {
		return ControllerResult{}, err
	}

	controllerResult, err := applyControllerSchemaChecks(ctx, conf, controller)
	if err != nil {
		return ControllerResult{}, err
	}

	result := ControllerResult{
		Kind:      controller.Kind,
		Name:      controller.ObjectMeta.GetName(),
		Namespace: controller.ObjectMeta.GetNamespace(),
		Results:   controllerResult,
		PodResult: podResult,
	}

	return result, nil
}

// ValidateControllers validates that each deployment conforms to the Polaris config,
// builds a list of ResourceResults organized by namespace.
func ValidateControllers(ctx context.Context, config *conf.Configuration, kubeResources *kube.ResourceProvider) ([]ControllerResult, error) {
	controllersToAudit := kubeResources.Controllers

	results := []ControllerResult{}
	for _, controller := range controllersToAudit {
		if !config.DisallowExemptions && hasExemptionAnnotation(controller) {
			continue
		}
		result, err := ValidateController(ctx, config, controller)
		if err != nil {
			logrus.Warn("An error occured validating controller:", err)
			return nil, err
		}
		results = append(results, result)
	}

	return results, nil
}

func hasExemptionAnnotation(ctrl kube.GenericWorkload) bool {
	annot := ctrl.ObjectMeta.GetAnnotations()
	val := annot[exemptionAnnotationKey]
	return strings.ToLower(val) == "true"
}
