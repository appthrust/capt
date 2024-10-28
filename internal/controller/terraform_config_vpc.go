package controller

import (
	"fmt"

	infrastructurev1beta1 "github.com/appthrust/capt/api/v1beta1"
	"github.com/appthrust/capt/internal/tf_module/vpc"
)

func generateVPCWorkspaceModule(vpcTemplate *infrastructurev1beta1.CAPTVPCTemplate) (string, error) {
	// Create VPC config builder with default values
	vpcConfigBuilder := vpc.NewVPCConfig()

	// Set VPC module version to latest
	vpcConfigBuilder.SetVersion("~> 5.0")

	// Configure VPC settings from template
	vpcConfigBuilder.
		SetName(vpcTemplate.Name).
		SetCIDR(vpcTemplate.Spec.CIDR).
		SetEnableNATGateway(vpcTemplate.Spec.EnableNatGateway).
		SetSingleNATGateway(vpcTemplate.Spec.SingleNatGateway)

	// Set AZs using dynamic expression
	vpcConfigBuilder.SetAZsExpression("local.azs")

	// Set subnets using dynamic expressions based on CIDR
	vpcConfigBuilder.SetPrivateSubnetsExpression(
		fmt.Sprintf("[for k, v in local.azs : cidrsubnet(\"%s\", 4, k)]", vpcTemplate.Spec.CIDR),
	)
	vpcConfigBuilder.SetPublicSubnetsExpression(
		fmt.Sprintf("[for k, v in local.azs : cidrsubnet(\"%s\", 8, k + 48)]", vpcTemplate.Spec.CIDR),
	)

	// Set subnet tags
	if vpcTemplate.Spec.PublicSubnetTags != nil {
		for key, value := range vpcTemplate.Spec.PublicSubnetTags {
			vpcConfigBuilder.AddPublicSubnetTag(key, value)
		}
	}
	if vpcTemplate.Spec.PrivateSubnetTags != nil {
		// Create a new map to avoid modifying the original
		privateTags := make(map[string]string)
		for k, v := range vpcTemplate.Spec.PrivateSubnetTags {
			// Handle special case for CLUSTER_NAME variable
			if v == "${CLUSTER_NAME}" {
				privateTags[k] = "${local.name}"
			} else {
				privateTags[k] = v
			}
		}
		vpcConfigBuilder.SetPrivateSubnetTags(privateTags)
	}

	// Set VPC tags
	if vpcTemplate.Spec.Tags != nil {
		for key, value := range vpcTemplate.Spec.Tags {
			vpcConfigBuilder.AddTag(key, value)
		}
	} else {
		// Set default tags if none are provided
		vpcConfigBuilder.
			AddTag("Environment", "dev").
			AddTag("Terraform", "true")
	}

	// Build VPC config
	vpcConfig, err := vpcConfigBuilder.Build()
	if err != nil {
		return "", fmt.Errorf("failed to build VPC config: %w", err)
	}

	// Generate HCL
	hcl, err := vpcConfig.GenerateHCL()
	if err != nil {
		return "", fmt.Errorf("failed to generate HCL: %w", err)
	}

	// Add required data sources and locals
	dataSources := `
data "aws_availability_zones" "available" {
  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

locals {
  azs = slice(data.aws_availability_zones.available.names, 0, 3)
  name = try(var.name, basename(path.cwd))
  tags = {
    Module     = basename(path.cwd)
    GithubRepo = "github.com/labthrust/terraform-aws"
  }
}

variable "name" {
  type        = string
  description = "Name of the VPC"
}
`

	return dataSources + hcl, nil
}
