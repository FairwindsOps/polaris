package kube

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	kubeAPICoreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	kubeAPIMetaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// GenericWorkload is a base implementation with some free methods for inherited structs
type GenericWorkload struct {
	Kind               string
	PodSpec            kubeAPICoreV1.PodSpec
	ObjectMeta         kubeAPIMetaV1.Object
	OriginalObjectJSON []byte
}

// NewGenericWorkloadFromPod builds a new workload for a given Pod without looking at parents
func NewGenericWorkloadFromPod(podResource kubeAPICoreV1.Pod, originalObject interface{}) (GenericWorkload, error) {
	workload := GenericWorkload{}
	workload.PodSpec = podResource.Spec
	workload.ObjectMeta = podResource.ObjectMeta.GetObjectMeta()
	workload.Kind = "Pod"
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
func NewGenericWorkload(podResource kubeAPICoreV1.Pod, dynamicClientPointer *dynamic.Interface, restMapperPointer *meta.RESTMapper, objectCache map[string]unstructured.Unstructured) (GenericWorkload, error) {
	workload, err := NewGenericWorkloadFromPod(podResource, nil)
	if err != nil {
		return workload, err
	}
	dynamicClient := *dynamicClientPointer
	restMapper := *restMapperPointer
	// If an owner exists then set the name to the workload.
	// This allows us to handle CRDs creating Workloads or DeploymentConfigs in OpenShift.
	owners := workload.ObjectMeta.GetOwnerReferences()
	lastKey := ""
	for len(owners) > 0 {
		if len(owners) > 1 {
			logrus.Warn("More than 1 owner found")
		}
		firstOwner := owners[0]
		if firstOwner.Kind == "Node" {
			break
		}
		workload.Kind = firstOwner.Kind
		key := fmt.Sprintf("%s/%s/%s", firstOwner.Kind, workload.ObjectMeta.GetNamespace(), firstOwner.Name)
		lastKey = key
		abstractObject, ok := objectCache[key]
		if ok {

			objMeta, err := meta.Accessor(&abstractObject)
			if err != nil {
				logrus.Warnf("Error retrieving parent metadata %s of API %s and Kind %s because of error: %v ", firstOwner.Name, firstOwner.APIVersion, firstOwner.Kind, err)
				return workload, err
			}
			workload.ObjectMeta = objMeta
			owners = abstractObject.GetOwnerReferences()

			continue
		}
		fqKind := schema.FromAPIVersionAndKind(firstOwner.APIVersion, firstOwner.Kind)
		mapping, err := restMapper.RESTMapping(fqKind.GroupKind(), fqKind.Version)
		if err != nil {
			logrus.Warnf("Error retrieving mapping %s of API %s and Kind %s because of error: %v ", firstOwner.Name, firstOwner.APIVersion, firstOwner.Kind, err)
			return workload, nil
		}
		err = cacheAllObjectsOfKind(dynamicClient, mapping.Resource, objectCache)
		if err != nil {
			logrus.Warnf("Error getting objects of Kind %s %v", firstOwner.Kind, err)
			return workload, err
		}

		abstractObject, ok = objectCache[key]
		if ok {

			objMeta, err := meta.Accessor(&abstractObject)
			if err != nil {
				logrus.Warnf("Error retrieving parent metadata %s of API %s and Kind %s because of error: %v ", firstOwner.Name, firstOwner.APIVersion, firstOwner.Kind, err)
				return workload, err
			}
			workload.ObjectMeta = objMeta
			owners = abstractObject.GetOwnerReferences()

			continue
		} else {
			logrus.Errorf("Cache missed again %s", key)
			return workload, errors.New("Could not retrieve parent object")
		}

	}
	if lastKey != "" {
		bytes, err := json.Marshal(objectCache[lastKey])
		if err != nil {
			return workload, err
		}
		workload.OriginalObjectJSON = bytes
	} else {
		bytes, err := json.Marshal(podResource)
		if err != nil {
			return workload, err
		}
		workload.OriginalObjectJSON = bytes
	}
	return workload, nil
}
