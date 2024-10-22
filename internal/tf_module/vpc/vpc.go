package vpc

import (
	"fmt"
	"net"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

// AZsConfig represents either a static list of AZs or a dynamic expression
type AZsConfig struct {
	Type    AZsType  `hcl:"type" json:"type"`
	Static  []string `hcl:"static,optional" json:"static,omitempty"`
	Dynamic string   `hcl:"dynamic,optional" json:"dynamic,omitempty"`
}

type AZsType string

const (
	AZsTypeStatic  AZsType = "static"
	AZsTypeDynamic AZsType = "dynamic"
)

// VPCConfig represents the configuration for a VPC
type VPCConfig struct {
	Source            string            `hcl:"source"`
	Version           string            `hcl:"version"`
	Name              string            `hcl:"name"`
	CIDR              string            `hcl:"cidr"`
	AZs               AZsConfig         `hcl:"azs"`
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
			Source:           "terraform-aws-modules/vpc/aws",
			Version:          "5.0.0",
			Name:             "eks-vpc",
			CIDR:             "10.0.0.0/16",
			AZs:              AZsConfig{Type: AZsTypeStatic, Static: []string{"us-west-2a", "us-west-2b", "us-west-2c"}},
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

// SetSource sets the source of the VPC module
func (b *VPCConfigBuilder) SetSource(source string) *VPCConfigBuilder {
	b.config.Source = source
	return b
}

// SetVersion sets the version of the VPC module
func (b *VPCConfigBuilder) SetVersion(version string) *VPCConfigBuilder {
	b.config.Version = version
	return b
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
	b.config.AZs = AZsConfig{
		Type:   AZsTypeStatic,
		Static: azs,
	}

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

// SetAZsExpression sets the AZs as an HCL expression
func (b *VPCConfigBuilder) SetAZsExpression(expr string) *VPCConfigBuilder {
	b.config.AZs = AZsConfig{
		Type:    AZsTypeDynamic,
		Dynamic: expr,
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
	if c.Source == "" {
		return fmt.Errorf("VPC module source cannot be empty")
	}
	if c.Version == "" {
		return fmt.Errorf("VPC module version cannot be empty")
	}
	if c.Name == "" {
		return fmt.Errorf("VPC name cannot be empty")
	}
	if c.CIDR == "" {
		return fmt.Errorf("VPC CIDR cannot be empty")
	}
	if _, _, err := net.ParseCIDR(c.CIDR); err != nil {
		return fmt.Errorf("invalid CIDR address: %v", err)
	}
	if len(c.PrivateSubnets) == 0 && len(c.PublicSubnets) == 0 {
		return fmt.Errorf("at least one subnet (private or public) must be specified")
	}
	if c.AZs.Type == AZsTypeStatic && len(c.AZs.Static) == 0 {
		return fmt.Errorf("static AZs must be specified when using AZsTypeStatic")
	}
	if c.AZs.Type == AZsTypeDynamic && c.AZs.Dynamic == "" {
		return fmt.Errorf("dynamic AZs expression must be specified when using AZsTypeDynamic")
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
	rootBody := f.Body()

	moduleBlock := rootBody.AppendNewBlock("module", []string{"vpc"})
	moduleBody := moduleBlock.Body()

	moduleBody.SetAttributeValue("source", cty.StringVal(c.Source))
	moduleBody.SetAttributeValue("version", cty.StringVal(c.Version))
	moduleBody.SetAttributeValue("name", cty.StringVal(c.Name))
	moduleBody.SetAttributeValue("cidr", cty.StringVal(c.CIDR))

	if c.AZs.Type == AZsTypeDynamic {
		tokens, diags := hclsyntax.LexExpression([]byte(c.AZs.Dynamic), "", hcl.Pos{Line: 1, Column: 1})
		if diags.HasErrors() {
			return "", fmt.Errorf("failed to lex dynamic AZs expression: %v", diags)
		}
		moduleBody.SetAttributeRaw("azs", ConvertHCLSyntaxToHCLWrite(tokens))
	} else {
		moduleBody.SetAttributeValue("azs", cty.ListVal(stringsToValues(c.AZs.Static)))
	}

	moduleBody.SetAttributeValue("private_subnets", cty.ListVal(stringsToValues(c.PrivateSubnets)))
	moduleBody.SetAttributeValue("public_subnets", cty.ListVal(stringsToValues(c.PublicSubnets)))
	moduleBody.SetAttributeValue("enable_nat_gateway", cty.BoolVal(c.EnableNATGateway))
	moduleBody.SetAttributeValue("single_nat_gateway", cty.BoolVal(c.SingleNATGateway))

	if len(c.PublicSubnetTags) > 0 {
		moduleBody.SetAttributeValue("public_subnet_tags", mapToCtyValue(c.PublicSubnetTags))
	}
	if len(c.PrivateSubnetTags) > 0 {
		moduleBody.SetAttributeValue("private_subnet_tags", mapToCtyValue(c.PrivateSubnetTags))
	}
	if len(c.Tags) > 0 {
		moduleBody.SetAttributeValue("tags", mapToCtyValue(c.Tags))
	}

	return string(f.Bytes()), nil
}

func stringsToValues(strs []string) []cty.Value {
	values := make([]cty.Value, len(strs))
	for i, s := range strs {
		values[i] = cty.StringVal(s)
	}
	return values
}

func mapToCtyValue(m map[string]string) cty.Value {
	ctyMap := make(map[string]cty.Value)
	for k, v := range m {
		ctyMap[k] = cty.StringVal(v)
	}
	return cty.ObjectVal(ctyMap)
}

// ConvertHCLSyntaxToHCLWrite converts hclsyntax.Tokens to hclwrite.Tokens
func ConvertHCLSyntaxToHCLWrite(syntaxTokens hclsyntax.Tokens) hclwrite.Tokens {
	writeTokens := make(hclwrite.Tokens, len(syntaxTokens))
	for i, token := range syntaxTokens {
		writeTokens[i] = &hclwrite.Token{
			Type:         token.Type,
			Bytes:        token.Bytes,
			SpacesBefore: 0,
		}
	}
	return writeTokens
}
