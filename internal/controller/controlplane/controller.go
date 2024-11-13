package controlplane

import (
	"context"
	"fmt"

	controlplanev1beta1 "github.com/appthrust/capt/api/controlplane/v1beta1"
	infrastructurev1beta1 "github.com/appthrust/capt/api/v1beta1"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	// CAPTControlPlaneFinalizer is the finalizer added to CAPTControlPlane instances
	CAPTControlPlaneFinalizer = "controlplane.cluster.x-k8s.io/captcontrolplane"
)

// Reconcile handles CAPTControlPlane events
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling CAPTControlPlane")

	// Get CAPTControlPlane
	controlPlane := &controlplanev1beta1.CAPTControlPlane{}
	if err := r.Get(ctx, req.NamespacedName, controlPlane); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Get owner Cluster
	cluster := &clusterv1.Cluster{}
	if err := r.Get(ctx, types.NamespacedName{
		Name:      controlPlane.Name,
		Namespace: controlPlane.Namespace,
	}, cluster); err != nil {
		if !apierrors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
		cluster = nil
	}

	// Handle deletion
	if !controlPlane.DeletionTimestamp.IsZero() {
		if controllerutil.ContainsFinalizer(controlPlane, CAPTControlPlaneFinalizer) {
			// Clean up associated resources
			if err := r.cleanupResources(ctx, controlPlane); err != nil {
				logger.Error(err, "Failed to cleanup resources")
				return ctrl.Result{}, err
			}

			// Remove finalizer
			controllerutil.RemoveFinalizer(controlPlane, CAPTControlPlaneFinalizer)

			// Update the object to remove the finalizer
			if err := r.Update(ctx, controlPlane); err != nil {
				if !apierrors.IsNotFound(err) {
					return ctrl.Result{}, err
				}
			}

			logger.Info("Successfully removed finalizer")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, nil
	}

	// Add finalizer if it doesn't exist
	if !controllerutil.ContainsFinalizer(controlPlane, CAPTControlPlaneFinalizer) {
		controllerutil.AddFinalizer(controlPlane, CAPTControlPlaneFinalizer)
		if err := r.Update(ctx, controlPlane); err != nil {
			return ctrl.Result{}, err
		}

		// Fetch the updated object
		if err := r.Get(ctx, req.NamespacedName, controlPlane); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Handle missing cluster case
	if cluster == nil {
		logger.Info("Owner cluster not found")
		meta.SetStatusCondition(&controlPlane.Status.Conditions, metav1.Condition{
			Type:               controlplanev1beta1.ControlPlaneReadyCondition,
			Status:             metav1.ConditionFalse,
			Reason:             controlplanev1beta1.ReasonCreating,
			Message:            "Waiting for owner cluster",
			LastTransitionTime: metav1.Now(),
		})
		if err := r.Status().Update(ctx, controlPlane); err != nil {
			return ctrl.Result{}, err
		}

		// Fetch the updated object
		if err := r.Get(ctx, req.NamespacedName, controlPlane); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{RequeueAfter: initializationRequeueInterval}, nil
	}

	// Set owner reference if cluster exists
	if err := r.setOwnerReference(ctx, controlPlane, cluster); err != nil {
		return ctrl.Result{}, err
	}

	// Fetch the updated object after setting owner reference
	if err := r.Get(ctx, req.NamespacedName, controlPlane); err != nil {
		return ctrl.Result{}, err
	}

	// Get WorkspaceTemplate
	workspaceTemplate := &infrastructurev1beta1.WorkspaceTemplate{}
	if err := r.Get(ctx, types.NamespacedName{
		Name:      controlPlane.Spec.WorkspaceTemplateRef.Name,
		Namespace: controlPlane.Namespace,
	}, workspaceTemplate); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Error(err, "Failed to get WorkspaceTemplate")
			result, setErr := r.setFailedStatus(ctx, controlPlane, cluster, "WorkspaceTemplateNotFound", fmt.Sprintf("Failed to get WorkspaceTemplate: %v", err))
			if setErr != nil {
				return ctrl.Result{}, fmt.Errorf("failed to set status: %v (original error: %v)", setErr, err)
			}
			return result, err
		}
		return ctrl.Result{}, err
	}

	// Get or create WorkspaceTemplateApply
	workspaceApply, err := r.getOrCreateWorkspaceTemplateApply(ctx, controlPlane, workspaceTemplate)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Set the WorkspaceTemplateApplyName if it's not set
	if controlPlane.Spec.WorkspaceTemplateApplyName == "" {
		controlPlane.Spec.WorkspaceTemplateApplyName = workspaceApply.Name
		if err := r.Update(ctx, controlPlane); err != nil {
			return ctrl.Result{}, err
		}

		// Fetch the updated object
		if err := r.Get(ctx, req.NamespacedName, controlPlane); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Update status based on WorkspaceTemplateApply conditions
	result, err := r.updateStatus(ctx, controlPlane, workspaceApply, cluster)
	if err != nil {
		return result, err
	}

	// Reconcile secrets after status update
	logger.Info("Reconciling secrets")
	if err := r.reconcileSecrets(ctx, controlPlane, cluster, workspaceApply); err != nil {
		logger.Error(err, "Failed to reconcile secrets")
		if _, setErr := r.setFailedStatus(ctx, controlPlane, cluster, "SecretReconciliationFailed", fmt.Sprintf("Failed to reconcile secrets: %v", err)); setErr != nil {
			return ctrl.Result{}, fmt.Errorf("failed to set status: %v (original error: %v)", setErr, err)
		}
		return ctrl.Result{RequeueAfter: errorRequeueInterval}, err
	}

	// Fetch the final updated object
	if err := r.Get(ctx, req.NamespacedName, controlPlane); err != nil {
		return ctrl.Result{}, err
	}

	// Return the result from updateStatus to maintain the requeue interval
	return result, nil
}

// setOwnerReference sets the owner reference to the parent Cluster
func (r *Reconciler) setOwnerReference(ctx context.Context, controlPlane *controlplanev1beta1.CAPTControlPlane, cluster *clusterv1.Cluster) error {
	if cluster == nil {
		return nil
	}

	// Check if owner reference already exists
	for _, ref := range controlPlane.OwnerReferences {
		if ref.Kind == "Cluster" && ref.APIVersion == clusterv1.GroupVersion.String() {
			return nil
		}
	}

	// Set owner reference
	if err := controllerutil.SetControllerReference(cluster, controlPlane, r.Scheme); err != nil {
		return errors.Wrap(err, "failed to set owner reference")
	}

	return r.Update(ctx, controlPlane)
}

// cleanupResources cleans up all resources associated with the CAPTControlPlane
func (r *Reconciler) cleanupResources(ctx context.Context, controlPlane *controlplanev1beta1.CAPTControlPlane) error {
	logger := log.FromContext(ctx)

	// Find and delete associated WorkspaceTemplateApply
	var applyName string
	if controlPlane.Spec.WorkspaceTemplateApplyName != "" {
		applyName = controlPlane.Spec.WorkspaceTemplateApplyName
	} else {
		applyName = fmt.Sprintf("%s-eks-controlplane-apply", controlPlane.Name)
	}

	workspaceApply := &infrastructurev1beta1.WorkspaceTemplateApply{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      applyName,
		Namespace: controlPlane.Namespace,
	}, workspaceApply)

	if err == nil {
		// WorkspaceTemplateApply exists, delete it
		if err := r.Delete(ctx, workspaceApply); err != nil {
			logger.Error(err, "Failed to delete WorkspaceTemplateApply")
			return fmt.Errorf("failed to delete WorkspaceTemplateApply: %v", err)
		}
		logger.Info("Successfully deleted WorkspaceTemplateApply")
	} else if !apierrors.IsNotFound(err) {
		return fmt.Errorf("failed to get WorkspaceTemplateApply: %v", err)
	}

	return nil
}
