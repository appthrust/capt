package secrets

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"strconv"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	controlplanev1beta1 "github.com/appthrust/capt/api/controlplane/v1beta1"
	corev1 "k8s.io/api/core/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

// Manager handles secret management for CAPTControlPlane
type Manager struct {
	client client.Client
}

// NewManager creates a new Manager instance
func NewManager(client client.Client) *Manager {
	return &Manager{
		client: client,
	}
}

// GetAndValidateSecret retrieves and validates the connection secret
func (m *Manager) GetAndValidateSecret(ctx context.Context, controlPlane *controlplanev1beta1.CAPTControlPlane) (*corev1.Secret, error) {
	logger := log.FromContext(ctx)
	secretName := fmt.Sprintf("%s-eks-connection", controlPlane.Name)
	logger.Info("Getting and validating connection secret", "secretName", secretName)

	secret := &corev1.Secret{}
	key := types.NamespacedName{
		Namespace: controlPlane.Namespace,
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

// GetClusterEndpoint gets the cluster endpoint from either workspace outputs or secret
func (m *Manager) GetClusterEndpoint(ctx context.Context, workspace *unstructured.Unstructured, secret *corev1.Secret) (*clusterv1.APIEndpoint, error) {
	logger := log.FromContext(ctx)
	logger.Info("Getting cluster endpoint")

	// Try to get endpoint from workspace outputs
	outputs, found, err := unstructured.NestedMap(workspace.Object, "status", "atProvider", "outputs")
	if err != nil {
		logger.Error(err, "Failed to get workspace outputs")
		return nil, fmt.Errorf("failed to get workspace outputs: %v", err)
	}

	if found && outputs != nil {
		if endpointValue, ok := outputs["cluster_endpoint"].(string); ok {
			logger.Info("Found cluster_endpoint in workspace outputs", "endpoint", endpointValue)
			return parseEndpoint(endpointValue)
		}
		logger.V(4).Info("cluster_endpoint not found in workspace outputs")
	}

	// Try to get endpoint from secret
	if secret != nil && secret.Data != nil {
		if endpointBytes, ok := secret.Data["cluster_endpoint"]; ok {
			endpointValue := string(endpointBytes)
			// Try to decode if it's base64 encoded
			if decoded, err := base64.StdEncoding.DecodeString(endpointValue); err == nil {
				endpointValue = string(decoded)
			}
			logger.Info("Found cluster_endpoint in secret", "endpoint", endpointValue)
			return parseEndpoint(endpointValue)
		}
		logger.V(4).Info("cluster_endpoint not found in secret")
	}

	// Return nil without error if endpoint is not found
	// This indicates the endpoint is not yet available
	logger.V(4).Info("Endpoint not found in both workspace outputs and secret, will retry later")
	return nil, nil
}

// GetCertificateAuthorityData gets the CA data from the secret
func (m *Manager) GetCertificateAuthorityData(ctx context.Context, secret *corev1.Secret) (string, error) {
	logger := log.FromContext(ctx)
	if secret == nil || secret.Data == nil {
		logger.Error(nil, "Secret or secret data is nil")
		return "", fmt.Errorf("secret or secret data is nil")
	}

	// Try to get CA data from secret
	if caData, ok := secret.Data["cluster_certificate_authority_data"]; ok {
		// Try to decode if it's base64 encoded
		if decoded, err := base64.StdEncoding.DecodeString(string(caData)); err == nil {
			logger.Info("Successfully retrieved and decoded certificate authority data from secret")
			return string(decoded), nil
		}
		logger.Info("Successfully retrieved certificate authority data from secret")
		return string(caData), nil
	}

	logger.Error(nil, "Certificate authority data not found in secret")
	return "", fmt.Errorf("cluster_certificate_authority_data not found in secret")
}

// ValidateEndpoint validates the endpoint configuration
func (m *Manager) ValidateEndpoint(endpoint *clusterv1.APIEndpoint) error {
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

// parseEndpoint parses an endpoint string into an APIEndpoint
func parseEndpoint(endpoint string) (*clusterv1.APIEndpoint, error) {
	if endpoint == "" {
		return nil, fmt.Errorf("empty endpoint")
	}

	// Parse URL
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse endpoint URL: %v", err)
	}

	// Extract host and port
	host := u.Hostname()
	portStr := u.Port()
	if portStr == "" {
		// Default to 443 for HTTPS
		portStr = "443"
	}

	port, err := strconv.ParseInt(portStr, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("failed to parse port: %v", err)
	}

	return &clusterv1.APIEndpoint{
		Host: host,
		Port: int32(port),
	}, nil
}
