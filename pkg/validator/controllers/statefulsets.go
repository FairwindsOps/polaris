package controllers

import (
	"github.com/fairwindsops/polaris/pkg/config"
	kubeAPIAppsV1 "k8s.io/api/apps/v1"
	kubeAPICoreV1 "k8s.io/api/core/v1"
)

// StatefulSetController is an implementation of controller for deployments
type StatefulSetController struct {
	GenericController
	K8SResource kubeAPIAppsV1.StatefulSet
}

// GetPodTemplate returns the kubernetes template spec
func (s StatefulSetController) GetPodTemplate() *kubeAPICoreV1.PodTemplateSpec {
	return &s.K8SResource.Spec.Template
}

// GetPodSpec returns the podspec from the original kubernetes resource
func (s StatefulSetController) GetPodSpec() *kubeAPICoreV1.PodSpec {
	return &s.K8SResource.Spec.Template.Spec
}

// GetType returns the supportedcontroller enum type
func (s StatefulSetController) GetType() config.SupportedController {
	return config.StatefulSets
}

// NewStatefulSetController builds a statefulset controller
func NewStatefulSetController(originalResource kubeAPIAppsV1.StatefulSet) Interface {
	controller := StatefulSetController{}
	controller.Name = originalResource.Name
	controller.Namespace = originalResource.Namespace
	controller.K8SResource = originalResource
	return controller
}
