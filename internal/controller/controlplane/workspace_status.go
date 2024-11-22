package controlplane

import (
	"context"
	"encoding/json"
	"fmt"

	controlplanev1beta1 "github.com/appthrust/capt/api/controlplane/v1beta1"
	infrastructurev1beta1 "github.com/appthrust/capt/api/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// getWorkspaceStatus retrieves and formats the status of the associated Workspace
func (r *Reconciler) getWorkspaceStatus(ctx context.Context, workspaceApply *infrastructurev1beta1.WorkspaceTemplateApply) (*controlplanev1beta1.WorkspaceStatus, error) {
	logger := log.FromContext(ctx)

	if workspaceApply == nil || workspaceApply.Status.WorkspaceName == "" {
		return nil, nil
	}

	// Get Workspace
	workspace := &unstructured.Unstructured{}
	workspace.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "tf.upbound.io",
		Version: "v1beta1",
		Kind:    "Workspace",
	})

	if err := r.Get(ctx, client.ObjectKey{
		Name:      workspaceApply.Status.WorkspaceName,
		Namespace: workspaceApply.Namespace,
	}, workspace); err != nil {
		return nil, fmt.Errorf("failed to get workspace: %v", err)
	}

	// Extract status fields
	status := workspace.Object["status"]
	if status == nil {
		return nil, nil
	}

	statusMap, ok := status.(map[string]interface{})
	if !ok {
		return nil, nil
	}

	// Check conditions for ready status
	conditions, exists := statusMap["conditions"]
	if !exists {
		return nil, nil
	}

	conditionsArray, ok := conditions.([]interface{})
	if !ok {
		return nil, nil
	}

	// Find Ready condition
	var ready bool
	var state string
	for _, c := range conditionsArray {
		condition, ok := c.(map[string]interface{})
		if !ok {
			continue
		}

		if typeStr, ok := condition["type"].(string); ok && typeStr == "Ready" {
			if status, ok := condition["status"].(string); ok {
				ready = (status == "True")
			}
			if reason, ok := condition["reason"].(string); ok {
				state = reason
			}
			break
		}
	}

	// Extract atProvider
	atProviderData, exists := statusMap["atProvider"]
	if !exists {
		return nil, nil
	}

	atProviderMap, ok := atProviderData.(map[string]interface{})
	if !ok {
		return nil, nil
	}

	atProvider := &runtime.RawExtension{
		Object: &unstructured.Unstructured{
			Object: atProviderMap,
		},
	}

	workspaceStatus := &controlplanev1beta1.WorkspaceStatus{
		Ready:      ready,
		State:      state,
		AtProvider: atProvider,
	}

	// Log only the final result
	workspaceStatusJSON, _ := json.MarshalIndent(workspaceStatus, "", "  ")
	logger.Info("Workspace status extracted",
		"ready", ready,
		"state", state,
		"hasAtProvider", atProvider != nil,
		"fullStatus", string(workspaceStatusJSON))

	return workspaceStatus, nil
}

// updateWorkspaceStatus updates the WorkspaceStatus in CAPTControlPlane
func (r *Reconciler) updateWorkspaceStatus(ctx context.Context, controlPlane *controlplanev1beta1.CAPTControlPlane, workspaceApply *infrastructurev1beta1.WorkspaceTemplateApply) error {
	logger := log.FromContext(ctx)

	// Initialize WorkspaceStatus if nil
	if controlPlane.Status.WorkspaceStatus == nil {
		controlPlane.Status.WorkspaceStatus = &controlplanev1beta1.WorkspaceStatus{}
	}

	// Create a deep copy for patching
	patchBase := controlPlane.DeepCopy()

	// Get new workspace status
	status, err := r.getWorkspaceStatus(ctx, workspaceApply)
	if err != nil {
		return err
	}

	// Keep a copy of the current workspace status if it exists
	var currentStatus *controlplanev1beta1.WorkspaceStatus
	if controlPlane.Status.WorkspaceStatus != nil {
		currentStatus = controlPlane.Status.WorkspaceStatus.DeepCopy()
	}

	// Update status
	if status != nil {
		controlPlane.Status.WorkspaceStatus = status
	} else {
		controlPlane.Status.WorkspaceStatus = &controlplanev1beta1.WorkspaceStatus{}
	}

	// Log the status after update
	logger.Info("Status after update",
		"workspaceStatus", controlPlane.Status.WorkspaceStatus)

	// Apply the patch
	if err := r.Status().Patch(ctx, controlPlane, client.MergeFrom(patchBase)); err != nil {
		// Restore the original status in case of error
		if currentStatus != nil {
			controlPlane.Status.WorkspaceStatus = currentStatus
		}
		return fmt.Errorf("failed to patch workspace status: %v", err)
	}

	return nil
}
