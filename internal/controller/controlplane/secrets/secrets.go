package secrets

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	corev1 "k8s.io/api/core/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

// SecretManager handles secret management for CAPTControlPlane
type SecretManager struct {
	client client.Client
}

// NewSecretManager creates a new SecretManager instance
func NewSecretManager(client client.Client) *SecretManager {
	return &SecretManager{
		client: client,
	}
}

// GetAndValidateSecret retrieves and validates the connection secret
func (m *SecretManager) GetAndValidateSecret(ctx context.Context, cluster *clusterv1.Cluster) (*corev1.Secret, error) {
	logger := log.FromContext(ctx)
	secretName := fmt.Sprintf("%s-eks-connection", cluster.Name)
	logger.Info("Getting and validating connection secret", "secretName", secretName)

	secret := &corev1.Secret{}
	key := types.NamespacedName{
		Namespace: cluster.Namespace,
		Name:      secretName,
	}

	if err := m.client.Get(ctx, key, secret); err != nil {
		logger.Error(err, "Failed to get secret", "secretName", secretName)
		return nil, fmt.Errorf("failed to get secret: %v", err)
	}

	// Validate required fields
	requiredFields := []string{
		"kubeconfig",
		"cluster_certificate_authority_data",
	}
	for _, field := range requiredFields {
		if _, ok := secret.Data[field]; !ok {
			logger.Error(nil, "Required field not found in secret", "field", field, "secretName", secretName)
			return nil, fmt.Errorf("required field %s not found in secret", field)
		}
		logger.Info("Found required field in secret", "field", field, "secretName", secretName)
	}

	logger.Info("Successfully validated connection secret", "name", secret.Name)
	return secret, nil
}

// GetClusterEndpoint retrieves the cluster endpoint with priority order
func (m *SecretManager) GetClusterEndpoint(ctx context.Context, workspace *unstructured.Unstructured, secret *corev1.Secret) (*clusterv1.APIEndpoint, error) {
	logger := log.FromContext(ctx)
	logger.Info("Getting cluster endpoint")

	// First try from workspace outputs
	outputs, found, err := unstructured.NestedMap(workspace.Object, "status", "atProvider", "outputs")
	if err != nil {
		logger.Error(err, "Failed to get workspace outputs")
		return nil, fmt.Errorf("failed to get workspace outputs: %v", err)
	}

	if found && outputs != nil {
		if endpoint, ok := outputs["cluster_endpoint"].(string); ok {
			logger.Info("Found endpoint in workspace outputs", "endpoint", endpoint)
			return &clusterv1.APIEndpoint{
				Host: endpoint,
				Port: 443,
			}, nil
		}
		logger.Info("Cluster endpoint not found in workspace outputs")
	}

	// Fallback to secret
	if secret != nil {
		if endpointData, ok := secret.Data["cluster_endpoint"]; ok {
			endpoint := string(endpointData)
			logger.Info("Found endpoint in secret", "endpoint", endpoint)
			return &clusterv1.APIEndpoint{
				Host: endpoint,
				Port: 443,
			}, nil
		}
		logger.Info("Cluster endpoint not found in secret")
	}

	return nil, fmt.Errorf("endpoint not found in both workspace outputs and secret")
}

// ValidateEndpoint validates the endpoint configuration
func (m *SecretManager) ValidateEndpoint(endpoint *clusterv1.APIEndpoint) error {
	if endpoint == nil {
		return fmt.Errorf("endpoint is nil")
	}

	if endpoint.Host == "" {
		return fmt.Errorf("endpoint host is empty")
	}

	if endpoint.Port == 0 {
		return fmt.Errorf("endpoint port is not set")
	}

	return nil
}

// GetKubeconfig retrieves the kubeconfig from the secret
func (m *SecretManager) GetKubeconfig(ctx context.Context, secret *corev1.Secret) (string, error) {
	logger := log.FromContext(ctx)
	if secret == nil {
		logger.Error(nil, "Secret is nil")
		return "", fmt.Errorf("secret is nil")
	}

	kubeconfigData, ok := secret.Data["kubeconfig"]
	if !ok {
		logger.Error(nil, "Kubeconfig not found in secret")
		return "", fmt.Errorf("kubeconfig not found in secret")
	}

	logger.Info("Successfully retrieved kubeconfig from secret")
	return string(kubeconfigData), nil
}

// GetCertificateAuthorityData retrieves the CA data from the secret
func (m *SecretManager) GetCertificateAuthorityData(ctx context.Context, secret *corev1.Secret) (string, error) {
	logger := log.FromContext(ctx)
	if secret == nil {
		logger.Error(nil, "Secret is nil")
		return "", fmt.Errorf("secret is nil")
	}

	caData, ok := secret.Data["cluster_certificate_authority_data"]
	if !ok {
		logger.Error(nil, "Certificate authority data not found in secret")
		return "", fmt.Errorf("certificate authority data not found in secret")
	}

	logger.Info("Successfully retrieved certificate authority data from secret")
	return string(caData), nil
}
