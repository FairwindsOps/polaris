package controllers

import (
	"github.com/fairwindsops/polaris/pkg/config"
	kubeAPICoreV1 "k8s.io/api/core/v1"
)

// NewNakedPodController builds a new controller interface for NakedPods
func NewNakedPodController(originalResource kubeAPICoreV1.Pod) GenericController {
	controller := GenericController{}
	controller.Name = originalResource.Name
	controller.Namespace = originalResource.Namespace
	controller.PodSpec = originalResource.Spec
	controller.ObjectMeta = originalResource.ObjectMeta
	controller.Kind = config.NakedPods.String()

	return controller
}
