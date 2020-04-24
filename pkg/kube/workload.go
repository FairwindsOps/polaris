package kube

import (
	"time"

	"github.com/sirupsen/logrus"
	kubeAPICoreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	kubeAPIMetaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// GenericWorkload is a base implementation with some free methods for inherited structs
type GenericWorkload struct {
	Name        string
	Namespace   string
	PodSpec     kubeAPICoreV1.PodSpec
	ObjectMeta  kubeAPIMetaV1.ObjectMeta
	Kind        string
	CreatedTime time.Time
}

// NewGenericWorkload builds a new workload for a given Pod
func NewGenericWorkload(originalResource kubeAPICoreV1.Pod, dynamicClientPointer *dynamic.Interface, restMapperPointer *meta.RESTMapper) GenericWorkload {
	workload := GenericWorkload{}
	workload.Name = originalResource.Name
	workload.Namespace = originalResource.Namespace
	workload.PodSpec = originalResource.Spec
	workload.ObjectMeta = originalResource.ObjectMeta
	workload.Kind = "Pod"
	workload.CreatedTime = workload.ObjectMeta.CreationTimestamp.Time

	if dynamicClientPointer == nil || restMapperPointer == nil {
		return workload
	}
	// If an owner exists then set the name to the workload.
	// This allows us to handle CRDs creating Workloads or DeploymentConfigs in OpenShift.
	owners := workload.ObjectMeta.OwnerReferences
	for len(owners) > 0 {
		if len(owners) > 1 {
			logrus.Warn("More than 1 owner found")
		}
		firstOwner := owners[0]
		if firstOwner.Kind == "Node" {
			break
		}
		workload.Kind = firstOwner.Kind
		workload.Name = firstOwner.Name

		dynamicClient := *dynamicClientPointer
		restMapper := *restMapperPointer
		fqKind := schema.FromAPIVersionAndKind(firstOwner.APIVersion, firstOwner.Kind)
		mapping, err := restMapper.RESTMapping(fqKind.GroupKind(), fqKind.Version)
		if err != nil {
			logrus.Warnf("Error retrieving mapping %s of API %s and Kind %s because of error: %v ", firstOwner.Name, firstOwner.APIVersion, firstOwner.Kind, err)
			return workload
		}
		getParents, err := dynamicClient.Resource(mapping.Resource).Namespace(workload.ObjectMeta.Namespace).Get(firstOwner.Name, kubeAPIMetaV1.GetOptions{})
		if err != nil {
			logrus.Warnf("Error retrieving parent object %s of API %s and Kind %s because of error: %v ", firstOwner.Name, firstOwner.APIVersion, firstOwner.Kind, err)
			return workload
		}
		owners = getParents.GetOwnerReferences()
	}

	return workload
}
