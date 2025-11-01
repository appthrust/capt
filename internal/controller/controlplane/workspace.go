package controlplane

import (
	"context"
	"fmt"
	"strings"

	controlplanev1beta1 "github.com/appthrust/capt/api/controlplane/v1beta1"
	infrastructurev1beta1 "github.com/appthrust/capt/api/v1beta1"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// reconcileWorkspace handles the reconciliation of WorkspaceTemplate and WorkspaceTemplateApply
func (r *Reconciler) reconcileWorkspace(
	ctx context.Context,
	controlPlane *controlplanev1beta1.CAPTControlPlane,
	cluster *clusterv1.Cluster,
) (ctrl.Result, error) {
	// Get the referenced WorkspaceTemplate
	workspaceTemplate := &infrastructurev1beta1.WorkspaceTemplate{}
	templateNamespacedName := types.NamespacedName{
		Name:      controlPlane.Spec.WorkspaceTemplateRef.Name,
		Namespace: controlPlane.Spec.WorkspaceTemplateRef.Namespace,
	}
	if err := r.Get(ctx, templateNamespacedName, workspaceTemplate); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get WorkspaceTemplate: %v", err)
	}

	// Get or create WorkspaceTemplateApply
	workspaceApply, err := r.getOrCreateWorkspaceTemplateApply(ctx, controlPlane)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Update status based on WorkspaceTemplateApply conditions
	// Note: updateStatus now avoids periodic requeue when ready to prevent loops
	return r.updateStatus(ctx, controlPlane, workspaceApply, cluster)
}

// getOrCreateWorkspaceTemplateApply gets an existing WorkspaceTemplateApply or creates a new one
func (r *Reconciler) getOrCreateWorkspaceTemplateApply(
	ctx context.Context,
	controlPlane *controlplanev1beta1.CAPTControlPlane,
) (*infrastructurev1beta1.WorkspaceTemplateApply, error) {
	// Determine the name for WorkspaceTemplateApply
	applyName := controlPlane.Spec.WorkspaceTemplateApplyName
	if applyName == "" {
		applyName = fmt.Sprintf("%s-eks-controlplane-apply", controlPlane.Name)
		// Update WorkspaceTemplateApplyName in Spec first
		controlPlaneCopy := controlPlane.DeepCopy()
		controlPlaneCopy.Spec.WorkspaceTemplateApplyName = applyName
		if err := r.Update(ctx, controlPlaneCopy); err != nil {
			return nil, fmt.Errorf("failed to update WorkspaceTemplateApplyName in spec: %v", err)
		}
		// Update the original object
		controlPlane.Spec.WorkspaceTemplateApplyName = applyName
	}

	// Try to find existing WorkspaceTemplateApply
	workspaceApply := &infrastructurev1beta1.WorkspaceTemplateApply{}
	err := r.Get(ctx, types.NamespacedName{Name: applyName, Namespace: controlPlane.Namespace}, workspaceApply)
	if err == nil {
		// Update existing WorkspaceTemplateApply
		workspaceApply.Spec = r.generateWorkspaceTemplateApplySpec(controlPlane)
		if err := r.Update(ctx, workspaceApply); err != nil {
			return nil, fmt.Errorf("failed to update WorkspaceTemplateApply: %v", err)
		}
		return workspaceApply, nil
	}

	if !apierrors.IsNotFound(err) {
		return nil, err
	}

	// Create new WorkspaceTemplateApply
	workspaceApply = &infrastructurev1beta1.WorkspaceTemplateApply{
		ObjectMeta: metav1.ObjectMeta{
			Name:      applyName,
			Namespace: controlPlane.Namespace,
		},
		Spec: r.generateWorkspaceTemplateApplySpec(controlPlane),
	}

	// Set owner reference
	if err := controllerutil.SetControllerReference(controlPlane, workspaceApply, r.Scheme); err != nil {
		return nil, fmt.Errorf("failed to set controller reference: %v", err)
	}

	if err := r.Create(ctx, workspaceApply); err != nil {
		return nil, fmt.Errorf("failed to create WorkspaceTemplateApply: %v", err)
	}

	return workspaceApply, nil
}

// generateWorkspaceTemplateApplySpec generates the spec for a WorkspaceTemplateApply
func (r *Reconciler) generateWorkspaceTemplateApplySpec(controlPlane *controlplanev1beta1.CAPTControlPlane) infrastructurev1beta1.WorkspaceTemplateApplySpec {
	spec := infrastructurev1beta1.WorkspaceTemplateApplySpec{
		TemplateRef: infrastructurev1beta1.WorkspaceTemplateReference{
			Name:      controlPlane.Spec.WorkspaceTemplateRef.Name,
			Namespace: controlPlane.Spec.WorkspaceTemplateRef.Namespace,
		},
		Variables: map[string]string{
			"cluster_name":       controlPlane.Name,
			"kubernetes_version": controlPlane.Spec.Version,
		},
		WriteConnectionSecretToRef: &xpv1.SecretReference{
			Name:      fmt.Sprintf("%s-eks-connection", controlPlane.Name),
			Namespace: controlPlane.Namespace,
		},
	}

	// Add region from ControlPlaneConfig
	if controlPlane.Spec.ControlPlaneConfig != nil {
		spec.Variables["region"] = controlPlane.Spec.ControlPlaneConfig.Region
	}

	// Add endpoint access configuration if specified
	if controlPlane.Spec.ControlPlaneConfig != nil && controlPlane.Spec.ControlPlaneConfig.EndpointAccess != nil {
		spec.Variables["endpoint_public_access"] = fmt.Sprintf("%v", controlPlane.Spec.ControlPlaneConfig.EndpointAccess.Public)
		spec.Variables["endpoint_private_access"] = fmt.Sprintf("%v", controlPlane.Spec.ControlPlaneConfig.EndpointAccess.Private)
	}

	// Add additional tags if specified
	if len(controlPlane.Spec.AdditionalTags) > 0 {
		for k, v := range controlPlane.Spec.AdditionalTags {
			spec.Variables[fmt.Sprintf("tags_%s", k)] = v
		}
	}

	// Add VPC workspace dependency: use owner cluster name if available
	// Determine associated CAPTCluster to derive the actual VPC WorkspaceTemplateApply name
	clusterName := controlPlane.Labels[clusterv1.ClusterNameLabel]
	// Prefer owner cluster name for cluster_name variable when available
	if clusterName != "" {
		spec.Variables["cluster_name"] = clusterName
	}
	vpcWorkspaceApplyName := ""
	if clusterName != "" {
		// List CAPTClusters owned by this cluster
		captClusters := &infrastructurev1beta1.CAPTClusterList{}
		if err := r.List(context.Background(), captClusters, client.InNamespace(controlPlane.Namespace), client.MatchingLabels{clusterv1.ClusterNameLabel: clusterName}); err == nil {
			if len(captClusters.Items) > 0 {
				cc := captClusters.Items[0]
				if cc.Spec.WorkspaceTemplateApplyName != "" {
					vpcWorkspaceApplyName = cc.Spec.WorkspaceTemplateApplyName
				} else {
					vpcWorkspaceApplyName = fmt.Sprintf("%s-vpc", cc.Name)
				}
			}
		}
	}
	if vpcWorkspaceApplyName == "" {
		// Fallback to controlPlane-based naming (best-effort)
		owner := controlPlane.Name
		if clusterName != "" {
			owner = clusterName
		}
		vpcWorkspaceApplyName = fmt.Sprintf("%s-vpc", owner)
	}

	// Wait for the VPC Workspace to be ready (its name equals applyName without "-apply")
	vpcWorkspaceName := strings.TrimSuffix(vpcWorkspaceApplyName, "-apply")
	spec.WaitForWorkspaces = []infrastructurev1beta1.WorkspaceReference{{
		Name:      vpcWorkspaceName,
		Namespace: controlPlane.Namespace,
	}}

	// Provide VPC workspace name to the template and wait for its connection secret as a required secret
	spec.Variables["vpc_workspace_name"] = vpcWorkspaceName
	vpcSecretName := fmt.Sprintf("%s-vpc-connection", vpcWorkspaceName)
	spec.WaitForSecrets = append(spec.WaitForSecrets, xpv1.SecretReference{
		Name:      vpcSecretName,
		Namespace: controlPlane.Namespace,
	})

	return spec
}
