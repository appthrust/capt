package controlplane

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	controlplanev1beta1 "github.com/appthrust/capt/api/controlplane/v1beta1"
	infrastructurev1beta1 "github.com/appthrust/capt/api/v1beta1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// reconcileEKSOutputsStatus collects EKS connection outputs from the Workspace connection Secret
// and updates CAPTControlPlane status fields without modifying the legacy secrets flow.
func (r *Reconciler) reconcileEKSOutputsStatus(
	ctx context.Context,
	controlPlane *controlplanev1beta1.CAPTControlPlane,
	workspaceApply *infrastructurev1beta1.WorkspaceTemplateApply,
) error {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling EKS outputs status (non-intrusive)")

	// Workspace must exist to resolve accurate secret reference
	if workspaceApply == nil || workspaceApply.Status.WorkspaceName == "" {
		return nil
	}

	// Get workspace (unstructured)
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
		// Best-effort; do not fail reconciliation chain
		logger.Error(err, "Failed to get workspace for EKS outputs status")
		return nil
	}

	// Resolve connection secret name (compatibility-first)
	var resolvedName, resolvedNamespace string
	if workspaceApply.Spec.WriteConnectionSecretToRef != nil {
		resolvedName = workspaceApply.Spec.WriteConnectionSecretToRef.Name
		resolvedNamespace = workspaceApply.Spec.WriteConnectionSecretToRef.Namespace
	}
	// Fallback to Workspace spec reference if not set
	if resolvedName == "" {
		if wsSecName, found, _ := unstructured.NestedString(workspace.Object, "spec", "writeConnectionSecretToRef", "name"); found && wsSecName != "" {
			resolvedName = wsSecName
			if wsSecNs, foundNs, _ := unstructured.NestedString(workspace.Object, "spec", "writeConnectionSecretToRef", "namespace"); foundNs && wsSecNs != "" {
				resolvedNamespace = wsSecNs
			} else {
				resolvedNamespace = workspace.GetNamespace()
			}
		}
	}
	// Derive candidates if still unknown
	candidates := []string{}
	if resolvedName == "" {
		wsName := workspace.GetName()
		if strings.HasSuffix(wsName, "-eks-controlplane") {
			candidates = append(candidates, strings.TrimSuffix(wsName, "-eks-controlplane")+"-eks-connection")
		}
		if strings.HasSuffix(wsName, "-controlplane") {
			candidates = append(candidates, strings.TrimSuffix(wsName, "-controlplane")+"-connection")
		}
		candidates = append(candidates, fmt.Sprintf("%s-eks-connection", controlPlane.Name))
		resolvedNamespace = workspace.GetNamespace()
	}

	// Try to get the secret
	secret := &corev1.Secret{}
	var getErr error
	if resolvedName != "" {
		getErr = r.Get(ctx, client.ObjectKey{
			Name:      resolvedName,
			Namespace: resolvedNamespace,
		}, secret)
	} else {
		// Probe candidates in order
		for _, name := range candidates {
			try := &corev1.Secret{}
			if err := r.Get(ctx, client.ObjectKey{Name: name, Namespace: resolvedNamespace}, try); err == nil {
				secret = try
				resolvedName = name
				getErr = nil
				break
			} else {
				getErr = err
			}
		}
	}

	// Handle not found: record ref and exit gracefully
	if apierrors.IsNotFound(getErr) || (resolvedName == "" && getErr == nil && secret.Name == "") {
		patchBase := controlPlane.DeepCopy()
		// Prefer showing first candidate if any, otherwise keep last resolved
		refName := resolvedName
		if refName == "" && len(candidates) > 0 {
			refName = candidates[0]
		}
		controlPlane.Status.WorkspaceOutputsRef = &controlplanev1beta1.WorkspaceOutputsRef{
			Name:      refName,
			Namespace: resolvedNamespace,
		}
		meta.SetStatusCondition(&controlPlane.Status.Conditions, metav1.Condition{
			Type:    controlplanev1beta1.EKSOutputsReadyCondition,
			Status:  metav1.ConditionFalse,
			Reason:  "NotFound",
			Message: "Waiting for workspace connection secret to be created",
		})
		_ = r.Status().Patch(ctx, controlPlane, client.MergeFrom(patchBase))
		return nil
	}
	if getErr != nil {
		// Transient error; log and exit
		logger.Error(getErr, "Failed to get EKS connection secret")
		return nil
	}

	// Build outputs and compute checksum
	outputs, checksum, err := buildEKSOutputsFromSecret(secret)
	if err != nil {
		patchBase := controlPlane.DeepCopy()
		meta.SetStatusCondition(&controlPlane.Status.Conditions, metav1.Condition{
			Type:    controlplanev1beta1.EKSOutputsReadyCondition,
			Status:  metav1.ConditionFalse,
			Reason:  "ParseError",
			Message: fmt.Sprintf("Failed to parse EKS outputs: %v", err),
		})
		_ = r.Status().Patch(ctx, controlPlane, client.MergeFrom(patchBase))
		return nil
	}

	// Patch status when changed
	patchBase := controlPlane.DeepCopy()
	if controlPlane.Status.EKSOutputsChecksum != checksum || controlPlane.Status.EKSOutputs == nil {
		controlPlane.Status.EKSOutputs = outputs
		controlPlane.Status.EKSOutputsChecksum = checksum
	}
	controlPlane.Status.WorkspaceOutputsRef = &controlplanev1beta1.WorkspaceOutputsRef{
		Name:      resolvedName,
		Namespace: resolvedNamespace,
	}
	meta.SetStatusCondition(&controlPlane.Status.Conditions, metav1.Condition{
		Type:    controlplanev1beta1.EKSOutputsReadyCondition,
		Status:  metav1.ConditionTrue,
		Reason:  "Success",
		Message: "EKS outputs parsed and stored",
	})
	if err := r.Status().Patch(ctx, controlPlane, client.MergeFrom(patchBase)); err != nil {
		logger.Error(err, "Failed to patch EKS outputs status")
	}

	return nil
}

// buildEKSOutputsFromSecret extracts and normalizes EKS outputs from the connection Secret.
func buildEKSOutputsFromSecret(s *corev1.Secret) (*controlplanev1beta1.EKSOutputsStatus, string, error) {
	get := func(key string) (string, bool) {
		if s == nil || s.Data == nil {
			return "", false
		}
		b, ok := s.Data[key]
		return string(b), ok
	}

	out := &controlplanev1beta1.EKSOutputsStatus{}

	if v, ok := get("cluster_name"); ok {
		out.ClusterName = v
	}
	if v, ok := get("cluster_endpoint"); ok {
		out.ClusterEndpoint = v
	}
	if v, ok := get("cluster_certificate_authority_data"); ok {
		out.ClusterCertificateAuthorityData = v
	}
	if v, ok := get("oidc_provider"); ok {
		out.OIDCProvider = v
	}
	if v, ok := get("oidc_provider_arn"); ok {
		out.OIDCProviderARN = v
	}

	if v, ok := get("external_dns"); ok && len(v) > 0 {
		var tmp struct {
			IAMRoleARN     string `json:"iam_role_arn"`
			ServiceAccount struct {
				Name      string `json:"name"`
				Namespace string `json:"namespace"`
			} `json:"service_account"`
		}
		if err := json.Unmarshal([]byte(v), &tmp); err != nil {
			return nil, "", fmt.Errorf("parse external_dns: %w", err)
		}
		out.ExternalDNS = &controlplanev1beta1.ExternalDNSStatus{
			IAMRoleARN: tmp.IAMRoleARN,
			ServiceAccount: &controlplanev1beta1.NamespacedNameStatus{
				Name:      tmp.ServiceAccount.Name,
				Namespace: tmp.ServiceAccount.Namespace,
			},
		}
	}

	if v, ok := get("karpenter"); ok && len(v) > 0 {
		var tmp struct {
			DiscoveryTag struct {
				Key   string `json:"key"`
				Value string `json:"value"`
			} `json:"discovery_tag"`
			EC2NodeClass struct {
				Role string `json:"role"`
			} `json:"ec2_node_class"`
			QueueName      string `json:"queue_name"`
			ServiceAccount struct {
				Annotations map[string]string `json:"annotations"`
			} `json:"service_account"`
		}
		if err := json.Unmarshal([]byte(v), &tmp); err != nil {
			return nil, "", fmt.Errorf("parse karpenter: %w", err)
		}
		out.Karpenter = &controlplanev1beta1.KarpenterStatus{
			DiscoveryTag: &controlplanev1beta1.KVPStatus{Key: tmp.DiscoveryTag.Key, Value: tmp.DiscoveryTag.Value},
			EC2NodeClass: &controlplanev1beta1.KarpenterNodeClass{Role: tmp.EC2NodeClass.Role},
			QueueName:    tmp.QueueName,
			ServiceAccount: &controlplanev1beta1.ServiceAccountStatus{
				Annotations: tmp.ServiceAccount.Annotations,
			},
		}
	}

	normalized, _ := json.Marshal(out)
	sum := sha256.Sum256(normalized)
	return out, hex.EncodeToString(sum[:]), nil
}
