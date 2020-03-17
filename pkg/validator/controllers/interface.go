package controllers

import (
	"github.com/fairwindsops/polaris/pkg/config"
	"github.com/fairwindsops/polaris/pkg/kube"
	kubeAPICoreV1 "k8s.io/api/core/v1"
	kubeAPIMetaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Interface is an interface for k8s controllers (e.g. Deployments and StatefulSets)
type Interface interface {
	GetName() string
	GetNamespace() string
	GetPodSpec() *kubeAPICoreV1.PodSpec
	GetKind() config.SupportedController
	GetObjectMeta() kubeAPIMetaV1.ObjectMeta
}

// GenericController is a base implementation with some free methods for inherited structs
type GenericController struct {
	Name      string
	Namespace string
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
func LoadControllers(kubeResources *kube.ResourceProvider) []Interface {
	interfaces := []Interface{}
	for _, pod := range kubeResources.Pods {
		interfaces = append(interfaces, NewNakedPodController(pod))
	}
	return interfaces
}
