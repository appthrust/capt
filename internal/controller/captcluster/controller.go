package captcluster

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	infrastructurev1beta1 "github.com/appthrust/capt/api/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

const (
	// requeueInterval is the interval to requeue when waiting for resources
	requeueInterval = 10 * time.Second

	// WaitingForClusterCondition represents the condition type for waiting for parent cluster
	WaitingForClusterCondition string = "WaitingForCluster"

	// CAPTClusterFinalizer is the finalizer added to CAPTCluster instances
	CAPTClusterFinalizer = "infrastructure.cluster.x-k8s.io/captcluster"

	// ClusterNameLabel is the label used to identify the cluster name
	ClusterNameLabel = "cluster.x-k8s.io/cluster-name"
)

// Reconciler reconciles a CAPTCluster object
type Reconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=captclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=captclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=captclusters/finalizers,verbs=update
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=workspacetemplates,verbs=get;list;watch
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=workspacetemplateapplies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters,verbs=get;list;watch;update;patch

// getOwnerCluster returns the owner Cluster for a CAPTCluster
func (r *Reconciler) getOwnerCluster(ctx context.Context, captCluster *infrastructurev1beta1.CAPTCluster) (*clusterv1.Cluster, error) {
	logger := log.FromContext(ctx)

	// OwnerReferencesから親Clusterの参照を取得
	for _, ref := range captCluster.OwnerReferences {
		if ref.APIVersion == clusterv1.GroupVersion.String() && ref.Kind == "Cluster" {
			// 親Clusterを取得
			cluster := &clusterv1.Cluster{}
			key := types.NamespacedName{
				Namespace: captCluster.Namespace,
				Name:      ref.Name,
			}
			if err := r.Get(ctx, key, cluster); err != nil {
				logger.Error(err, "Failed to get owner Cluster", "name", ref.Name)
				return nil, err
			}
			return cluster, nil
		}
	}
	return nil, fmt.Errorf("no owner cluster found")
}

// ensureClusterLabels ensures that the required Cluster API labels are set
func (r *Reconciler) ensureClusterLabels(ctx context.Context, captCluster *infrastructurev1beta1.CAPTCluster, cluster *clusterv1.Cluster) error {
	needsUpdate := false

	// Ensure labels map exists
	if captCluster.Labels == nil {
		captCluster.Labels = make(map[string]string)
		needsUpdate = true
	}

	// Set cluster name label if not present or incorrect
	if captCluster.Labels[ClusterNameLabel] != cluster.Name {
		captCluster.Labels[ClusterNameLabel] = cluster.Name
		needsUpdate = true
	}

	if needsUpdate {
		return r.Update(ctx, captCluster)
	}

	return nil
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling CAPTCluster")

	// Fetch the CAPTCluster instance
	captCluster := &infrastructurev1beta1.CAPTCluster{}
	if err := r.Get(ctx, req.NamespacedName, captCluster); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("CAPTCluster resource not found. Ignoring since object must be deleted.")
			return Result{}, nil
		}
		logger.Error(err, "Failed to get CAPTCluster")
		return Result{}, err
	}

	// Handle deletion
	if !captCluster.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, captCluster)
	}

	// Add finalizer if it doesn't exist
	if !controllerutil.ContainsFinalizer(captCluster, CAPTClusterFinalizer) {
		controllerutil.AddFinalizer(captCluster, CAPTClusterFinalizer)
		if err := r.Update(ctx, captCluster); err != nil {
			logger.Error(err, "Failed to add finalizer")
			return Result{}, err
		}
		return Result{Requeue: true}, nil
	}

	// Get owner Cluster using OwnerReferences
	cluster, err := r.getOwnerCluster(ctx, captCluster)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return r.handleMissingCluster(ctx, captCluster)
		}
		logger.Error(err, "Failed to get owner Cluster")
		return Result{}, err
	}

	// Ensure required labels are set
	if err := r.ensureClusterLabels(ctx, captCluster, cluster); err != nil {
		logger.Error(err, "Failed to ensure cluster labels")
		return Result{}, err
	}

	// Clear WaitingForCluster condition if it exists
	meta.RemoveStatusCondition(&captCluster.Status.Conditions, WaitingForClusterCondition)

	// Validate VPC configuration
	if err := captCluster.Spec.ValidateVPCConfiguration(); err != nil {
		logger.Error(err, "Invalid VPC configuration")
		return r.setFailedStatus(ctx, captCluster, cluster, "InvalidVPCConfig", err.Error())
	}

	// Handle VPC configuration
	result, err := r.reconcileVPC(ctx, captCluster, cluster)
	if err != nil {
		logger.Error(err, "Failed to reconcile VPC")
		return r.setFailedStatus(ctx, captCluster, cluster, "VPCReconciliationFailed", err.Error())
	}

	logger.Info("Successfully reconciled CAPTCluster", "name", captCluster.Name)
	return result, nil
}

// handleMissingCluster handles the case where the parent Cluster does not exist
func (r *Reconciler) handleMissingCluster(ctx context.Context, captCluster *infrastructurev1beta1.CAPTCluster) (Result, error) {
	logger := log.FromContext(ctx)

	// Set WaitingForCluster condition
	meta.SetStatusCondition(&captCluster.Status.Conditions, metav1.Condition{
		Type:               WaitingForClusterCondition,
		Status:             metav1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             "ClusterNotFound",
		Message:            "Waiting for owner Cluster to be created",
	})
	captCluster.Status.Ready = false

	// Clean up any existing WorkspaceTemplateApply
	if err := r.cleanupWorkspaceTemplateApply(ctx, captCluster); err != nil {
		return Result{}, err
	}

	if err := r.Status().Update(ctx, captCluster); err != nil {
		logger.Error(err, "Failed to update CAPTCluster status")
		return Result{}, err
	}

	logger.Info("Waiting for owner Cluster to be created")
	return Result{RequeueAfter: requeueInterval}, nil
}

// cleanupWorkspaceTemplateApply removes any existing WorkspaceTemplateApply and clears the reference
func (r *Reconciler) cleanupWorkspaceTemplateApply(ctx context.Context, captCluster *infrastructurev1beta1.CAPTCluster) error {
	logger := log.FromContext(ctx)

	if captCluster.Spec.WorkspaceTemplateApplyName == "" {
		return nil
	}

	workspaceApply := &infrastructurev1beta1.WorkspaceTemplateApply{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      captCluster.Spec.WorkspaceTemplateApplyName,
		Namespace: captCluster.Namespace,
	}, workspaceApply)

	if err == nil {
		// WorkspaceTemplateApply exists, delete it
		if err := r.Delete(ctx, workspaceApply); err != nil {
			logger.Error(err, "Failed to delete WorkspaceTemplateApply while waiting for parent Cluster")
			return err
		}
		logger.Info("Deleted WorkspaceTemplateApply while waiting for parent Cluster")
	} else if !apierrors.IsNotFound(err) {
		return err
	}

	// Clear the reference
	captCluster.Spec.WorkspaceTemplateApplyName = ""
	if err := r.Update(ctx, captCluster); err != nil {
		logger.Error(err, "Failed to clear WorkspaceTemplateApplyName")
		return err
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrastructurev1beta1.CAPTCluster{}).
		Owns(&infrastructurev1beta1.WorkspaceTemplateApply{}).
		// Watch Cluster deletions and map them to the corresponding CAPTCluster
		Watches(
			&clusterv1.Cluster{},
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, o client.Object) []reconcile.Request {
				cluster := o.(*clusterv1.Cluster)
				return []reconcile.Request{
					{
						NamespacedName: types.NamespacedName{
							Name:      cluster.Name,
							Namespace: cluster.Namespace,
						},
					},
				}
			}),
		).
		Complete(r)
}
