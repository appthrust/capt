package eks

import (
	"bytes"
	"fmt"

	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

// GenerateHCL converts the EKSConfig to Terraform HCL
func (c *EKSConfig) GenerateHCL() (string, error) {
	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()

	// Add provider block
	providerBlock := rootBody.AppendNewBlock("provider", []string{"aws"})
	providerBody := providerBlock.Body()
	providerBody.SetAttributeValue("region", cty.StringVal(c.Region))

	// Add module block for EKS
	eksBlock := rootBody.AppendNewBlock("module", []string{"eks"})
	eksBody := eksBlock.Body()

	// Set default values for source and version if not provided
	if c.Source == "" {
		c.Source = "terraform-aws-modules/eks/aws"
	}
	if c.Version == "" {
		c.Version = "~> 20.0"
	}

	// Encode EKSConfig into the eks module body
	gohcl.EncodeIntoBody(c, eksBody)

	// Add module block for VPC
	vpcBlock := rootBody.AppendNewBlock("module", []string{"vpc"})
	vpcBody := vpcBlock.Body()
	vpcBody.SetAttributeValue("source", cty.StringVal("terraform-aws-modules/vpc/aws"))
	vpcBody.SetAttributeValue("version", cty.StringVal("~> 5.0"))
	vpcBody.SetAttributeValue("name", cty.StringVal(fmt.Sprintf("%s-vpc", c.ClusterName)))

	// Encode VPCConfig into the vpc module body
	gohcl.EncodeIntoBody(&c.VPC, vpcBody)

	// Format the generated HCL
	var buf bytes.Buffer
	_, err := f.WriteTo(&buf)
	if err != nil {
		return "", fmt.Errorf("failed to generate HCL: %w", err)
	}

	return buf.String(), nil
}
