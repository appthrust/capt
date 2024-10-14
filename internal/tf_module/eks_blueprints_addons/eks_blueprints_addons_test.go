package eks_blueprints_addons

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNewEKSBlueprintsAddonsConfig(t *testing.T) {
	config := NewEKSBlueprintsAddonsConfig()

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

	if len(config.EKSAddons) != 3 {
		t.Errorf("Expected 3 EKS addons, got %d", len(config.EKSAddons))
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
	config := NewEKSBlueprintsAddonsConfig()
	hcl, err := config.GenerateHCL()

	if err != nil {
		t.Fatalf("GenerateHCL failed: %v", err)
	}

	// Print the generated HCL for debugging
	t.Logf("Generated HCL:\n%s\n", hcl)

	expectedContent := []string{
		"module \"eks_blueprints_addons\"",
		"source = \"aws-ia/eks-blueprints-addons/aws\"",
		"version = \"~> 1.16\"",
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
		"cacheDir = \"/tmp/.helmcache\"",
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
	coreDNSConfig := config.EKSAddons["coredns"].ConfigurationValues
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
