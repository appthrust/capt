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

// NewVPCConfig creates a new VPCConfig with default values
func NewVPCConfig() *VPCConfig {
	return &VPCConfig{
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
	}
}

// SetName sets the name of the VPC
func (c *VPCConfig) SetName(name string) {
	c.Name = name
}

// SetCIDR sets the CIDR block for the VPC
func (c *VPCConfig) SetCIDR(cidr string) {
	c.CIDR = cidr
}

// SetAZs sets the availability zones for the VPC
func (c *VPCConfig) SetAZs(azs []string) {
	c.AZs = azs
}

// SetPrivateSubnets sets the private subnets for the VPC
func (c *VPCConfig) SetPrivateSubnets(subnets []string) {
	c.PrivateSubnets = subnets
}

// SetPublicSubnets sets the public subnets for the VPC
func (c *VPCConfig) SetPublicSubnets(subnets []string) {
	c.PublicSubnets = subnets
}

// SetEnableNATGateway sets whether to enable NAT Gateway
func (c *VPCConfig) SetEnableNATGateway(enable bool) {
	c.EnableNATGateway = enable
}

// SetSingleNATGateway sets whether to use a single NAT Gateway
func (c *VPCConfig) SetSingleNATGateway(single bool) {
	c.SingleNATGateway = single
}

// AddPublicSubnetTag adds a tag to the public subnets
func (c *VPCConfig) AddPublicSubnetTag(key, value string) {
	c.PublicSubnetTags[key] = value
}

// AddPrivateSubnetTag adds a tag to the private subnets
func (c *VPCConfig) AddPrivateSubnetTag(key, value string) {
	c.PrivateSubnetTags[key] = value
}

// AddTag adds a tag to the VPC
func (c *VPCConfig) AddTag(key, value string) {
	c.Tags[key] = value
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
