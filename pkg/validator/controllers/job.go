package controllers

import (
	"github.com/reactiveops/polaris/pkg/config"
	kubeAPIBatchV1 "k8s.io/api/batch/v1"
	kubeAPICoreV1 "k8s.io/api/core/v1"
)

// JobController is an implementation of controller for deployments
type JobController struct {
	GenericController
	K8SResource kubeAPIBatchV1.Job
}

// GetPodTemplate returns the original template spec
func (d JobController) GetPodTemplate() *kubeAPICoreV1.PodTemplateSpec {
	return &d.K8SResource.Spec.Template
}

// GetPodSpec returns the original kubernetes template pod spec
func (d JobController) GetPodSpec() *kubeAPICoreV1.PodSpec {
	return &d.K8SResource.Spec.Template.Spec
}

// GetType returns the supportedcontroller enum type
func (d JobController) GetType() config.SupportedController {
	return config.Jobs
}

// NewJobController builds a new controller interface for Deployments
func NewJobController(originalDeploymentResource kubeAPIBatchV1.Job) Interface {
	controller := JobController{}
	controller.Name = originalDeploymentResource.Name
	controller.Namespace = originalDeploymentResource.Namespace
	controller.K8SResource = originalDeploymentResource
	return controller
}
