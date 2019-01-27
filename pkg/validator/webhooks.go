package validator

import (
	"os"

	conf "github.com/reactiveops/fairwinds/pkg/config"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission/builder"
)

// NewDeployWebhook creates a validating admission webhook for deploy creation and updates.
func NewDeployWebhook(mgr manager.Manager, c conf.Configuration) *admission.Webhook {
	webhook, err := builder.NewWebhookBuilder().
		Name("deploy.k8s.io").
		Validating().
		Path("/validating-deployment").
		Operations(admissionregistrationv1beta1.Create, admissionregistrationv1beta1.Update).
		WithManager(mgr).
		ForType(&appsv1.Deployment{}).
		Handlers(&DeployValidator{Config: c}).
		Build()
	if err != nil {
		log.Error(err, "unable to setup deploy validating webhook")
		os.Exit(1)
	}
	return webhook
}

// NewPodWebhook creates a validating admission webhook for pod creation and updates.
func NewPodWebhook(mgr manager.Manager, c conf.Configuration) *admission.Webhook {
	webhook, err := builder.NewWebhookBuilder().
		Name("pod.k8s.io").
		Validating().
		Path("/validating-pod").
		Operations(admissionregistrationv1beta1.Create, admissionregistrationv1beta1.Update).
		WithManager(mgr).
		ForType(&corev1.Pod{}).
		Handlers(&PodValidator{Config: c}).
		Build()
	if err != nil {
		log.Error(err, "unable to setup pod validating webhook")
		os.Exit(1)
	}
	return webhook
}
