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
)

// reconcileSecrets handles secret management for CAPTControlPlane
func (r *Reconciler) reconcileSecrets(ctx context.Context, controlPlane *controlplanev1beta1.CAPTControlPlane, cluster *clusterv1.Cluster, workspaceApply *infrastructurev1beta1.WorkspaceTemplateApply) error {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling secrets")

	// Verify WorkspaceTemplateApply is ready and has a workspace name
	if workspaceApply.Status.WorkspaceName == "" {
		logger.Info("Workspace name not set, waiting for WorkspaceTemplateApply to be ready")
		_, err := r.setFailedStatus(ctx, controlPlane, cluster, ReasonWorkspaceNotReady, "Waiting for workspace to be created")
		if err != nil {
			return fmt.Errorf("failed to set status: %v", err)
		}
		return fmt.Errorf("workspace name is not set")
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
		_, setErr := r.setFailedStatus(ctx, controlPlane, cluster, ReasonEndpointError, fmt.Sprintf("Failed to get workspace: %v", err))
		if setErr != nil {
			return fmt.Errorf("failed to set status: %v (original error: %v)", setErr, err)
		}
		return err
	}

	// Get connection secret
	secret := &corev1.Secret{}
	secretRef := workspaceApply.Spec.WriteConnectionSecretToRef
	if secretRef == nil {
		err := fmt.Errorf("workspace connection secret reference is not set")
		logger.Error(err, "Failed to get secret reference")
		_, setErr := r.setFailedStatus(ctx, controlPlane, cluster, ReasonSecretError, err.Error())
		if setErr != nil {
			return fmt.Errorf("failed to set status: %v (original error: %v)", setErr, err)
		}
		return err
	}

	if err := r.Get(ctx, client.ObjectKey{
		Name:      secretRef.Name,
		Namespace: secretRef.Namespace,
	}, secret); err != nil {
		if !apierrors.IsNotFound(err) {
			logger.Error(err, "Failed to get secret")
			_, setErr := r.setFailedStatus(ctx, controlPlane, cluster, ReasonSecretError, fmt.Sprintf("Failed to get secret: %v", err))
			if setErr != nil {
				return fmt.Errorf("failed to set status: %v (original error: %v)", setErr, err)
			}
			return err
		}
		// If secret is not found, wait for it to be created
		logger.Info("Waiting for secret to be created")
		_, err := r.setFailedStatus(ctx, controlPlane, cluster, ReasonSecretError, "Waiting for connection secret to be created")
		if err != nil {
			return fmt.Errorf("failed to set status: %v", err)
		}
		return fmt.Errorf("connection secret not found")
	}

	// Initialize secret manager
	secretManager := secrets.NewManager(r.Client)

	// Get endpoint
	endpoint, err := secretManager.GetClusterEndpoint(ctx, workspace, secret)
	if err != nil {
		logger.Error(err, "Failed to get cluster endpoint")
		_, setErr := r.setFailedStatus(ctx, controlPlane, cluster, ReasonEndpointError, fmt.Sprintf("Failed to get cluster endpoint: %v", err))
		if setErr != nil {
			return fmt.Errorf("failed to set status: %v (original error: %v)", setErr, err)
		}
		return err
	}

	// Get CA data
	caData, err := secretManager.GetCertificateAuthorityData(ctx, secret)
	if err != nil {
		logger.Error(err, "Failed to get certificate authority data")
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

	// Set controller reference for kubeconfig secret
	if err := controllerutil.SetControllerReference(controlPlane, kubeconfigSecret, r.Scheme); err != nil {
		logger.Error(err, "Failed to set controller reference for kubeconfig secret")
		return err
	}

	// Create or update the kubeconfig secret
	if err := r.Create(ctx, kubeconfigSecret); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			logger.Error(err, "Failed to create kubeconfig secret")
			return err
		}
		// Update existing secret
		existing := &corev1.Secret{}
		if err := r.Get(ctx, client.ObjectKey{Name: kubeconfigSecret.Name, Namespace: kubeconfigSecret.Namespace}, existing); err != nil {
			logger.Error(err, "Failed to get existing kubeconfig secret")
			return err
		}
		existing.Data = kubeconfigSecret.Data
		if err := r.Update(ctx, existing); err != nil {
			logger.Error(err, "Failed to update kubeconfig secret")
			return err
		}
	}

	// Create Cluster API CA certificate secret
	caSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-ca", controlPlane.Name),
			Namespace: controlPlane.Namespace,
		},
		Data: map[string][]byte{
			"tls.crt": []byte(caData),
			"ca.crt":  []byte(caData),
		},
	}

	// Set controller reference for CA secret
	if err := controllerutil.SetControllerReference(controlPlane, caSecret, r.Scheme); err != nil {
		logger.Error(err, "Failed to set controller reference for CA secret")
		return err
	}

	// Create or update the CA secret
	if err := r.Create(ctx, caSecret); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			logger.Error(err, "Failed to create CA secret")
			return err
		}
		// Update existing secret
		existing := &corev1.Secret{}
		if err := r.Get(ctx, client.ObjectKey{Name: caSecret.Name, Namespace: caSecret.Namespace}, existing); err != nil {
			logger.Error(err, "Failed to get existing CA secret")
			return err
		}
		existing.Data = caSecret.Data
		if err := r.Update(ctx, existing); err != nil {
			logger.Error(err, "Failed to update CA secret")
			return err
		}
	}

	// Update CAPTControlPlane endpoint
	patchBase := controlPlane.DeepCopy()
	controlPlane.Spec.ControlPlaneEndpoint = *endpoint

	if err := r.Patch(ctx, controlPlane, client.MergeFrom(patchBase)); err != nil {
		logger.Error(err, "Failed to patch CAPTControlPlane endpoint")
		return err
	}

	// Update owner cluster endpoint if it exists
	if cluster != nil {
		patchBase := cluster.DeepCopy()
		cluster.Spec.ControlPlaneEndpoint = controlPlane.Spec.ControlPlaneEndpoint
		if err := r.Patch(ctx, cluster, client.MergeFrom(patchBase)); err != nil {
			logger.Error(err, "Failed to patch Cluster endpoint")
			return err
		}
	}

	return nil
}
