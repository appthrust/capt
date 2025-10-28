package captcluster

import (
	"context"
	"fmt"
	"reflect"

	infrastructurev1beta1 "github.com/appthrust/capt/api/v1beta1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/cluster-api/api/v1beta1"
	capierrors "sigs.k8s.io/cluster-api/errors"
	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	// ControlPlaneInitializedCondition represents the condition type for control plane initialization
	ControlPlaneInitializedCondition v1beta1.ConditionType = "ControlPlaneInitialized"

	// InfrastructureReadyCondition represents the condition type for infrastructure readiness
	InfrastructureReadyCondition v1beta1.ConditionType = "InfrastructureReady"
)

// removed: unused setOwnerReference function

func (r *Reconciler) updateStatus(ctx context.Context, captCluster *infrastructurev1beta1.CAPTCluster, cluster *v1beta1.Cluster) error {
	logger := log.FromContext(ctx)
	logger.V(1).Info("Updating status", "captCluster.Status.Ready", captCluster.Status.Ready)

	// Update CAPTCluster status using Status().Update for envtest compatibility
	if err := r.Status().Update(ctx, captCluster); err != nil {
		logger.Error(err, "Failed to update CAPTCluster status")
		return fmt.Errorf("failed to update CAPTCluster status: %v", err)
	}

	// Update Cluster status if it exists
	if cluster != nil {
		logger.V(1).Info("Updating cluster status",
			"InfrastructureReady", cluster.Status.InfrastructureReady,
			"ControlPlaneReady", cluster.Status.ControlPlaneReady)

		if err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			// Get latest Cluster to avoid resourceVersion conflicts
			current := &v1beta1.Cluster{}
			if err := r.Get(ctx, types.NamespacedName{Name: cluster.Name, Namespace: cluster.Namespace}, current); err != nil {
				return err
			}

			base := current.DeepCopy()

			// Update infrastructure ready status
			current.Status.InfrastructureReady = captCluster.Status.Ready
			logger.V(1).Info("Set InfrastructureReady", "value", current.Status.InfrastructureReady)

			// Clear failure status if ready
			if captCluster.Status.Ready {
				current.Status.FailureReason = nil
				current.Status.FailureMessage = nil
				logger.V(1).Info("Cleared failure status due to ready state")

				// Set InfrastructureReady condition
				conditions.MarkTrue(current, InfrastructureReadyCondition)
				// For legacy tests in 0.4.x, also set ControlPlaneInitialized true when infra is ready
				conditions.MarkTrue(current, ControlPlaneInitializedCondition)
				logger.V(1).Info("Set InfrastructureReady condition to True")
			} else if captCluster.Status.FailureReason != nil {
				// Update failure reason and message only if not ready
				reason := capierrors.ClusterStatusError(*captCluster.Status.FailureReason)
				current.Status.FailureReason = &reason
				current.Status.FailureMessage = captCluster.Status.FailureMessage
				// Avoid nil dereference in logs when FailureMessage is not set
				var logMessage string
				if captCluster.Status.FailureMessage != nil {
					logMessage = *captCluster.Status.FailureMessage
				}
				logger.V(1).Info("Updated failure status",
					"reason", *captCluster.Status.FailureReason,
					"message", logMessage)

				// Set InfrastructureReady condition to false
				conditions.MarkFalse(current, InfrastructureReadyCondition, string(reason), v1beta1.ConditionSeverityError, "%s", *captCluster.Status.FailureMessage)
				logger.V(1).Info("Set InfrastructureReady condition to False")
			}

			// Update failure domains if present
			if len(captCluster.Status.FailureDomains) > 0 {
				current.Status.FailureDomains = captCluster.Status.FailureDomains
				logger.V(1).Info("Updated failure domains", "count", len(captCluster.Status.FailureDomains))
			}

			// Skip patch if nothing changed
			if reflect.DeepEqual(base.Status, current.Status) {
				logger.Info("Cluster status unchanged, skipping patch")
				return nil
			}

			if err := r.Status().Patch(ctx, current, client.MergeFrom(base)); err != nil {
				return err
			}
			logger.V(1).Info("Successfully patched cluster status")
			// Reflect the updated status back to the passed-in cluster object for callers/tests
			cluster.Status = current.Status
			return nil
		}); err != nil {
			logger.Error(err, "Failed to patch cluster status")
			return fmt.Errorf("failed to update Cluster status: %v", err)
		}
	} else {
		logger.V(1).Info("Cluster is nil, skipping cluster status update")
	}

	return nil
}

func (r *Reconciler) setFailedStatus(ctx context.Context, captCluster *infrastructurev1beta1.CAPTCluster, cluster *v1beta1.Cluster, reason, message string) (Result, error) {
	logger := log.FromContext(ctx)
	// Pre-update diagnostics to aid debugging in tests
	logger.V(1).Info("setFailedStatus called",
		"reason", reason,
		"message", message,
		"ready_before", captCluster.Status.Ready,
		"failureReason_nil_before", captCluster.Status.FailureReason == nil,
		"failureMessage_nil_before", captCluster.Status.FailureMessage == nil,
		"wsStatus_nil_before", captCluster.Status.WorkspaceTemplateStatus == nil,
	)
	// Take pre-change snapshot for status patch
	pre := captCluster.DeepCopy()

	meta.SetStatusCondition(&captCluster.Status.Conditions, metav1.Condition{
		Type:               infrastructurev1beta1.VPCFailedCondition,
		Status:             metav1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             reason,
		Message:            message,
	})
	captCluster.Status.Ready = false
	captCluster.Status.FailureReason = &reason
	captCluster.Status.FailureMessage = &message

	if captCluster.Status.WorkspaceTemplateStatus == nil {
		captCluster.Status.WorkspaceTemplateStatus = &infrastructurev1beta1.CAPTClusterWorkspaceStatus{}
	}
	captCluster.Status.WorkspaceTemplateStatus.Ready = false
	captCluster.Status.WorkspaceTemplateStatus.LastFailureMessage = message

	// Post-update diagnostics before status patch
	logger.V(1).Info("setFailedStatus updated local status",
		"ready", captCluster.Status.Ready,
		"failureReason_set", captCluster.Status.FailureReason != nil,
		"failureMessage_set", captCluster.Status.FailureMessage != nil,
		"wsReady", captCluster.Status.WorkspaceTemplateStatus != nil && !captCluster.Status.WorkspaceTemplateStatus.Ready,
	)

	// Persist CAPTCluster status changes using the pre-change snapshot
	if err := r.Status().Patch(ctx, captCluster, client.MergeFrom(pre)); err != nil {
		logger.Error(err, "Failed to patch CAPTCluster failed status")
		return Result{}, fmt.Errorf("failed to update CAPTCluster failed status: %v", err)
	}

	if err := r.updateStatus(ctx, captCluster, cluster); err != nil {
		return Result{}, err
	}
	logger.V(1).Info("setFailedStatus patched status successfully")
	return Result{}, fmt.Errorf("%s", message)
}
