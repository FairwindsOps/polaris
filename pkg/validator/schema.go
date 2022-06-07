// Copyright 2022 FairwindsOps, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package validator

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/qri-io/jsonschema"
	"github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	"gomodules.xyz/jsonpatch/v2"
	corev1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/fairwindsops/polaris/pkg/config"
	"github.com/fairwindsops/polaris/pkg/kube"
)

type schemaTestCase struct {
	Target           config.TargetKind
	Resource         kube.GenericResource
	IsInitContianer  bool
	Container        *corev1.Container
	ResourceProvider *kube.ResourceProvider
}

func resolveCheck(conf *config.Configuration, checkID string, test schemaTestCase) (*config.SchemaCheck, error) {
	if !conf.DisallowExemptions &&
		!conf.DisallowAnnotationExemptions &&
		hasExemptionAnnotation(test.Resource.ObjectMeta, checkID) {
		return nil, nil
	}
	check, ok := conf.CustomChecks[checkID]
	if !ok {
		check, ok = config.BuiltInChecks[checkID]
	}
	if !ok {
		return nil, fmt.Errorf("Check %s not found", checkID)
	}

	containerName := ""
	if test.Container != nil {
		containerName = test.Container.Name
	}
	if !conf.IsActionable(check.ID, test.Resource.ObjectMeta, containerName) {
		return nil, nil
	}
	if !check.IsActionable(test.Target, test.Resource.Kind, test.IsInitContianer) {
		return nil, nil
	}
	checkPtr, err := check.TemplateForResource(test.Resource.Resource.Object)
	if err != nil {
		return nil, err
	}
	return checkPtr, nil
}

func makeResult(conf *config.Configuration, check *config.SchemaCheck, passes bool, issues []jsonschema.ValError) ResultMessage {
	details := []string{}
	for _, issue := range issues {
		details = append(details, issue.Message)
	}
	result := ResultMessage{
		ID:       check.ID,
		Severity: conf.Checks[check.ID],
		Category: check.Category,
		Success:  passes,
		// FIXME: need to fix the tests before adding this back
		//Details: details,
	}
	if passes {
		result.Message = check.SuccessMessage
	} else {
		result.Message = check.FailureMessage
	}
	return result
}

const exemptionAnnotationKey = "polaris.fairwinds.com/exempt"
const exemptionAnnotationPattern = "polaris.fairwinds.com/%s-exempt"

func hasExemptionAnnotation(objMeta metaV1.Object, checkID string) bool {
	annot := objMeta.GetAnnotations()
	val := annot[exemptionAnnotationKey]
	if strings.ToLower(val) == "true" {
		return true
	}
	checkKey := fmt.Sprintf(exemptionAnnotationPattern, checkID)
	val = annot[checkKey]
	if strings.ToLower(val) == "true" {
		return true
	}
	return false
}

// ApplyAllSchemaChecksToResourceProvider applies all available checks to a ResourceProvider
func ApplyAllSchemaChecksToResourceProvider(conf *config.Configuration, resourceProvider *kube.ResourceProvider) ([]Result, error) {
	results := []Result{}
	for _, resources := range resourceProvider.Resources {
		kindResults, err := ApplyAllSchemaChecksToAllResources(conf, resourceProvider, resources)
		if err != nil {
			return results, err
		}
		results = append(results, kindResults...)
	}
	return results, nil
}

// ApplyAllSchemaChecksToAllResources applies available checks to a list of resources
func ApplyAllSchemaChecksToAllResources(conf *config.Configuration, resourceProvider *kube.ResourceProvider, resources []kube.GenericResource) ([]Result, error) {
	results := []Result{}
	for _, resource := range resources {
		result, err := ApplyAllSchemaChecks(conf, resourceProvider, resource)
		if err != nil {
			return results, err
		}
		results = append(results, result)
	}
	return results, nil
}

// ApplyAllSchemaChecks applies available checks to a single resource
func ApplyAllSchemaChecks(conf *config.Configuration, resourceProvider *kube.ResourceProvider, resource kube.GenericResource) (Result, error) {
	if resource.PodSpec == nil {
		return applyNonControllerSchemaChecks(conf, resourceProvider, resource)
	}
	return applyControllerSchemaChecks(conf, resourceProvider, resource)
}

func applyNonControllerSchemaChecks(conf *config.Configuration, resourceProvider *kube.ResourceProvider, resource kube.GenericResource) (Result, error) {
	finalResult := Result{
		Kind:      resource.Kind,
		Name:      resource.ObjectMeta.GetName(),
		Namespace: resource.ObjectMeta.GetNamespace(),
	}
	resultSet, err := applyTopLevelSchemaChecks(conf, resourceProvider, resource, false)
	finalResult.Results = resultSet
	return finalResult, err
}

func applyControllerSchemaChecks(conf *config.Configuration, resourceProvider *kube.ResourceProvider, resource kube.GenericResource) (Result, error) {
	finalResult := Result{
		Kind:      resource.Kind,
		Name:      resource.ObjectMeta.GetName(),
		Namespace: resource.ObjectMeta.GetNamespace(),
	}
	resultSet, err := applyTopLevelSchemaChecks(conf, resourceProvider, resource, true)
	if err != nil {
		return finalResult, err
	}
	finalResult.Results = resultSet

	nonControllerResults, err := applyTopLevelSchemaChecks(conf, resourceProvider, resource, false)
	if err != nil {
		return finalResult, err
	}
	for key, val := range nonControllerResults {
		if _, ok := finalResult.Results[key]; ok {
			return finalResult, errors.New("Duplicate finding for check " + key)
		}
		finalResult.Results[key] = val
	}

	podRS, err := applyPodSchemaChecks(conf, resourceProvider, resource)
	if err != nil {
		return finalResult, err
	}
	podRes := PodResult{
		Results:          podRS,
		ContainerResults: []ContainerResult{},
	}
	finalResult.PodResult = &podRes

	for _, container := range resource.PodSpec.InitContainers {
		results, err := applyContainerSchemaChecks(conf, resourceProvider, resource, &container, true)
		if err != nil {
			return finalResult, err
		}
		cRes := ContainerResult{
			Name:    container.Name,
			Results: results,
		}
		podRes.ContainerResults = append(podRes.ContainerResults, cRes)
	}
	for _, container := range resource.PodSpec.Containers {
		results, err := applyContainerSchemaChecks(conf, resourceProvider, resource, &container, false)
		if err != nil {
			return finalResult, err
		}
		cRes := ContainerResult{
			Name:    container.Name,
			Results: results,
		}
		podRes.ContainerResults = append(podRes.ContainerResults, cRes)
	}

	return finalResult, nil
}

func applyTopLevelSchemaChecks(conf *config.Configuration, resources *kube.ResourceProvider, res kube.GenericResource, isController bool) (ResultSet, error) {
	test := schemaTestCase{
		ResourceProvider: resources,
		Resource:         res,
	}
	if isController {
		test.Target = config.TargetController
	}
	return applySchemaChecks(conf, test)
}

func applyPodSchemaChecks(conf *config.Configuration, resources *kube.ResourceProvider, controller kube.GenericResource) (ResultSet, error) {
	test := schemaTestCase{
		Target:           config.TargetPodSpec,
		ResourceProvider: resources,
		Resource:         controller,
	}
	return applySchemaChecks(conf, test)
}

func applyContainerSchemaChecks(conf *config.Configuration, resources *kube.ResourceProvider, controller kube.GenericResource, container *corev1.Container, isInit bool) (ResultSet, error) {
	test := schemaTestCase{
		Target:           config.TargetContainer,
		ResourceProvider: resources,
		Resource:         controller,
		Container:        container,
		IsInitContianer:  isInit,
	}
	return applySchemaChecks(conf, test)
}

func applySchemaChecks(conf *config.Configuration, test schemaTestCase) (ResultSet, error) {
	results := ResultSet{}
	checkIDs := getSortedKeys(conf.Checks)
	for _, checkID := range checkIDs {
		result, err := applySchemaCheck(conf, checkID, test)
		if err != nil {
			return results, err
		}
		if result != nil {
			results[checkID] = *result
		}
	}
	return results, nil
}

func applySchemaCheck(conf *config.Configuration, checkID string, test schemaTestCase) (*ResultMessage, error) {
	check, err := resolveCheck(conf, checkID, test)
	if err != nil {
		return nil, err
	} else if check == nil {
		return nil, nil
	}
	var passes bool
	var issues []jsonschema.ValError
	var prefix string
	if check.SchemaTarget != "" {
		if check.SchemaTarget == config.TargetPodSpec && check.Target == config.TargetContainer {
			podCopy := *test.Resource.PodSpec
			podCopy.InitContainers = []corev1.Container{}
			podCopy.Containers = []corev1.Container{*test.Container}
			containerIndex := funk.IndexOf(test.Resource.PodSpec.Containers, func(value corev1.Container) bool {
				return value.Name == test.Container.Name
			})
			prefix = getJSONSchemaPrefix(test.Resource.Kind)
			if prefix != "" {
				prefix += "/containers/" + strconv.Itoa(containerIndex)
			}
			passes, issues, err = check.CheckPodSpec(&podCopy)
		} else {
			return nil, fmt.Errorf("Unknown combination of target (%s) and schema target (%s)", check.Target, check.SchemaTarget)
		}
	} else if check.Target == config.TargetPodSpec {
		passes, issues, err = check.CheckPodSpec(test.Resource.PodSpec)
		prefix = getJSONSchemaPrefix(test.Resource.Kind)
	} else if check.Target == config.TargetContainer {
		containerIndex := funk.IndexOf(test.Resource.PodSpec.Containers, func(value corev1.Container) bool {
			return value.Name == test.Container.Name
		})
		prefix = getJSONSchemaPrefix(test.Resource.Kind)
		if prefix != "" {
			prefix += "/containers/" + strconv.Itoa(containerIndex)
		}
		passes, issues, err = check.CheckContainer(test.Container)
	} else {
		passes, issues, err = check.CheckObject(test.Resource.Resource.Object)
	}
	if err != nil {
		return nil, err
	}
	for groupkind := range check.AdditionalValidators {
		if !passes {
			break
		}
		resources := test.ResourceProvider.Resources[groupkind]
		namespace := test.Resource.ObjectMeta.GetNamespace()
		if test.Resource.Kind == "Namespace" {
			namespace = test.Resource.ObjectMeta.GetName()
		}
		resources = funk.Filter(resources, func(res kube.GenericResource) bool {
			return res.ObjectMeta.GetNamespace() == namespace
		}).([]kube.GenericResource)
		objects := funk.Map(resources, func(res kube.GenericResource) interface{} {
			return res.Resource.Object
		}).([]interface{})
		passes, err = check.CheckAdditionalObjects(groupkind, objects)
		if err != nil {
			return nil, err
		}
	}
	result := makeResult(conf, check, passes, issues)
	if !passes {
		if funk.Contains(conf.Mutations, checkID) {
			mutations := funk.Map(check.Mutations, func(mutation jsonpatch.Operation) jsonpatch.Operation {
				mutationCopy := deepCopyMutation(mutation)
				mutationCopy.Path = prefix + mutationCopy.Path
				return mutationCopy
			}).([]jsonpatch.Operation)
			result.Mutations = mutations
			result.Comments = check.Comments
		}
	}
	return &result, nil
}

func getSortedKeys(m map[string]config.Severity) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func deepCopyMutation(source jsonpatch.Operation) jsonpatch.Operation {
	destination := jsonpatch.Operation{
		Operation: source.Operation,
		Path:      source.Path,
		Value:     source.Value,
	}
	return destination
}

func getJSONSchemaPrefix(kind string) (prefix string) {
	if kind == "CronJob" {
		prefix = "/spec/jobTemplate/spec/template/spec"
	} else if kind == "Pod" {
		prefix = "/spec"
	} else if (kind == "Deployment") || (kind == "DaemonSet") ||
		(kind == "StatefulSet") || (kind == "Job") || (kind == "ReplicationController") {
		prefix = "/spec/template/spec"
	} else {
		logrus.Warningf("Mutation for this this resource (%s) is not supported", kind)
	}
	return prefix
}
