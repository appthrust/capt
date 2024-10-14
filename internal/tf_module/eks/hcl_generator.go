package eks

import (
	"bytes"
	"fmt"
	"strings"

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
	eksBody.SetAttributeValue("source", cty.StringVal("terraform-aws-modules/eks/aws"))
	eksBody.SetAttributeValue("version", cty.StringVal("~> 20.0"))
	eksBody.SetAttributeValue("cluster_name", cty.StringVal(c.ClusterName))
	eksBody.SetAttributeValue("cluster_version", cty.StringVal(c.ClusterVersion))
	eksBody.SetAttributeValue("vpc_id", cty.StringVal(escapeInterpolation("${module.vpc.vpc_id}")))
	eksBody.SetAttributeValue("subnet_ids", cty.StringVal(escapeInterpolation("${module.vpc.private_subnets}")))

	// Add module block for VPC
	vpcBlock := rootBody.AppendNewBlock("module", []string{"vpc"})
	vpcBody := vpcBlock.Body()
	vpcBody.SetAttributeValue("source", cty.StringVal("terraform-aws-modules/vpc/aws"))
	vpcBody.SetAttributeValue("version", cty.StringVal("~> 5.0"))
	vpcBody.SetAttributeValue("name", cty.StringVal(fmt.Sprintf("%s-vpc", c.ClusterName)))
	vpcBody.SetAttributeValue("cidr", cty.StringVal(c.VPC.CIDR))
	vpcBody.SetAttributeValue("azs", cty.ListVal(stringsToValueSlice(c.VPC.PrivateSubnets)))
	vpcBody.SetAttributeValue("private_subnets", cty.ListVal(stringsToValueSlice(c.VPC.PrivateSubnets)))
	vpcBody.SetAttributeValue("public_subnets", cty.ListVal(stringsToValueSlice(c.VPC.PublicSubnets)))
	vpcBody.SetAttributeValue("enable_nat_gateway", cty.BoolVal(c.VPC.EnableNATGateway))
	vpcBody.SetAttributeValue("single_nat_gateway", cty.BoolVal(c.VPC.SingleNATGateway))

	// Add node groups
	nodeGroupsBlock := eksBody.AppendNewBlock("node_groups", nil)
	nodeGroupsBody := nodeGroupsBlock.Body()
	for _, ng := range c.NodeGroups {
		ngBlock := nodeGroupsBody.AppendNewBlock(ng.Name, nil)
		ngBody := ngBlock.Body()
		ngBody.SetAttributeValue("instance_types", cty.ListVal([]cty.Value{cty.StringVal(ng.InstanceType)}))
		ngBody.SetAttributeValue("min_size", cty.NumberIntVal(int64(ng.MinSize)))
		ngBody.SetAttributeValue("max_size", cty.NumberIntVal(int64(ng.MaxSize)))
		ngBody.SetAttributeValue("desired_size", cty.NumberIntVal(int64(ng.DesiredSize)))
		ngBody.SetAttributeValue("disk_size", cty.NumberIntVal(int64(ng.DiskSize)))
	}

	// Add add-ons
	addOnsBlock := eksBody.AppendNewBlock("cluster_addons", nil)
	addOnsBody := addOnsBlock.Body()
	for _, addon := range c.AddOns {
		addonBlock := addOnsBody.AppendNewBlock(addon.Name, nil)
		addonBody := addonBlock.Body()
		addonBody.SetAttributeValue("version", cty.StringVal(addon.Version))
	}

	// Format the generated HCL
	var buf bytes.Buffer
	_, err := f.WriteTo(&buf)
	if err != nil {
		return "", fmt.Errorf("failed to generate HCL: %w", err)
	}

	return buf.String(), nil
}

func stringsToValueSlice(strs []string) []cty.Value {
	vals := make([]cty.Value, len(strs))
	for i, s := range strs {
		vals[i] = cty.StringVal(s)
	}
	return vals
}

func escapeInterpolation(s string) string {
	return strings.ReplaceAll(s, "$", "$$")
}
