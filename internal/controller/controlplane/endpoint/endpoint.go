package endpoint

import (
	"context"
	"encoding/base64"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	corev1 "k8s.io/api/core/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

// GetEndpointFromWorkspace attempts to get the cluster endpoint from a Workspace
func GetEndpointFromWorkspace(ctx context.Context, c client.Client, workspaceName string) (*clusterv1.APIEndpoint, error) {
	logger := log.FromContext(ctx)

	// Define Workspace GVK
	workspaceGVK := schema.GroupVersionKind{
		Group:   "tf.upbound.io",
		Version: "v1beta1",
		Kind:    "Workspace",
	}

	// Get Workspace
	workspace := &unstructured.Unstructured{}
	workspace.SetGroupVersionKind(workspaceGVK)
	if err := c.Get(ctx, types.NamespacedName{Name: workspaceName}, workspace); err != nil {
		return nil, fmt.Errorf("failed to get Workspace: %w", err)
	}

	logger.Info("Found Workspace", "name", workspace.GetName())

	// Try to get endpoint from outputs first
	outputs, found, err := unstructured.NestedMap(workspace.Object, "status", "atProvider", "outputs")
	if err != nil {
		return nil, fmt.Errorf("failed to get outputs from Workspace: %w", err)
	}

	if found && outputs != nil {
		if endpoint, ok := outputs["cluster_endpoint"].(string); ok {
			logger.Info("Found cluster_endpoint in Workspace outputs", "endpoint", endpoint)
			return &clusterv1.APIEndpoint{
				Host: endpoint,
				Port: 443, // EKS API server always uses port 443
			}, nil
		}
		logger.Info("cluster_endpoint not found in Workspace outputs")
	}

	// Try to get endpoint from secret if outputs don't have it
	secretRef, found, err := unstructured.NestedMap(workspace.Object, "spec", "writeConnectionSecretToRef")
	if err != nil {
		return nil, fmt.Errorf("failed to get writeConnectionSecretToRef from Workspace: %w", err)
	}

	if !found || secretRef == nil {
		logger.Info("writeConnectionSecretToRef not found in Workspace")
		return nil, nil
	}

	secretName, ok := secretRef["name"].(string)
	if !ok {
		return nil, fmt.Errorf("secret name not found in writeConnectionSecretToRef")
	}

	secretNamespace, ok := secretRef["namespace"].(string)
	if !ok {
		return nil, fmt.Errorf("secret namespace not found in writeConnectionSecretToRef")
	}

	// Get secret
	secret := &corev1.Secret{}
	if err := c.Get(ctx, types.NamespacedName{Name: secretName, Namespace: secretNamespace}, secret); err != nil {
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}

	logger.Info("Successfully retrieved secret",
		"name", secretName,
		"namespace", secretNamespace)

	if endpointData, ok := secret.Data["cluster_endpoint"]; ok {
		logger.Info("Found cluster_endpoint in secret",
			"raw_length", len(endpointData),
			"raw_data", string(endpointData))

		endpoint, err := base64.StdEncoding.DecodeString(string(endpointData))
		if err != nil {
			return nil, fmt.Errorf("failed to decode endpoint data: %w", err)
		}

		logger.Info("Successfully decoded endpoint", "endpoint", string(endpoint))
		return &clusterv1.APIEndpoint{
			Host: string(endpoint),
			Port: 443, // EKS API server always uses port 443
		}, nil
	}

	logger.Info("cluster_endpoint not found in secret")
	return nil, nil
}
