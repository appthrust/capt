package vpc

import (
	"fmt"
)

// VPCConfig represents the configuration for a VPC
type VPCConfig struct {
	Name              string
	CIDR              string
	AZs               []string
	PrivateSubnets    []string
	PublicSubnets     []string
	EnableNATGateway  bool
	SingleNATGateway  bool
	PublicSubnetTags  map[string]string
	PrivateSubnetTags map[string]string
	Tags              map[string]string
}

// VPCConfigBuilder is a builder for VPCConfig
type VPCConfigBuilder struct {
	config *VPCConfig
}

// NewVPCConfig creates a new VPCConfigBuilder with default values
func NewVPCConfig() *VPCConfigBuilder {
	return &VPCConfigBuilder{
		config: &VPCConfig{
			Name:             "eks-vpc",
			CIDR:             "10.0.0.0/16",
			AZs:              []string{"a", "b", "c"},
			EnableNATGateway: true,
			SingleNATGateway: true,
			PublicSubnetTags: map[string]string{
				"kubernetes.io/role/elb": "1",
			},
			PrivateSubnetTags: map[string]string{
				"kubernetes.io/role/internal-elb": "1",
			},
			Tags: map[string]string{},
		},
	}
}

// SetName sets the name of the VPC
func (b *VPCConfigBuilder) SetName(name string) *VPCConfigBuilder {
	b.config.Name = name
	return b
}

// SetCIDR sets the CIDR block for the VPC
func (b *VPCConfigBuilder) SetCIDR(cidr string) *VPCConfigBuilder {
	b.config.CIDR = cidr
	return b
}

// SetAZs sets the availability zones for the VPC
func (b *VPCConfigBuilder) SetAZs(azs []string) *VPCConfigBuilder {
	b.config.AZs = azs
	return b
}

// SetPrivateSubnets sets the private subnets for the VPC
func (b *VPCConfigBuilder) SetPrivateSubnets(subnets []string) *VPCConfigBuilder {
	b.config.PrivateSubnets = subnets
	return b
}

// SetPublicSubnets sets the public subnets for the VPC
func (b *VPCConfigBuilder) SetPublicSubnets(subnets []string) *VPCConfigBuilder {
	b.config.PublicSubnets = subnets
	return b
}

// SetEnableNATGateway sets whether to enable NAT Gateway
func (b *VPCConfigBuilder) SetEnableNATGateway(enable bool) *VPCConfigBuilder {
	b.config.EnableNATGateway = enable
	return b
}

// SetSingleNATGateway sets whether to use a single NAT Gateway
func (b *VPCConfigBuilder) SetSingleNATGateway(single bool) *VPCConfigBuilder {
	b.config.SingleNATGateway = single
	return b
}

// AddPublicSubnetTag adds a tag to the public subnets
func (b *VPCConfigBuilder) AddPublicSubnetTag(key, value string) *VPCConfigBuilder {
	if b.config.PublicSubnetTags == nil {
		b.config.PublicSubnetTags = make(map[string]string)
	}
	b.config.PublicSubnetTags[key] = value
	return b
}

// AddPrivateSubnetTag adds a tag to the private subnets
func (b *VPCConfigBuilder) AddPrivateSubnetTag(key, value string) *VPCConfigBuilder {
	if b.config.PrivateSubnetTags == nil {
		b.config.PrivateSubnetTags = make(map[string]string)
	}
	b.config.PrivateSubnetTags[key] = value
	return b
}

// AddTag adds a tag to the VPC
func (b *VPCConfigBuilder) AddTag(key, value string) *VPCConfigBuilder {
	if b.config.Tags == nil {
		b.config.Tags = make(map[string]string)
	}
	b.config.Tags[key] = value
	return b
}

// Build creates the final VPCConfig
func (b *VPCConfigBuilder) Build() (*VPCConfig, error) {
	if err := b.config.Validate(); err != nil {
		return nil, err
	}
	return b.config, nil
}

// Validate checks if the VPCConfig is valid
func (c *VPCConfig) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("VPC name cannot be empty")
	}
	if c.CIDR == "" {
		return fmt.Errorf("VPC CIDR cannot be empty")
	}
	if len(c.AZs) == 0 {
		return fmt.Errorf("at least one availability zone must be specified")
	}
	return nil
}
