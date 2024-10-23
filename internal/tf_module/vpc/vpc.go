package vpc

import (
	"fmt"
	"net"
)

// ValueType represents the type of a field in HCL
type ValueType string

const (
	ValueTypeString     ValueType = "string"
	ValueTypeBool       ValueType = "bool"
	ValueTypeStringMap  ValueType = "string_map"
	ValueTypeStringList ValueType = "string_list"
)

// ConfigType represents whether the config is static or dynamic
type ConfigType string

func (t ConfigType) String() string { return string(t) }

const (
	ConfigTypeStatic  ConfigType = "static"
	ConfigTypeDynamic ConfigType = "dynamic"
)

// HclField represents a field in HCL that can be either static or dynamic
type HclField struct {
	Type      ConfigType  `hcl:"type" json:"type"`
	Static    interface{} `hcl:"static,optional" json:"static,omitempty"`
	Dynamic   string      `hcl:"dynamic,optional" json:"dynamic,omitempty"`
	ValueType ValueType   `hcl:"value_type" json:"value_type"`
}

// VPCConfig represents the configuration for a VPC
type VPCConfig struct {
	Source            *HclField `hcl:"source"`
	Version           *HclField `hcl:"version"`
	Name              *HclField `hcl:"name"`
	CIDR              *HclField `hcl:"cidr"`
	AZs               *HclField `hcl:"azs,optional"`
	PrivateSubnets    *HclField `hcl:"private_subnets,optional"`
	PublicSubnets     *HclField `hcl:"public_subnets,optional"`
	EnableNATGateway  *HclField `hcl:"enable_nat_gateway"`
	SingleNATGateway  *HclField `hcl:"single_nat_gateway"`
	PublicSubnetTags  *HclField `hcl:"public_subnet_tags,optional"`
	PrivateSubnetTags *HclField `hcl:"private_subnet_tags,optional"`
	Tags              *HclField `hcl:"tags,optional"`
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
			Source:           &HclField{Type: ConfigTypeStatic, Static: "terraform-aws-modules/vpc/aws", ValueType: ValueTypeString},
			Version:          &HclField{Type: ConfigTypeStatic, Static: "5.0.0", ValueType: ValueTypeString},
			Name:             &HclField{Type: ConfigTypeStatic, Static: "eks-vpc", ValueType: ValueTypeString},
			CIDR:             &HclField{Type: ConfigTypeStatic, Static: "10.0.0.0/16", ValueType: ValueTypeString},
			AZs:              &HclField{Type: ConfigTypeStatic, Static: []string{"us-west-2a", "us-west-2b", "us-west-2c"}, ValueType: ValueTypeStringList},
			PrivateSubnets:   &HclField{Type: ConfigTypeStatic, Static: []string{"10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"}, ValueType: ValueTypeStringList},
			PublicSubnets:    &HclField{Type: ConfigTypeStatic, Static: []string{"10.0.101.0/24", "10.0.102.0/24", "10.0.103.0/24"}, ValueType: ValueTypeStringList},
			EnableNATGateway: &HclField{Type: ConfigTypeStatic, Static: true, ValueType: ValueTypeBool},
			SingleNATGateway: &HclField{Type: ConfigTypeStatic, Static: true, ValueType: ValueTypeBool},
			PublicSubnetTags: &HclField{Type: ConfigTypeStatic, Static: map[string]string{"kubernetes.io/role/elb": "1"}, ValueType: ValueTypeStringMap},
			PrivateSubnetTags: &HclField{
				Type:      ConfigTypeStatic,
				Static:    defaultPrivateSubnetTags,
				ValueType: ValueTypeStringMap,
			},
			Tags: &HclField{Type: ConfigTypeStatic, Static: map[string]string{}, ValueType: ValueTypeStringMap},
		},
	}
}

// Builder methods
func (b *VPCConfigBuilder) SetSource(source string) *VPCConfigBuilder {
	b.config.Source = &HclField{
		Type:      ConfigTypeStatic,
		Static:    source,
		ValueType: ValueTypeString,
	}
	return b
}

func (b *VPCConfigBuilder) SetVersion(version string) *VPCConfigBuilder {
	b.config.Version = &HclField{
		Type:      ConfigTypeStatic,
		Static:    version,
		ValueType: ValueTypeString,
	}
	return b
}

func (b *VPCConfigBuilder) SetName(name string) *VPCConfigBuilder {
	b.config.Name = &HclField{
		Type:      ConfigTypeStatic,
		Static:    name,
		ValueType: ValueTypeString,
	}
	return b
}

func (b *VPCConfigBuilder) SetCIDR(cidr string) *VPCConfigBuilder {
	b.config.CIDR = &HclField{
		Type:      ConfigTypeStatic,
		Static:    cidr,
		ValueType: ValueTypeString,
	}
	return b
}

func (b *VPCConfigBuilder) SetAZs(azs []string) *VPCConfigBuilder {
	b.config.AZs = &HclField{
		Type:      ConfigTypeStatic,
		Static:    azs,
		ValueType: ValueTypeStringList,
	}
	return b
}

func (b *VPCConfigBuilder) SetAZsExpression(expr string) *VPCConfigBuilder {
	b.config.AZs = &HclField{
		Type:      ConfigTypeDynamic,
		Dynamic:   expr,
		ValueType: ValueTypeStringList,
	}
	return b
}

func (b *VPCConfigBuilder) SetPrivateSubnets(subnets []string) *VPCConfigBuilder {
	b.config.PrivateSubnets = &HclField{
		Type:      ConfigTypeStatic,
		Static:    subnets,
		ValueType: ValueTypeStringList,
	}
	return b
}

func (b *VPCConfigBuilder) SetPrivateSubnetsExpression(expr string) *VPCConfigBuilder {
	b.config.PrivateSubnets = &HclField{
		Type:      ConfigTypeDynamic,
		Dynamic:   expr,
		ValueType: ValueTypeStringList,
	}
	return b
}

func (b *VPCConfigBuilder) SetPublicSubnets(subnets []string) *VPCConfigBuilder {
	b.config.PublicSubnets = &HclField{
		Type:      ConfigTypeStatic,
		Static:    subnets,
		ValueType: ValueTypeStringList,
	}
	return b
}

func (b *VPCConfigBuilder) SetPublicSubnetsExpression(expr string) *VPCConfigBuilder {
	b.config.PublicSubnets = &HclField{
		Type:      ConfigTypeDynamic,
		Dynamic:   expr,
		ValueType: ValueTypeStringList,
	}
	return b
}

func (b *VPCConfigBuilder) SetPrivateSubnetTags(tags map[string]string) *VPCConfigBuilder {
	b.config.PrivateSubnetTags = &HclField{
		Type:      ConfigTypeStatic,
		Static:    tags,
		ValueType: ValueTypeStringMap,
	}
	return b
}

func (b *VPCConfigBuilder) SetPrivateSubnetTagsExpression(expr string) *VPCConfigBuilder {
	b.config.PrivateSubnetTags = &HclField{
		Type:      ConfigTypeDynamic,
		Dynamic:   expr,
		ValueType: ValueTypeStringMap,
	}
	return b
}

func (b *VPCConfigBuilder) SetEnableNATGateway(enable bool) *VPCConfigBuilder {
	b.config.EnableNATGateway = &HclField{
		Type:      ConfigTypeStatic,
		Static:    enable,
		ValueType: ValueTypeBool,
	}
	return b
}

func (b *VPCConfigBuilder) SetSingleNATGateway(single bool) *VPCConfigBuilder {
	b.config.SingleNATGateway = &HclField{
		Type:      ConfigTypeStatic,
		Static:    single,
		ValueType: ValueTypeBool,
	}
	return b
}

func (b *VPCConfigBuilder) AddPublicSubnetTag(key, value string) *VPCConfigBuilder {
	tags := make(map[string]string)
	if b.config.PublicSubnetTags != nil && b.config.PublicSubnetTags.Type == ConfigTypeStatic {
		tags = b.config.PublicSubnetTags.Static.(map[string]string)
	}
	tags[key] = value
	b.config.PublicSubnetTags = &HclField{
		Type:      ConfigTypeStatic,
		Static:    tags,
		ValueType: ValueTypeStringMap,
	}
	return b
}

func (b *VPCConfigBuilder) AddTag(key, value string) *VPCConfigBuilder {
	tags := make(map[string]string)
	if b.config.Tags != nil && b.config.Tags.Type == ConfigTypeStatic {
		tags = b.config.Tags.Static.(map[string]string)
	}
	tags[key] = value
	b.config.Tags = &HclField{
		Type:      ConfigTypeStatic,
		Static:    tags,
		ValueType: ValueTypeStringMap,
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
	if c.Source == nil || c.Source.Type == ConfigTypeStatic && c.Source.Static.(string) == "" {
		return fmt.Errorf("VPC module source cannot be empty")
	}
	if c.Version == nil || c.Version.Type == ConfigTypeStatic && c.Version.Static.(string) == "" {
		return fmt.Errorf("VPC module version cannot be empty")
	}
	if c.Name == nil || c.Name.Type == ConfigTypeStatic && c.Name.Static.(string) == "" {
		return fmt.Errorf("VPC name cannot be empty")
	}
	if c.CIDR == nil || c.CIDR.Type == ConfigTypeStatic && c.CIDR.Static.(string) == "" {
		return fmt.Errorf("VPC CIDR cannot be empty")
	}
	if c.CIDR.Type == ConfigTypeStatic {
		if _, _, err := net.ParseCIDR(c.CIDR.Static.(string)); err != nil {
			return fmt.Errorf("invalid CIDR address: %v", err)
		}
	}
	return nil
}
