package v1beta1

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
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var captmachinelog = logf.Log.WithName("captmachine-resource")

func (r *CaptMachine) SetupWebhookWithManager(mgr ctrl.Manager) error {
	if webhookClient == nil {
		webhookClient = mgr.GetClient()
	}
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-infrastructure-cluster-x-k8s-io-v1beta1-captmachine,mutating=true,failurePolicy=fail,sideEffects=None,groups=infrastructure.cluster.x-k8s.io,resources=captmachines,verbs=create;update,versions=v1beta1,name=mcaptmachine.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &CaptMachine{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *CaptMachine) Default() {
	captmachinelog.Info("default", "name", r.Name)
}

// +kubebuilder:webhook:path=/validate-infrastructure-cluster-x-k8s-io-v1beta1-captmachine,mutating=false,failurePolicy=fail,sideEffects=None,groups=infrastructure.cluster.x-k8s.io,resources=captmachines,verbs=create;update,versions=v1beta1,name=vcaptmachine.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &CaptMachine{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *CaptMachine) ValidateCreate() (admission.Warnings, error) {
	captmachinelog.Info("validate create", "name", r.Name)
	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *CaptMachine) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	captmachinelog.Info("validate update", "name", r.Name)
	oldMachine, ok := old.(*CaptMachine)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected a CaptMachine but got a %T", old))
	}

	cluster, err := getOwnerClusterForMachine(context.Background(), r)
	if err != nil {
		return nil, nil // Could not get owner cluster, allowing update.
	}

	if cluster != nil && cluster.Spec.Topology != nil {
		if !equality.Semantic.DeepEqual(r.Spec, oldMachine.Spec) {
			gr := schema.GroupResource{Group: r.GroupVersionKind().Group, Resource: "captmachines"}
			return nil, apierrors.NewForbidden(gr, r.Name, fmt.Errorf("CaptMachine.spec is immutable when managed by a ClusterTopology"))
		}
	}

	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *CaptMachine) ValidateDelete() (admission.Warnings, error) {
	captmachinelog.Info("validate delete", "name", r.Name)
	return nil, nil
}

func getOwnerClusterForMachine(ctx context.Context, r *CaptMachine) (*clusterv1.Cluster, error) {
	// CaptMachine is owned by a Machine.
	// Machine is owned by a MachineSet.
	// MachineSet is owned by a Cluster.
	// We need to traverse up the ownership chain.

	// First, find the owning Machine.
	var owningMachine *clusterv1.Machine
	for _, owner := range r.OwnerReferences {
		if owner.APIVersion == clusterv1.GroupVersion.String() && owner.Kind == "Machine" {
			machine := &clusterv1.Machine{}
			key := types.NamespacedName{Namespace: r.Namespace, Name: owner.Name}
			if err := webhookClient.Get(ctx, key, machine); err != nil {
				return nil, err
			}
			owningMachine = machine
			break
		}
	}

	if owningMachine == nil {
		return nil, nil // No owner Machine found
	}

	// Now get the cluster by (1) label, (2) owner references chain fallback
	if clusterName, ok := owningMachine.Labels[clusterv1.ClusterNameLabel]; ok && clusterName != "" {
		if c, err := getClusterByName(ctx, r.Namespace, clusterName); err == nil {
			return c, nil
		}
	}

	// Fallback: try Cluster with same name as owningMachine.Spec.ClusterName (string)
	if owningMachine.Spec.ClusterName != "" {
		if c, err := getClusterByName(ctx, r.Namespace, owningMachine.Spec.ClusterName); err == nil {
			return c, nil
		}
	}

	// Last resort: no cluster found
	return nil, nil
}
