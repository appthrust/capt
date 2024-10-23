package vpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewVPCConfig(t *testing.T) {
	builder := NewVPCConfig()
	config := builder.config

	assert.Equal(t, "terraform-aws-modules/vpc/aws", config.Source)
	assert.Equal(t, "5.0.0", config.Version)
	assert.Equal(t, "eks-vpc", config.Name)
	assert.Equal(t, "10.0.0.0/16", config.CIDR)
	assert.True(t, config.EnableNATGateway)
	assert.True(t, config.SingleNATGateway)

	// Check default private subnet tags
	assert.NotNil(t, config.PrivateSubnetTags)
	staticTags := config.PrivateSubnetTags.Static.(map[string]string)
	assert.Equal(t, 1, len(staticTags))
	assert.Equal(t, "1", staticTags["kubernetes.io/role/internal-elb"])

	// Check default AZs, private subnets, and public subnets
	assert.Equal(t, AZsTypeStatic, config.AZs.Type)
	assert.Equal(t, []string{"us-west-2a", "us-west-2b", "us-west-2c"}, config.AZs.Static.([]string))

	assert.Equal(t, PrivateSubnetsTypeStatic, config.PrivateSubnets.Type)
	assert.Equal(t, []string{"10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"}, config.PrivateSubnets.Static.([]string))

	assert.Equal(t, PublicSubnetsTypeStatic, config.PublicSubnets.Type)
	assert.Equal(t, []string{"10.0.101.0/24", "10.0.102.0/24", "10.0.103.0/24"}, config.PublicSubnets.Static.([]string))
}

func TestVPCConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *VPCConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: &VPCConfig{
				Source:  "source",
				Version: "version",
				Name:    "name",
				CIDR:    "10.0.0.0/16",
			},
			wantErr: false,
		},
		{
			name: "empty source",
			config: &VPCConfig{
				Version: "version",
				Name:    "name",
				CIDR:    "10.0.0.0/16",
			},
			wantErr: true,
		},
		{
			name: "empty version",
			config: &VPCConfig{
				Source: "source",
				Name:   "name",
				CIDR:   "10.0.0.0/16",
			},
			wantErr: true,
		},
		{
			name: "empty name",
			config: &VPCConfig{
				Source:  "source",
				Version: "version",
				CIDR:    "10.0.0.0/16",
			},
			wantErr: true,
		},
		{
			name: "empty CIDR",
			config: &VPCConfig{
				Source:  "source",
				Version: "version",
				Name:    "name",
			},
			wantErr: true,
		},
		{
			name: "invalid CIDR",
			config: &VPCConfig{
				Source:  "source",
				Version: "version",
				Name:    "name",
				CIDR:    "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBuilderMethods(t *testing.T) {
	builder := NewVPCConfig()

	// Test SetAZs
	azs := []string{"us-east-1a", "us-east-1b"}
	builder.SetAZs(azs)
	assert.Equal(t, AZsTypeStatic, builder.config.AZs.Type)
	assert.Equal(t, azs, builder.config.AZs.Static.([]string))

	// Test SetAZsExpression
	expr := "data.aws_availability_zones.available.names"
	builder.SetAZsExpression(expr)
	assert.Equal(t, AZsTypeDynamic, builder.config.AZs.Type)
	assert.Equal(t, expr, builder.config.AZs.Dynamic)

	// Test SetPrivateSubnets
	privateSubnets := []string{"10.0.1.0/24", "10.0.2.0/24"}
	builder.SetPrivateSubnets(privateSubnets)
	assert.Equal(t, PrivateSubnetsTypeStatic, builder.config.PrivateSubnets.Type)
	assert.Equal(t, privateSubnets, builder.config.PrivateSubnets.Static.([]string))

	// Test SetPrivateSubnetsExpression
	privateSubnetsExpr := "local.private_subnets"
	builder.SetPrivateSubnetsExpression(privateSubnetsExpr)
	assert.Equal(t, PrivateSubnetsTypeDynamic, builder.config.PrivateSubnets.Type)
	assert.Equal(t, privateSubnetsExpr, builder.config.PrivateSubnets.Dynamic)

	// Test SetPublicSubnets
	publicSubnets := []string{"10.0.101.0/24", "10.0.102.0/24"}
	builder.SetPublicSubnets(publicSubnets)
	assert.Equal(t, PublicSubnetsTypeStatic, builder.config.PublicSubnets.Type)
	assert.Equal(t, publicSubnets, builder.config.PublicSubnets.Static.([]string))

	// Test SetPublicSubnetsExpression
	publicSubnetsExpr := "local.public_subnets"
	builder.SetPublicSubnetsExpression(publicSubnetsExpr)
	assert.Equal(t, PublicSubnetsTypeDynamic, builder.config.PublicSubnets.Type)
	assert.Equal(t, publicSubnetsExpr, builder.config.PublicSubnets.Dynamic)

	// Test SetPrivateSubnetTags
	privateTags := map[string]string{"key": "value"}
	builder.SetPrivateSubnetTags(privateTags)
	assert.Equal(t, PrivateSubnetTagsTypeStatic, builder.config.PrivateSubnetTags.Type)
	assert.Equal(t, privateTags, builder.config.PrivateSubnetTags.Static.(map[string]string))

	// Test SetPrivateSubnetTagsExpression
	privateTagsExpr := "local.private_subnet_tags"
	builder.SetPrivateSubnetTagsExpression(privateTagsExpr)
	assert.Equal(t, PrivateSubnetTagsTypeDynamic, builder.config.PrivateSubnetTags.Type)
	assert.Equal(t, privateTagsExpr, builder.config.PrivateSubnetTags.Dynamic)

	// Test AddPublicSubnetTag
	builder.AddPublicSubnetTag("key", "value")
	assert.Equal(t, "value", builder.config.PublicSubnetTags["key"])

	// Test AddTag
	builder.AddTag("key", "value")
	assert.Equal(t, "value", builder.config.Tags["key"])
}

func TestChainMethods(t *testing.T) {
	builder := NewVPCConfig()
	config, err := builder.
		SetName("chain-vpc").
		SetCIDR("192.168.0.0/16").
		SetAZs([]string{"us-east-1a", "us-east-1b"}).
		SetPrivateSubnets([]string{"192.168.1.0/24", "192.168.2.0/24"}).
		SetPublicSubnets([]string{"192.168.101.0/24", "192.168.102.0/24"}).
		SetEnableNATGateway(true).
		SetSingleNATGateway(false).
		AddPublicSubnetTag("public-key", "public-value").
		AddTag("tag-key", "tag-value").
		Build()

	assert.NoError(t, err)
	assert.Equal(t, "chain-vpc", config.Name)
	assert.Equal(t, "192.168.0.0/16", config.CIDR)
	assert.Equal(t, []string{"us-east-1a", "us-east-1b"}, config.AZs.Static.([]string))
	assert.Equal(t, []string{"192.168.1.0/24", "192.168.2.0/24"}, config.PrivateSubnets.Static.([]string))
	assert.Equal(t, []string{"192.168.101.0/24", "192.168.102.0/24"}, config.PublicSubnets.Static.([]string))
	assert.True(t, config.EnableNATGateway)
	assert.False(t, config.SingleNATGateway)
	assert.Equal(t, "public-value", config.PublicSubnetTags["public-key"])
	assert.Equal(t, "tag-value", config.Tags["tag-key"])
}

func TestErrorCases(t *testing.T) {
	builder := NewVPCConfig()

	// Test invalid CIDR
	_, err := builder.SetCIDR("invalid").Build()
	assert.Error(t, err)

	// Test empty name
	_, err = builder.SetName("").Build()
	assert.Error(t, err)

	// Test empty source
	builder.config.Source = ""
	_, err = builder.Build()
	assert.Error(t, err)

	// Test empty version
	builder.config.Version = ""
	_, err = builder.Build()
	assert.Error(t, err)
}
