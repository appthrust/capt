package vpc

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

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
