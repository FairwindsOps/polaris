package controllers

import (
	"github.com/fairwindsops/polaris/pkg/config"
	"github.com/sirupsen/logrus"
	kubeAPIBatchV1beta1 "k8s.io/api/batch/v1beta1"
)

// NewCronJobController builds a new controller interface for Deployments
func NewCronJobController(originalResource kubeAPIBatchV1beta1.CronJob) GenericController {
	controller := GenericController{}
	controller.Name = originalResource.Name
	controller.Namespace = originalResource.Namespace
	controller.PodSpec = originalResource.Spec.JobTemplate.Spec.Template.Spec
	controller.ObjectMeta = originalResource.ObjectMeta
	controller.Kind = config.CronJobs
	if controller.Name == "" {
		logrus.Warn("Name is missing from controller", originalResource.Namespace)
	}
	return controller
}
