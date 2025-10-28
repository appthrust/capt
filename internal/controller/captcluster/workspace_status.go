package captcluster

import (
	"context"
	"encoding/json"
	"fmt"

	infrastructurev1beta1 "github.com/appthrust/capt/api/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// getWorkspaceStatus retrieves and formats the status of the associated Workspace for CAPTCluster
func (r *Reconciler) getWorkspaceStatus(ctx context.Context, workspaceApply *infrastructurev1beta1.WorkspaceTemplateApply) (*infrastructurev1beta1.WorkspaceStatus, error) {
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

	workspaceStatus := &infrastructurev1beta1.WorkspaceStatus{
		Ready:      ready,
		State:      state,
		AtProvider: atProvider,
	}

	// Log only the final result
	workspaceStatusJSON, _ := json.MarshalIndent(workspaceStatus, "", "  ")
	logger.Info("Workspace status extracted (CAPTCluster)",
		"ready", ready,
		"state", state,
		"hasAtProvider", atProvider != nil,
		"fullStatus", string(workspaceStatusJSON))

	return workspaceStatus, nil
}

// updateWorkspaceStatus updates the WorkspaceStatus in CAPTCluster
func (r *Reconciler) updateWorkspaceStatus(ctx context.Context, captCluster *infrastructurev1beta1.CAPTCluster, workspaceApply *infrastructurev1beta1.WorkspaceTemplateApply) error {
	logger := log.FromContext(ctx)

	// Initialize WorkspaceStatus if nil
	if captCluster.Status.WorkspaceStatus == nil {
		captCluster.Status.WorkspaceStatus = &infrastructurev1beta1.WorkspaceStatus{}
	}

	// Get new workspace status
	status, err := r.getWorkspaceStatus(ctx, workspaceApply)
	if err != nil {
		return err
	}

	// Update status
	if status != nil {
		captCluster.Status.WorkspaceStatus = status
	} else {
		captCluster.Status.WorkspaceStatus = &infrastructurev1beta1.WorkspaceStatus{}
	}

	// Log the status after update
	logger.Info("CAPTCluster workspaceStatus after update",
		"workspaceStatus", captCluster.Status.WorkspaceStatus)
	return nil
}
