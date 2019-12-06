package controllers

import (
	"github.com/fairwindsops/polaris/pkg/config"
	kubeAPICoreV1 "k8s.io/api/core/v1"
)

// NOTE: Maybe this name of ReplicationController is duplicative but it's more explicit since
//       that's how kubernetes refers the the object.

// ReplicationControllerController is an implementation of controller for deployments
type ReplicationControllerController struct {
	GenericController
	K8SResource kubeAPICoreV1.ReplicationController
}

// GetPodTemplate returns the original template spec
func (r ReplicationControllerController) GetPodTemplate() *kubeAPICoreV1.PodTemplateSpec {
	return r.K8SResource.Spec.Template
}

// GetPodSpec returns the original kubernetes template pod spec
func (r ReplicationControllerController) GetPodSpec() *kubeAPICoreV1.PodSpec {
	return &r.K8SResource.Spec.Template.Spec
}

// GetAnnotations returns the controller's annotations
func (r ReplicationControllerController) GetAnnotations() map[string]string {
	return r.K8SResource.ObjectMeta.Annotations
}

// GetType returns the supportedcontroller enum type
func (r ReplicationControllerController) GetType() config.SupportedController {
	return config.ReplicationControllers
}

// NewReplicationControllerController builds a new controller interface for Deployments
func NewReplicationControllerController(originalResource kubeAPICoreV1.ReplicationController) Interface {
	controller := ReplicationControllerController{}
	controller.Name = originalResource.Name
	controller.Namespace = originalResource.Namespace
	controller.K8SResource = originalResource
	return controller
}
