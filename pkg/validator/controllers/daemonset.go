package controllers

import (
	"github.com/fairwindsops/polaris/pkg/config"
	kubeAPIAppsV1 "k8s.io/api/apps/v1"
	kubeAPICoreV1 "k8s.io/api/core/v1"
)

// DaemonSetController is an implementation of controller for deployments
type DaemonSetController struct {
	GenericController
	K8SResource kubeAPIAppsV1.DaemonSet
}

// GetPodTemplate returns the original template spec
func (d DaemonSetController) GetPodTemplate() *kubeAPICoreV1.PodTemplateSpec {
	return &d.K8SResource.Spec.Template
}

// GetPodSpec returns the original kubernetes template pod spec
func (d DaemonSetController) GetPodSpec() *kubeAPICoreV1.PodSpec {
	return &d.K8SResource.Spec.Template.Spec
}

// GetAnnotations returns the controller's annotations
func (d DaemonSetController) GetAnnotations() map[string]string {
	return d.K8SResource.ObjectMeta.Annotations
}

// GetType returns the supportedcontroller enum type
func (d DaemonSetController) GetType() config.SupportedController {
	return config.DaemonSets
}

// NewDaemonSetController builds a new controller interface for Deployments
func NewDaemonSetController(originalResource kubeAPIAppsV1.DaemonSet) Interface {
	controller := DaemonSetController{}
	controller.Name = originalResource.Name
	controller.Namespace = originalResource.Namespace
	controller.K8SResource = originalResource
	return controller
}
