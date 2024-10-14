package eks

import (
	"fmt"
	"strings"
	"testing"
)

func TestNewEKSConfig(t *testing.T) {
	config := NewEKSConfig()

	if config.ClusterName != "eks-cluster" {
		t.Errorf("Expected cluster name to be 'eks-cluster', got '%s'", config.ClusterName)
	}

	if config.ClusterVersion != "1.31" {
		t.Errorf("Expected cluster version to be '1.31', got '%s'", config.ClusterVersion)
	}

	if config.Region != "us-west-2" {
		t.Errorf("Expected region to be 'us-west-2', got '%s'", config.Region)
	}

	if len(config.NodeGroups) != 1 {
		t.Errorf("Expected 1 node group, got %d", len(config.NodeGroups))
	}

	if len(config.AddOns) != 3 {
		t.Errorf("Expected 3 add-ons, got %d", len(config.AddOns))
	}
}

func TestEKSConfigValidate(t *testing.T) {
	config := NewEKSConfig()

	if err := config.Validate(); err != nil {
		t.Errorf("Validation failed for valid config: %v", err)
	}

	// Test invalid configurations
	invalidConfigs := []struct {
		name string
		mod  func(*EKSConfig)
	}{
		{"Empty cluster name", func(c *EKSConfig) { c.ClusterName = "" }},
		{"Empty cluster version", func(c *EKSConfig) { c.ClusterVersion = "" }},
		{"Empty region", func(c *EKSConfig) { c.Region = "" }},
		{"Empty VPC CIDR", func(c *EKSConfig) { c.VPC.CIDR = "" }},
		{"No private subnets", func(c *EKSConfig) { c.VPC.PrivateSubnets = []string{} }},
		{"No public subnets", func(c *EKSConfig) { c.VPC.PublicSubnets = []string{} }},
	}

	for _, ic := range invalidConfigs {
		t.Run(ic.name, func(t *testing.T) {
			invalidConfig := NewEKSConfig()
			ic.mod(invalidConfig)
			if err := invalidConfig.Validate(); err == nil {
				t.Errorf("Expected validation to fail for %s", ic.name)
			}
		})
	}
}

func TestGenerateHCL(t *testing.T) {
	config := NewEKSConfig()
	hcl, err := config.GenerateHCL()

	if err != nil {
		t.Fatalf("GenerateHCL failed: %v", err)
	}

	// Print the generated HCL for debugging
	fmt.Printf("Generated HCL:\n%s\n", hcl)

	expectedContent := []string{
		"provider aws",
		"region = us-west-2",
		"module eks",
		"source = terraform-aws-modules/eks/aws",
		"cluster_name = eks-cluster",
		"cluster_version = 1.31",
		"module vpc",
		"cidr = 10.0.0.0/16",
		"node_groups",
		"cluster_addons",
	}

	normalizedHCL := strings.ReplaceAll(strings.ReplaceAll(hcl, "\"", ""), " ", "")

	for _, expected := range expectedContent {
		normalizedExpected := strings.ReplaceAll(strings.ReplaceAll(expected, "\"", ""), " ", "")
		if !strings.Contains(normalizedHCL, normalizedExpected) {
			t.Errorf("Generated HCL does not contain expected string: %s", expected)
		}
	}
}
