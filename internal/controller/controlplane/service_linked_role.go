package controlplane

import (
	"context"
	"fmt"

	controlplanev1beta1 "github.com/appthrust/capt/api/controlplane/v1beta1"
	infrastructurev1beta1 "github.com/appthrust/capt/api/v1beta1"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	terraformv1beta1 "github.com/upbound/provider-terraform/apis/v1beta1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	errGetWorkspaceTemplateApply    = "failed to get WorkspaceTemplateApply"
	errCreateWorkspaceTemplateApply = "failed to create WorkspaceTemplateApply"
	errGetWorkspace                 = "failed to get Workspace"
)

// reconcileSpotServiceLinkedRole ensures the EC2 Spot Service-Linked Role exists
func (r *Reconciler) reconcileSpotServiceLinkedRole(ctx context.Context, controlPlane *controlplanev1beta1.CAPTControlPlane) error {
	logger := log.FromContext(ctx)

	// Create check workspace name
	checkWorkspaceName := fmt.Sprintf("%s-spot-role-check", controlPlane.Name)

	// Try to find existing check workspace apply
	checkWorkspaceApply := &infrastructurev1beta1.WorkspaceTemplateApply{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      checkWorkspaceName,
		Namespace: controlPlane.Namespace,
	}, checkWorkspaceApply)

	if err != nil {
		if !apierrors.IsNotFound(err) {
			logger.Error(err, errGetWorkspaceTemplateApply, "workspace", checkWorkspaceName)
			return fmt.Errorf("%s: %w", errGetWorkspaceTemplateApply, err)
		}

		// Create new check workspace apply
		checkWorkspaceApply = &infrastructurev1beta1.WorkspaceTemplateApply{
			ObjectMeta: metav1.ObjectMeta{
				Name:      checkWorkspaceName,
				Namespace: controlPlane.Namespace,
			},
			Spec: infrastructurev1beta1.WorkspaceTemplateApplySpec{
				TemplateRef: infrastructurev1beta1.WorkspaceTemplateReference{
					Name:      "spot-role-check",
					Namespace: controlPlane.Namespace,
				},
			},
		}

		// Set owner reference
		if err := controllerutil.SetControllerReference(controlPlane, checkWorkspaceApply, r.Scheme); err != nil {
			logger.Error(err, "failed to set controller reference", "workspace", checkWorkspaceName)
			return fmt.Errorf("failed to set controller reference: %w", err)
		}

		if err := r.Create(ctx, checkWorkspaceApply); err != nil {
			logger.Error(err, errCreateWorkspaceTemplateApply, "workspace", checkWorkspaceName)
			return fmt.Errorf("%s: %w", errCreateWorkspaceTemplateApply, err)
		}

		logger.Info("Created Spot Role check workspace apply", "workspace", checkWorkspaceName)
		return nil
	}

	// Check if the workspace apply is ready
	if !checkWorkspaceApply.Status.Applied {
		logger.Info("Waiting for Spot Role check workspace apply to be applied", "workspace", checkWorkspaceName)
		return nil
	}

	// Get the actual workspace
	checkWorkspace := &terraformv1beta1.Workspace{}
	if err := r.Get(ctx, types.NamespacedName{
		Name:      checkWorkspaceApply.Status.WorkspaceName,
		Namespace: controlPlane.Namespace,
	}, checkWorkspace); err != nil {
		logger.Error(err, errGetWorkspace, "workspace", checkWorkspaceApply.Status.WorkspaceName)
		return fmt.Errorf("%s: %w", errGetWorkspace, err)
	}

	// Check if the workspace is ready
	readyCondition := FindStatusCondition(checkWorkspace.Status.Conditions, xpv1.TypeReady)
	if readyCondition == nil || readyCondition.Status != corev1.ConditionTrue {
		logger.Info("Waiting for Spot Role check workspace to be ready", "workspace", checkWorkspaceApply.Status.WorkspaceName)
		return nil
	}

	// Get role_exists from outputs
	roleExists := false
	if checkWorkspace.Status.AtProvider.Outputs != nil {
		if val, ok := checkWorkspace.Status.AtProvider.Outputs["role_exists"]; ok {
			if string(val.Raw) == "true" {
				roleExists = true
			}
		}
	}

	if !roleExists {
		// Role doesn't exist, create it
		createWorkspaceName := fmt.Sprintf("%s-spot-role-create", controlPlane.Name)
		createWorkspaceApply := &infrastructurev1beta1.WorkspaceTemplateApply{}
		err := r.Get(ctx, types.NamespacedName{
			Name:      createWorkspaceName,
			Namespace: controlPlane.Namespace,
		}, createWorkspaceApply)

		if err != nil {
			if !apierrors.IsNotFound(err) {
				logger.Error(err, errGetWorkspaceTemplateApply, "workspace", createWorkspaceName)
				return fmt.Errorf("%s: %w", errGetWorkspaceTemplateApply, err)
			}

			// Create new create workspace apply
			createWorkspaceApply = &infrastructurev1beta1.WorkspaceTemplateApply{
				ObjectMeta: metav1.ObjectMeta{
					Name:      createWorkspaceName,
					Namespace: controlPlane.Namespace,
				},
				Spec: infrastructurev1beta1.WorkspaceTemplateApplySpec{
					TemplateRef: infrastructurev1beta1.WorkspaceTemplateReference{
						Name:      "spot-role-create",
						Namespace: controlPlane.Namespace,
					},
				},
			}

			// Set owner reference
			if err := controllerutil.SetControllerReference(controlPlane, createWorkspaceApply, r.Scheme); err != nil {
				logger.Error(err, "failed to set controller reference", "workspace", createWorkspaceName)
				return fmt.Errorf("failed to set controller reference: %w", err)
			}

			if err := r.Create(ctx, createWorkspaceApply); err != nil {
				logger.Error(err, errCreateWorkspaceTemplateApply, "workspace", createWorkspaceName)
				return fmt.Errorf("%s: %w", errCreateWorkspaceTemplateApply, err)
			}

			logger.Info("Created Spot Role create workspace apply", "workspace", createWorkspaceName)
			return nil
		}

		// Check if the workspace apply is ready
		if !createWorkspaceApply.Status.Applied {
			logger.Info("Waiting for Spot Role create workspace apply to be applied", "workspace", createWorkspaceName)
			return nil
		}

		// Get the actual workspace
		createWorkspace := &terraformv1beta1.Workspace{}
		if err := r.Get(ctx, types.NamespacedName{
			Name:      createWorkspaceApply.Status.WorkspaceName,
			Namespace: controlPlane.Namespace,
		}, createWorkspace); err != nil {
			logger.Error(err, errGetWorkspace, "workspace", createWorkspaceApply.Status.WorkspaceName)
			return fmt.Errorf("%s: %w", errGetWorkspace, err)
		}

		// Check if the create workspace is ready
		readyCondition = FindStatusCondition(createWorkspace.Status.Conditions, xpv1.TypeReady)
		if readyCondition == nil || readyCondition.Status != corev1.ConditionTrue {
			logger.Info("Waiting for Spot Role create workspace to be ready", "workspace", createWorkspaceApply.Status.WorkspaceName)
			return nil
		}

		// Check if role_arn is in outputs
		if createWorkspace.Status.AtProvider.Outputs != nil {
			if _, ok := createWorkspace.Status.AtProvider.Outputs["role_arn"]; !ok {
				logger.Error(nil, "role_arn not found in outputs", "workspace", createWorkspaceApply.Status.WorkspaceName)
				return fmt.Errorf("role_arn not found in outputs for workspace %s", createWorkspaceApply.Status.WorkspaceName)
			}
		}
	}

	return nil
}

// FindStatusCondition finds the condition that matches the given type in the condition slice
func FindStatusCondition(conditions []xpv1.Condition, conditionType xpv1.ConditionType) *xpv1.Condition {
	for i := range conditions {
		if conditions[i].Type == conditionType {
			return &conditions[i]
		}
	}
	return nil
}
