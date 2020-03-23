package controllers

import (
	"github.com/fairwindsops/polaris/pkg/config"
	kubeAPIBatchV1 "k8s.io/api/batch/v1"
)

// NewJobController builds a new controller interface for Deployments
func NewJobController(originalResource kubeAPIBatchV1.Job) GenericController {
	controller := GenericController{}
	controller.Name = originalResource.Name
	controller.Namespace = originalResource.Namespace
	controller.PodSpec = originalResource.Spec.Template.Spec
	controller.ObjectMeta = originalResource.ObjectMeta
	controller.Kind = config.Jobs
	return controller
}
