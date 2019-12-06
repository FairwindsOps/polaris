package controllers

import (
	"github.com/fairwindsops/polaris/pkg/config"
	kubeAPIAppsV1 "k8s.io/api/apps/v1"
	kubeAPICoreV1 "k8s.io/api/core/v1"
)

// DeploymentController is an implementation of controller for deployments
type DeploymentController struct {
	GenericController
	K8SResource kubeAPIAppsV1.Deployment
}

// GetPodTemplate returns the original template spec
func (d DeploymentController) GetPodTemplate() *kubeAPICoreV1.PodTemplateSpec {
	return &d.K8SResource.Spec.Template
}

// GetPodSpec returns the original kubernetes template pod spec
func (d DeploymentController) GetPodSpec() *kubeAPICoreV1.PodSpec {
	return &d.K8SResource.Spec.Template.Spec
}

// GetAnnotations returns the controller's annotations
func (d DeploymentController) GetAnnotations() map[string]string {
	return d.K8SResource.ObjectMeta.Annotations
}

// GetType returns the supportedcontroller enum type
func (d DeploymentController) GetType() config.SupportedController {
	return config.Deployments
}

// NewDeploymentController builds a new controller interface for Deployments
func NewDeploymentController(originalDeploymentResource kubeAPIAppsV1.Deployment) Interface {
	controller := DeploymentController{}
	controller.Name = originalDeploymentResource.Name
	controller.Namespace = originalDeploymentResource.Namespace
	controller.K8SResource = originalDeploymentResource
	return controller
}
