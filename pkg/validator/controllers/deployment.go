package controllers

import (
	"github.com/fairwindsops/polaris/pkg/config"
	kubeAPIAppsV1 "k8s.io/api/apps/v1"
	kubeAPICoreV1 "k8s.io/api/core/v1"
	kubeAPIMetaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DeploymentController is an implementation of controller for deployments
type DeploymentController struct {
	GenericController
	K8SResource kubeAPIAppsV1.Deployment
}

// GetPodSpec returns the original kubernetes template pod spec
func (d DeploymentController) GetPodSpec() *kubeAPICoreV1.PodSpec {
	return &d.K8SResource.Spec.Template.Spec
}

// GetObjectMeta returns the metadata
func (d DeploymentController) GetObjectMeta() kubeAPIMetaV1.ObjectMeta {
	return d.K8SResource.ObjectMeta
}

// GetKind returns the supportedcontroller enum type
func (d DeploymentController) GetKind() config.SupportedController {
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
