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

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
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
	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()

	// Add terraform block
	tfBlock := rootBody.AppendNewBlock("terraform", nil)
	tfBody := tfBlock.Body()
	tfBody.SetAttributeValue("required_version", cty.StringVal("~> 1.0"))

	// Add required_providers block
	reqProvidersBlock := tfBody.AppendNewBlock("required_providers", nil)
	reqProvidersBody := reqProvidersBlock.Body()
	awsBlock := reqProvidersBody.AppendNewBlock("aws", nil)
	awsBody := awsBlock.Body()
	awsBody.SetAttributeValue("source", cty.StringVal("hashicorp/aws"))
	awsBody.SetAttributeValue("version", cty.StringVal("~> 4.0"))

	// Add provider block
	providerBlock := rootBody.AppendNewBlock("provider", []string{"aws"})
	providerBody := providerBlock.Body()
	providerBody.SetAttributeValue("region", cty.StringVal(cluster.Spec.Region))

	// Add VPC resource
	vpcBlock := rootBody.AppendNewBlock("resource", []string{"aws_vpc", "main"})
	vpcBody := vpcBlock.Body()
	vpcBody.SetAttributeValue("cidr_block", cty.StringVal(cluster.Spec.VPC.CIDR))
	vpcBody.SetAttributeValue("enable_dns_hostnames", cty.BoolVal(true))
	vpcBody.SetAttributeValue("enable_dns_support", cty.BoolVal(true))

	// Add EKS cluster resource
	eksBlock := rootBody.AppendNewBlock("resource", []string{"aws_eks_cluster", "main"})
	eksBody := eksBlock.Body()
	eksBody.SetAttributeValue("name", cty.StringVal(cluster.Name))
	eksBody.SetAttributeValue("version", cty.StringVal(cluster.Spec.EKS.Version))
	eksBody.SetAttributeValue("role_arn", cty.StringVal("${aws_iam_role.eks_cluster_role.arn}"))

	vpcConfigBlock := eksBody.AppendNewBlock("vpc_config", nil)
	vpcConfigBody := vpcConfigBlock.Body()
	vpcConfigBody.SetAttributeValue("endpoint_private_access", cty.BoolVal(cluster.Spec.EKS.PrivateAccess))
	vpcConfigBody.SetAttributeValue("endpoint_public_access", cty.BoolVal(cluster.Spec.EKS.PublicAccess))
	// Note: Subnet IDs should be added here, but they're not directly available in the CAPTClusterSpec

	// Add IAM role for EKS cluster
	iamRoleBlock := rootBody.AppendNewBlock("resource", []string{"aws_iam_role", "eks_cluster_role"})
	iamRoleBody := iamRoleBlock.Body()
	iamRoleBody.SetAttributeValue("name", cty.StringVal(fmt.Sprintf("%s-eks-cluster-role", cluster.Name)))
	iamRoleBody.SetAttributeValue("assume_role_policy", cty.StringVal(`{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Principal": {
					"Service": "eks.amazonaws.com"
				},
				"Action": "sts:AssumeRole"
			}
		]
	}`))

	// Add EKS node groups
	for _, ng := range cluster.Spec.EKS.NodeGroups {
		ngBlock := rootBody.AppendNewBlock("resource", []string{"aws_eks_node_group", ng.Name})
		ngBody := ngBlock.Body()
		ngBody.SetAttributeValue("cluster_name", cty.StringVal(cluster.Name))
		ngBody.SetAttributeValue("node_group_name", cty.StringVal(ng.Name))
		ngBody.SetAttributeValue("node_role_arn", cty.StringVal("${aws_iam_role.eks_node_role.arn}"))
		// Note: Subnet IDs should be added here, but they're not directly available in the CAPTClusterSpec

		scalingConfigBlock := ngBody.AppendNewBlock("scaling_config", nil)
		scalingConfigBody := scalingConfigBlock.Body()
		scalingConfigBody.SetAttributeValue("desired_size", cty.NumberIntVal(int64(ng.DesiredSize)))
		scalingConfigBody.SetAttributeValue("max_size", cty.NumberIntVal(int64(ng.MaxSize)))
		scalingConfigBody.SetAttributeValue("min_size", cty.NumberIntVal(int64(ng.MinSize)))

		ngBody.SetAttributeValue("instance_types", cty.ListVal([]cty.Value{cty.StringVal(ng.InstanceType)}))
	}

	// Add IAM role for EKS nodes
	nodeIamRoleBlock := rootBody.AppendNewBlock("resource", []string{"aws_iam_role", "eks_node_role"})
	nodeIamRoleBody := nodeIamRoleBlock.Body()
	nodeIamRoleBody.SetAttributeValue("name", cty.StringVal(fmt.Sprintf("%s-eks-node-role", cluster.Name)))
	nodeIamRoleBody.SetAttributeValue("assume_role_policy", cty.StringVal(`{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Principal": {
					"Service": "ec2.amazonaws.com"
				},
				"Action": "sts:AssumeRole"
			}
		]
	}`))

	// Add other necessary resources and data sources
	// TODO: Add VPC subnets, security groups, and other required resources

	return f.Bytes(), nil
}
