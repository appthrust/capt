package controller

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	infrastructurev1beta1 "github.com/appthrust/capt/api/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

// CAPTClusterReconciler reconciles a CAPTCluster object
type CAPTClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=captclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=captclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=captclusters/finalizers,verbs=update
// +kubebuilder:rbac:groups=tf.upbound.io,resources=workspaces,verbs=get;list;watch;create;update;patch;delete

func (r *CAPTClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Fetch the CAPTCluster instance
	captCluster := &infrastructurev1beta1.CAPTCluster{}
	if err := r.Get(ctx, req.NamespacedName, captCluster); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("CAPTCluster resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get CAPTCluster")
		return ctrl.Result{}, err
	}

	// Handle finalizer
	if err := handleFinalizer(ctx, r.Client, captCluster); err != nil {
		return ctrl.Result{}, err
	}

	// Handle Terraform Workspace
	workspaceName := types.NamespacedName{
		Name:      fmt.Sprintf("%s-workspace", captCluster.Name),
		Namespace: captCluster.Namespace,
	}
	if err := reconcileTerraformWorkspace(ctx, r.Client, captCluster, workspaceName); err != nil {
		return ctrl.Result{}, err
	}

	// Update CAPTCluster status
	captCluster.Status.WorkspaceName = workspaceName.Name
	if err := r.Status().Update(ctx, captCluster); err != nil {
		log.Error(err, "Failed to update CAPTCluster status")
		return ctrl.Result{}, err
	}

	log.Info("Successfully reconciled CAPTCluster", "name", captCluster.Name)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CAPTClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrastructurev1beta1.CAPTCluster{}).
		Complete(r)
}
