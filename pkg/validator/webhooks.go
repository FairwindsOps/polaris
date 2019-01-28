package validator

import (
	"fmt"
	"os"

	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission/builder"
)

// NewWebhook creates a validating admission webhook for the apiType.
func NewWebhook(name string, mgr manager.Manager, validator Validator, apiType runtime.Object) *admission.Webhook {
	name = fmt.Sprintf("%s.k8s.io", name)
	path := fmt.Sprintf("/validating-%s", name)

	webhook, err := builder.NewWebhookBuilder().
		Name(name).
		Validating().
		Path(path).
		Operations(admissionregistrationv1beta1.Create, admissionregistrationv1beta1.Update).
		WithManager(mgr).
		ForType(apiType).
		Handlers(&validator).
		Build()
	if err != nil {
		log.Error(err, "unable to setup validating webhook:", name)
		os.Exit(1)
	}

	return webhook
}
