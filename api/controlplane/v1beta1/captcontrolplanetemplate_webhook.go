package v1beta1

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var captcontrolplanetemplatelog = logf.Log.WithName("captcontrolplanetemplate-resource")

func (r *CAPTControlPlaneTemplate) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-controlplane-cluster-x-k8s-io-v1beta1-captcontrolplanetemplate,mutating=true,failurePolicy=fail,sideEffects=None,groups=controlplane.cluster.x-k8s.io,resources=captcontrolplanetemplates,verbs=create;update,versions=v1beta1,name=mcaptcontrolplanetemplate.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &CAPTControlPlaneTemplate{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *CAPTControlPlaneTemplate) Default() {
	captcontrolplanetemplatelog.Info("default", "name", r.Name)
}

//+kubebuilder:webhook:path=/validate-controlplane-cluster-x-k8s-io-v1beta1-captcontrolplanetemplate,mutating=false,failurePolicy=fail,sideEffects=None,groups=controlplane.cluster.x-k8s.io,resources=captcontrolplanetemplates,verbs=create;update,versions=v1beta1,name=vcaptcontrolplanetemplate.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &CAPTControlPlaneTemplate{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *CAPTControlPlaneTemplate) ValidateCreate() (admission.Warnings, error) {
	captcontrolplanetemplatelog.Info("validate create", "name", r.Name)
	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *CAPTControlPlaneTemplate) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	captcontrolplanetemplatelog.Info("validate update", "name", r.Name)
	oldTpl, ok := old.(*CAPTControlPlaneTemplate)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected a CAPTControlPlaneTemplate but got a %T", old))
	}

	// Enforce immutability of spec.template when updated
	if !equality.Semantic.DeepEqual(r.Spec.Template, oldTpl.Spec.Template) {
		gr := schema.GroupResource{Group: r.GroupVersionKind().Group, Resource: "captcontrolplanetemplates"}
		return nil, apierrors.NewForbidden(gr, r.Name, fmt.Errorf("CAPTControlPlaneTemplate.spec.template is immutable"))
	}

	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *CAPTControlPlaneTemplate) ValidateDelete() (admission.Warnings, error) {
	captcontrolplanetemplatelog.Info("validate delete", "name", r.Name)
	return nil, nil
}
