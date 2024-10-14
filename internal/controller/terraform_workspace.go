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
