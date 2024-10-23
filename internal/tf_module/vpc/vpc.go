package vpc

import (
	"fmt"
	"net"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

type DynamicStaticType interface {
	AZsType | PrivateSubnetsType | PublicSubnetsType | PrivateSubnetTagsType
	String() string
}

// AZsType represents the type of AZs configuration
type AZsType string

func (t AZsType) String() string { return string(t) }

const (
	AZsTypeStatic  AZsType = "static"
	AZsTypeDynamic AZsType = "dynamic"
)

// PrivateSubnetsType represents the type of private subnets configuration
type PrivateSubnetsType string

func (t PrivateSubnetsType) String() string { return string(t) }

const (
	PrivateSubnetsTypeStatic  PrivateSubnetsType = "static"
	PrivateSubnetsTypeDynamic PrivateSubnetsType = "dynamic"
)

// PublicSubnetsType represents the type of public subnets configuration
type PublicSubnetsType string

func (t PublicSubnetsType) String() string { return string(t) }

const (
	PublicSubnetsTypeStatic  PublicSubnetsType = "static"
	PublicSubnetsTypeDynamic PublicSubnetsType = "dynamic"
)

// PrivateSubnetTagsType represents the type of private subnet tags configuration
type PrivateSubnetTagsType string

func (t PrivateSubnetTagsType) String() string { return string(t) }

const (
	PrivateSubnetTagsTypeStatic  PrivateSubnetTagsType = "static"
	PrivateSubnetTagsTypeDynamic PrivateSubnetTagsType = "dynamic"
)

type DynamicStaticConfig[T DynamicStaticType] struct {
	Type    T           `hcl:"type" json:"type"`
	Static  interface{} `hcl:"static,optional" json:"static,omitempty"`
	Dynamic string      `hcl:"dynamic,optional" json:"dynamic,omitempty"`
}

// AZs represents the configuration for availability zones
type AZs = DynamicStaticConfig[AZsType]

// PrivateSubnets represents the configuration for private subnets
type PrivateSubnets = DynamicStaticConfig[PrivateSubnetsType]

// PublicSubnets represents the configuration for public subnets
type PublicSubnets = DynamicStaticConfig[PublicSubnetsType]

// PrivateSubnetTags represents the configuration for private subnet tags
type PrivateSubnetTags = DynamicStaticConfig[PrivateSubnetTagsType]

// VPCConfig represents the configuration for a VPC
type VPCConfig struct {
	Source            string             `hcl:"source"`
	Version           string             `hcl:"version"`
	Name              string             `hcl:"name"`
	CIDR              string             `hcl:"cidr"`
	AZs               *AZs               `hcl:"azs,optional"`
	PrivateSubnets    *PrivateSubnets    `hcl:"private_subnets,optional"`
	PublicSubnets     *PublicSubnets     `hcl:"public_subnets,optional"`
	EnableNATGateway  bool               `hcl:"enable_nat_gateway"`
	SingleNATGateway  bool               `hcl:"single_nat_gateway"`
	PublicSubnetTags  map[string]string  `hcl:"public_subnet_tags,optional"`
	PrivateSubnetTags *PrivateSubnetTags `hcl:"private_subnet_tags,optional"`
	Tags              map[string]string  `hcl:"tags,optional"`
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
			Source:           "terraform-aws-modules/vpc/aws",
			Version:          "5.0.0",
			Name:             "eks-vpc",
			CIDR:             "10.0.0.0/16",
			AZs:              &AZs{Type: AZsTypeStatic, Static: []string{"us-west-2a", "us-west-2b", "us-west-2c"}},
			PrivateSubnets:   &PrivateSubnets{Type: PrivateSubnetsTypeStatic, Static: []string{"10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"}},
			PublicSubnets:    &PublicSubnets{Type: PublicSubnetsTypeStatic, Static: []string{"10.0.101.0/24", "10.0.102.0/24", "10.0.103.0/24"}},
			EnableNATGateway: true,
			SingleNATGateway: true,
			PublicSubnetTags: map[string]string{
				"kubernetes.io/role/elb": "1",
			},
			PrivateSubnetTags: &PrivateSubnetTags{
				Type:   PrivateSubnetTagsTypeStatic,
				Static: defaultPrivateSubnetTags,
			},
			Tags: map[string]string{},
		},
	}
}

func handleDynamicStaticAttribute[T DynamicStaticType](
	moduleBody *hclwrite.Body,
	attrName string,
	config *DynamicStaticConfig[T],
	staticConverter func(interface{}) cty.Value,
) error {
	if config == nil {
		return nil
	}

	if config.Type.String() == "dynamic" {
		tokens, diags := hclsyntax.LexExpression([]byte(config.Dynamic), "", hcl.Pos{Line: 1, Column: 1})
		if diags.HasErrors() {
			return fmt.Errorf("failed to lex dynamic %s expression: %v", attrName, diags)
		}
		moduleBody.SetAttributeRaw(attrName, ConvertHCLSyntaxToHCLWrite(tokens))
	} else {
		moduleBody.SetAttributeValue(attrName, staticConverter(config.Static))
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

	// Handle AZs
	if err := handleDynamicStaticAttribute(moduleBody, "azs", c.AZs, func(v interface{}) cty.Value {
		return cty.ListVal(stringsToValues(v.([]string)))
	}); err != nil {
		return "", err
	}

	// Handle PrivateSubnets
	if err := handleDynamicStaticAttribute(moduleBody, "private_subnets", c.PrivateSubnets, func(v interface{}) cty.Value {
		return cty.ListVal(stringsToValues(v.([]string)))
	}); err != nil {
		return "", err
	}

	// Handle PublicSubnets
	if err := handleDynamicStaticAttribute(moduleBody, "public_subnets", c.PublicSubnets, func(v interface{}) cty.Value {
		return cty.ListVal(stringsToValues(v.([]string)))
	}); err != nil {
		return "", err
	}

	// Handle PrivateSubnetTags
	if err := handleDynamicStaticAttribute(moduleBody, "private_subnet_tags", c.PrivateSubnetTags, func(v interface{}) cty.Value {
		return cty.MapVal(stringMapToValues(v.(map[string]string)))
	}); err != nil {
		return "", err
	}

	moduleBody.SetAttributeValue("enable_nat_gateway", cty.BoolVal(c.EnableNATGateway))
	moduleBody.SetAttributeValue("single_nat_gateway", cty.BoolVal(c.SingleNATGateway))
	moduleBody.SetAttributeValue("enable_dns_hostnames", cty.BoolVal(true))
	moduleBody.SetAttributeValue("enable_dns_support", cty.BoolVal(true))

	if len(c.PublicSubnetTags) > 0 {
		moduleBody.SetAttributeValue("public_subnet_tags", cty.MapVal(stringMapToValues(c.PublicSubnetTags)))
	}

	if len(c.Tags) > 0 {
		moduleBody.SetAttributeValue("tags", cty.MapVal(stringMapToValues(c.Tags)))
	}

	var formattedhcl = hclwrite.Format(f.Bytes())

	return string(formattedhcl), nil
}

func stringsToValues(strs []string) []cty.Value {
	values := make([]cty.Value, len(strs))
	for i, s := range strs {
		values[i] = cty.StringVal(s)
	}
	return values
}

func stringMapToValues(m map[string]string) map[string]cty.Value {
	values := make(map[string]cty.Value, len(m))
	for k, v := range m {
		values[k] = cty.StringVal(v)
	}
	return values
}

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

// Builder methods
func (b *VPCConfigBuilder) SetSource(source string) *VPCConfigBuilder {
	b.config.Source = source
	return b
}

func (b *VPCConfigBuilder) SetVersion(version string) *VPCConfigBuilder {
	b.config.Version = version
	return b
}

func (b *VPCConfigBuilder) SetName(name string) *VPCConfigBuilder {
	b.config.Name = name
	return b
}

func (b *VPCConfigBuilder) SetCIDR(cidr string) *VPCConfigBuilder {
	b.config.CIDR = cidr
	return b
}

func (b *VPCConfigBuilder) SetAZs(azs []string) *VPCConfigBuilder {
	b.config.AZs = &AZs{
		Type:   AZsTypeStatic,
		Static: azs,
	}
	return b
}

func (b *VPCConfigBuilder) SetAZsExpression(expr string) *VPCConfigBuilder {
	b.config.AZs = &AZs{
		Type:    AZsTypeDynamic,
		Dynamic: expr,
	}
	return b
}

func (b *VPCConfigBuilder) SetPrivateSubnets(subnets []string) *VPCConfigBuilder {
	b.config.PrivateSubnets = &PrivateSubnets{
		Type:   PrivateSubnetsTypeStatic,
		Static: subnets,
	}
	return b
}

func (b *VPCConfigBuilder) SetPrivateSubnetsExpression(expr string) *VPCConfigBuilder {
	b.config.PrivateSubnets = &PrivateSubnets{
		Type:    PrivateSubnetsTypeDynamic,
		Dynamic: expr,
	}
	return b
}

func (b *VPCConfigBuilder) SetPublicSubnets(subnets []string) *VPCConfigBuilder {
	b.config.PublicSubnets = &PublicSubnets{
		Type:   PublicSubnetsTypeStatic,
		Static: subnets,
	}
	return b
}

func (b *VPCConfigBuilder) SetPublicSubnetsExpression(expr string) *VPCConfigBuilder {
	b.config.PublicSubnets = &PublicSubnets{
		Type:    PublicSubnetsTypeDynamic,
		Dynamic: expr,
	}
	return b
}

func (b *VPCConfigBuilder) SetPrivateSubnetTags(tags map[string]string) *VPCConfigBuilder {
	b.config.PrivateSubnetTags = &PrivateSubnetTags{
		Type:   PrivateSubnetTagsTypeStatic,
		Static: tags,
	}
	return b
}

func (b *VPCConfigBuilder) SetPrivateSubnetTagsExpression(expr string) *VPCConfigBuilder {
	b.config.PrivateSubnetTags = &PrivateSubnetTags{
		Type:    PrivateSubnetTagsTypeDynamic,
		Dynamic: expr,
	}
	return b
}

func (b *VPCConfigBuilder) SetEnableNATGateway(enable bool) *VPCConfigBuilder {
	b.config.EnableNATGateway = enable
	return b
}

func (b *VPCConfigBuilder) SetSingleNATGateway(single bool) *VPCConfigBuilder {
	b.config.SingleNATGateway = single
	return b
}

func (b *VPCConfigBuilder) AddPublicSubnetTag(key, value string) *VPCConfigBuilder {
	if b.config.PublicSubnetTags == nil {
		b.config.PublicSubnetTags = make(map[string]string)
	}
	b.config.PublicSubnetTags[key] = value
	return b
}

func (b *VPCConfigBuilder) AddTag(key, value string) *VPCConfigBuilder {
	if b.config.Tags == nil {
		b.config.Tags = make(map[string]string)
	}
	b.config.Tags[key] = value
	return b
}

func (b *VPCConfigBuilder) Build() (*VPCConfig, error) {
	if err := b.config.Validate(); err != nil {
		return nil, err
	}
	return b.config, nil
}

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
	return nil
}
