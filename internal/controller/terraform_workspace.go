package controller

import (
	"context"
	"fmt"

	infrastructurev1beta1 "github.com/appthrust/capt/api/v1beta1"
	tfv1beta1 "github.com/upbound/provider-terraform/apis/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func reconcileTerraformWorkspace(ctx context.Context, c client.Client, captCluster *infrastructurev1beta1.CAPTCluster, workspaceName types.NamespacedName) error {
	log := log.FromContext(ctx)

	// If networkRef is specified, ensure VPC workspace exists
	if captCluster.Spec.NetworkRef != nil {
		vpcTemplate := &infrastructurev1beta1.CAPTVPCTemplate{}
		vpcNamespacedName := types.NamespacedName{
			Name:      captCluster.Spec.NetworkRef.Name,
			Namespace: captCluster.Spec.NetworkRef.Namespace,
		}

		if err := c.Get(ctx, vpcNamespacedName, vpcTemplate); err != nil {
			if apierrors.IsNotFound(err) {
				log.Info("Referenced VPC template not found", "vpc", vpcNamespacedName)
				return fmt.Errorf("referenced VPC template not found: %s", vpcNamespacedName)
			}
			return fmt.Errorf("failed to get VPC template: %w", err)
		}

		vpcWorkspaceName := types.NamespacedName{
			Name:      fmt.Sprintf("%s-vpc-workspace", vpcTemplate.Name),
			Namespace: vpcTemplate.Namespace,
		}

		if err := reconcileVPCWorkspace(ctx, c, vpcTemplate, vpcWorkspaceName); err != nil {
			return fmt.Errorf("failed to reconcile VPC workspace: %w", err)
		}

		// Update CAPTCluster status with VPC workspace name
		captCluster.Status.NetworkWorkspaceName = vpcWorkspaceName.Name
		if err := c.Status().Update(ctx, captCluster); err != nil {
			return fmt.Errorf("failed to update CAPTCluster status with VPC workspace name: %w", err)
		}
	}

	workspace := &tfv1beta1.Workspace{}
	if err := c.Get(ctx, workspaceName, workspace); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Terraform Workspace not found. Creating a new one.")
			if err := createWorkspace(ctx, c, captCluster, workspaceName); err != nil {
				log.Error(err, "Failed to create Terraform Workspace")
				return err
			}
			log.Info("Created Terraform Workspace", "workspace", workspaceName)
		} else {
			log.Error(err, "Failed to get Terraform Workspace")
			return err
		}
	} else {
		log.Info("Updating existing Terraform Workspace", "workspace", workspaceName)
		if err := updateWorkspace(ctx, c, captCluster, workspace); err != nil {
			log.Error(err, "Failed to update Terraform Workspace")
			return err
		}
		log.Info("Updated Terraform Workspace", "workspace", workspaceName)
	}

	return nil
}

func reconcileVPCWorkspace(ctx context.Context, c client.Client, vpcTemplate *infrastructurev1beta1.CAPTVPCTemplate, workspaceName types.NamespacedName) error {
	log := log.FromContext(ctx)

	workspace := &tfv1beta1.Workspace{}
	if err := c.Get(ctx, workspaceName, workspace); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("VPC Terraform Workspace not found. Creating a new one.")
			if err := createVPCWorkspace(ctx, c, vpcTemplate, workspaceName); err != nil {
				log.Error(err, "Failed to create VPC Terraform Workspace")
				return err
			}
			log.Info("Created VPC Terraform Workspace", "workspace", workspaceName)
		} else {
			log.Error(err, "Failed to get VPC Terraform Workspace")
			return err
		}
	} else {
		log.Info("Updating existing VPC Terraform Workspace", "workspace", workspaceName)
		if err := updateVPCWorkspace(ctx, c, vpcTemplate, workspace); err != nil {
			log.Error(err, "Failed to update VPC Terraform Workspace")
			return err
		}
		log.Info("Updated VPC Terraform Workspace", "workspace", workspaceName)
	}

	return nil
}

func createVPCWorkspace(ctx context.Context, c client.Client, vpcTemplate *infrastructurev1beta1.CAPTVPCTemplate, name types.NamespacedName) error {
	hclCode, err := generateVPCWorkspaceModule(vpcTemplate)
	if err != nil {
		return fmt.Errorf("failed to generate VPC Terraform code: %w", err)
	}

	workspace := &tfv1beta1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name.Name,
			Namespace: name.Namespace,
		},
		Spec: tfv1beta1.WorkspaceSpec{
			ForProvider: tfv1beta1.WorkspaceParameters{
				Module: hclCode,
				Source: tfv1beta1.ModuleSourceInline,
				Vars: []tfv1beta1.Var{
					{
						Key:   "name",
						Value: vpcTemplate.Name,
					},
				},
			},
		},
	}

	return c.Create(ctx, workspace)
}

func updateVPCWorkspace(ctx context.Context, c client.Client, vpcTemplate *infrastructurev1beta1.CAPTVPCTemplate, workspace *tfv1beta1.Workspace) error {
	hclCode, err := generateVPCWorkspaceModule(vpcTemplate)
	if err != nil {
		return fmt.Errorf("failed to generate VPC Terraform code: %w", err)
	}

	workspace.Spec.ForProvider.Module = hclCode
	workspace.Spec.ForProvider.Vars = []tfv1beta1.Var{
		{
			Key:   "name",
			Value: vpcTemplate.Name,
		},
	}

	return c.Update(ctx, workspace)
}

func createWorkspace(ctx context.Context, c client.Client, cluster *infrastructurev1beta1.CAPTCluster, name types.NamespacedName) error {
	hclCode, err := generateStructuredTerraformCode(cluster)
	if err != nil {
		return fmt.Errorf("failed to generate Terraform code: %w", err)
	}

	workspace := &tfv1beta1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name.Name,
			Namespace: name.Namespace,
		},
		Spec: tfv1beta1.WorkspaceSpec{
			ForProvider: tfv1beta1.WorkspaceParameters{
				Module: string(hclCode),
				Source: tfv1beta1.ModuleSourceInline,
			},
		},
	}

	return c.Create(ctx, workspace)
}

func updateWorkspace(ctx context.Context, c client.Client, cluster *infrastructurev1beta1.CAPTCluster, workspace *tfv1beta1.Workspace) error {
	hclCode, err := generateStructuredTerraformCode(cluster)
	if err != nil {
		return fmt.Errorf("failed to generate Terraform code: %w", err)
	}

	workspace.Spec.ForProvider.Module = string(hclCode)
	return c.Update(ctx, workspace)
}

func deleteWorkspace(ctx context.Context, c client.Client, captCluster *infrastructurev1beta1.CAPTCluster) error {
	log := log.FromContext(ctx)

	// Delete VPC workspace if it exists
	if captCluster.Status.NetworkWorkspaceName != "" {
		vpcWorkspace := &tfv1beta1.Workspace{}
		vpcWorkspaceName := types.NamespacedName{
			Name:      captCluster.Status.NetworkWorkspaceName,
			Namespace: captCluster.Namespace,
		}

		if err := c.Get(ctx, vpcWorkspaceName, vpcWorkspace); err != nil {
			if !apierrors.IsNotFound(err) {
				return fmt.Errorf("failed to get VPC workspace: %w", err)
			}
		} else {
			if err := c.Delete(ctx, vpcWorkspace); err != nil {
				return fmt.Errorf("failed to delete VPC workspace: %w", err)
			}
			log.Info("Deleted VPC Terraform Workspace", "workspace", vpcWorkspaceName)
		}
	}

	// Delete cluster workspace
	workspace := &tfv1beta1.Workspace{}
	workspaceName := types.NamespacedName{
		Name:      fmt.Sprintf("%s-workspace", captCluster.Name),
		Namespace: captCluster.Namespace,
	}

	if err := c.Get(ctx, workspaceName, workspace); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Terraform Workspace not found. Skipping deletion.")
			return nil
		}
		return err
	}

	if err := c.Delete(ctx, workspace); err != nil {
		log.Error(err, "Failed to delete Terraform Workspace")
		return err
	}

	log.Info("Deleted Terraform Workspace", "workspace", workspaceName)
	return nil
}

func generateStructuredTerraformCode(cluster *infrastructurev1beta1.CAPTCluster) ([]byte, error) {
	config := generateTerraformConfig(cluster)

	jsonData, err := convertToJSON(config)
	if err != nil {
		return nil, fmt.Errorf("failed to convert config to JSON: %w", err)
	}

	hclCode, err := convertJSONToHCL(jsonData)
	if err != nil {
		return nil, fmt.Errorf("failed to convert JSON to HCL: %w", err)
	}

	return hclCode, nil
}

// New function to switch between configurations
func generateConfigBasedOnCluster(cluster *infrastructurev1beta1.CAPTCluster) (interface{}, error) {
	// You can add a condition here to determine which config to use
	// For example, based on a feature flag or cluster specification
	useEKSConfig := false // This should be determined based on your requirements

	if useEKSConfig {
		return generateEKSTerraformConfig(cluster), nil
	}
	return generateTerraformConfig(cluster), nil
}
