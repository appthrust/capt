package controlplane

import (
	"context"
	"fmt"

	controlplanev1beta1 "github.com/appthrust/capt/api/controlplane/v1beta1"
	infrastructurev1beta1 "github.com/appthrust/capt/api/v1beta1"
	"github.com/appthrust/capt/internal/controller/controlplane/secrets"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	// Reason constants for status conditions
	ReasonSecretError       = "SecretError"
	ReasonEndpointError     = "EndpointError"
	ReasonWorkspaceNotReady = "WorkspaceNotReady"
	ReasonEndpointNotReady  = "EndpointNotReady"
)

// reconcileSecrets handles secret management for CAPTControlPlane
func (r *Reconciler) reconcileSecrets(ctx context.Context, controlPlane *controlplanev1beta1.CAPTControlPlane, cluster *clusterv1.Cluster, workspaceApply *infrastructurev1beta1.WorkspaceTemplateApply) error {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling secrets")

	// Skip if secrets are already reconciled
	if controlPlane.Status.SecretsReady {
		logger.V(4).Info("Secrets already reconciled")
		return nil
	}

	// Verify WorkspaceTemplateApply is ready and has a workspace name
	if workspaceApply.Status.WorkspaceName == "" {
		logger.Info("Workspace name not set, waiting for WorkspaceTemplateApply to be ready")
		return nil
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
		logger.Error(err, "Failed to get workspace")
		return nil
	}

	// Get connection secret
	secret := &corev1.Secret{}
	secretRef := workspaceApply.Spec.WriteConnectionSecretToRef
	if secretRef == nil {
		err := fmt.Errorf("workspace connection secret reference is not set")
		logger.Error(err, "Failed to get secret reference")
		return nil
	}

	if err := r.Get(ctx, client.ObjectKey{
		Name:      secretRef.Name,
		Namespace: secretRef.Namespace,
	}, secret); err != nil {
		if !apierrors.IsNotFound(err) {
			logger.Error(err, "Failed to get secret")
			return nil
		}
		// If secret is not found, wait for it to be created
		logger.Info("Waiting for secret to be created")
		return nil
	}

	// Initialize secret manager
	secretManager := secrets.NewManager(r.Client)

	// Get endpoint
	endpoint, err := secretManager.GetClusterEndpoint(ctx, workspace, secret)
	if err != nil {
		logger.Error(err, "Failed to get cluster endpoint")
		return nil
	}

	// If endpoint is not ready yet, requeue
	if endpoint == nil {
		logger.Info("Endpoint not ready yet, will retry")
		return nil
	}

	// Get CA data
	caData, err := secretManager.GetCertificateAuthorityData(ctx, secret)
	if err != nil {
		logger.Error(err, "Failed to get certificate authority data")
		return nil
	}

	// Check if kubeconfig secret already exists
	kubeconfigSecretName := fmt.Sprintf("%s-control-plane-kubeconfig", controlPlane.Name)
	existingKubeconfigSecret := &corev1.Secret{}
	kubeconfigSecretExists := true
	if err := r.Get(ctx, client.ObjectKey{
		Name:      kubeconfigSecretName,
		Namespace: controlPlane.Namespace,
	}, existingKubeconfigSecret); err != nil {
		if !apierrors.IsNotFound(err) {
			logger.Error(err, "Failed to get existing kubeconfig secret")
			return nil
		}
		kubeconfigSecretExists = false
	}

	// Create or update kubeconfig secret only if needed
	if !kubeconfigSecretExists {
		kubeconfigSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      kubeconfigSecretName,
				Namespace: controlPlane.Namespace,
			},
			Data: map[string][]byte{
				"value": secret.Data["kubeconfig"],
			},
		}

		if err := controllerutil.SetControllerReference(controlPlane, kubeconfigSecret, r.Scheme); err != nil {
			logger.Error(err, "Failed to set controller reference for kubeconfig secret")
			return nil
		}

		if err := r.Create(ctx, kubeconfigSecret); err != nil {
			logger.Error(err, "Failed to create kubeconfig secret")
			return nil
		}
		logger.Info("Created kubeconfig secret")
	}

	// Check if CA secret already exists
	caSecretName := fmt.Sprintf("%s-ca", controlPlane.Name)
	existingCASecret := &corev1.Secret{}
	caSecretExists := true
	if err := r.Get(ctx, client.ObjectKey{
		Name:      caSecretName,
		Namespace: controlPlane.Namespace,
	}, existingCASecret); err != nil {
		if !apierrors.IsNotFound(err) {
			logger.Error(err, "Failed to get existing CA secret")
			return nil
		}
		caSecretExists = false
	}

	// Create or update CA secret only if needed
	if !caSecretExists {
		caSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      caSecretName,
				Namespace: controlPlane.Namespace,
			},
			Data: map[string][]byte{
				"tls.crt": []byte(caData),
				"ca.crt":  []byte(caData),
			},
		}

		if err := controllerutil.SetControllerReference(controlPlane, caSecret, r.Scheme); err != nil {
			logger.Error(err, "Failed to set controller reference for CA secret")
			return nil
		}

		if err := r.Create(ctx, caSecret); err != nil {
			logger.Error(err, "Failed to create CA secret")
			return nil
		}
		logger.Info("Created CA secret")
	}

	// Mark secrets as reconciled
	if !controlPlane.Status.SecretsReady {
		patchBase := controlPlane.DeepCopy()
		controlPlane.Status.SecretsReady = true
		if err := r.Status().Patch(ctx, controlPlane, client.MergeFrom(patchBase)); err != nil {
			logger.Error(err, "Failed to update secrets status")
			return nil
		}
		logger.Info("Marked secrets as ready")
	}

	return nil
}
