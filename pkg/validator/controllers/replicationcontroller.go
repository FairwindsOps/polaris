package controllers

import (
	"github.com/fairwindsops/polaris/pkg/config"
	kubeAPICoreV1 "k8s.io/api/core/v1"
)

// NewReplicationControllerController builds a new controller interface for Deployments
func NewReplicationControllerController(originalResource kubeAPICoreV1.ReplicationController) GenericController {
	controller := GenericController{}
	controller.Name = originalResource.Name
	controller.Namespace = originalResource.Namespace
	controller.PodSpec = originalResource.Spec.Template.Spec
	controller.ObjectMeta = originalResource.ObjectMeta
	controller.Kind = config.ReplicationControllers
	return controller
}
