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
var captmachinetemplatelog = logf.Log.WithName("captmachinetemplate-resource")

func (r *CaptMachineTemplate) SetupWebhookWithManager(mgr ctrl.Manager) error {
	if webhookClient == nil {
		webhookClient = mgr.GetClient()
	}
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-infrastructure-cluster-x-k8s-io-v1beta1-captmachinetemplate,mutating=true,failurePolicy=fail,sideEffects=None,groups=infrastructure.cluster.x-k8s.io,resources=captmachinetemplates,verbs=create;update,versions=v1beta1,name=mcaptmachinetemplate.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &CaptMachineTemplate{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *CaptMachineTemplate) Default() {
	captmachinetemplatelog.Info("default", "name", r.Name)
}

//+kubebuilder:webhook:path=/validate-infrastructure-cluster-x-k8s-io-v1beta1-captmachinetemplate,mutating=false,failurePolicy=fail,sideEffects=None,groups=infrastructure.cluster.x-k8s.io,resources=captmachinetemplates,verbs=create;update,versions=v1beta1,name=vcaptmachinetemplate.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &CaptMachineTemplate{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *CaptMachineTemplate) ValidateCreate() (admission.Warnings, error) {
	captmachinetemplatelog.Info("validate create", "name", r.Name)
	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *CaptMachineTemplate) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	captmachinetemplatelog.Info("validate update", "name", r.Name)
	oldMachine, ok := old.(*CaptMachineTemplate)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected a CaptMachineTemplate but got a %T", old))
	}

	if !equality.Semantic.DeepEqual(r.Spec.Template, oldMachine.Spec.Template) {
		gr := schema.GroupResource{Group: r.GroupVersionKind().Group, Resource: "captmachinetemplates"}
		return nil, apierrors.NewForbidden(gr, r.Name, fmt.Errorf("CaptMachineTemplate.spec.template is immutable"))
	}

	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *CaptMachineTemplate) ValidateDelete() (admission.Warnings, error) {
	captmachinetemplatelog.Info("validate delete", "name", r.Name)
	return nil, nil
}
