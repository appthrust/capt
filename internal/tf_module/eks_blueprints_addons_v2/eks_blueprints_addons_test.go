package eks_blueprints_addons_v2

import (
	"fmt"
	"testing"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

func TestEKSBlueprintsAddonsConfig(t *testing.T) {
	builder := NewEKSBlueprintsAddonsConfig()
	builder.SetClusterName("test-cluster")
	builder.SetClusterEndpoint("https://test-endpoint")
	builder.SetClusterVersion("1.24")
	builder.SetOIDCProviderARN("arn:aws:iam::123456789012:oidc-provider")
	builder.SetEnableKarpenter(true)
	builder.SetKarpenterHelmCacheDir("/tmp/.helmcache")
	builder.SetKarpenterUseNamePrefix(true)
	builder.AddTag("environment", "test")

	config, err := builder.Build()
	if err != nil {
		t.Fatalf("Failed to build config: %v", err)
	}

	hcl, err := config.GenerateHCL()
	if err != nil {
		t.Fatalf("Failed to generate HCL: %v", err)
	}

	// 期待値のHCLを生成
	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()
	moduleBlock := rootBody.AppendNewBlock("module", []string{"eks_blueprints_addons"})
	moduleBody := moduleBlock.Body()

	// 属性を生成されたHCLと同じ順序で設定
	moduleBody.SetAttributeValue("source", cty.StringVal("aws-ia/eks-blueprints-addons/aws"))
	moduleBody.SetAttributeValue("version", cty.StringVal("4.32.1"))
	moduleBody.SetAttributeValue("cluster_name", cty.StringVal("test-cluster"))
	moduleBody.SetAttributeValue("cluster_endpoint", cty.StringVal("https://test-endpoint"))
	moduleBody.SetAttributeValue("cluster_version", cty.StringVal("1.24"))
	moduleBody.SetAttributeValue("oidc_provider_arn", cty.StringVal("arn:aws:iam::123456789012:oidc-provider"))
	moduleBody.SetAttributeValue("enable_karpenter", cty.BoolVal(true))

	// Karpenterブロックを追加
	karpenterBlock := moduleBody.AppendNewBlock("karpenter", nil)
	karpenterBody := karpenterBlock.Body()
	helmConfigBlock := karpenterBody.AppendNewBlock("helm_config", nil)
	helmConfigBody := helmConfigBlock.Body()
	helmConfigBody.SetAttributeValue("cacheDir", cty.StringVal("/tmp/.helmcache"))

	moduleBody.SetAttributeValue("karpenter_use_name_prefix", cty.BoolVal(true))
	moduleBody.SetAttributeValue("coredns_config_values", cty.StringVal(`{"computeType":"Fargate","resources":{"limits":{"cpu":"0.25","memory":"256M"},"requests":{"cpu":"0.25","memory":"256M"}}}`))

	tagsMap := map[string]cty.Value{
		"environment": cty.StringVal("test"),
	}
	moduleBody.SetAttributeValue("tags", cty.MapVal(tagsMap))

	expectedHCL := string(hclwrite.Format(f.Bytes()))
	formattedHCL := string(hclwrite.Format([]byte(hcl)))

	// デバッグ出力
	fmt.Printf("Expected HCL:\n%s\n", expectedHCL)
	fmt.Printf("Generated HCL:\n%s\n", formattedHCL)

	if expectedHCL != formattedHCL {
		t.Errorf("Generated HCL does not match expected HCL:\nExpected:\n%s\nGot:\n%s",
			expectedHCL, formattedHCL)
	}
}
