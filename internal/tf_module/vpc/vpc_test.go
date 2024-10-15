package vpc

import (
	"strings"
	"testing"
)

func TestNewVPCConfig(t *testing.T) {
	builder := NewVPCConfig()
	config, err := builder.Build()
	if err != nil {
		t.Fatalf("Unexpected error building VPCConfig: %v", err)
	}

	if config.Name != "eks-vpc" {
		t.Errorf("Expected VPC name to be 'eks-vpc', got '%s'", config.Name)
	}

	if config.CIDR != "10.0.0.0/16" {
		t.Errorf("Expected CIDR to be '10.0.0.0/16', got '%s'", config.CIDR)
	}

	if len(config.AZs) != 3 {
		t.Errorf("Expected 3 AZs, got %d", len(config.AZs))
	}

	if !config.EnableNATGateway {
		t.Errorf("Expected EnableNATGateway to be true")
	}

	if !config.SingleNATGateway {
		t.Errorf("Expected SingleNATGateway to be true")
	}

	if len(config.PublicSubnetTags) != 1 {
		t.Errorf("Expected 1 public subnet tag, got %d", len(config.PublicSubnetTags))
	}

	if len(config.PrivateSubnetTags) != 1 {
		t.Errorf("Expected 1 private subnet tag, got %d", len(config.PrivateSubnetTags))
	}
}

func TestVPCConfigValidate(t *testing.T) {
	_, err := NewVPCConfig().Build()
	if err != nil {
		t.Fatalf("Unexpected error building valid VPCConfig: %v", err)
	}

	// Test invalid configurations
	invalidConfigs := []struct {
		name string
		mod  func(*VPCConfigBuilder)
	}{
		{"Empty VPC name", func(b *VPCConfigBuilder) { b.SetName("") }},
		{"Empty CIDR", func(b *VPCConfigBuilder) { b.SetCIDR("") }},
		{"No AZs", func(b *VPCConfigBuilder) { b.SetAZs([]string{}) }},
	}

	for _, ic := range invalidConfigs {
		t.Run(ic.name, func(t *testing.T) {
			invalidBuilder := NewVPCConfig()
			ic.mod(invalidBuilder)
			_, err := invalidBuilder.Build()
			if err == nil {
				t.Errorf("Expected validation to fail for %s", ic.name)
			}
		})
	}
}

func TestGenerateHCL(t *testing.T) {
	builder := NewVPCConfig()
	builder.SetPrivateSubnets([]string{"10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"})
	builder.SetPublicSubnets([]string{"10.0.101.0/24", "10.0.102.0/24", "10.0.103.0/24"})
	builder.AddTag("Environment", "dev")

	config, err := builder.Build()
	if err != nil {
		t.Fatalf("Unexpected error building VPCConfig: %v", err)
	}

	hcl, err := config.GenerateHCL()

	if err != nil {
		t.Fatalf("GenerateHCL failed: %v", err)
	}

	// Print the generated HCL for debugging
	t.Logf("Generated HCL:\n%s\n", hcl)

	expectedContent := []string{
		`module "vpc"`,
		`source`,
		`terraform-aws-modules/vpc/aws`,
		`version`,
		`~> 5.0`,
		`name`,
		`eks-vpc`,
		`cidr`,
		`10.0.0.0/16`,
		`azs`,
		`a`,
		`b`,
		`c`,
		`private_subnets`,
		`10.0.1.0/24`,
		`10.0.2.0/24`,
		`10.0.3.0/24`,
		`public_subnets`,
		`10.0.101.0/24`,
		`10.0.102.0/24`,
		`10.0.103.0/24`,
		`enable_nat_gateway`,
		`true`,
		`single_nat_gateway`,
		`true`,
		`public_subnet_tags`,
		`kubernetes.io/role/elb`,
		`1`,
		`private_subnet_tags`,
		`kubernetes.io/role/internal-elb`,
		`1`,
		`tags`,
		`Environment`,
		`dev`,
	}

	for _, expected := range expectedContent {
		if !strings.Contains(hcl, expected) {
			t.Errorf("Generated HCL does not contain expected string: %s", expected)
		}
	}
}
