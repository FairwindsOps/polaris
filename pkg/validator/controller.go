// Copyright 2019 ReactiveOps
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
	conf "github.com/reactiveops/polaris/pkg/config"
	"github.com/reactiveops/polaris/pkg/kube"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
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
		applyLimitRanges(&controller, kubeResources.LimitRanges)
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

// applyLimitRanges modifies the Limits and Requests of the input controller
// based on the presense an applicable LimitRange
func applyLimitRanges(controller *Controller, limits []corev1.LimitRange) {
	for _, limit := range limits {
		if limit.ObjectMeta.Namespace != controller.Namespace {
			continue
		}
		for _, limitItem := range limit.Spec.Limits {
			if limitItem.Type != corev1.LimitTypeContainer {
				continue
			}
			for containerIdx, container := range controller.Spec.Template.Spec.Containers {
				if container.Resources.Limits == nil {
					controller.Spec.Template.Spec.Containers[containerIdx].Resources.Limits = make(map[corev1.ResourceName]resource.Quantity)
				}
				if container.Resources.Requests == nil {
					controller.Spec.Template.Spec.Containers[containerIdx].Resources.Requests = make(map[corev1.ResourceName]resource.Quantity)
				}
				for resType, resLimit := range limitItem.Default {
					if _, ok := container.Resources.Limits[resType]; !ok {
						controller.Spec.Template.Spec.Containers[containerIdx].Resources.Limits[resType] = resLimit
					}
				}
				for resType, resLimit := range limitItem.DefaultRequest {
					if _, ok := container.Resources.Requests[resType]; !ok {
						controller.Spec.Template.Spec.Containers[containerIdx].Resources.Requests[resType] = resLimit
					}
				}
			}
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
