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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// ControllerSpec is a generic type for k8s controller specs
type ControllerSpec struct {
	Template corev1.PodTemplateSpec
}

// Controller is a generic type for k8s controllers (e.g. Deployments and StatefulSets)
type Controller struct {
	Type      string
	Name      string
	Namespace string
	Spec      ControllerSpec
}

// ValidateController validates a single controller, returns a ControllerResult.
func ValidateController(conf conf.Configuration, controller Controller) ControllerResult {
	pod := controller.Spec.Template.Spec
	podResult := ValidatePod(conf, &pod)
	return ControllerResult{
		Type:      controller.Type,
		Name:      controller.Name,
		PodResult: podResult,
	}
}

// ValidateControllers validates that each deployment conforms to the Polaris config,
// returns a list of ResourceResults organized by namespace.
func ValidateControllers(config conf.Configuration, kubeResources *kube.ResourceProvider, nsResults *NamespacedResults) {
	controllers := []Controller{}
	for _, deploy := range kubeResources.Deployments {
		controllers = append(controllers, ControllerFromDeployment(deploy))
	}
	for _, deploy := range kubeResources.StatefulSets {
		controllers = append(controllers, ControllerFromStatefulSet(deploy))
	}
	for _, controller := range controllers {
		controllerResult := ValidateController(config, controller)
		nsResult := nsResults.getNamespaceResult(controller.Namespace)
		nsResult.Summary.appendResults(*controllerResult.PodResult.Summary)
		if controller.Type == "Deployment" {
			nsResult.DeploymentResults = append(nsResult.DeploymentResults, controllerResult)
		} else if controller.Type == "StatefulSet" {
			nsResult.StatefulSetResults = append(nsResult.StatefulSetResults, controllerResult)
		}
	}
}

// ControllerFrom* functions are 100% boilerplate

// ControllerFromDeployment creates a controller
func ControllerFromDeployment(c appsv1.Deployment) Controller {
	spec := ControllerSpec{
		Template: c.Spec.Template,
	}
	return Controller{
		Type:      "Deployment",
		Name:      c.Name,
		Namespace: c.Namespace,
		Spec:      spec,
	}
}

// ControllerFromStatefulSet creates a controller
func ControllerFromStatefulSet(c appsv1.StatefulSet) Controller {
	spec := ControllerSpec{
		Template: c.Spec.Template,
	}
	return Controller{
		Type:      "StatefulSet",
		Name:      c.Name,
		Namespace: c.Namespace,
		Spec:      spec,
	}
}
