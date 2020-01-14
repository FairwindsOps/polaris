package controllers

import (
	"github.com/fairwindsops/polaris/pkg/config"
	kubeAPICoreV1 "k8s.io/api/core/v1"
	kubeAPIMetaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NakedPodController is an implementation of controller for deployments
type NakedPodController struct {
	GenericController
	K8SResource kubeAPICoreV1.Pod
}

// GetPodTemplate returns the original template spec
func (n NakedPodController) GetPodTemplate() *kubeAPICoreV1.PodTemplateSpec {
	return nil
}

// GetPodSpec returns the original kubernetes template pod spec
func (n NakedPodController) GetPodSpec() *kubeAPICoreV1.PodSpec {
	return &n.K8SResource.Spec
}

// GetObjectMeta returns the metadata
func (n NakedPodController) GetObjectMeta() kubeAPIMetaV1.ObjectMeta {
	return n.K8SResource.ObjectMeta
}

// GetKind returns the supportedcontroller enum type
func (n NakedPodController) GetKind() config.SupportedController {
	return config.NakedPods
}

// NewNakedPodController builds a new controller interface for NakedPods
func NewNakedPodController(originalNakedPodResource kubeAPICoreV1.Pod) Interface {
	controller := NakedPodController{}
	controller.Name = originalNakedPodResource.Name
	controller.Namespace = originalNakedPodResource.Namespace
	controller.K8SResource = originalNakedPodResource
	return controller
}
