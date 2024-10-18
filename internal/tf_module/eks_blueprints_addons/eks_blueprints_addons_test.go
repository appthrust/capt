package eks_blueprints_addons

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNewEKSBlueprintsAddonsConfig(t *testing.T) {
	builder := NewEKSBlueprintsAddonsConfig()
	config, err := builder.Build()

	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if config.ClusterName != "${module.eks.cluster_name}" {
		t.Errorf("Expected cluster name to be '${module.eks.cluster_name}', got '%s'", config.ClusterName)
	}

	if config.ClusterEndpoint != "${module.eks.cluster_endpoint}" {
		t.Errorf("Expected cluster endpoint to be '${module.eks.cluster_endpoint}', got '%s'", config.ClusterEndpoint)
	}

	if config.ClusterVersion != "${module.eks.cluster_version}" {
		t.Errorf("Expected cluster version to be '${module.eks.cluster_version}', got '%s'", config.ClusterVersion)
	}

	if config.OIDCProviderARN != "${module.eks.oidc_provider_arn}" {
		t.Errorf("Expected OIDC provider ARN to be '${module.eks.oidc_provider_arn}', got '%s'", config.OIDCProviderARN)
	}

	if config.EKSAddons.CoreDNS == nil || config.EKSAddons.VPCCni == nil || config.EKSAddons.KubeProxy == nil {
		t.Errorf("Expected all EKS addons to be initialized")
	}

	if !config.EnableKarpenter {
		t.Errorf("Expected EnableKarpenter to be true")
	}

	if config.Karpenter.HelmConfig.CacheDir != "/tmp/.helmcache" {
		t.Errorf("Expected Karpenter Helm cache dir to be '/tmp/.helmcache', got '%s'", config.Karpenter.HelmConfig.CacheDir)
	}

	if config.KarpenterNode.IAMRoleUseNamePrefix {
		t.Errorf("Expected KarpenterNode IAMRoleUseNamePrefix to be false")
	}
}

func TestGenerateHCL(t *testing.T) {
	builder := NewEKSBlueprintsAddonsConfig()
	config, err := builder.Build()

	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	hcl, err := config.GenerateHCL()

	if err != nil {
		t.Fatalf("GenerateHCL failed: %v", err)
	}

	// Print the generated HCL for debugging
	t.Logf("Generated HCL:\n%s\n", hcl)

	expectedContent := []string{
		"cluster_name",
		"cluster_endpoint",
		"cluster_version",
		"oidc_provider_arn",
		"eks_addons",
		"coredns",
		"vpc-cni",
		"kube-proxy",
		"enable_karpenter = true",
		"karpenter",
		"helm_config",
		"cache_dir = \"/tmp/.helmcache\"",
		"karpenter_node",
		"iam_role_use_name_prefix = false",
	}

	normalizedHCL := strings.ReplaceAll(strings.ReplaceAll(hcl, " ", ""), "\n", "")

	for _, expected := range expectedContent {
		normalizedExpected := strings.ReplaceAll(strings.ReplaceAll(expected, " ", ""), "\n", "")
		if !strings.Contains(normalizedHCL, normalizedExpected) {
			t.Errorf("Generated HCL does not contain expected string: %s", expected)
		}
	}

	// Test CoreDNS configuration
	coreDNSConfig := config.EKSAddons.CoreDNS.ConfigurationValues
	var coreDNSConfigMap map[string]interface{}
	err = json.Unmarshal([]byte(coreDNSConfig), &coreDNSConfigMap)
	if err != nil {
		t.Fatalf("Failed to unmarshal CoreDNS config: %v", err)
	}

	if computeType, ok := coreDNSConfigMap["computeType"].(string); !ok || computeType != "Fargate" {
		t.Errorf("Expected CoreDNS computeType to be 'Fargate', got '%v'", coreDNSConfigMap["computeType"])
	}

	resources, ok := coreDNSConfigMap["resources"].(map[string]interface{})
	if !ok {
		t.Fatalf("CoreDNS resources not found or not a map")
	}

	limits, ok := resources["limits"].(map[string]interface{})
	if !ok {
		t.Fatalf("CoreDNS resources.limits not found or not a map")
	}

	if cpu, ok := limits["cpu"].(string); !ok || cpu != "0.25" {
		t.Errorf("Expected CoreDNS CPU limit to be '0.25', got '%v'", limits["cpu"])
	}

	if memory, ok := limits["memory"].(string); !ok || memory != "256M" {
		t.Errorf("Expected CoreDNS memory limit to be '256M', got '%v'", limits["memory"])
	}
}

func TestEKSBlueprintsAddonsConfigBuilder(t *testing.T) {
	builder := NewEKSBlueprintsAddonsConfig()

	builder.SetClusterName("test-cluster")
	builder.SetClusterEndpoint("https://test-endpoint.eks.amazonaws.com")
	builder.SetClusterVersion("1.21")
	builder.SetOIDCProviderARN("arn:aws:iam::123456789012:oidc-provider/test-provider")
	builder.SetEnableKarpenter(false)
	builder.SetKarpenterHelmCacheDir("/tmp/test-cache")
	builder.SetKarpenterNodeIAMRoleUseNamePrefix(true)
	builder.AddTag("environment", "test")

	config, err := builder.Build()

	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if config.ClusterName != "test-cluster" {
		t.Errorf("Expected cluster name to be 'test-cluster', got '%s'", config.ClusterName)
	}

	if config.ClusterEndpoint != "https://test-endpoint.eks.amazonaws.com" {
		t.Errorf("Expected cluster endpoint to be 'https://test-endpoint.eks.amazonaws.com', got '%s'", config.ClusterEndpoint)
	}

	if config.ClusterVersion != "1.21" {
		t.Errorf("Expected cluster version to be '1.21', got '%s'", config.ClusterVersion)
	}

	if config.OIDCProviderARN != "arn:aws:iam::123456789012:oidc-provider/test-provider" {
		t.Errorf("Expected OIDC provider ARN to be 'arn:aws:iam::123456789012:oidc-provider/test-provider', got '%s'", config.OIDCProviderARN)
	}

	if config.EnableKarpenter {
		t.Errorf("Expected EnableKarpenter to be false")
	}

	if config.Karpenter.HelmConfig.CacheDir != "/tmp/test-cache" {
		t.Errorf("Expected Karpenter Helm cache dir to be '/tmp/test-cache', got '%s'", config.Karpenter.HelmConfig.CacheDir)
	}

	if !config.KarpenterNode.IAMRoleUseNamePrefix {
		t.Errorf("Expected KarpenterNode IAMRoleUseNamePrefix to be true")
	}

	if config.Tags["environment"] != "test" {
		t.Errorf("Expected 'environment' tag to be 'test', got '%s'", config.Tags["environment"])
	}
}
