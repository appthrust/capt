package v1beta1

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var captclustertemplatelog = logf.Log.WithName("captclustertemplate-resource")

func (r *CAPTClusterTemplate) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-infrastructure-cluster-x-k8s-io-v1beta1-captclustertemplate,mutating=true,failurePolicy=fail,sideEffects=None,groups=infrastructure.cluster.x-k8s.io,resources=captclustertemplates,verbs=create;update,versions=v1beta1,name=mcaptclustertemplate.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &CAPTClusterTemplate{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *CAPTClusterTemplate) Default() {
	captclustertemplatelog.Info("default", "name", r.Name)
	// TODO(user): fill in your defaulting logic.
}

// +kubebuilder:webhook:path=/validate-infrastructure-cluster-x-k8s-io-v1beta1-captclustertemplate,mutating=false,failurePolicy=fail,sideEffects=None,groups=infrastructure.cluster.x-k8s.io,resources=captclustertemplates,verbs=create;update,versions=v1beta1,name=vcaptclustertemplate.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &CAPTClusterTemplate{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *CAPTClusterTemplate) ValidateCreate() (admission.Warnings, error) {
	captclustertemplatelog.Info("validate create", "name", r.Name)

	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *CAPTClusterTemplate) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	captclustertemplatelog.Info("validate update", "name", r.Name)
	oldTemplate, ok := old.(*CAPTClusterTemplate)
	if !ok {
		return nil, errors.NewBadRequest(fmt.Sprintf("expected a CAPTClusterTemplate but got a %T", old))
	}

	if !equality.Semantic.DeepEqual(r.Spec.Template, oldTemplate.Spec.Template) {
		gr := schema.GroupResource{Group: r.GroupVersionKind().Group, Resource: "captclustertemplates"}
		return nil, errors.NewForbidden(gr, r.Name, fmt.Errorf("CAPTClusterTemplate.spec.template is immutable"))
	}

	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *CAPTClusterTemplate) ValidateDelete() (admission.Warnings, error) {
	captclustertemplatelog.Info("validate delete", "name", r.Name)

	return nil, nil
}
