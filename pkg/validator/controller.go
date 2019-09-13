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
	conf "github.com/fairwindsops/polaris/pkg/config"
	"github.com/fairwindsops/polaris/pkg/kube"
	"github.com/fairwindsops/polaris/pkg/validator/controllers"
	controller "github.com/fairwindsops/polaris/pkg/validator/controllers"
	"github.com/sirupsen/logrus"
)

// ValidateController validates a single controller, returns a ControllerResult.
func ValidateController(conf conf.Configuration, controller controller.Interface) ControllerResult {
	pod := controller.GetPodSpec()
	podResult := ValidatePod(conf, controller.GetName(), pod)
	return ControllerResult{
		Type:      controller.GetType().String(),
		Name:      controller.GetName(),
		PodResult: podResult,
	}
}

// ValidateControllers validates that each deployment conforms to the Polaris config,
// builds a list of ResourceResults organized by namespace.
func ValidateControllers(config conf.Configuration, kubeResources *kube.ResourceProvider, nsResults *NamespacedResults) {
	var controllersToAudit []controller.Interface
	for _, supportedControllers := range config.ControllersToScan {
		loadedControllers, _ := controllers.LoadControllersByType(supportedControllers, kubeResources)
		controllersToAudit = append(controllersToAudit, loadedControllers...)
	}

	for _, controller := range controllersToAudit {
		controllerResult := ValidateController(config, controller)
		nsResult := nsResults.getNamespaceResult(controller.GetNamespace())
		nsResult.Summary.appendResults(*controllerResult.PodResult.Summary)
		if err := nsResult.AddResult(controller.GetType(), controllerResult); err != nil {
			logrus.Errorf("Internal Error: Failed to add a grouped result: %s", err)
		}
	}
}
