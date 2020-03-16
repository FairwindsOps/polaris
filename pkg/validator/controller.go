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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	conf "github.com/fairwindsops/polaris/pkg/config"
	"github.com/fairwindsops/polaris/pkg/kube"
	"github.com/fairwindsops/polaris/pkg/validator/controllers"
	controller "github.com/fairwindsops/polaris/pkg/validator/controllers"
)

const exemptionAnnotationKey = "polaris.fairwinds.com/exempt"

// ValidateController validates a single controller, returns a ControllerResult.
func ValidateController(conf *conf.Configuration, controller controller.Interface, kubeResources *kube.ResourceProvider) (ControllerResult, error) {
	podResult, err := ValidatePod(conf, controller)
	if err != nil {
		return ControllerResult{}, err
	}
	result := ControllerResult{
		Kind:        controller.GetKind().String(),
		Name:        controller.GetName(),
		Namespace:   controller.GetObjectMeta().Namespace,
		Results:     ResultSet{},
		PodResult:   podResult,
		CreatedTime: controller.GetObjectMeta().CreationTimestamp.Time,
	}
	owners := controller.GetObjectMeta().OwnerReferences
	// If an owner exists then set the name to the controller.
	// This allows us to handle CRDs creating Controllers or DeploymentConfigs in OpenShift.
	for len(owners) > 0 {
		firstOwner := owners[0]
		result.Kind = firstOwner.Kind
		result.Name = firstOwner.Name
		if kubeResources.DynamicClient != nil {

			dynamicClient := *kubeResources.DynamicClient
			restMapper := *kubeResources.RestMapper
			fqKind := schema.FromAPIVersionAndKind(firstOwner.APIVersion, firstOwner.Kind)
			mapping, err := restMapper.RESTMapping(fqKind.GroupKind(), fqKind.Version)
			if err != nil {
				logrus.Warnf("Error retrieving mapping %s of API %s and Kind %s because of error: %v ", firstOwner.Name, firstOwner.APIVersion, firstOwner.Kind, err)
				return result, nil
			}
			getParents, err := dynamicClient.Resource(mapping.Resource).Namespace(controller.GetObjectMeta().Namespace).Get(firstOwner.Name, metav1.GetOptions{})
			if err != nil {
				logrus.Warnf("Error retrieving parent object %s of API %s and Kind %s because of error: %v ", firstOwner.Name, firstOwner.APIVersion, firstOwner.Kind, err)
				return result, nil
			}
			owners = getParents.GetOwnerReferences()
		} else {
			break
		}
	}

	return result, nil
}

// Because the controllers with an Owner take on the name of the Owner, this eliminates any duplicates.
// In cases like CronJobs older children can hang around, so this takes the most recent.
func deduplicateControllers(controllerResults []ControllerResult) []ControllerResult {
	controllerMap := make(map[string][]ControllerResult)
	for _, controller := range controllerResults {
		key := controller.Namespace + "/" + controller.Kind + "/" + controller.Name
		controllerMap[key] = append(controllerMap[key], controller)
	}
	results := make([]ControllerResult, 0)
	for _, controllers := range controllerMap {
		if len(controllers) == 1 {
			results = append(results, controllers[0])
		} else {
			latestController := controllers[0]
			for _, controller := range controllers[1:] {
				if controller.CreatedTime.After(latestController.CreatedTime) {
					latestController = controller
				}
			}
			results = append(results, latestController)
		}
	}
	return results
}

// ValidateControllers validates that each deployment conforms to the Polaris config,
// builds a list of ResourceResults organized by namespace.
func ValidateControllers(config *conf.Configuration, kubeResources *kube.ResourceProvider) ([]ControllerResult, error) {
	var controllersToAudit []controller.Interface
	loadedControllers, err := controllers.LoadControllersByKind(conf.NakedPods, kubeResources)
	if err != nil {
		logrus.Warn(err)
	}
	controllersToAudit = append(controllersToAudit, loadedControllers...)
	results := []ControllerResult{}
	for _, controller := range controllersToAudit {
		if !config.DisallowExemptions && hasExemptionAnnotation(controller) {
			continue
		}
		result, err := ValidateController(config, controller, kubeResources)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return deduplicateControllers(results), nil
}

func hasExemptionAnnotation(ctrl controller.Interface) bool {
	annot := ctrl.GetObjectMeta().Annotations
	val := annot[exemptionAnnotationKey]
	return strings.ToLower(val) == "true"
}
