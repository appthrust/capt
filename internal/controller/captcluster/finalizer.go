package captcluster

import (
	"context"

	infrastructurev1beta1 "github.com/appthrust/capt/api/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// reconcileDelete handles the deletion of a CAPTCluster
func (r *Reconciler) reconcileDelete(ctx context.Context, captCluster *infrastructurev1beta1.CAPTCluster) (Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Handling deletion of CAPTCluster")

	// If finalizer is not present, return immediately
	if !controllerutil.ContainsFinalizer(captCluster, CAPTClusterFinalizer) {
		return Result{}, nil
	}

	// Find and delete associated WorkspaceTemplateApply if it exists
	if captCluster.Spec.WorkspaceTemplateApplyName != "" {
		workspaceApply := &infrastructurev1beta1.WorkspaceTemplateApply{}
		err := r.Get(ctx, types.NamespacedName{
			Name:      captCluster.Spec.WorkspaceTemplateApplyName,
			Namespace: captCluster.Namespace,
		}, workspaceApply)

		if err == nil {
			// Check if we should retain the VPC
			if !captCluster.Spec.RetainVPCOnDelete || captCluster.Spec.VPCTemplateRef == nil {
				// Delete WorkspaceTemplateApply if it exists and should not be retained
				if err := r.Delete(ctx, workspaceApply); err != nil {
					logger.Error(err, "Failed to delete WorkspaceTemplateApply")
					return Result{}, err
				}
				logger.Info("Successfully requested deletion of WorkspaceTemplateApply")

				// Check if WorkspaceTemplateApply still exists
				err = r.Get(ctx, types.NamespacedName{
					Name:      captCluster.Spec.WorkspaceTemplateApplyName,
					Namespace: captCluster.Namespace,
				}, workspaceApply)
				if err == nil {
					// WorkspaceTemplateApply still exists, requeue
					logger.Info("Waiting for WorkspaceTemplateApply to be deleted")
					return Result{RequeueAfter: requeueInterval}, nil
				} else if !apierrors.IsNotFound(err) {
					// Error other than NotFound occurred
					logger.Error(err, "Failed to get WorkspaceTemplateApply")
					return Result{}, err
				}
			} else {
				logger.Info("RetainVPCOnDelete is true, skipping WorkspaceTemplateApply deletion",
					"vpcId", captCluster.Status.VPCID,
					"workspaceTemplateApplyName", captCluster.Spec.WorkspaceTemplateApplyName)
			}
		} else if !apierrors.IsNotFound(err) {
			// Error other than NotFound occurred
			logger.Error(err, "Failed to get WorkspaceTemplateApply")
			return Result{}, err
		}
	}

	// Remove finalizer
	controllerutil.RemoveFinalizer(captCluster, CAPTClusterFinalizer)
	if err := r.Update(ctx, captCluster); err != nil {
		logger.Error(err, "Failed to remove finalizer")
		return Result{}, err
	}

	logger.Info("Successfully cleaned up CAPTCluster")
	return Result{}, nil
}
