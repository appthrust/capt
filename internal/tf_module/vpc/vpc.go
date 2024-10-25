package vpc

import (
	"fmt"
	"net"

	"github.com/appthrust/capt/internal/tf_module/hcl"
)

// VPCConfig represents the configuration for a VPC
type VPCConfig struct {
	Source            *hcl.HclField `hcl:"source"`
	Version           *hcl.HclField `hcl:"version"`
	Name              *hcl.HclField `hcl:"name"`
	CIDR              *hcl.HclField `hcl:"cidr"`
	AZs               *hcl.HclField `hcl:"azs,optional"`
	PrivateSubnets    *hcl.HclField `hcl:"private_subnets,optional"`
	PublicSubnets     *hcl.HclField `hcl:"public_subnets,optional"`
	EnableNATGateway  *hcl.HclField `hcl:"enable_nat_gateway"`
	SingleNATGateway  *hcl.HclField `hcl:"single_nat_gateway"`
	PublicSubnetTags  *hcl.HclField `hcl:"public_subnet_tags,optional"`
	PrivateSubnetTags *hcl.HclField `hcl:"private_subnet_tags,optional"`
	Tags              *hcl.HclField `hcl:"tags,optional"`
}

// VPCConfigBuilder is a builder for VPCConfig
type VPCConfigBuilder struct {
	config *VPCConfig
}

// NewVPCConfig creates a new VPCConfigBuilder with default values
func NewVPCConfig() *VPCConfigBuilder {
	defaultPrivateSubnetTags := map[string]string{
		"kubernetes.io/role/internal-elb": "1",
	}

	return &VPCConfigBuilder{
		config: &VPCConfig{
			Source:           &hcl.HclField{Type: hcl.ConfigTypeStatic, Static: "terraform-aws-modules/vpc/aws", ValueType: hcl.ValueTypeString},
			Version:          &hcl.HclField{Type: hcl.ConfigTypeStatic, Static: "5.0.0", ValueType: hcl.ValueTypeString},
			Name:             &hcl.HclField{Type: hcl.ConfigTypeStatic, Static: "eks-vpc", ValueType: hcl.ValueTypeString},
			CIDR:             &hcl.HclField{Type: hcl.ConfigTypeStatic, Static: "10.0.0.0/16", ValueType: hcl.ValueTypeString},
			AZs:              &hcl.HclField{Type: hcl.ConfigTypeStatic, Static: []string{"us-west-2a", "us-west-2b", "us-west-2c"}, ValueType: hcl.ValueTypeStringList},
			PrivateSubnets:   &hcl.HclField{Type: hcl.ConfigTypeStatic, Static: []string{"10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"}, ValueType: hcl.ValueTypeStringList},
			PublicSubnets:    &hcl.HclField{Type: hcl.ConfigTypeStatic, Static: []string{"10.0.101.0/24", "10.0.102.0/24", "10.0.103.0/24"}, ValueType: hcl.ValueTypeStringList},
			EnableNATGateway: &hcl.HclField{Type: hcl.ConfigTypeStatic, Static: true, ValueType: hcl.ValueTypeBool},
			SingleNATGateway: &hcl.HclField{Type: hcl.ConfigTypeStatic, Static: true, ValueType: hcl.ValueTypeBool},
			PublicSubnetTags: &hcl.HclField{Type: hcl.ConfigTypeStatic, Static: map[string]string{"kubernetes.io/role/elb": "1"}, ValueType: hcl.ValueTypeStringMap},
			PrivateSubnetTags: &hcl.HclField{
				Type:      hcl.ConfigTypeStatic,
				Static:    defaultPrivateSubnetTags,
				ValueType: hcl.ValueTypeStringMap,
			},
			Tags: &hcl.HclField{Type: hcl.ConfigTypeStatic, Static: map[string]string{}, ValueType: hcl.ValueTypeStringMap},
		},
	}
}

// Builder methods
func (b *VPCConfigBuilder) SetSource(source string) *VPCConfigBuilder {
	b.config.Source = &hcl.HclField{
		Type:      hcl.ConfigTypeStatic,
		Static:    source,
		ValueType: hcl.ValueTypeString,
	}
	return b
}

func (b *VPCConfigBuilder) SetVersion(version string) *VPCConfigBuilder {
	b.config.Version = &hcl.HclField{
		Type:      hcl.ConfigTypeStatic,
		Static:    version,
		ValueType: hcl.ValueTypeString,
	}
	return b
}

func (b *VPCConfigBuilder) SetName(name string) *VPCConfigBuilder {
	b.config.Name = &hcl.HclField{
		Type:      hcl.ConfigTypeStatic,
		Static:    name,
		ValueType: hcl.ValueTypeString,
	}
	return b
}

func (b *VPCConfigBuilder) SetCIDR(cidr string) *VPCConfigBuilder {
	b.config.CIDR = &hcl.HclField{
		Type:      hcl.ConfigTypeStatic,
		Static:    cidr,
		ValueType: hcl.ValueTypeString,
	}
	return b
}

func (b *VPCConfigBuilder) SetAZs(azs []string) *VPCConfigBuilder {
	b.config.AZs = &hcl.HclField{
		Type:      hcl.ConfigTypeStatic,
		Static:    azs,
		ValueType: hcl.ValueTypeStringList,
	}
	return b
}

func (b *VPCConfigBuilder) SetAZsExpression(expr string) *VPCConfigBuilder {
	b.config.AZs = &hcl.HclField{
		Type:      hcl.ConfigTypeDynamic,
		Dynamic:   expr,
		ValueType: hcl.ValueTypeStringList,
	}
	return b
}

func (b *VPCConfigBuilder) SetPrivateSubnets(subnets []string) *VPCConfigBuilder {
	b.config.PrivateSubnets = &hcl.HclField{
		Type:      hcl.ConfigTypeStatic,
		Static:    subnets,
		ValueType: hcl.ValueTypeStringList,
	}
	return b
}

func (b *VPCConfigBuilder) SetPrivateSubnetsExpression(expr string) *VPCConfigBuilder {
	b.config.PrivateSubnets = &hcl.HclField{
		Type:      hcl.ConfigTypeDynamic,
		Dynamic:   expr,
		ValueType: hcl.ValueTypeStringList,
	}
	return b
}

func (b *VPCConfigBuilder) SetPublicSubnets(subnets []string) *VPCConfigBuilder {
	b.config.PublicSubnets = &hcl.HclField{
		Type:      hcl.ConfigTypeStatic,
		Static:    subnets,
		ValueType: hcl.ValueTypeStringList,
	}
	return b
}

func (b *VPCConfigBuilder) SetPublicSubnetsExpression(expr string) *VPCConfigBuilder {
	b.config.PublicSubnets = &hcl.HclField{
		Type:      hcl.ConfigTypeDynamic,
		Dynamic:   expr,
		ValueType: hcl.ValueTypeStringList,
	}
	return b
}

func (b *VPCConfigBuilder) SetPrivateSubnetTags(tags map[string]string) *VPCConfigBuilder {
	b.config.PrivateSubnetTags = &hcl.HclField{
		Type:      hcl.ConfigTypeStatic,
		Static:    tags,
		ValueType: hcl.ValueTypeStringMap,
	}
	return b
}

func (b *VPCConfigBuilder) SetPrivateSubnetTagsExpression(expr string) *VPCConfigBuilder {
	b.config.PrivateSubnetTags = &hcl.HclField{
		Type:      hcl.ConfigTypeDynamic,
		Dynamic:   expr,
		ValueType: hcl.ValueTypeStringMap,
	}
	return b
}

func (b *VPCConfigBuilder) SetEnableNATGateway(enable bool) *VPCConfigBuilder {
	b.config.EnableNATGateway = &hcl.HclField{
		Type:      hcl.ConfigTypeStatic,
		Static:    enable,
		ValueType: hcl.ValueTypeBool,
	}
	return b
}

func (b *VPCConfigBuilder) SetSingleNATGateway(single bool) *VPCConfigBuilder {
	b.config.SingleNATGateway = &hcl.HclField{
		Type:      hcl.ConfigTypeStatic,
		Static:    single,
		ValueType: hcl.ValueTypeBool,
	}
	return b
}

func (b *VPCConfigBuilder) AddPublicSubnetTag(key, value string) *VPCConfigBuilder {
	tags := make(map[string]string)
	if b.config.PublicSubnetTags != nil && b.config.PublicSubnetTags.Type == hcl.ConfigTypeStatic {
		tags = b.config.PublicSubnetTags.Static.(map[string]string)
	}
	tags[key] = value
	b.config.PublicSubnetTags = &hcl.HclField{
		Type:      hcl.ConfigTypeStatic,
		Static:    tags,
		ValueType: hcl.ValueTypeStringMap,
	}
	return b
}

func (b *VPCConfigBuilder) AddTag(key, value string) *VPCConfigBuilder {
	tags := make(map[string]string)
	if b.config.Tags != nil && b.config.Tags.Type == hcl.ConfigTypeStatic {
		tags = b.config.Tags.Static.(map[string]string)
	}
	tags[key] = value
	b.config.Tags = &hcl.HclField{
		Type:      hcl.ConfigTypeStatic,
		Static:    tags,
		ValueType: hcl.ValueTypeStringMap,
	}
	return b
}

func (b *VPCConfigBuilder) Build() (*VPCConfig, error) {
	if err := b.config.Validate(); err != nil {
		return nil, err
	}
	return b.config, nil
}

func (c *VPCConfig) Validate() error {
	if c.Source == nil || c.Source.Type == hcl.ConfigTypeStatic && c.Source.Static.(string) == "" {
		return fmt.Errorf("VPC module source cannot be empty")
	}
	if c.Version == nil || c.Version.Type == hcl.ConfigTypeStatic && c.Version.Static.(string) == "" {
		return fmt.Errorf("VPC module version cannot be empty")
	}
	if c.Name == nil || c.Name.Type == hcl.ConfigTypeStatic && c.Name.Static.(string) == "" {
		return fmt.Errorf("VPC name cannot be empty")
	}
	if c.CIDR == nil || c.CIDR.Type == hcl.ConfigTypeStatic && c.CIDR.Static.(string) == "" {
		return fmt.Errorf("VPC CIDR cannot be empty")
	}
	if c.CIDR.Type == hcl.ConfigTypeStatic {
		if _, _, err := net.ParseCIDR(c.CIDR.Static.(string)); err != nil {
			return fmt.Errorf("invalid CIDR address: %v", err)
		}
	}
	return nil
}

func (c *VPCConfig) GenerateHCL() (string, error) {
	generator := hcl.NewHclGenerator("module", []string{"vpc"})
	return generator.GenerateHCL(c)
}
