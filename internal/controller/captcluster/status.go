package captcluster

import (
	"context"
	"fmt"

	infrastructurev1beta1 "github.com/appthrust/capt/api/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/cluster-api/api/v1beta1"
	capierrors "sigs.k8s.io/cluster-api/errors"
	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	// ControlPlaneInitializedCondition represents the condition type for control plane initialization
	ControlPlaneInitializedCondition v1beta1.ConditionType = "ControlPlaneInitialized"

	// InfrastructureReadyCondition represents the condition type for infrastructure readiness
	InfrastructureReadyCondition v1beta1.ConditionType = "InfrastructureReady"
)

func (r *Reconciler) setOwnerReference(ctx context.Context, captCluster *infrastructurev1beta1.CAPTCluster, cluster *v1beta1.Cluster) error {
	if cluster == nil {
		return nil
	}

	// Check if owner reference is already set
	for _, ref := range captCluster.OwnerReferences {
		if ref.Kind == "Cluster" && ref.APIVersion == v1beta1.GroupVersion.String() {
			return nil
		}
	}

	// Set owner reference
	if err := controllerutil.SetControllerReference(cluster, captCluster, r.Scheme); err != nil {
		return fmt.Errorf("failed to set owner reference: %v", err)
	}

	return r.Update(ctx, captCluster)
}

func (r *Reconciler) updateStatus(ctx context.Context, captCluster *infrastructurev1beta1.CAPTCluster, cluster *v1beta1.Cluster) error {
	logger := log.FromContext(ctx)
	logger.Info("Updating status", "captCluster.Status.Ready", captCluster.Status.Ready)

	// Update CAPTCluster status
	if err := r.Status().Update(ctx, captCluster); err != nil {
		logger.Error(err, "Failed to update CAPTCluster status")
		return fmt.Errorf("failed to update CAPTCluster status: %v", err)
	}

	// Update Cluster status if it exists
	if cluster != nil {
		logger.Info("Updating cluster status",
			"InfrastructureReady", cluster.Status.InfrastructureReady,
			"ControlPlaneReady", cluster.Status.ControlPlaneReady)

		patch := client.MergeFrom(cluster.DeepCopy())

		// Update infrastructure ready status
		cluster.Status.InfrastructureReady = captCluster.Status.Ready
		logger.Info("Set InfrastructureReady", "value", cluster.Status.InfrastructureReady)

		// Clear failure status if ready
		if captCluster.Status.Ready {
			cluster.Status.FailureReason = nil
			cluster.Status.FailureMessage = nil
			logger.Info("Cleared failure status due to ready state")

			// Set InfrastructureReady condition
			conditions.Set(cluster, &v1beta1.Condition{
				Type:               InfrastructureReadyCondition,
				Status:             corev1.ConditionTrue,
				LastTransitionTime: metav1.Now(),
				Reason:             "InfrastructureReady",
				Message:            "Infrastructure is ready",
			})
			logger.Info("Set InfrastructureReady condition to True")
		} else if captCluster.Status.FailureReason != nil {
			// Update failure reason and message only if not ready
			reason := capierrors.ClusterStatusError(*captCluster.Status.FailureReason)
			cluster.Status.FailureReason = &reason
			cluster.Status.FailureMessage = captCluster.Status.FailureMessage
			logger.Info("Updated failure status",
				"reason", *captCluster.Status.FailureReason,
				"message", *captCluster.Status.FailureMessage)

			// Set InfrastructureReady condition to false
			conditions.Set(cluster, &v1beta1.Condition{
				Type:               InfrastructureReadyCondition,
				Status:             corev1.ConditionFalse,
				LastTransitionTime: metav1.Now(),
				Reason:             string(reason),
				Message:            *captCluster.Status.FailureMessage,
			})
			logger.Info("Set InfrastructureReady condition to False")
		}

		// Update failure domains if present
		if len(captCluster.Status.FailureDomains) > 0 {
			cluster.Status.FailureDomains = captCluster.Status.FailureDomains
			logger.Info("Updated failure domains", "count", len(captCluster.Status.FailureDomains))
		}

		if err := r.Status().Patch(ctx, cluster, patch); err != nil {
			logger.Error(err, "Failed to patch cluster status")
			return fmt.Errorf("failed to update Cluster status: %v", err)
		}
		logger.Info("Successfully patched cluster status")
	} else {
		logger.Info("Cluster is nil, skipping cluster status update")
	}

	return nil
}

func (r *Reconciler) setFailedStatus(ctx context.Context, captCluster *infrastructurev1beta1.CAPTCluster, cluster *v1beta1.Cluster, reason, message string) (Result, error) {
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

	if captCluster.Status.WorkspaceTemplateStatus != nil {
		captCluster.Status.WorkspaceTemplateStatus.Ready = false
		captCluster.Status.WorkspaceTemplateStatus.LastFailureMessage = message
	}

	if err := r.updateStatus(ctx, captCluster, cluster); err != nil {
		return Result{}, err
	}
	return Result{}, fmt.Errorf(message)
}
