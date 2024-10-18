package vpc

import (
	"reflect"
	"strings"
	"testing"
)

// テストヘルパー関数
func setupVPCConfig() *VPCConfigBuilder {
	return NewVPCConfig()
}

func TestNewVPCConfig(t *testing.T) {
	config, err := setupVPCConfig().Build()
	if err != nil {
		t.Fatalf("Unexpected error building VPCConfig: %v", err)
	}

	tests := []struct {
		name     string
		got      interface{}
		expected interface{}
	}{
		{"VPC name", config.Name, "eks-vpc"},
		{"CIDR", config.CIDR, "10.0.0.0/16"},
		{"AZs count", len(config.AZs), 3},
		{"EnableNATGateway", config.EnableNATGateway, true},
		{"SingleNATGateway", config.SingleNATGateway, true},
		{"PublicSubnetTags count", len(config.PublicSubnetTags), 1},
		{"PrivateSubnetTags count", len(config.PrivateSubnetTags), 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !reflect.DeepEqual(tt.got, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, tt.got)
			}
		})
	}
}

func TestVPCConfigValidate(t *testing.T) {
	tests := []struct {
		name        string
		mod         func(*VPCConfigBuilder)
		expectError bool
	}{
		{"Valid config", func(b *VPCConfigBuilder) {
			b.SetPrivateSubnets([]string{"10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"})
			b.SetPublicSubnets([]string{"10.0.101.0/24", "10.0.102.0/24", "10.0.103.0/24"})
		}, false},
		{"Empty VPC name", func(b *VPCConfigBuilder) {
			b.SetName("")
			b.SetPrivateSubnets([]string{"10.0.1.0/24"})
		}, true},
		{"Empty CIDR", func(b *VPCConfigBuilder) {
			b.SetCIDR("")
			b.SetPrivateSubnets([]string{"10.0.1.0/24"})
		}, true},
		{"No AZs", func(b *VPCConfigBuilder) {
			b.SetAZs([]string{})
			b.SetPrivateSubnets([]string{"10.0.1.0/24"})
		}, true},
		{"Invalid CIDR", func(b *VPCConfigBuilder) {
			b.SetCIDR("invalid-cidr")
			b.SetPrivateSubnets([]string{"10.0.1.0/24"})
		}, true},
		{"Empty subnets", func(b *VPCConfigBuilder) {
			b.SetPrivateSubnets([]string{})
			b.SetPublicSubnets([]string{})
		}, true},
		{"Mismatched subnets and AZs", func(b *VPCConfigBuilder) {
			b.SetAZs([]string{"us-west-2a", "us-west-2b"})
			b.SetPrivateSubnets([]string{"10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"})
		}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := setupVPCConfig()
			tt.mod(builder)
			_, err := builder.Build()
			if (err != nil) != tt.expectError {
				t.Errorf("Expected error: %v, got: %v", tt.expectError, err)
			}
		})
	}
}

func TestBuilderMethods(t *testing.T) {
	builder := setupVPCConfig()

	tests := []struct {
		name     string
		mod      func(*VPCConfigBuilder)
		check    func(*VPCConfig) bool
		expected bool
	}{
		{
			"SetName",
			func(b *VPCConfigBuilder) { b.SetName("test-vpc") },
			func(c *VPCConfig) bool { return c.Name == "test-vpc" },
			true,
		},
		{
			"SetCIDR",
			func(b *VPCConfigBuilder) { b.SetCIDR("172.16.0.0/16") },
			func(c *VPCConfig) bool { return c.CIDR == "172.16.0.0/16" },
			true,
		},
		{
			"SetAZs",
			func(b *VPCConfigBuilder) { b.SetAZs([]string{"us-west-2a", "us-west-2b"}) },
			func(c *VPCConfig) bool { return reflect.DeepEqual(c.AZs, []string{"us-west-2a", "us-west-2b"}) },
			true,
		},
		{
			"SetPrivateSubnets",
			func(b *VPCConfigBuilder) { b.SetPrivateSubnets([]string{"172.16.1.0/24", "172.16.2.0/24"}) },
			func(c *VPCConfig) bool {
				return reflect.DeepEqual(c.PrivateSubnets, []string{"172.16.1.0/24", "172.16.2.0/24"})
			},
			true,
		},
		{
			"SetPublicSubnets",
			func(b *VPCConfigBuilder) { b.SetPublicSubnets([]string{"172.16.101.0/24", "172.16.102.0/24"}) },
			func(c *VPCConfig) bool {
				return reflect.DeepEqual(c.PublicSubnets, []string{"172.16.101.0/24", "172.16.102.0/24"})
			},
			true,
		},
		{
			"SetEnableNATGateway",
			func(b *VPCConfigBuilder) { b.SetEnableNATGateway(false) },
			func(c *VPCConfig) bool { return c.EnableNATGateway == false },
			true,
		},
		{
			"SetSingleNATGateway",
			func(b *VPCConfigBuilder) { b.SetSingleNATGateway(false) },
			func(c *VPCConfig) bool { return c.SingleNATGateway == false },
			true,
		},
		{
			"AddPublicSubnetTag",
			func(b *VPCConfigBuilder) { b.AddPublicSubnetTag("test-key", "test-value") },
			func(c *VPCConfig) bool { return c.PublicSubnetTags["test-key"] == "test-value" },
			true,
		},
		{
			"AddPrivateSubnetTag",
			func(b *VPCConfigBuilder) { b.AddPrivateSubnetTag("test-key", "test-value") },
			func(c *VPCConfig) bool { return c.PrivateSubnetTags["test-key"] == "test-value" },
			true,
		},
		{
			"AddTag",
			func(b *VPCConfigBuilder) { b.AddTag("test-key", "test-value") },
			func(c *VPCConfig) bool { return c.Tags["test-key"] == "test-value" },
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mod(builder)
			config, err := builder.Build()
			if err != nil {
				t.Fatalf("Unexpected error building VPCConfig: %v", err)
			}
			if tt.check(config) != tt.expected {
				t.Errorf("Test %s failed", tt.name)
			}
		})
	}
}

func TestChainMethods(t *testing.T) {
	builder := setupVPCConfig()
	config, err := builder.
		SetName("chain-vpc").
		SetCIDR("192.168.0.0/16").
		SetAZs([]string{"us-east-1a", "us-east-1b"}).
		SetPrivateSubnets([]string{"192.168.1.0/24", "192.168.2.0/24"}).
		SetPublicSubnets([]string{"192.168.101.0/24", "192.168.102.0/24"}).
		SetEnableNATGateway(true).
		SetSingleNATGateway(false).
		AddPublicSubnetTag("public-key", "public-value").
		AddPrivateSubnetTag("private-key", "private-value").
		AddTag("global-key", "global-value").
		Build()

	if err != nil {
		t.Fatalf("Unexpected error building VPCConfig: %v", err)
	}

	expectedConfig := &VPCConfig{
		Name:              "chain-vpc",
		CIDR:              "192.168.0.0/16",
		AZs:               []string{"us-east-1a", "us-east-1b"},
		PrivateSubnets:    []string{"192.168.1.0/24", "192.168.2.0/24"},
		PublicSubnets:     []string{"192.168.101.0/24", "192.168.102.0/24"},
		EnableNATGateway:  true,
		SingleNATGateway:  false,
		PublicSubnetTags:  map[string]string{"public-key": "public-value", "kubernetes.io/role/elb": "1"},
		PrivateSubnetTags: map[string]string{"private-key": "private-value", "kubernetes.io/role/internal-elb": "1"},
		Tags:              map[string]string{"global-key": "global-value"},
	}

	if !reflect.DeepEqual(config, expectedConfig) {
		t.Errorf("Chain methods test failed. Expected %+v, got %+v", expectedConfig, config)
	}
}

func TestErrorCases(t *testing.T) {
	tests := []struct {
		name        string
		mod         func(*VPCConfigBuilder)
		expectedErr string
	}{
		{
			"Invalid CIDR",
			func(b *VPCConfigBuilder) {
				b.SetCIDR("invalid-cidr")
				b.SetPrivateSubnets([]string{"10.0.1.0/24"})
			},
			"invalid CIDR address",
		},
		{
			"Mismatched subnets and AZs",
			func(b *VPCConfigBuilder) {
				b.SetAZs([]string{"us-west-2a", "us-west-2b"})
				b.SetPrivateSubnets([]string{"10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"})
			},
			"number of private subnets must match the number of AZs",
		},
		{
			"Empty name",
			func(b *VPCConfigBuilder) {
				b.SetName("")
				b.SetPrivateSubnets([]string{"10.0.1.0/24"})
			},
			"VPC name cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := setupVPCConfig()
			tt.mod(builder)
			_, err := builder.Build()
			if err == nil {
				t.Errorf("Expected error containing '%s', but got no error", tt.expectedErr)
			} else if !strings.Contains(err.Error(), tt.expectedErr) {
				t.Errorf("Expected error containing '%s', but got '%s'", tt.expectedErr, err.Error())
			}
		})
	}
}