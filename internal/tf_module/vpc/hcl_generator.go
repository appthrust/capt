package vpc

import (
	"bytes"
	"fmt"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

// GenerateHCL converts the VPCConfig to Terraform HCL
func (c *VPCConfig) GenerateHCL() (string, error) {
	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()

	// Add module block for VPC
	vpcBlock := rootBody.AppendNewBlock("module", []string{"vpc"})
	vpcBody := vpcBlock.Body()

	vpcBody.SetAttributeValue("source", cty.StringVal("terraform-aws-modules/vpc/aws"))
	vpcBody.SetAttributeValue("version", cty.StringVal("~> 5.0"))
	vpcBody.SetAttributeValue("name", cty.StringVal(c.Name))
	vpcBody.SetAttributeValue("cidr", cty.StringVal(c.CIDR))

	azs := make([]cty.Value, len(c.AZs))
	for i, az := range c.AZs {
		azs[i] = cty.StringVal(az)
	}
	vpcBody.SetAttributeValue("azs", cty.ListVal(azs))

	if len(c.PrivateSubnets) > 0 {
		privateSubnets := make([]cty.Value, len(c.PrivateSubnets))
		for i, subnet := range c.PrivateSubnets {
			privateSubnets[i] = cty.StringVal(subnet)
		}
		vpcBody.SetAttributeValue("private_subnets", cty.ListVal(privateSubnets))
	}

	if len(c.PublicSubnets) > 0 {
		publicSubnets := make([]cty.Value, len(c.PublicSubnets))
		for i, subnet := range c.PublicSubnets {
			publicSubnets[i] = cty.StringVal(subnet)
		}
		vpcBody.SetAttributeValue("public_subnets", cty.ListVal(publicSubnets))
	}

	vpcBody.SetAttributeValue("enable_nat_gateway", cty.BoolVal(c.EnableNATGateway))
	vpcBody.SetAttributeValue("single_nat_gateway", cty.BoolVal(c.SingleNATGateway))

	if len(c.PublicSubnetTags) > 0 {
		publicSubnetTagsBlock := vpcBody.AppendNewBlock("public_subnet_tags", nil)
		publicSubnetTagsBody := publicSubnetTagsBlock.Body()
		for key, value := range c.PublicSubnetTags {
			publicSubnetTagsBody.SetAttributeValue(key, cty.StringVal(value))
		}
	}

	if len(c.PrivateSubnetTags) > 0 {
		privateSubnetTagsBlock := vpcBody.AppendNewBlock("private_subnet_tags", nil)
		privateSubnetTagsBody := privateSubnetTagsBlock.Body()
		for key, value := range c.PrivateSubnetTags {
			privateSubnetTagsBody.SetAttributeValue(key, cty.StringVal(value))
		}
	}

	if len(c.Tags) > 0 {
		tagsBlock := vpcBody.AppendNewBlock("tags", nil)
		tagsBody := tagsBlock.Body()
		for key, value := range c.Tags {
			tagsBody.SetAttributeValue(key, cty.StringVal(value))
		}
	}

	// Format the generated HCL
	var buf bytes.Buffer
	_, err := f.WriteTo(&buf)
	if err != nil {
		return "", fmt.Errorf("failed to generate HCL: %w", err)
	}

	return buf.String(), nil
}
