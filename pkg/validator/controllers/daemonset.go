package controllers

import (
	"github.com/fairwindsops/polaris/pkg/config"
	kubeAPIAppsV1 "k8s.io/api/apps/v1"
)

// NewDaemonSetController builds a new controller interface for Deployments
func NewDaemonSetController(originalResource kubeAPIAppsV1.DaemonSet) Interface {
	controller := GenericController{}
	controller.Name = originalResource.Name
	controller.Namespace = originalResource.Namespace
	controller.PodSpec = originalResource.Spec.Template.Spec
	controller.ObjectMeta = originalResource.ObjectMeta
	controller.Kind = config.DaemonSets
	return controller
}
