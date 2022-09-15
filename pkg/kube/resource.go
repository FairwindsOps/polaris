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

package kube

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	kubeAPICoreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	kubeAPIMetaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// GenericResource is a base implementation with some free methods for inherited structs
type GenericResource struct {
	Kind               string
	ObjectMeta         kubeAPIMetaV1.Object
	Resource           unstructured.Unstructured
	PodSpec            *kubeAPICoreV1.PodSpec
	PodTemplate        interface{}
	OriginalObjectJSON []byte
	OriginalObjectYAML []byte
}

// NewGenericResourceFromUnstructured creates a workload from an unstructured.Unstructured
func NewGenericResourceFromUnstructured(unst unstructured.Unstructured, podSpecMap interface{}) (GenericResource, error) {
	if unst.GetCreationTimestamp().Time.IsZero() {
		unstructured.RemoveNestedField(unst.Object, "metadata", "creationTimestamp")
		unstructured.RemoveNestedField(unst.Object, "status")
	}
	workload := GenericResource{
		Kind:     unst.GetKind(),
		Resource: unst,
	}
	objMeta, err := meta.Accessor(&unst)
	if err != nil {
		return workload, err
	}
	workload.ObjectMeta = objMeta
	workload.PodTemplate, err = GetPodTemplate(unst.UnstructuredContent())
	if err != nil {
		return workload, err
	}

	b, err := json.Marshal(&unst)
	if err != nil {
		return workload, err
	}
	workload.OriginalObjectJSON = b
	m := make(map[string]interface{})
	err = json.Unmarshal(b, &m)
	if err != nil {
		return workload, err
	}
	if podSpecMap == nil {
		podSpecMap = GetPodSpec(m)
	}
	if podSpecMap != nil {
		b, err = json.Marshal(podSpecMap)
		if err != nil {
			return workload, err
		}
		podSpec := kubeAPICoreV1.PodSpec{}
		err = json.Unmarshal(b, &podSpec)
		if err != nil {
			return workload, err
		}
		workload.PodSpec = &podSpec
	}
	return workload, nil
}

// NewGenericResourceFromPod builds a new workload for a given Pod without looking at parents
func NewGenericResourceFromPod(podResource kubeAPICoreV1.Pod, originalObject interface{}) (GenericResource, error) {
	podMap, err := SerializePod(&podResource)
	if err != nil {
		return GenericResource{}, err
	}
	workload := GenericResource{
		Kind:        "Pod",
		PodSpec:     &podResource.Spec,
		PodTemplate: podMap,
		ObjectMeta:  podResource.ObjectMeta.GetObjectMeta(),
	}
	if originalObject != nil {
		bytes, err := json.Marshal(originalObject)
		if err != nil {
			return workload, err
		}
		workload.OriginalObjectJSON = bytes

		err = json.Unmarshal(bytes, &workload.Resource.Object)
		if err != nil {
			logrus.Error("Couldn't marshal JSON for pod ", err)
			return workload, err
		}
		objMeta, err := meta.Accessor(&workload.Resource)
		if err != nil {
			logrus.Error("Couldn't create meta accessor for unstructred ", err)
			return workload, err
		}
		workload.ObjectMeta = objMeta
	}
	return workload, nil
}

// NewGenericResourceFromBytes parses a generic kubernetes resource
func NewGenericResourceFromBytes(contentBytes []byte) (GenericResource, error) {
	unst := unstructured.Unstructured{}
	err := yaml.Unmarshal(contentBytes, &unst.Object)
	if err != nil {
		return GenericResource{}, err
	}
	res, err := NewGenericResourceFromUnstructured(unst, nil)
	res.OriginalObjectYAML = contentBytes
	return res, err
}

// ResolveControllerFromPod builds a new workload for a given Pod
func ResolveControllerFromPod(ctx context.Context, podResource kubeAPICoreV1.Pod, dynamicClient *dynamic.Interface, restMapper *meta.RESTMapper, objectCache map[string]unstructured.Unstructured) (GenericResource, error) {
	workload, err := resolveControllerFromPod(ctx, podResource, dynamicClient, restMapper, objectCache)
	if err != nil {
		return workload, err
	}
	if len(workload.OriginalObjectJSON) == 0 {
		return NewGenericResourceFromPod(podResource, podResource)
	}
	return workload, err
}

func resolveControllerFromPod(ctx context.Context, podResource kubeAPICoreV1.Pod, dynamicClient *dynamic.Interface, restMapper *meta.RESTMapper, objectCache map[string]unstructured.Unstructured) (GenericResource, error) {
	podWorkload, err := NewGenericResourceFromPod(podResource, nil)
	if err != nil {
		return podWorkload, err
	}
	topKind := "Pod"
	topMeta := podWorkload.ObjectMeta
	var topPodSpec interface{}
	topPodSpec = podWorkload.Resource.Object
	owners := podResource.ObjectMeta.GetOwnerReferences()
	lastKey := ""
	for len(owners) > 0 {
		if len(owners) > 1 {
			logrus.Warn("More than 1 owner found")
		}
		firstOwner := owners[0]
		if firstOwner.Kind == "Node" {
			break
		}
		topKind = firstOwner.Kind
		key := fmt.Sprintf("%s/%s/%s", firstOwner.Kind, topMeta.GetNamespace(), firstOwner.Name)
		lastKey = key
		abstractObject, ok := objectCache[key]
		if !ok {
			err := cacheAllObjectsOfKind(ctx, firstOwner.APIVersion, firstOwner.Kind, dynamicClient, restMapper, objectCache)
			if err != nil {
				logrus.Warnf("Error caching objects of Kind %s %v", firstOwner.Kind, err)
				break
			}
			abstractObject, ok = objectCache[key]
			if !ok {
				logrus.Errorf("Cache missed %s again", key)
				break
			}
		}

		objMeta, err := meta.Accessor(&abstractObject)
		if err != nil {
			logrus.Warnf("Error retrieving parent metadata %s of API %s and Kind %s because of error: %v ", firstOwner.Name, firstOwner.APIVersion, firstOwner.Kind, err)
			return GenericResource{}, err
		}
		podSpec := GetPodSpec(abstractObject.Object)
		if podSpec != nil {
			topPodSpec = podSpec
		}
		topMeta = objMeta
		owners = abstractObject.GetOwnerReferences()
	}

	if lastKey != "" {
		unst := objectCache[lastKey]
		return NewGenericResourceFromUnstructured(unst, topPodSpec)
	}
	workload, err := NewGenericResourceFromPod(podResource, podResource)
	if err != nil {
		return workload, err
	}
	workload.Kind = topKind
	workload.ObjectMeta = topMeta
	return workload, nil
}

func cacheAllObjectsOfKind(ctx context.Context, apiVersion, kind string, dynamicClient *dynamic.Interface, restMapper *meta.RESTMapper, objectCache map[string]unstructured.Unstructured) error {
	fqKind := schema.FromAPIVersionAndKind(apiVersion, kind)
	mapping, err := (*restMapper).RESTMapping(fqKind.GroupKind(), fqKind.Version)
	if err != nil {
		logrus.Warnf("Error retrieving mapping of API %s and Kind %s because of error: %v", apiVersion, kind, err)
		return err
	}

	objects, err := (*dynamicClient).Resource(mapping.Resource).Namespace("").List(ctx, kubeAPIMetaV1.ListOptions{})
	if err != nil {
		logrus.Warnf("Error retrieving parent object API %s and Kind %s because of error: %v", mapping.Resource.Version, mapping.Resource.Resource, err)
		return err
	}
	for idx, object := range objects.Items {
		key := fmt.Sprintf("%s/%s/%s", object.GetKind(), object.GetNamespace(), object.GetName())
		objectCache[key] = objects.Items[idx]
	}
	return nil
}

func getObject(ctx context.Context, namespace, kind, version, name string, dynamicClient *dynamic.Interface, restMapper *meta.RESTMapper) (*unstructured.Unstructured, error) {
	fqKind := schema.ParseGroupKind(kind)
	mapping, err := (*restMapper).RESTMapping(fqKind, version)
	if err != nil {
		return nil, err
	}
	object, err := (*dynamicClient).Resource(mapping.Resource).Namespace(namespace).Get(ctx, name, kubeAPIMetaV1.GetOptions{})
	return object, err
}

// GetPodSpec looks inside arbitrary YAML for a PodSpec
func GetPodSpec(yaml map[string]interface{}) interface{} {
	for _, child := range podSpecFields {
		if childYaml, ok := yaml[child]; ok {
			return GetPodSpec(childYaml.(map[string]interface{}))
		}
	}
	if _, ok := yaml["containers"]; ok {
		return yaml
	}
	return nil
}

// GetPodTemplate looks inside arbitrary YAML for a Pod template, containing
// fields `spec.containers`.
// For example, it returns the `spec.template` level of a Kubernetes Deployment yaml.
func GetPodTemplate(yaml map[string]interface{}) (podTemplate interface{}, err error) {
	if yamlSpec, ok := yaml["spec"]; ok {
		if yamlSpecMap, ok := yamlSpec.(map[string]interface{}); ok {
			if _, ok := yamlSpecMap["containers"]; ok {
				// This is a hack around unstructured.SetNestedField using DeepCopy which does
				// not support the type int, and panics.
				// Related: https://github.com/kubernetes/kubernetes/issues/62769
				podTemplateJSON, err := json.Marshal(yaml)
				if err != nil {
					return nil, err
				}
				podTemplateMap := make(map[string]interface{})
				err = json.Unmarshal(podTemplateJSON, &podTemplateMap)
				if err != nil {
					return nil, err
				}
				return podTemplateMap, nil
			}
		}
	}
	for _, podSpecField := range podSpecFields {
		if childYaml, ok := yaml[podSpecField]; ok {
			return GetPodTemplate(childYaml.(map[string]interface{}))
		}
	}
	return nil, nil
}
