package v1beta2

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var captcontrolplanelog = logf.Log.WithName("captcontrolplane-resource")
var webhookClient client.Client

func (r *CAPTControlPlane) SetupWebhookWithManager(mgr ctrl.Manager) error {
	if webhookClient == nil {
		webhookClient = mgr.GetClient()
	}
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-controlplane-cluster-x-k8s-io-v1beta2-captcontrolplane,mutating=true,failurePolicy=fail,sideEffects=None,groups=controlplane.cluster.x-k8s.io,resources=captcontrolplanes,verbs=create;update,versions=v1beta2,name=mcaptcontrolplane.v1beta2.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &CAPTControlPlane{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *CAPTControlPlane) Default() {
	captcontrolplanelog.Info("default", "name", r.Name)
}

// +kubebuilder:webhook:path=/validate-controlplane-cluster-x-k8s-io-v1beta2-captcontrolplane,mutating=false,failurePolicy=fail,sideEffects=None,groups=controlplane.cluster.x-k8s.io,resources=captcontrolplanes,verbs=create;update,versions=v1beta2,name=vcaptcontrolplane.v1beta2.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &CAPTControlPlane{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *CAPTControlPlane) ValidateCreate() (admission.Warnings, error) {
	captcontrolplanelog.Info("validate create", "name", r.Name)
	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *CAPTControlPlane) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	captcontrolplanelog.Info("validate update", "name", r.Name)
	oldCP, ok := old.(*CAPTControlPlane)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected a CAPTControlPlane but got a %T", old))
	}

	cluster, err := r.getOwnerCluster(context.Background())
	if err != nil {
		// Could not get owner cluster, might be in the process of deletion.
		// Allow the update to proceed.
		return nil, nil
	}

	if cluster != nil && cluster.Spec.Topology != nil {
		if !equality.Semantic.DeepEqual(r.Spec, oldCP.Spec) {
			gr := schema.GroupResource{Group: r.GroupVersionKind().Group, Resource: "captcontrolplanes"}
			return nil, apierrors.NewForbidden(gr, r.Name, fmt.Errorf("CAPTControlPlane.spec is immutable when managed by a ClusterTopology"))
		}
	}

	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *CAPTControlPlane) ValidateDelete() (admission.Warnings, error) {
	captcontrolplanelog.Info("validate delete", "name", r.Name)
	return nil, nil
}

func (r *CAPTControlPlane) getOwnerCluster(ctx context.Context) (*clusterv1.Cluster, error) {
	for _, owner := range r.OwnerReferences {
		if owner.APIVersion == clusterv1.GroupVersion.String() && owner.Kind == "Cluster" {
			return getClusterByName(ctx, r.Namespace, owner.Name)
		}
	}
	return nil, nil // No owner cluster found
}

func getClusterByName(ctx context.Context, namespace, name string) (*clusterv1.Cluster, error) {
	cluster := &clusterv1.Cluster{}
	key := types.NamespacedName{Namespace: namespace, Name: name}
	if err := webhookClient.Get(ctx, key, cluster); err != nil {
		return nil, err
	}
	return cluster, nil
}
