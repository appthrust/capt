package captcluster

import (
	"context"
	"fmt"

	infrastructurev1beta1 "github.com/appthrust/capt/api/v1beta1"
	"github.com/appthrust/capt/internal/controller/controlplane/endpoint"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *Reconciler) reconcileVPC(ctx context.Context, captCluster *infrastructurev1beta1.CAPTCluster, cluster *clusterv1.Cluster) (Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Starting VPC reconciliation", "clusterName", captCluster.Name)

	// Ensure parent cluster exists
	if cluster == nil {
		logger.Info("Parent cluster is nil, cannot proceed with VPC reconciliation")
		return Result{}, fmt.Errorf("parent cluster is required for VPC reconciliation")
	}

	// Handle existing VPC case
	if captCluster.Spec.ExistingVPCID != "" {
		return r.handleExistingVPC(ctx, captCluster, cluster)
	}

	// Handle VPC template case
	if captCluster.Spec.VPCTemplateRef == nil {
		logger.Info("No VPC configuration provided")
		return Result{}, nil
	}

	return r.handleVPCTemplate(ctx, captCluster, cluster)
}

func (r *Reconciler) handleExistingVPC(ctx context.Context, captCluster *infrastructurev1beta1.CAPTCluster, cluster *clusterv1.Cluster) (Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Using existing VPC", "vpcId", captCluster.Spec.ExistingVPCID)

	captCluster.Status.VPCID = captCluster.Spec.ExistingVPCID
	captCluster.Status.Ready = true
	meta.SetStatusCondition(&captCluster.Status.Conditions, metav1.Condition{
		Type:               infrastructurev1beta1.VPCReadyCondition,
		Status:             metav1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             infrastructurev1beta1.ReasonExistingVPCUsed,
		Message:            "Using existing VPC",
	})

	if err := r.updateStatus(ctx, captCluster, cluster); err != nil {
		return Result{}, err
	}
	return Result{}, nil
}

func (r *Reconciler) handleVPCTemplate(ctx context.Context, captCluster *infrastructurev1beta1.CAPTCluster, cluster *clusterv1.Cluster) (Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Creating new VPC using template", "templateRef", captCluster.Spec.VPCTemplateRef)

	// Get the referenced WorkspaceTemplate
	vpcTemplate := &infrastructurev1beta1.WorkspaceTemplate{}
	templateName := types.NamespacedName{
		Name:      captCluster.Spec.VPCTemplateRef.Name,
		Namespace: captCluster.Namespace,
	}
	if err := r.Get(ctx, templateName, vpcTemplate); err != nil {
		logger.Error(err, "Failed to get VPC WorkspaceTemplate")
		return Result{}, fmt.Errorf("failed to get VPC WorkspaceTemplate: %v", err)
	}

	// Get or create WorkspaceTemplateApply
	workspaceApply, err := r.getOrCreateWorkspaceTemplateApply(ctx, captCluster)
	if err != nil {
		return Result{}, err
	}

	// Update status based on WorkspaceTemplateApply conditions
	if result, err := r.updateVPCStatus(ctx, captCluster, cluster, workspaceApply); err != nil || result.RequeueAfter > 0 {
		return result, err
	}

	// Get and verify VPC ID
	if result, err := r.verifyVPCID(ctx, captCluster, cluster, workspaceApply); err != nil || result.RequeueAfter > 0 {
		return result, err
	}

	return Result{}, nil
}

func (r *Reconciler) getOrCreateWorkspaceTemplateApply(ctx context.Context, captCluster *infrastructurev1beta1.CAPTCluster) (*infrastructurev1beta1.WorkspaceTemplateApply, error) {
	logger := log.FromContext(ctx)

	// Determine the name for WorkspaceTemplateApply
	applyName := captCluster.Spec.WorkspaceTemplateApplyName
	if applyName == "" {
		applyName = fmt.Sprintf("%s-vpc", captCluster.Name)
	}

	// Try to find existing WorkspaceTemplateApply
	workspaceApply := &infrastructurev1beta1.WorkspaceTemplateApply{}
	err := r.Get(ctx, types.NamespacedName{Name: applyName, Namespace: captCluster.Namespace}, workspaceApply)
	if err == nil {
		// Update existing WorkspaceTemplateApply if needed
		workspaceApply.Spec = infrastructurev1beta1.WorkspaceTemplateApplySpec{
			TemplateRef: *captCluster.Spec.VPCTemplateRef,
			Variables: map[string]string{
				"name":        fmt.Sprintf("%s-vpc", captCluster.Name),
				"environment": "production", // TODO: Make this configurable
			},
		}
		if err := r.Update(ctx, workspaceApply); err != nil {
			logger.Error(err, "Failed to update WorkspaceTemplateApply")
			return nil, fmt.Errorf("failed to update WorkspaceTemplateApply: %v", err)
		}
		return workspaceApply, nil
	}

	if err != nil && !apierrors.IsNotFound(err) {
		return nil, err
	}

	// Create new WorkspaceTemplateApply
	workspaceApply = &infrastructurev1beta1.WorkspaceTemplateApply{
		ObjectMeta: metav1.ObjectMeta{
			Name:      applyName,
			Namespace: captCluster.Namespace,
		},
		Spec: infrastructurev1beta1.WorkspaceTemplateApplySpec{
			TemplateRef: *captCluster.Spec.VPCTemplateRef,
			Variables: map[string]string{
				"name":        fmt.Sprintf("%s-vpc", captCluster.Name),
				"environment": "production", // TODO: Make this configurable
			},
		},
	}

	// Set owner reference
	if err := controllerutil.SetControllerReference(captCluster, workspaceApply, r.Scheme); err != nil {
		logger.Error(err, "Failed to set owner reference")
		return nil, fmt.Errorf("failed to set owner reference: %v", err)
	}

	if err := r.Create(ctx, workspaceApply); err != nil {
		logger.Error(err, "Failed to create WorkspaceTemplateApply")
		return nil, fmt.Errorf("failed to create WorkspaceTemplateApply: %v", err)
	}

	// Update WorkspaceTemplateApplyName in Spec
	patch := client.MergeFrom(captCluster.DeepCopy())
	captCluster.Spec.WorkspaceTemplateApplyName = applyName
	if err := r.Patch(ctx, captCluster, patch); err != nil {
		logger.Error(err, "Failed to update WorkspaceTemplateApplyName in spec")
		return nil, fmt.Errorf("failed to update WorkspaceTemplateApplyName in spec: %v", err)
	}

	return workspaceApply, nil
}

func (r *Reconciler) updateVPCStatus(ctx context.Context, captCluster *infrastructurev1beta1.CAPTCluster, cluster *clusterv1.Cluster, workspaceApply *infrastructurev1beta1.WorkspaceTemplateApply) (Result, error) {
	logger := log.FromContext(ctx)

	// Initialize WorkspaceTemplateStatus if not exists
	if captCluster.Status.WorkspaceTemplateStatus == nil {
		captCluster.Status.WorkspaceTemplateStatus = &infrastructurev1beta1.CAPTClusterWorkspaceStatus{}
	}

	// Check conditions
	var syncedCondition, readyCondition bool
	var errorMessage string

	for _, condition := range workspaceApply.Status.Conditions {
		logger.Info("Checking condition", "type", condition.Type, "status", condition.Status)
		if condition.Type == xpv1.TypeSynced {
			syncedCondition = condition.Status == corev1.ConditionTrue
			if !syncedCondition {
				errorMessage = condition.Message
			}
		}
		if condition.Type == xpv1.TypeReady {
			readyCondition = condition.Status == corev1.ConditionTrue
			if !readyCondition {
				errorMessage = condition.Message
			}
		}
	}

	logger.Info("Status check", "applied", workspaceApply.Status.Applied, "synced", syncedCondition, "ready", readyCondition)

	if !workspaceApply.Status.Applied || !syncedCondition || !readyCondition {
		// Update status based on workspace conditions
		if errorMessage != "" {
			meta.SetStatusCondition(&captCluster.Status.Conditions, metav1.Condition{
				Type:               infrastructurev1beta1.VPCReadyCondition,
				Status:             metav1.ConditionFalse,
				LastTransitionTime: metav1.Now(),
				Reason:             infrastructurev1beta1.ReasonVPCCreationFailed,
				Message:            errorMessage,
			})
		} else {
			meta.SetStatusCondition(&captCluster.Status.Conditions, metav1.Condition{
				Type:               infrastructurev1beta1.VPCReadyCondition,
				Status:             metav1.ConditionFalse,
				LastTransitionTime: metav1.Now(),
				Reason:             infrastructurev1beta1.ReasonVPCCreating,
				Message:            "VPC is being created",
			})
		}

		captCluster.Status.Ready = false
		captCluster.Status.WorkspaceTemplateStatus.Ready = false
		if errorMessage != "" {
			captCluster.Status.WorkspaceTemplateStatus.LastFailureMessage = errorMessage
			if workspaceApply.Status.LastAppliedTime != nil {
				captCluster.Status.WorkspaceTemplateStatus.LastFailedRevision = workspaceApply.Status.LastAppliedTime.String()
			}
		}

		if err := r.updateStatus(ctx, captCluster, cluster); err != nil {
			return Result{}, err
		}
		return Result{RequeueAfter: requeueInterval}, nil
	}

	return Result{}, nil
}

func (r *Reconciler) verifyVPCID(ctx context.Context, captCluster *infrastructurev1beta1.CAPTCluster, cluster *clusterv1.Cluster, workspaceApply *infrastructurev1beta1.WorkspaceTemplateApply) (Result, error) {
	logger := log.FromContext(ctx)

	if workspaceApply.Status.WorkspaceName == "" {
		logger.Info("WorkspaceName not set in WorkspaceTemplateApply status", "workspaceApplyName", workspaceApply.Name)
		return Result{RequeueAfter: requeueInterval}, nil
	}

	vpcID, err := endpoint.GetVPCIDFromWorkspace(ctx, r.Client, workspaceApply.Status.WorkspaceName)
	if err != nil {
		logger.Error(err, "Failed to get VPC ID from workspace")
		return Result{RequeueAfter: requeueInterval}, nil
	}

	if vpcID == "" {
		logger.Info("VPC ID not found in workspace outputs, requeueing")
		return Result{RequeueAfter: requeueInterval}, nil
	}

	captCluster.Status.VPCID = vpcID
	logger.Info("Set VPC ID in status", "vpcId", vpcID)

	// Update final status
	meta.SetStatusCondition(&captCluster.Status.Conditions, metav1.Condition{
		Type:               infrastructurev1beta1.VPCReadyCondition,
		Status:             metav1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             infrastructurev1beta1.ReasonVPCCreated,
		Message:            "VPC has been successfully created",
	})

	captCluster.Status.Ready = true
	captCluster.Status.WorkspaceTemplateStatus.Ready = true
	captCluster.Status.WorkspaceTemplateStatus.LastAppliedRevision = workspaceApply.Status.LastAppliedTime.String()

	logger.Info("Updating final status",
		"ready", captCluster.Status.Ready,
		"vpcId", captCluster.Status.VPCID,
		"workspaceTemplateStatus", captCluster.Status.WorkspaceTemplateStatus)

	if err := r.updateStatus(ctx, captCluster, cluster); err != nil {
		return Result{}, err
	}

	return Result{}, nil
}
