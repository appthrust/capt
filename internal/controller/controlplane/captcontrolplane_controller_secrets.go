package controlplane

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	controlplanev1beta1 "github.com/appthrust/capt/api/controlplane/v1beta1"
	infrastructurev1beta1 "github.com/appthrust/capt/api/v1beta1"
	"github.com/appthrust/capt/internal/controller/controlplane/secrets"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

const (
	ReasonSecretError   = "SecretError"
	ReasonEndpointError = "EndpointError"
)

// reconcileSecrets handles secret management for CAPTControlPlane
func (r *CAPTControlPlaneReconciler) reconcileSecrets(ctx context.Context, controlPlane *controlplanev1beta1.CAPTControlPlane, cluster *clusterv1.Cluster, workspaceApply *infrastructurev1beta1.WorkspaceTemplateApply) error {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling secrets")

	// Initialize secret manager
	secretManager := secrets.NewSecretManager(r.Client)

	// Get and validate secret
	secret, err := secretManager.GetAndValidateSecret(ctx, cluster)
	if err != nil {
		logger.Error(err, "Failed to get or validate secret")
		_, setErr := r.setFailedStatus(ctx, controlPlane, cluster, ReasonSecretError, fmt.Sprintf("Failed to get or validate secret: %v", err))
		if setErr != nil {
			return fmt.Errorf("failed to set status: %v (original error: %v)", setErr, err)
		}
		return err
	}

	// Get workspace
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
		logger.Error(err, "Failed to get workspace",
			"workspaceName", workspaceApply.Status.WorkspaceName,
			"namespace", workspaceApply.Namespace)
		_, setErr := r.setFailedStatus(ctx, controlPlane, cluster, ReasonEndpointError, fmt.Sprintf("Failed to get workspace: %v", err))
		if setErr != nil {
			return fmt.Errorf("failed to set status: %v (original error: %v)", setErr, err)
		}
		return err
	}

	logger.Info("Found workspace",
		"workspaceName", workspaceApply.Status.WorkspaceName,
		"namespace", workspaceApply.Namespace)

	// Get endpoint from workspace or secret
	endpoint, err := secretManager.GetClusterEndpoint(ctx, workspace, secret)
	if err != nil {
		logger.Error(err, "Failed to get cluster endpoint")
		_, setErr := r.setFailedStatus(ctx, controlPlane, cluster, ReasonEndpointError, fmt.Sprintf("Failed to get cluster endpoint: %v", err))
		if setErr != nil {
			return fmt.Errorf("failed to set status: %v (original error: %v)", setErr, err)
		}
		return err
	}

	// Create Cluster API kubeconfig secret
	kubeconfigSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-control-plane-kubeconfig", controlPlane.Name),
			Namespace: controlPlane.Namespace,
		},
		Data: map[string][]byte{
			"value": secret.Data["kubeconfig"],
		},
	}

	// Set controller reference
	if err := controllerutil.SetControllerReference(controlPlane, kubeconfigSecret, r.Scheme); err != nil {
		logger.Error(err, "Failed to set controller reference")
		return err
	}

	// Create or update the kubeconfig secret
	existingSecret := &corev1.Secret{}
	err = r.Get(ctx, client.ObjectKey{Name: kubeconfigSecret.Name, Namespace: kubeconfigSecret.Namespace}, existingSecret)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Create new secret
			if err := r.Create(ctx, kubeconfigSecret); err != nil {
				logger.Error(err, "Failed to create kubeconfig secret")
				return err
			}
			logger.Info("Successfully created kubeconfig secret",
				"name", kubeconfigSecret.Name,
				"namespace", kubeconfigSecret.Namespace)
		} else {
			logger.Error(err, "Failed to get existing kubeconfig secret")
			return err
		}
	} else {
		// Update existing secret
		existingSecret.Data = kubeconfigSecret.Data
		if err := r.Update(ctx, existingSecret); err != nil {
			logger.Error(err, "Failed to update kubeconfig secret")
			return err
		}
		logger.Info("Successfully updated kubeconfig secret",
			"name", existingSecret.Name,
			"namespace", existingSecret.Namespace)
	}

	// Update CAPTControlPlane endpoint
	patchBase := controlPlane.DeepCopy()
	controlPlane.Spec.ControlPlaneEndpoint = *endpoint

	if err := r.Patch(ctx, controlPlane, client.MergeFrom(patchBase)); err != nil {
		logger.Error(err, "Failed to patch CAPTControlPlane endpoint")
		return err
	}

	logger.Info("Successfully patched CAPTControlPlane endpoint")

	// Update owner cluster endpoint if it exists
	if cluster != nil {
		patchBase := cluster.DeepCopy()
		cluster.Spec.ControlPlaneEndpoint = controlPlane.Spec.ControlPlaneEndpoint
		if err := r.Patch(ctx, cluster, client.MergeFrom(patchBase)); err != nil {
			logger.Error(err, "Failed to patch Cluster endpoint")
			return err
		}
		logger.Info("Successfully patched Cluster endpoint")
	}

	return nil
}