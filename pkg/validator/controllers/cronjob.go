package controllers

import (
	"github.com/fairwindsops/polaris/pkg/config"
	kubeAPIBatchV1beta1 "k8s.io/api/batch/v1beta1"
	kubeAPICoreV1 "k8s.io/api/core/v1"
)

// CronJobController is an implementation of controller for deployments
type CronJobController struct {
	GenericController
	K8SResource kubeAPIBatchV1beta1.CronJob
}

// GetPodTemplate returns the original template spec
func (c CronJobController) GetPodTemplate() *kubeAPICoreV1.PodTemplateSpec {
	return &c.K8SResource.Spec.JobTemplate.Spec.Template
}

// GetPodSpec returns the original kubernetes template pod spec
func (c CronJobController) GetPodSpec() *kubeAPICoreV1.PodSpec {
	return &c.K8SResource.Spec.JobTemplate.Spec.Template.Spec
}

// GetType returns the supportedcontroller enum type
func (c CronJobController) GetType() config.SupportedController {
	return config.CronJobs
}

// GetAnnotations returns the controller's annotations
func (c CronJobController) GetAnnotations() map[string]string {
	return c.K8SResource.ObjectMeta.Annotations
}

// NewCronJobController builds a new controller interface for Deployments
func NewCronJobController(originalDeploymentResource kubeAPIBatchV1beta1.CronJob) Interface {
	controller := CronJobController{}
	controller.Name = originalDeploymentResource.Name
	controller.Namespace = originalDeploymentResource.Namespace
	controller.K8SResource = originalDeploymentResource
	return controller
}
