package kube

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/fairwindsops/controller-utils/pkg/controller"
	"github.com/fairwindsops/controller-utils/pkg/podspec"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	kubeAPICoreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	kubeAPIMetaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8sYaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
)

// GenericWorkload is a base implementation with some free methods for inherited structs
type GenericWorkload struct {
	Kind               string
	PodSpec            kubeAPICoreV1.PodSpec
	ObjectMeta         kubeAPIMetaV1.Object
	OriginalObjectJSON []byte
}

// NewGenericWorkloadFromUnstructured creates a workload from an unstructured.Unstructured
func NewGenericWorkloadFromUnstructured(kind string, unst *unstructured.Unstructured) (GenericWorkload, error) {
	workload := GenericWorkload{
		Kind: kind,
	}

	objMeta, err := meta.Accessor(unst)
	if err != nil {
		return workload, err
	}
	workload.ObjectMeta = objMeta

	b, err := json.Marshal(unst)
	if err != nil {
		return workload, err
	}
	workload.OriginalObjectJSON = b

	m := make(map[string]interface{})
	err = json.Unmarshal(b, &m)
	if err != nil {
		return workload, err
	}
	podSpecMap := podspec.GetPodSpec(m)
	b, err = json.Marshal(podSpecMap)
	if err != nil {
		return workload, err
	}
	podSpecObject := kubeAPICoreV1.PodSpec{}
	err = json.Unmarshal(b, &podSpecObject)
	if err != nil {
		return workload, err
	}
	workload.PodSpec = podSpecObject

	return workload, nil
}

// NewGenericWorkloadFromPod builds a new workload for a given Pod without looking at parents
func NewGenericWorkloadFromPod(podResource kubeAPICoreV1.Pod, originalObject interface{}) (GenericWorkload, error) {
	workload := GenericWorkload{
		Kind:       "Pod",
		PodSpec:    podResource.Spec,
		ObjectMeta: podResource.ObjectMeta.GetObjectMeta(),
	}
	if originalObject != nil {
		bytes, err := json.Marshal(originalObject)
		if err != nil {
			return workload, err
		}
		workload.OriginalObjectJSON = bytes
	}
	return workload, nil
}

// NewGenericWorkload builds a new workload for a given Pod
func NewGenericWorkload(ctx context.Context, podResource kubeAPICoreV1.Pod, dynamicClient *dynamic.Interface, restMapper *meta.RESTMapper, objectCache map[string]unstructured.Unstructured) (GenericWorkload, error) {
	workload, err := newGenericWorkload(ctx, podResource, dynamicClient, restMapper, objectCache)
	if err != nil {
		return workload, err
	}
	if len(workload.OriginalObjectJSON) == 0 {
		return NewGenericWorkloadFromPod(podResource, podResource)
	}
	return workload, err
}

func newGenericWorkload(ctx context.Context, podResource kubeAPICoreV1.Pod, dynamicClient *dynamic.Interface, restMapper *meta.RESTMapper, objectCache map[string]unstructured.Unstructured) (GenericWorkload, error) {
	workload, err := NewGenericWorkloadFromPod(podResource, nil)
	if err != nil {
		return workload, err
	}
	objMeta, err := meta.Accessor(&podResource)
	if err != nil {
		return workload, err
	}
	controllerObject, err := controller.GetTopController(ctx, *dynamicClient, *restMapper, objMeta)
	if err != nil {
		return workload, err
	}

	bytes, err := json.Marshal(controllerObject)
	if err != nil {
		return workload, err
	}
	workload.OriginalObjectJSON = bytes
	return workload, nil
}

func cacheAllObjectsOfKind(ctx context.Context, apiVersion, kind string, dynamicClient *dynamic.Interface, restMapper *meta.RESTMapper, objectCache map[string]unstructured.Unstructured) error {
	fqKind := schema.FromAPIVersionAndKind(apiVersion, kind)
	mapping, err := (*restMapper).RESTMapping(fqKind.GroupKind(), fqKind.Version)
	if err != nil {
		logrus.Warnf("Error retrieving mapping of API %s and Kind %s because of error: %v ", apiVersion, kind, err)
		return err
	}

	objects, err := (*dynamicClient).Resource(mapping.Resource).Namespace("").List(ctx, kubeAPIMetaV1.ListOptions{})
	if err != nil {
		logrus.Warnf("Error retrieving parent object API %s and Kind %s because of error: %v ", mapping.Resource.Version, mapping.Resource.Resource, err)
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

// GetWorkloadFromBytes parses a GenericWorkload
func GetWorkloadFromBytes(contentBytes []byte) (*GenericWorkload, error) {
	yamlNode := make(map[string]interface{})
	err := yaml.Unmarshal(contentBytes, &yamlNode)
	if err != nil {
		logrus.Errorf("Invalid YAML: %s", string(contentBytes))
		return nil, err
	}
	finalDoc := make(map[string]interface{})
	finalDoc["metadata"] = yamlNode["metadata"]
	finalDoc["apiVersion"] = "v1"
	finalDoc["kind"] = "Pod"
	podSpecObject := podspec.GetPodSpec(yamlNode)
	if podSpecObject == nil {
		return nil, nil
	}
	finalDoc["spec"] = podSpecObject
	marshaledYaml, err := yaml.Marshal(finalDoc)
	if err != nil {
		logrus.Errorf("Could not marshal yaml: %v", err)
		return nil, err
	}
	decoder := k8sYaml.NewYAMLOrJSONDecoder(bytes.NewReader(marshaledYaml), 1000)
	pod := kubeAPICoreV1.Pod{}
	err = decoder.Decode(&pod)
	newController, err := NewGenericWorkloadFromPod(pod, yamlNode)

	if err != nil {
		return nil, err
	}
	newController.Kind = yamlNode["kind"].(string)
	return &newController, nil
}
