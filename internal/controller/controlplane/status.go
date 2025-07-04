package controlplane

import (
	"context"
	"fmt"

	controlplanev1beta1 "github.com/appthrust/capt/api/controlplane/v1beta1"
	infrastructurev1beta1 "github.com/appthrust/capt/api/v1beta1"
	"github.com/appthrust/capt/internal/controller/controlplane/endpoint"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/errors"
	"sigs.k8s.io/cluster-api/util/conditions"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	// ReasonEndpointUpdateFailed indicates that the endpoint update failed
	ReasonEndpointUpdateFailed = "EndpointUpdateFailed"
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

// initializeStatus initializes the status fields if they are nil
func initializeStatus(controlPlane *controlplanev1beta1.CAPTControlPlane) {
	if controlPlane.Status.WorkspaceTemplateStatus == nil {
		controlPlane.Status.WorkspaceTemplateStatus = &controlplanev1beta1.WorkspaceTemplateStatus{}
	}
	if controlPlane.Status.WorkspaceStatus == nil {
		controlPlane.Status.WorkspaceStatus = &controlplanev1beta1.WorkspaceStatus{}
	}
	if controlPlane.Status.Conditions == nil {
		controlPlane.Status.Conditions = []metav1.Condition{}
	}
}

// updateStatus updates the status of the CAPTControlPlane and its owner Cluster
func (r *Reconciler) updateStatus(
	ctx context.Context,
	controlPlane *controlplanev1beta1.CAPTControlPlane,
	workspaceApply *infrastructurev1beta1.WorkspaceTemplateApply,
	cluster *clusterv1.Cluster,
) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Initialize status fields
	initializeStatus(controlPlane)

	// Create a patch base before any updates
	patchBase := controlPlane.DeepCopy()

	// Update Workspace status
	if err := r.updateWorkspaceStatus(ctx, controlPlane, workspaceApply); err != nil {
		return r.setFailedStatus(ctx, controlPlane, cluster, "WorkspaceStatusUpdateFailed", fmt.Sprintf("Failed to update workspace status: %v", err))
	}

	// Log the current status
	logger.Info("Status after workspace update",
		"workspaceStatus", controlPlane.Status.WorkspaceStatus,
		"workspaceTemplateStatus", controlPlane.Status.WorkspaceTemplateStatus)

	// Update status based on workspace conditions
	ready := isWorkspaceReady(workspaceApply)
	errorMessage := getWorkspaceError(workspaceApply)

	if !ready {
		return r.handleNotReadyStatus(ctx, controlPlane, cluster, errorMessage)
	}

	// Update endpoint from workspace first
	if workspaceApply.Status.WorkspaceName != "" {
		if apiEndpoint, err := endpoint.GetEndpointFromWorkspace(ctx, r.Client, workspaceApply.Status.WorkspaceName); err != nil {
			errMsg := fmt.Sprintf("Failed to get endpoint from workspace: %v", err)
			return r.setFailedStatus(ctx, controlPlane, cluster, ReasonEndpointUpdateFailed, errMsg)
		} else if apiEndpoint != nil {
			logger.Info("Updating control plane endpoint", "endpoint", apiEndpoint)

			// Update CAPTControlPlane endpoint
			controlPlane.Spec.ControlPlaneEndpoint = *apiEndpoint
			if err := r.Update(ctx, controlPlane); err != nil {
				errMsg := fmt.Sprintf("Failed to update control plane endpoint: %v", err)
				return r.setFailedStatus(ctx, controlPlane, cluster, ReasonEndpointUpdateFailed, errMsg)
			}

			// Update parent Cluster endpoint
			if cluster != nil {
				patchBase := cluster.DeepCopy()
				cluster.Spec.ControlPlaneEndpoint = *apiEndpoint
				if err := r.Patch(ctx, cluster, client.MergeFrom(patchBase)); err != nil {
					errMsg := fmt.Sprintf("Failed to update cluster endpoint: %v", err)
					return r.setFailedStatus(ctx, controlPlane, cluster, ReasonEndpointUpdateFailed, errMsg)
				}
				logger.Info("Updated parent cluster endpoint", "endpoint", apiEndpoint)
			}

			// Re-initialize status fields after endpoint update
			initializeStatus(controlPlane)
		}
	}

	// Only proceed with ready status if endpoint update was successful
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

	// Log the status before update
	logger.Info("Status before final update",
		"workspaceStatus", controlPlane.Status.WorkspaceStatus,
		"workspaceTemplateStatus", controlPlane.Status.WorkspaceTemplateStatus)

	// Update status
	if err := r.Status().Patch(ctx, controlPlane, client.MergeFrom(patchBase)); err != nil {
		return ctrl.Result{}, err
	}

	// Update Cluster status if it exists
	if cluster != nil {
		patchBase := cluster.DeepCopy()
		cluster.Status.ControlPlaneReady = true

		// Set ControlPlaneInitialized condition when all required conditions are met
		if controlPlane.Status.Ready && controlPlane.Status.Initialized && controlPlane.Status.SecretsReady {
			conditions.MarkTrue(cluster, clusterv1.ControlPlaneInitializedCondition)
		}

		if err := r.Status().Patch(ctx, cluster, client.MergeFrom(patchBase)); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Use default interval for ready state
	return ctrl.Result{RequeueAfter: defaultRequeueInterval}, nil
}

// handleNotReadyStatus updates the status for a not-ready control plane
func (r *Reconciler) handleNotReadyStatus(
	ctx context.Context,
	controlPlane *controlplanev1beta1.CAPTControlPlane,
	cluster *clusterv1.Cluster,
	errorMessage string,
) (ctrl.Result, error) {
	// Initialize status fields
	initializeStatus(controlPlane)

	// Create a patch base before any updates
	patchBase := controlPlane.DeepCopy()

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

	// Update status
	if err := r.Status().Patch(ctx, controlPlane, client.MergeFrom(patchBase)); err != nil {
		return ctrl.Result{}, err
	}

	// Update Cluster status if it exists
	if cluster != nil {
		patchBase := cluster.DeepCopy()
		cluster.Status.ControlPlaneReady = false

		// Set ControlPlaneInitialized condition to false when control plane is not ready
		conditions.MarkFalse(cluster, clusterv1.ControlPlaneInitializedCondition,
			"ControlPlaneNotReady", clusterv1.ConditionSeverityInfo,
			"Control plane is not ready")

		if err := r.Status().Patch(ctx, cluster, client.MergeFrom(patchBase)); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Use initialization interval for not-ready state
	return ctrl.Result{RequeueAfter: initializationRequeueInterval}, nil
}

// setFailedStatus sets the status to failed with the given reason and message
func (r *Reconciler) setFailedStatus(
	ctx context.Context,
	controlPlane *controlplanev1beta1.CAPTControlPlane,
	cluster *clusterv1.Cluster,
	reason string,
	message string,
) (ctrl.Result, error) {
	// Initialize status fields
	initializeStatus(controlPlane)

	// Create a patch base before any updates
	patchBase := controlPlane.DeepCopy()

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
	controlPlane.Status.WorkspaceTemplateStatus.Ready = false
	controlPlane.Status.WorkspaceTemplateStatus.LastFailureMessage = message

	// Set failure reason and message
	failureReason := reason
	controlPlane.Status.FailureReason = &failureReason
	controlPlane.Status.FailureMessage = &message

	// Update status
	if err := r.Status().Patch(ctx, controlPlane, client.MergeFrom(patchBase)); err != nil {
		return ctrl.Result{}, err
	}

	// Update Cluster status if it exists
	if cluster != nil {
		patchBase := cluster.DeepCopy()
		cluster.Status.ControlPlaneReady = false
		cluster.Status.FailureMessage = &message
		clusterStatusError := errors.ClusterStatusError(reason)
		cluster.Status.FailureReason = &clusterStatusError

		// Set ControlPlaneInitialized condition to false when control plane fails
		conditions.MarkFalse(cluster, clusterv1.ControlPlaneInitializedCondition,
			reason, clusterv1.ConditionSeverityError,
			"%s", message)

		if err := r.Status().Patch(ctx, cluster, client.MergeFrom(patchBase)); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Use error interval for failed state
	return ctrl.Result{RequeueAfter: errorRequeueInterval}, nil
}
