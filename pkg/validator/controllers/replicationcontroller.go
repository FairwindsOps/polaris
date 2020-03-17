package controllers

import (
	"github.com/fairwindsops/polaris/pkg/config"
	kubeAPICoreV1 "k8s.io/api/core/v1"
	kubeAPIMetaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: Maybe this name of ReplicationController is duplicative but it's more explicit since
//       that's how kubernetes refers the the object.

// ReplicationControllerController is an implementation of controller for deployments
type ReplicationControllerController struct {
	GenericController
	K8SResource kubeAPICoreV1.ReplicationController
}

// GetPodSpec returns the original kubernetes template pod spec
func (r ReplicationControllerController) GetPodSpec() *kubeAPICoreV1.PodSpec {
	return &r.K8SResource.Spec.Template.Spec
}

// GetObjectMeta returns the metadata
func (r ReplicationControllerController) GetObjectMeta() kubeAPIMetaV1.ObjectMeta {
	return r.K8SResource.ObjectMeta
}

// GetKind returns the supportedcontroller enum type
func (r ReplicationControllerController) GetKind() config.SupportedController {
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
