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
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var captclusterlog = logf.Log.WithName("captcluster-resource")
var webhookClient client.Client

func (r *CAPTCluster) SetupWebhookWithManager(mgr ctrl.Manager) error {
	webhookClient = mgr.GetClient()
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-infrastructure-cluster-x-k8s-io-v1beta1-captcluster,mutating=true,failurePolicy=fail,sideEffects=None,groups=infrastructure.cluster.x-k8s.io,resources=captclusters,verbs=create;update,versions=v1beta1,name=mcaptcluster.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &CAPTCluster{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *CAPTCluster) Default() {
	captclusterlog.Info("default", "name", r.Name)
}

// +kubebuilder:webhook:path=/validate-infrastructure-cluster-x-k8s-io-v1beta1-captcluster,mutating=false,failurePolicy=fail,sideEffects=None,groups=infrastructure.cluster.x-k8s.io,resources=captclusters,verbs=create;update,versions=v1beta1,name=vcaptcluster.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &CAPTCluster{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *CAPTCluster) ValidateCreate() (admission.Warnings, error) {
	captclusterlog.Info("validate create", "name", r.Name)
	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *CAPTCluster) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	captclusterlog.Info("validate update", "name", r.Name)
	oldCluster, ok := old.(*CAPTCluster)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected a CAPTCluster but got a %T", old))
	}

	cluster, err := r.getOwnerCluster(context.Background())
	if err != nil {
		// Could not get owner cluster, might be in the process of deletion.
		// Allow the update to proceed.
		return nil, nil
	}

	if cluster != nil && cluster.Spec.Topology != nil {
		if !equality.Semantic.DeepEqual(r.Spec, oldCluster.Spec) {
			gr := schema.GroupResource{Group: r.GroupVersionKind().Group, Resource: "captclusters"}
			return nil, apierrors.NewForbidden(gr, r.Name, fmt.Errorf("CAPTCluster.spec is immutable when managed by a ClusterTopology"))
		}
	}

	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *CAPTCluster) ValidateDelete() (admission.Warnings, error) {
	captclusterlog.Info("validate delete", "name", r.Name)
	return nil, nil
}

func (r *CAPTCluster) getOwnerCluster(ctx context.Context) (*clusterv1.Cluster, error) {
	// Prefer explicit owner reference if present
	for _, owner := range r.OwnerReferences {
		if owner.APIVersion == clusterv1.GroupVersion.String() && owner.Kind == "Cluster" {
			return getClusterByName(ctx, r.Namespace, owner.Name)
		}
	}

	// Fallback: cluster-name label
	if name, ok := r.Labels[clusterv1.ClusterNameLabel]; ok && name != "" {
		if c, err := getClusterByName(ctx, r.Namespace, name); err == nil {
			return c, nil
		}
	}

	// Last resort: same-name Cluster in the namespace
	if r.Name != "" {
		if c, err := getClusterByName(ctx, r.Namespace, r.Name); err == nil {
			return c, nil
		}
	}

	return nil, nil
}

func getClusterByName(ctx context.Context, namespace, name string) (*clusterv1.Cluster, error) {
	cluster := &clusterv1.Cluster{}
	key := types.NamespacedName{Namespace: namespace, Name: name}
	if err := webhookClient.Get(ctx, key, cluster); err != nil {
		return nil, err
	}
	return cluster, nil
}
