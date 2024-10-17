package vpc

import (
	"fmt"
	"net"

	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

// VPCConfig represents the configuration for a VPC
type VPCConfig struct {
	Name              string            `hcl:"name"`
	CIDR              string            `hcl:"cidr"`
	AZs               []string          `hcl:"azs"`
	PrivateSubnets    []string          `hcl:"private_subnets,optional"`
	PublicSubnets     []string          `hcl:"public_subnets,optional"`
	EnableNATGateway  bool              `hcl:"enable_nat_gateway"`
	SingleNATGateway  bool              `hcl:"single_nat_gateway"`
	PublicSubnetTags  map[string]string `hcl:"public_subnet_tags,optional"`
	PrivateSubnetTags map[string]string `hcl:"private_subnet_tags,optional"`
	Tags              map[string]string `hcl:"tags,optional"`
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
			PrivateSubnets:   []string{"10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"},
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
	// Adjust private subnets to match the number of AZs
	if len(b.config.PrivateSubnets) > 0 {
		newPrivateSubnets := make([]string, len(azs))
		for i := range azs {
			if i < len(b.config.PrivateSubnets) {
				newPrivateSubnets[i] = b.config.PrivateSubnets[i]
			} else {
				// Generate a new subnet CIDR if needed
				newPrivateSubnets[i] = fmt.Sprintf("10.0.%d.0/24", i+1)
			}
		}
		b.config.PrivateSubnets = newPrivateSubnets
	}
	// Adjust public subnets to match the number of AZs if they exist
	if len(b.config.PublicSubnets) > 0 {
		newPublicSubnets := make([]string, len(azs))
		for i := range azs {
			if i < len(b.config.PublicSubnets) {
				newPublicSubnets[i] = b.config.PublicSubnets[i]
			} else {
				// Generate a new subnet CIDR if needed
				newPublicSubnets[i] = fmt.Sprintf("10.0.%d.0/24", 101+i)
			}
		}
		b.config.PublicSubnets = newPublicSubnets
	}
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
	if _, _, err := net.ParseCIDR(c.CIDR); err != nil {
		return fmt.Errorf("invalid CIDR address: %v", err)
	}
	if len(c.AZs) == 0 {
		return fmt.Errorf("at least one availability zone must be specified")
	}
	if len(c.PrivateSubnets) == 0 && len(c.PublicSubnets) == 0 {
		return fmt.Errorf("at least one subnet (private or public) must be specified")
	}
	if len(c.PrivateSubnets) > 0 && len(c.PrivateSubnets) != len(c.AZs) {
		return fmt.Errorf("number of private subnets must match the number of AZs")
	}
	if len(c.PublicSubnets) > 0 && len(c.PublicSubnets) != len(c.AZs) {
		return fmt.Errorf("number of public subnets must match the number of AZs")
	}
	return c.validateSubnets()
}

func (c *VPCConfig) validateSubnets() error {
	allSubnets := append(c.PrivateSubnets, c.PublicSubnets...)
	for _, subnet := range allSubnets {
		if _, _, err := net.ParseCIDR(subnet); err != nil {
			return fmt.Errorf("invalid subnet CIDR address: %v", err)
		}
	}
	return nil
}

func (c *VPCConfig) GenerateHCL() (string, error) {
	f := hclwrite.NewEmptyFile()
	gohcl.EncodeIntoBody(c, f.Body())
	return string(f.Bytes()), nil
}
