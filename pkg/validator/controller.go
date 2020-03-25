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
	"strings"

	"github.com/sirupsen/logrus"

	conf "github.com/fairwindsops/polaris/pkg/config"
	"github.com/fairwindsops/polaris/pkg/kube"
	controller "github.com/fairwindsops/polaris/pkg/validator/controllers"
)

const exemptionAnnotationKey = "polaris.fairwinds.com/exempt"

// ValidateController validates a single controller, returns a ControllerResult.
func ValidateController(conf *conf.Configuration, controller controller.GenericController, kubeResources *kube.ResourceProvider) (ControllerResult, error) {
	podResult, err := ValidatePod(conf, controller)
	if err != nil {
		return ControllerResult{}, err
	}
	result := ControllerResult{
		Kind:      controller.GetKindString(),
		Name:      controller.GetName(),
		Namespace: controller.GetObjectMeta().Namespace,
		Results:   ResultSet{},
		PodResult: podResult,
	}

	return result, nil
}

// ValidateControllers validates that each deployment conforms to the Polaris config,
// builds a list of ResourceResults organized by namespace.
func ValidateControllers(config *conf.Configuration, kubeResources *kube.ResourceProvider) ([]ControllerResult, error) {
	var controllersToAudit []controller.GenericController
	loadedControllers := kubeResources.Controllers
	controllersToAudit = append(controllersToAudit, loadedControllers...)

	results := []ControllerResult{}
	for _, controller := range controllersToAudit {
		if !config.DisallowExemptions && hasExemptionAnnotation(controller) {
			continue
		}
		result, err := ValidateController(config, controller, kubeResources)
		if err != nil {
			logrus.Warn("An error occured validating controller:", err)
			return nil, err
		}
		results = append(results, result)
	}

	return results, nil
}

func hasExemptionAnnotation(ctrl controller.GenericController) bool {
	annot := ctrl.GetObjectMeta().Annotations
	val := annot[exemptionAnnotationKey]
	return strings.ToLower(val) == "true"
}
