package eks_blueprints_addons_v2

import (
	"fmt"
	"testing"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

func TestEKSBlueprintsAddonsConfig(t *testing.T) {
	// Create and configure the builder
	builder := NewEKSBlueprintsAddonsConfig()

	config, err := builder.Build()
	if err != nil {
		t.Fatalf("Failed to build config: %v", err)
	}

	hcl, err := config.GenerateHCL()
	if err != nil {
		t.Fatalf("Failed to generate HCL: %v", err)
	}

	// Generate expected HCL
	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()
	moduleBlock := rootBody.AppendNewBlock("module", []string{"eks_blueprints_addons"})
	moduleBody := moduleBlock.Body()

	moduleBody.SetAttributeValue("source", cty.StringVal("aws-ia/eks-blueprints-addons/aws"))
	moduleBody.SetAttributeValue("version", cty.StringVal("~> 1.16"))
	moduleBody.SetAttributeRaw("cluster_name", hclwrite.TokensForIdentifier("module.eks.cluster_name"))
	moduleBody.SetAttributeRaw("cluster_endpoint", hclwrite.TokensForIdentifier("module.eks.cluster_endpoint"))
	moduleBody.SetAttributeRaw("cluster_version", hclwrite.TokensForIdentifier("module.eks.cluster_version"))
	moduleBody.SetAttributeRaw("oidc_provider_arn", hclwrite.TokensForIdentifier("module.eks.oidc_provider_arn"))
	moduleBody.SetAttributeRaw("create_delay_dependencies", hclwrite.TokensForIdentifier("[for prof in module.eks.fargate_profiles : prof.fargate_profile_arn]"))

	// Add eks_addons block
	eksAddonsBlock := moduleBody.AppendNewBlock("eks_addons", nil)
	eksAddonsBody := eksAddonsBlock.Body()

	// CoreDNS block with properly formatted configuration
	coreDNSBlock := eksAddonsBody.AppendNewBlock("coredns", nil)
	coreDNSBody := coreDNSBlock.Body()
	coreDNSBody.SetAttributeValue("configuration_values", cty.StringVal(`{
  "computeType": "Fargate",
  "resources": {
    "limits": {
      "cpu": "0.25",
      "memory": "256M"
    },
    "requests": {
      "cpu": "0.25",
      "memory": "256M"
    }
  }
}`))

	// vpc-cni and kube-proxy blocks
	eksAddonsBody.AppendNewBlock("vpc-cni", nil)
	eksAddonsBody.AppendNewBlock("kube-proxy", nil)

	// Enable Karpenter
	moduleBody.SetAttributeValue("enable_karpenter", cty.BoolVal(true))

	// Karpenter block
	karpenterBlock := moduleBody.AppendNewBlock("karpenter", nil)
	karpenterBody := karpenterBlock.Body()
	helmConfigBlock := karpenterBody.AppendNewBlock("helm_config", nil)
	helmConfigBody := helmConfigBlock.Body()
	helmConfigBody.SetAttributeValue("cacheDir", cty.StringVal("/tmp/.helmcache"))

	// Karpenter node block
	karpenterNodeBlock := moduleBody.AppendNewBlock("karpenter_node", nil)
	karpenterNodeBody := karpenterNodeBlock.Body()
	karpenterNodeBody.SetAttributeValue("iam_role_use_name_prefix", cty.BoolVal(false))

	// Tags
	moduleBody.SetAttributeRaw("tags", hclwrite.TokensForIdentifier("local.tags"))

	expectedHCL := string(hclwrite.Format(f.Bytes()))
	formattedHCL := string(hclwrite.Format([]byte(hcl)))

	// Debug output
	fmt.Printf("Expected HCL:\n%s\n", expectedHCL)
	fmt.Printf("Generated HCL:\n%s\n", formattedHCL)

	if expectedHCL != formattedHCL {
		t.Errorf("Generated HCL does not match expected HCL:\nExpected:\n%s\nGot:\n%s",
			expectedHCL, formattedHCL)
	}
}
