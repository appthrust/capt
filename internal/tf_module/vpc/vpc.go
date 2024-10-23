package vpc

import (
	"fmt"
	"net"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
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

// PrivateSubnetsConfig represents either a static list of subnets or a dynamic expression
type PrivateSubnetsConfig struct {
	Type    PrivateSubnetsType `hcl:"type" json:"type"`
	Static  []string           `hcl:"static,optional" json:"static,omitempty"`
	Dynamic string             `hcl:"dynamic,optional" json:"dynamic,omitempty"`
}

type PrivateSubnetsType string

const (
	PrivateSubnetsTypeStatic  PrivateSubnetsType = "static"
	PrivateSubnetsTypeDynamic PrivateSubnetsType = "dynamic"
)

// PublicSubnetsConfig represents either a static list of subnets or a dynamic expression
type PublicSubnetsConfig struct {
	Type    PublicSubnetsType `hcl:"type" json:"type"`
	Static  []string          `hcl:"static,optional" json:"static,omitempty"`
	Dynamic string            `hcl:"dynamic,optional" json:"dynamic,omitempty"`
}

type PublicSubnetsType string

const (
	PublicSubnetsTypeStatic  PublicSubnetsType = "static"
	PublicSubnetsTypeDynamic PublicSubnetsType = "dynamic"
)

// VPCConfig represents the configuration for a VPC
type VPCConfig struct {
	Source            string                `hcl:"source"`
	Version           string                `hcl:"version"`
	Name              string                `hcl:"name"`
	CIDR              string                `hcl:"cidr"`
	AZs               *AZsConfig            `hcl:"azs,optional"`
	PrivateSubnets    *PrivateSubnetsConfig `hcl:"private_subnets,optional"`
	PublicSubnets     *PublicSubnetsConfig  `hcl:"public_subnets,optional"`
	EnableNATGateway  bool                  `hcl:"enable_nat_gateway"`
	SingleNATGateway  bool                  `hcl:"single_nat_gateway"`
	PublicSubnetTags  map[string]string     `hcl:"public_subnet_tags,optional"`
	PrivateSubnetTags map[string]string     `hcl:"private_subnet_tags,optional"`
	Tags              map[string]string     `hcl:"tags,optional"`
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
			AZs:              &AZsConfig{Type: AZsTypeStatic, Static: []string{"us-west-2a", "us-west-2b", "us-west-2c"}},
			PrivateSubnets:   &PrivateSubnetsConfig{Type: PrivateSubnetsTypeStatic, Static: []string{"10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"}},
			PublicSubnets:    &PublicSubnetsConfig{Type: PublicSubnetsTypeStatic, Static: []string{"10.0.101.0/24", "10.0.102.0/24", "10.0.103.0/24"}},
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
	b.config.AZs = &AZsConfig{
		Type:   AZsTypeStatic,
		Static: azs,
	}

	// Adjust private subnets to match the number of AZs
	if b.config.PrivateSubnets != nil && b.config.PrivateSubnets.Type == PrivateSubnetsTypeStatic {
		newPrivateSubnets := make([]string, len(azs))
		for i := range azs {
			if i < len(b.config.PrivateSubnets.Static) {
				newPrivateSubnets[i] = b.config.PrivateSubnets.Static[i]
			} else {
				// Generate a new subnet CIDR if needed
				newPrivateSubnets[i] = fmt.Sprintf("10.0.%d.0/24", i+1)
			}
		}
		b.config.PrivateSubnets.Static = newPrivateSubnets
	}
	// Adjust public subnets to match the number of AZs if they exist
	if b.config.PublicSubnets != nil && b.config.PublicSubnets.Type == PublicSubnetsTypeStatic {
		newPublicSubnets := make([]string, len(azs))
		for i := range azs {
			if i < len(b.config.PublicSubnets.Static) {
				newPublicSubnets[i] = b.config.PublicSubnets.Static[i]
			} else {
				// Generate a new subnet CIDR if needed
				newPublicSubnets[i] = fmt.Sprintf("10.0.%d.0/24", 101+i)
			}
		}
		b.config.PublicSubnets.Static = newPublicSubnets
	}
	return b
}

// SetAZsExpression sets the AZs as an HCL expression
func (b *VPCConfigBuilder) SetAZsExpression(expr string) *VPCConfigBuilder {
	b.config.AZs = &AZsConfig{
		Type:    AZsTypeDynamic,
		Dynamic: expr,
	}
	return b
}

// SetPrivateSubnets sets the private subnets for the VPC
func (b *VPCConfigBuilder) SetPrivateSubnets(subnets []string) *VPCConfigBuilder {
	b.config.PrivateSubnets = &PrivateSubnetsConfig{
		Type:   PrivateSubnetsTypeStatic,
		Static: subnets,
	}
	return b
}

// SetPrivateSubnetsExpression sets the private subnets as an HCL expression
func (b *VPCConfigBuilder) SetPrivateSubnetsExpression(expr string) *VPCConfigBuilder {
	b.config.PrivateSubnets = &PrivateSubnetsConfig{
		Type:    PrivateSubnetsTypeDynamic,
		Dynamic: expr,
	}
	return b
}

// SetPublicSubnets sets the public subnets for the VPC
func (b *VPCConfigBuilder) SetPublicSubnets(subnets []string) *VPCConfigBuilder {
	b.config.PublicSubnets = &PublicSubnetsConfig{
		Type:   PublicSubnetsTypeStatic,
		Static: subnets,
	}
	return b
}

// SetPublicSubnetsExpression sets the public subnets as an HCL expression
func (b *VPCConfigBuilder) SetPublicSubnetsExpression(expr string) *VPCConfigBuilder {
	b.config.PublicSubnets = &PublicSubnetsConfig{
		Type:    PublicSubnetsTypeDynamic,
		Dynamic: expr,
	}
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
	if c.PrivateSubnets == nil && c.PublicSubnets == nil {
		return fmt.Errorf("at least one subnet (private or public) must be specified")
	}
	if c.AZs == nil {
		return fmt.Errorf("AZs configuration cannot be nil")
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
	// Check if at least one subnet type has valid configuration
	hasValidSubnets := false

	if c.PrivateSubnets != nil {
		if c.PrivateSubnets.Type == PrivateSubnetsTypeStatic {
			if len(c.PrivateSubnets.Static) == 0 {
				return fmt.Errorf("static private subnets must not be empty when using PrivateSubnetsTypeStatic")
			}
			hasValidSubnets = true
		} else if c.PrivateSubnets.Type == PrivateSubnetsTypeDynamic {
			if c.PrivateSubnets.Dynamic == "" {
				return fmt.Errorf("dynamic private subnets expression must not be empty when using PrivateSubnetsTypeDynamic")
			}
			hasValidSubnets = true
		}
	}

	if c.PublicSubnets != nil {
		if c.PublicSubnets.Type == PublicSubnetsTypeStatic {
			if len(c.PublicSubnets.Static) == 0 {
				return fmt.Errorf("static public subnets must not be empty when using PublicSubnetsTypeStatic")
			}
			hasValidSubnets = true
		} else if c.PublicSubnets.Type == PublicSubnetsTypeDynamic {
			if c.PublicSubnets.Dynamic == "" {
				return fmt.Errorf("dynamic public subnets expression must not be empty when using PublicSubnetsTypeDynamic")
			}
			hasValidSubnets = true
		}
	}

	if !hasValidSubnets {
		return fmt.Errorf("at least one subnet type (private or public) must have valid configuration")
	}

	// Validate CIDR format for static subnets
	var allSubnets []string
	if c.PrivateSubnets != nil && c.PrivateSubnets.Type == PrivateSubnetsTypeStatic {
		allSubnets = append(allSubnets, c.PrivateSubnets.Static...)
	}
	if c.PublicSubnets != nil && c.PublicSubnets.Type == PublicSubnetsTypeStatic {
		allSubnets = append(allSubnets, c.PublicSubnets.Static...)
	}
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

	// Create a local copy of the config
	configCopy := *c
	azs := configCopy.AZs
	privateSubnets := configCopy.PrivateSubnets
	publicSubnets := configCopy.PublicSubnets
	configCopy.AZs = nil
	configCopy.PrivateSubnets = nil
	configCopy.PublicSubnets = nil

	// Encode most fields using gohcl.EncodeIntoBody
	gohcl.EncodeIntoBody(&configCopy, moduleBody)

	// Handle AZs separately
	if azs.Type == AZsTypeDynamic {
		tokens, diags := hclsyntax.LexExpression([]byte(azs.Dynamic), "", hcl.Pos{Line: 1, Column: 1})
		if diags.HasErrors() {
			return "", fmt.Errorf("failed to lex dynamic AZs expression: %v", diags)
		}
		moduleBody.SetAttributeRaw("azs", ConvertHCLSyntaxToHCLWrite(tokens))
	} else {
		moduleBody.SetAttributeValue("azs", cty.ListVal(stringsToValues(azs.Static)))
	}

	// Handle PrivateSubnets separately
	if privateSubnets != nil {
		if privateSubnets.Type == PrivateSubnetsTypeDynamic {
			tokens, diags := hclsyntax.LexExpression([]byte(privateSubnets.Dynamic), "", hcl.Pos{Line: 1, Column: 1})
			if diags.HasErrors() {
				return "", fmt.Errorf("failed to lex dynamic private subnets expression: %v", diags)
			}
			moduleBody.SetAttributeRaw("private_subnets", ConvertHCLSyntaxToHCLWrite(tokens))
		} else {
			moduleBody.SetAttributeValue("private_subnets", cty.ListVal(stringsToValues(privateSubnets.Static)))
		}
	}

	// Handle PublicSubnets separately
	if publicSubnets != nil {
		if publicSubnets.Type == PublicSubnetsTypeDynamic {
			tokens, diags := hclsyntax.LexExpression([]byte(publicSubnets.Dynamic), "", hcl.Pos{Line: 1, Column: 1})
			if diags.HasErrors() {
				return "", fmt.Errorf("failed to lex dynamic public subnets expression: %v", diags)
			}
			moduleBody.SetAttributeRaw("public_subnets", ConvertHCLSyntaxToHCLWrite(tokens))
		} else {
			moduleBody.SetAttributeValue("public_subnets", cty.ListVal(stringsToValues(publicSubnets.Static)))
		}
	}

	// Format the generated HCL
	return string(hclwrite.Format(f.Bytes())), nil
}

func stringsToValues(strs []string) []cty.Value {
	values := make([]cty.Value, len(strs))
	for i, s := range strs {
		values[i] = cty.StringVal(s)
	}
	return values
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
