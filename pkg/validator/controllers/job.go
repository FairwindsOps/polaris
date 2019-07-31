package controllers

import (
	"github.com/fairwindsops/polaris/pkg/config"
	kubeAPIBatchV1 "k8s.io/api/batch/v1"
	kubeAPICoreV1 "k8s.io/api/core/v1"
)

// JobController is an implementation of controller for deployments
type JobController struct {
	GenericController
	K8SResource kubeAPIBatchV1.Job
}

// GetPodTemplate returns the original template spec
func (j JobController) GetPodTemplate() *kubeAPICoreV1.PodTemplateSpec {
	return &j.K8SResource.Spec.Template
}

// GetPodSpec returns the original kubernetes template pod spec
func (j JobController) GetPodSpec() *kubeAPICoreV1.PodSpec {
	return &j.K8SResource.Spec.Template.Spec
}

// GetType returns the supportedcontroller enum type
func (j JobController) GetType() config.SupportedController {
	return config.Jobs
}

// NewJobController builds a new controller interface for Deployments
func NewJobController(originalResource kubeAPIBatchV1.Job) Interface {
	controller := JobController{}
	controller.Name = originalResource.Name
	controller.Namespace = originalResource.Namespace
	controller.K8SResource = originalResource
	return controller
}
