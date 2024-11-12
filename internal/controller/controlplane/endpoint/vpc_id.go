package endpoint

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// GetVPCIDFromWorkspace attempts to get the VPC ID from a Workspace
func GetVPCIDFromWorkspace(ctx context.Context, c client.Client, namespace, workspaceName string) (string, error) {
	logger := log.FromContext(ctx)
	logger.Info("Attempting to get VPC ID from workspace", "namespace", namespace, "workspaceName", workspaceName)

	// Define Workspace GVK
	workspaceGVK := schema.GroupVersionKind{
		Group:   "tf.upbound.io",
		Version: "v1beta1",
		Kind:    "Workspace",
	}

	// Get Workspace
	workspace := &unstructured.Unstructured{}
	workspace.SetGroupVersionKind(workspaceGVK)
	if err := c.Get(ctx, types.NamespacedName{Namespace: namespace, Name: workspaceName}, workspace); err != nil {
		logger.Error(err, "Failed to get Workspace", "namespace", namespace, "workspaceName", workspaceName)
		return "", fmt.Errorf("failed to get Workspace: %w", err)
	}

	logger.Info("Found Workspace", "name", workspace.GetName())

	// Try to get vpc_id from outputs
	outputs, found, err := unstructured.NestedMap(workspace.Object, "status", "atProvider", "outputs")
	if err != nil {
		logger.Error(err, "Failed to get outputs from Workspace")
		return "", fmt.Errorf("failed to get outputs from Workspace: %w", err)
	}

	logger.Info("Workspace outputs status", "found", found, "outputs", outputs)

	if found && outputs != nil {
		if vpcID, ok := outputs["vpc_id"].(string); ok {
			logger.Info("Found vpc_id in Workspace outputs", "vpc_id", vpcID)
			return vpcID, nil
		}
		logger.Info("vpc_id not found in Workspace outputs or not a string")
	}

	// Log the entire workspace object for debugging
	logger.Info("Workspace object", "workspace", workspace.Object)

	return "", nil
}
