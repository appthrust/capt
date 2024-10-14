package controller

import (
	"fmt"

	infrastructurev1beta1 "github.com/appthrust/capt/api/v1beta1"
)

func generateTerraformCode(cluster *infrastructurev1beta1.CAPTCluster) string {
	return fmt.Sprintf(`
module "eks" {
  source  = "terraform-aws-modules/eks/aws"
  version = "~> 19.0"

  cluster_name    = "%s"
  cluster_version = "%s"

  vpc_id     = module.vpc.vpc_id
  subnet_ids = module.vpc.private_subnets

  cluster_endpoint_public_access  = %t
  cluster_endpoint_private_access = %t

  # Add other EKS configurations based on CAPTCluster spec
  %s
}

module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "~> 5.0"

  name = "%s-vpc"
  cidr = "%s"

  azs             = ["${data.aws_availability_zones.available.names[0]}", "${data.aws_availability_zones.available.names[1]}", "${data.aws_availability_zones.available.names[2]}"]
  private_subnets = [for k, v in data.aws_availability_zones.available.names : cidrsubnet(var.vpc_cidr, 4, k)]
  public_subnets  = [for k, v in data.aws_availability_zones.available.names : cidrsubnet(var.vpc_cidr, 8, k+48)]

  enable_nat_gateway   = %t
  single_nat_gateway   = %t
  enable_dns_hostnames = true

  public_subnet_tags = %s
  private_subnet_tags = %s
}

data "aws_availability_zones" "available" {}

variable "vpc_cidr" {
  default = "%s"
}
`, cluster.Name, cluster.Spec.EKS.Version,
		cluster.Spec.EKS.PublicAccess,
		cluster.Spec.EKS.PrivateAccess,
		generateNodeGroupsConfig(cluster.Spec.EKS.NodeGroups),
		cluster.Name, cluster.Spec.VPC.CIDR,
		cluster.Spec.VPC.EnableNatGateway,
		cluster.Spec.VPC.SingleNatGateway,
		formatTags(cluster.Spec.VPC.PublicSubnetTags),
		formatTags(cluster.Spec.VPC.PrivateSubnetTags),
		cluster.Spec.VPC.CIDR)
}

func generateNodeGroupsConfig(nodeGroups []infrastructurev1beta1.NodeGroupConfig) string {
	if len(nodeGroups) == 0 {
		return ""
	}

	result := "eks_managed_node_groups = {\n"
	for _, ng := range nodeGroups {
		result += fmt.Sprintf(`    %s = {
      instance_types = ["%s"]
      min_size     = %d
      max_size     = %d
      desired_size = %d
    }
`, ng.Name, ng.InstanceType, ng.MinSize, ng.MaxSize, ng.DesiredSize)
	}
	result += "  }"
	return result
}

func formatTags(tags map[string]string) string {
	if len(tags) == 0 {
		return "{}"
	}

	result := "{\n"
	for k, v := range tags {
		result += fmt.Sprintf("    %s = \"%s\"\n", k, v)
	}
	result += "  }"

	return result
}
