package controllers

import (
	"github.com/fairwindsops/polaris/pkg/config"
	kubeAPIAppsV1 "k8s.io/api/apps/v1"
)

// NewStatefulSetController builds a statefulset controller
func NewStatefulSetController(originalResource kubeAPIAppsV1.StatefulSet) Interface {
	controller := GenericController{}
	controller.Name = originalResource.Name
	controller.Namespace = originalResource.Namespace
	controller.PodSpec = originalResource.Spec.Template.Spec
	controller.ObjectMeta = originalResource.ObjectMeta
	controller.Kind = config.StatefulSets

	return controller
}
