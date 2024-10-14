package eks_blueprints_addons

import (
	"bytes"
	"fmt"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

// GenerateHCL converts the EKSBlueprintsAddonsConfig to Terraform HCL
func (c *EKSBlueprintsAddonsConfig) GenerateHCL() (string, error) {
	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()

	// Add module block for EKS Blueprints Addons
	moduleBlock := rootBody.AppendNewBlock("module", []string{"eks_blueprints_addons"})
	moduleBody := moduleBlock.Body()

	moduleBody.SetAttributeValue("source", cty.StringVal("aws-ia/eks-blueprints-addons/aws"))
	moduleBody.SetAttributeValue("version", cty.StringVal("~> 1.16"))

	moduleBody.SetAttributeValue("cluster_name", cty.StringVal(c.ClusterName))
	moduleBody.SetAttributeValue("cluster_endpoint", cty.StringVal(c.ClusterEndpoint))
	moduleBody.SetAttributeValue("cluster_version", cty.StringVal(c.ClusterVersion))
	moduleBody.SetAttributeValue("oidc_provider_arn", cty.StringVal(c.OIDCProviderARN))

	// Add EKS Addons
	eksAddonsBlock := moduleBody.AppendNewBlock("eks_addons", nil)
	eksAddonsBody := eksAddonsBlock.Body()
	for addonName, addon := range c.EKSAddons {
		addonBlock := eksAddonsBody.AppendNewBlock(addonName, nil)
		addonBody := addonBlock.Body()
		if addon.ConfigurationValues != "" {
			addonBody.SetAttributeValue("configuration_values", cty.StringVal(addon.ConfigurationValues))
		}
	}

	moduleBody.SetAttributeValue("enable_karpenter", cty.BoolVal(c.EnableKarpenter))

	// Add Karpenter configuration
	karpenterBlock := moduleBody.AppendNewBlock("karpenter", nil)
	karpenterBody := karpenterBlock.Body()
	helmConfigBlock := karpenterBody.AppendNewBlock("helm_config", nil)
	helmConfigBody := helmConfigBlock.Body()
	helmConfigBody.SetAttributeValue("cacheDir", cty.StringVal(c.Karpenter.HelmConfig.CacheDir))

	// Add Karpenter Node configuration
	karpenterNodeBlock := moduleBody.AppendNewBlock("karpenter_node", nil)
	karpenterNodeBody := karpenterNodeBlock.Body()
	karpenterNodeBody.SetAttributeValue("iam_role_use_name_prefix", cty.BoolVal(c.KarpenterNode.IAMRoleUseNamePrefix))

	// Add tags
	if len(c.Tags) > 0 {
		tagsBody := moduleBody.AppendNewBlock("tags", nil).Body()
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
