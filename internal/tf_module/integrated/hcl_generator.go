package integrated

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

// GenerateHCL converts the IntegratedConfig to Terraform HCL
func (c *IntegratedConfig) GenerateHCL() (string, error) {
	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()

	// Generate data sources
	generateDataSources(rootBody, c.DataSources)
	rootBody.AppendNewline()

	// Generate locals
	generateLocals(rootBody, c.Locals)
	rootBody.AppendNewline()

	// Generate VPC HCL
	vpcConfig, err := c.VPCConfig.Build()
	if err != nil {
		return "", fmt.Errorf("failed to build VPC config: %w", err)
	}
	vpcHCL, err := vpcConfig.GenerateHCL()
	if err != nil {
		return "", fmt.Errorf("failed to generate VPC HCL: %w", err)
	}
	rootBody.AppendUnstructuredTokens(hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte("module")},
		{Type: hclsyntax.TokenOBrace, Bytes: []byte("{")},
		{Type: hclsyntax.TokenIdent, Bytes: []byte(vpcHCL)},
		{Type: hclsyntax.TokenCBrace, Bytes: []byte("}")},
	})
	rootBody.AppendNewline()

	// Generate EKS HCL
	eksHCL, err := c.EKSConfig.GenerateHCL()
	if err != nil {
		return "", fmt.Errorf("failed to generate EKS HCL: %w", err)
	}
	rootBody.AppendUnstructuredTokens(hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte("module")},
		{Type: hclsyntax.TokenOBrace, Bytes: []byte("{")},
		{Type: hclsyntax.TokenIdent, Bytes: []byte(eksHCL)},
		{Type: hclsyntax.TokenCBrace, Bytes: []byte("}")},
	})
	rootBody.AppendNewline()

	// Generate EKS Blueprints Addons HCL
	addonsHCL, err := c.AddonsConfig.GenerateHCL()
	if err != nil {
		return "", fmt.Errorf("failed to generate EKS Blueprints Addons HCL: %w", err)
	}
	rootBody.AppendUnstructuredTokens(hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte("module")},
		{Type: hclsyntax.TokenOBrace, Bytes: []byte("{")},
		{Type: hclsyntax.TokenIdent, Bytes: []byte(addonsHCL)},
		{Type: hclsyntax.TokenCBrace, Bytes: []byte("}")},
	})
	rootBody.AppendNewline()

	// Generate variables
	generateVariables(rootBody, c.Variables)

	return string(f.Bytes()), nil
}

func generateDataSources(body *hclwrite.Body, dataSources DataSources) {
	dataBlock := body.AppendNewBlock("data", []string{"aws_availability_zones", "available"})
	dataBody := dataBlock.Body()
	filterBlock := dataBody.AppendNewBlock("filter", nil)
	filterBody := filterBlock.Body()
	filterBody.SetAttributeValue("name", cty.StringVal(dataSources.AvailabilityZones.Filter.Name))
	filterBody.SetAttributeValue("values", cty.ListVal(stringsToValueSlice(dataSources.AvailabilityZones.Filter.Values)))
}

func generateLocals(body *hclwrite.Body, locals Locals) {
	localsBlock := body.AppendNewBlock("locals", nil)
	localsBody := localsBlock.Body()
	localsBody.SetAttributeValue("azs", cty.StringVal(locals.AZs))
	localsBody.SetAttributeValue("name", cty.StringVal(locals.Name))

	tagsBlock := localsBody.AppendNewBlock("tags", nil)
	tagsBody := tagsBlock.Body()
	for key, value := range locals.Tags {
		tagsBody.SetAttributeValue(key, cty.StringVal(value))
	}
}

func generateVariables(body *hclwrite.Body, variables Variables) {
	generateVariable(body, "name", variables.Name)
	generateVariable(body, "vpc_cidr", variables.VpcCIDR)
	generateVariable(body, "region", variables.Region)
}

func generateVariable(body *hclwrite.Body, name string, variable Variable) {
	varBlock := body.AppendNewBlock("variable", []string{name})
	varBody := varBlock.Body()
	varBody.SetAttributeValue("type", cty.StringVal(variable.Type))
	varBody.SetAttributeValue("description", cty.StringVal(variable.Description))
	varBody.SetAttributeValue("default", cty.StringVal(variable.Default))
}

func stringsToValueSlice(strs []string) []cty.Value {
	vals := make([]cty.Value, len(strs))
	for i, s := range strs {
		vals[i] = cty.StringVal(s)
	}
	return vals
}
