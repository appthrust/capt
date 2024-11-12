package controlplane

import (
	"context"

	controlplanev1beta1 "github.com/appthrust/capt/api/controlplane/v1beta1"
	infrastructurev1beta1 "github.com/appthrust/capt/api/v1beta1"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Helper functions

// isWorkspaceReady checks if the workspace is ready
func isWorkspaceReady(workspaceApply *infrastructurev1beta1.WorkspaceTemplateApply) bool {
	if workspaceApply == nil {
		return false
	}

	for _, condition := range workspaceApply.Status.Conditions {
		if condition.Type == xpv1.TypeReady && condition.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

// getWorkspaceError gets the error message from the workspace if any
func getWorkspaceError(workspaceApply *infrastructurev1beta1.WorkspaceTemplateApply) string {
	if workspaceApply == nil {
		return ""
	}

	for _, condition := range workspaceApply.Status.Conditions {
		if condition.Type == xpv1.TypeReady && condition.Status == corev1.ConditionFalse {
			return condition.Message
		}
	}
	return ""
}

// updateStatus updates the status of the CAPTControlPlane and its owner Cluster
func (r *Reconciler) updateStatus(
	ctx context.Context,
	controlPlane *controlplanev1beta1.CAPTControlPlane,
	workspaceApply *infrastructurev1beta1.WorkspaceTemplateApply,
	cluster *clusterv1.Cluster,
) (ctrl.Result, error) {
	// Initialize WorkspaceTemplateStatus if not exists
	if controlPlane.Status.WorkspaceTemplateStatus == nil {
		controlPlane.Status.WorkspaceTemplateStatus = &controlplanev1beta1.WorkspaceTemplateStatus{}
	}

	// Update status based on workspace conditions
	ready := isWorkspaceReady(workspaceApply)
	errorMessage := getWorkspaceError(workspaceApply)

	if !ready {
		return r.handleNotReadyStatus(ctx, controlPlane, cluster, errorMessage)
	}

	return r.handleReadyStatus(ctx, controlPlane, cluster, workspaceApply)
}

// handleNotReadyStatus updates the status for a not-ready control plane
func (r *Reconciler) handleNotReadyStatus(
	ctx context.Context,
	controlPlane *controlplanev1beta1.CAPTControlPlane,
	cluster *clusterv1.Cluster,
	errorMessage string,
) (ctrl.Result, error) {
	if errorMessage != "" {
		meta.SetStatusCondition(&controlPlane.Status.Conditions, metav1.Condition{
			Type:               controlplanev1beta1.ControlPlaneReadyCondition,
			Status:             metav1.ConditionFalse,
			LastTransitionTime: metav1.Now(),
			Reason:             controlplanev1beta1.ReasonWorkspaceError,
			Message:            errorMessage,
		})
		controlPlane.Status.Phase = controlplanev1beta1.ControlPlaneFailedCondition
	} else {
		meta.SetStatusCondition(&controlPlane.Status.Conditions, metav1.Condition{
			Type:               controlplanev1beta1.ControlPlaneReadyCondition,
			Status:             metav1.ConditionFalse,
			LastTransitionTime: metav1.Now(),
			Reason:             controlplanev1beta1.ReasonCreating,
			Message:            "Control plane is being created",
		})
		controlPlane.Status.Phase = controlplanev1beta1.ControlPlaneCreatingCondition
	}

	controlPlane.Status.Ready = false
	controlPlane.Status.Initialized = false
	controlPlane.Status.WorkspaceTemplateStatus.Ready = false

	if errorMessage != "" {
		controlPlane.Status.WorkspaceTemplateStatus.LastFailureMessage = errorMessage
	}

	if err := r.Status().Update(ctx, controlPlane); err != nil {
		return ctrl.Result{}, err
	}

	// Update Cluster status if it exists
	if cluster != nil {
		patchBase := cluster.DeepCopy()
		cluster.Status.ControlPlaneReady = false
		if err := r.Status().Patch(ctx, cluster, client.MergeFrom(patchBase)); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{RequeueAfter: requeueInterval}, nil
}

// handleReadyStatus updates the status for a ready control plane
func (r *Reconciler) handleReadyStatus(
	ctx context.Context,
	controlPlane *controlplanev1beta1.CAPTControlPlane,
	cluster *clusterv1.Cluster,
	workspaceApply *infrastructurev1beta1.WorkspaceTemplateApply,
) (ctrl.Result, error) {
	meta.SetStatusCondition(&controlPlane.Status.Conditions, metav1.Condition{
		Type:               controlplanev1beta1.ControlPlaneReadyCondition,
		Status:             metav1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             controlplanev1beta1.ReasonReady,
		Message:            "Control plane is ready",
	})

	controlPlane.Status.Phase = controlplanev1beta1.ControlPlaneReadyCondition
	controlPlane.Status.Ready = true
	controlPlane.Status.Initialized = true
	controlPlane.Status.WorkspaceTemplateStatus.Ready = true
	controlPlane.Status.FailureReason = nil
	controlPlane.Status.FailureMessage = nil
	controlPlane.Status.WorkspaceTemplateStatus.LastFailureMessage = ""

	if workspaceApply.Status.LastAppliedTime != nil {
		controlPlane.Status.WorkspaceTemplateStatus.LastAppliedRevision = workspaceApply.Status.LastAppliedTime.String()
	}

	if err := r.Status().Update(ctx, controlPlane); err != nil {
		return ctrl.Result{}, err
	}

	// Update Cluster status if it exists
	if cluster != nil {
		patchBase := cluster.DeepCopy()
		cluster.Status.ControlPlaneReady = true
		if err := r.Status().Patch(ctx, cluster, client.MergeFrom(patchBase)); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// setFailedStatus sets the status to failed with the given reason and message
func (r *Reconciler) setFailedStatus(
	ctx context.Context,
	controlPlane *controlplanev1beta1.CAPTControlPlane,
	cluster *clusterv1.Cluster,
	reason string,
	message string,
) (ctrl.Result, error) {
	meta.SetStatusCondition(&controlPlane.Status.Conditions, metav1.Condition{
		Type:               controlplanev1beta1.ControlPlaneReadyCondition,
		Status:             metav1.ConditionFalse,
		LastTransitionTime: metav1.Now(),
		Reason:             reason,
		Message:            message,
	})

	controlPlane.Status.Phase = controlplanev1beta1.ControlPlaneFailedCondition
	controlPlane.Status.Ready = false
	controlPlane.Status.Initialized = false
	if controlPlane.Status.WorkspaceTemplateStatus != nil {
		controlPlane.Status.WorkspaceTemplateStatus.Ready = false
		controlPlane.Status.WorkspaceTemplateStatus.LastFailureMessage = message
	}

	if err := r.Status().Update(ctx, controlPlane); err != nil {
		return ctrl.Result{}, err
	}

	// Update Cluster status if it exists
	if cluster != nil {
		patchBase := cluster.DeepCopy()
		cluster.Status.ControlPlaneReady = false
		cluster.Status.FailureMessage = &message
		if err := r.Status().Patch(ctx, cluster, client.MergeFrom(patchBase)); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{RequeueAfter: requeueInterval}, nil
}