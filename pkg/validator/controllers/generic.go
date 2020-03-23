package controllers

import (
	"time"

	"github.com/fairwindsops/polaris/pkg/config"
	"github.com/sirupsen/logrus"
	kubeAPICoreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	kubeAPIMetaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// GenericController is a base implementation with some free methods for inherited structs
type GenericController struct {
	Name        string
	Namespace   string
	PodSpec     kubeAPICoreV1.PodSpec
	ObjectMeta  kubeAPIMetaV1.ObjectMeta
	Kind        config.SupportedController
	KindString  string
	CreatedTime time.Time
}

// GetPodSpec returns the original kubernetes template pod spec
func (g GenericController) GetPodSpec() *kubeAPICoreV1.PodSpec {
	return &g.PodSpec
}

// GetObjectMeta returns the metadata
func (g GenericController) GetObjectMeta() kubeAPIMetaV1.ObjectMeta {
	return g.ObjectMeta
}

// GetKind returns the supportedcontroller enum type
func (g GenericController) GetKind() config.SupportedController {
	return g.Kind
}

// GetKindString returns a string representing what kind of object the top level controller is.
func (g GenericController) GetKindString() string {
	return g.KindString
}

// GetName is inherited by all controllers using generic controller to get the name of the controller
func (g GenericController) GetName() string {
	return g.Name
}

// GetNamespace is inherited by all controllers using generic controller to get the namespace of the controller
func (g GenericController) GetNamespace() string {
	return g.Namespace
}

// LoadControllers loads a list of controllers from the kubeResources Pods
func LoadControllers(pods []kubeAPICoreV1.Pod, dynamicClientPointer *dynamic.Interface, restMapperPointer *meta.RESTMapper) []GenericController {
	interfaces := []GenericController{}
	for _, pod := range pods {
		interfaces = append(interfaces, NewGenericPodController(pod, dynamicClientPointer, restMapperPointer))
	}
	// TODO DeDupe
	return interfaces
}

// NewGenericPodController builds a new controller interface for anytype of Pod
func NewGenericPodController(originalResource kubeAPICoreV1.Pod, dynamicClientPointer *dynamic.Interface, restMapperPointer *meta.RESTMapper) GenericController {
	controller := GenericController{}
	controller.Name = originalResource.Name
	controller.Namespace = originalResource.Namespace
	controller.PodSpec = originalResource.Spec
	controller.ObjectMeta = originalResource.ObjectMeta
	controller.Kind = config.NakedPods
	controller.KindString = "Pod"
	controller.CreatedTime = controller.GetObjectMeta().CreationTimestamp.Time

	owners := controller.GetObjectMeta().OwnerReferences
	if dynamicClientPointer == nil || restMapperPointer == nil {
		return controller
	}
	// If an owner exists then set the name to the controller.
	// This allows us to handle CRDs creating Controllers or DeploymentConfigs in OpenShift.
	for len(owners) > 0 {
		firstOwner := owners[0]
		controller.KindString = firstOwner.Kind
		controller.Name = firstOwner.Name

		dynamicClient := *dynamicClientPointer
		restMapper := *restMapperPointer
		fqKind := schema.FromAPIVersionAndKind(firstOwner.APIVersion, firstOwner.Kind)
		mapping, err := restMapper.RESTMapping(fqKind.GroupKind(), fqKind.Version)
		if err != nil {
			logrus.Warnf("Error retrieving mapping %s of API %s and Kind %s because of error: %v ", firstOwner.Name, firstOwner.APIVersion, firstOwner.Kind, err)
			return controller
		}
		getParents, err := dynamicClient.Resource(mapping.Resource).Namespace(controller.GetObjectMeta().Namespace).Get(firstOwner.Name, kubeAPIMetaV1.GetOptions{})
		if err != nil {
			logrus.Warnf("Error retrieving parent object %s of API %s and Kind %s because of error: %v ", firstOwner.Name, firstOwner.APIVersion, firstOwner.Kind, err)
			return controller
		}
		owners = getParents.GetOwnerReferences()

	}

	return controller
}
