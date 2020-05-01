package kube

import (
	"fmt"

	"github.com/sirupsen/logrus"
	kubeAPICoreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	kubeAPIMetaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// GenericWorkload is a base implementation with some free methods for inherited structs
type GenericWorkload struct {
	Kind       string
	PodSpec    kubeAPICoreV1.PodSpec
	ObjectMeta kubeAPIMetaV1.Object
}

// NewGenericWorkloadFromPod builds a new workload for a given Pod without looking at parents
func NewGenericWorkloadFromPod(originalResource kubeAPICoreV1.Pod) GenericWorkload {
	workload := GenericWorkload{}
	workload.PodSpec = originalResource.Spec
	workload.ObjectMeta = originalResource.ObjectMeta.GetObjectMeta()
	workload.Kind = "Pod"
	return workload
}

// NewGenericWorkload builds a new workload for a given Pod
func NewGenericWorkload(originalResource kubeAPICoreV1.Pod, dynamicClientPointer *dynamic.Interface, restMapperPointer *meta.RESTMapper, objectCache map[string]kubeAPIMetaV1.Object) GenericWorkload {
	workload := NewGenericWorkloadFromPod(originalResource)

	dynamicClient := *dynamicClientPointer
	restMapper := *restMapperPointer
	// If an owner exists then set the name to the workload.
	// This allows us to handle CRDs creating Workloads or DeploymentConfigs in OpenShift.
	owners := workload.ObjectMeta.GetOwnerReferences()
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
		objMeta, ok := objectCache[key]
		if ok {
			workload.ObjectMeta = objMeta
			owners = objMeta.GetOwnerReferences()

			continue
		}
		fqKind := schema.FromAPIVersionAndKind(firstOwner.APIVersion, firstOwner.Kind)
		mapping, err := restMapper.RESTMapping(fqKind.GroupKind(), fqKind.Version)
		if err != nil {
			logrus.Warnf("Error retrieving mapping %s of API %s and Kind %s because of error: %v ", firstOwner.Name, firstOwner.APIVersion, firstOwner.Kind, err)
			return workload
		}
		err = cacheAllObjectsOfKind(dynamicClient, mapping.Resource, objectCache)
		if err != nil {
			logrus.Warnf("Error getting objects of Kind %s %v", firstOwner.Kind, err)
			return workload
		}

		objMeta, ok = objectCache[key]
		if ok {
			workload.ObjectMeta = objMeta
			owners = objMeta.GetOwnerReferences()

			continue
		} else {
			logrus.Errorf("Cache missed again %s", key)
		}
		parent, err := dynamicClient.Resource(mapping.Resource).Namespace(workload.ObjectMeta.GetNamespace()).Get(firstOwner.Name, kubeAPIMetaV1.GetOptions{})
		if err != nil {
			logrus.Warnf("Error retrieving parent object %s of API %s and Kind %s because of error: %v ", firstOwner.Name, firstOwner.APIVersion, firstOwner.Kind, err)
			return workload
		}
		objMeta, err = meta.Accessor(parent)
		if err != nil {
			logrus.Warnf("Error retrieving parent metadata %s of API %s and Kind %s because of error: %v ", firstOwner.Name, firstOwner.APIVersion, firstOwner.Kind, err)
			return workload
		}
		workload.ObjectMeta = objMeta
		objectCache[key] = objMeta
		owners = parent.GetOwnerReferences()
	}

	return workload
}
