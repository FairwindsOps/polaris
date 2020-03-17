package controllers

import (
	"github.com/fairwindsops/polaris/pkg/config"
	kubeAPIBatchV1beta1 "k8s.io/api/batch/v1beta1"
)

// NewCronJobController builds a new controller interface for Deployments
func NewCronJobController(originalResource kubeAPIBatchV1beta1.CronJob) Interface {
	controller := GenericController{}
	controller.Name = originalResource.Name
	controller.Namespace = originalResource.Namespace
	controller.PodSpec = originalResource.Spec.JobTemplate.Spec.Template.Spec
	controller.ObjectMeta = originalResource.ObjectMeta
	controller.Kind = config.CronJobs

	return controller
}
