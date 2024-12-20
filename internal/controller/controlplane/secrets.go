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

//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch

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
		_, setErr := r.setFailedStatus(ctx, controlPlane, cluster, ReasonEndpointError, fmt.Sprintf("Failed to get workspace: %v", err))
		if setErr != nil {
			logger.Error(setErr, "Failed to set status")
		}
		return nil
	}

	// Get connection secret
	secret := &corev1.Secret{}
	secretRef := workspaceApply.Spec.WriteConnectionSecretToRef
	if secretRef == nil {
		err := fmt.Errorf("workspace connection secret reference is not set")
		logger.Error(err, "Failed to get secret reference")
		_, setErr := r.setFailedStatus(ctx, controlPlane, cluster, ReasonSecretError, err.Error())
		if setErr != nil {
			logger.Error(setErr, "Failed to set status")
		}
		return nil
	}

	if err := r.Get(ctx, client.ObjectKey{
		Name:      secretRef.Name,
		Namespace: secretRef.Namespace,
	}, secret); err != nil {
		if !apierrors.IsNotFound(err) {
			logger.Error(err, "Failed to get secret")
			_, setErr := r.setFailedStatus(ctx, controlPlane, cluster, ReasonSecretError, fmt.Sprintf("Failed to get secret: %v", err))
			if setErr != nil {
				logger.Error(setErr, "Failed to set status")
			}
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
		_, setErr := r.setFailedStatus(ctx, controlPlane, cluster, ReasonEndpointError, fmt.Sprintf("Failed to get cluster endpoint: %v", err))
		if setErr != nil {
			logger.Error(setErr, "Failed to set status")
		}
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
		_, setErr := r.setFailedStatus(ctx, controlPlane, cluster, ReasonSecretError, fmt.Sprintf("Failed to get certificate authority data: %v", err))
		if setErr != nil {
			logger.Error(setErr, "Failed to set status")
		}
		return nil
	}

	// Create CA secret
	if err := r.reconcileCASecret(ctx, controlPlane, caData); err != nil {
		logger.Error(err, "Failed to reconcile CA secret")
		return err
	}

	// Create kubeconfig secret
	if err := r.reconcileKubeconfigSecret(ctx, controlPlane, cluster); err != nil {
		logger.Error(err, "Failed to reconcile kubeconfig secret")
		return err
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

// reconcileCASecret creates or updates the CA secret
func (r *Reconciler) reconcileCASecret(ctx context.Context, controlPlane *controlplanev1beta1.CAPTControlPlane, caData string) error {
	logger := log.FromContext(ctx)

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
			return err
		}
		caSecretExists = false
	}

	// Create CA secret if it doesn't exist
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
			return err
		}

		if err := r.Create(ctx, caSecret); err != nil {
			logger.Error(err, "Failed to create CA secret")
			return err
		}
		logger.Info("Created CA secret")
	}

	return nil
}

// reconcileKubeconfigSecret creates or updates the kubeconfig secret
func (r *Reconciler) reconcileKubeconfigSecret(ctx context.Context, controlPlane *controlplanev1beta1.CAPTControlPlane, cluster *clusterv1.Cluster) error {
	logger := log.FromContext(ctx)

	// Get outputs secret first
	outputsSecretName := fmt.Sprintf("%s-outputs-kubeconfig", cluster.Name)
	outputsSecret := &corev1.Secret{}
	if err := r.Get(ctx, client.ObjectKey{
		Name:      outputsSecretName,
		Namespace: "default", // outputs-kubeconfigはdefaultネームスペースにある
	}, outputsSecret); err != nil {
		if !apierrors.IsNotFound(err) {
			logger.Error(err, "Failed to get outputs secret")
			return err
		}
		logger.Info("Waiting for outputs secret to be created", "name", outputsSecretName)
		return nil
	}

	// Prepare kubeconfig secret
	kubeconfigSecretName := fmt.Sprintf("%s-kubeconfig", cluster.Name)
	kubeconfigSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kubeconfigSecretName,
			Namespace: controlPlane.Namespace,
			Labels: map[string]string{
				"cluster.x-k8s.io/cluster-name": cluster.Name,
			},
		},
		Type: "cluster.x-k8s.io/secret",
		Data: map[string][]byte{
			"value": outputsSecret.Data["kubeconfig"],
		},
	}

	// Set controller reference
	if err := controllerutil.SetControllerReference(controlPlane, kubeconfigSecret, r.Scheme); err != nil {
		logger.Error(err, "Failed to set controller reference for kubeconfig secret")
		return err
	}

	// Create or update kubeconfig secret
	existingKubeconfigSecret := &corev1.Secret{}
	err := r.Get(ctx, client.ObjectKey{
		Name:      kubeconfigSecretName,
		Namespace: controlPlane.Namespace,
	}, existingKubeconfigSecret)

	if err != nil {
		if !apierrors.IsNotFound(err) {
			logger.Error(err, "Failed to get existing kubeconfig secret")
			return err
		}
		// Create new secret
		if err := r.Create(ctx, kubeconfigSecret); err != nil {
			logger.Error(err, "Failed to create kubeconfig secret")
			return err
		}
		logger.Info("Created kubeconfig secret")
	} else {
		// Update existing secret
		existingKubeconfigSecret.Data = kubeconfigSecret.Data
		existingKubeconfigSecret.Labels = kubeconfigSecret.Labels
		if err := r.Update(ctx, existingKubeconfigSecret); err != nil {
			logger.Error(err, "Failed to update kubeconfig secret")
			return err
		}
		logger.Info("Updated kubeconfig secret")
	}

	return nil
}
