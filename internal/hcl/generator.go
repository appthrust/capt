package hcl

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

// HclGenerator provides functionality to generate HCL from structs
type HclGenerator struct {
	blockType   string
	blockLabels []string
}

// NewHclGenerator creates a new HclGenerator
func NewHclGenerator(blockType string, blockLabels []string) *HclGenerator {
	return &HclGenerator{
		blockType:   blockType,
		blockLabels: blockLabels,
	}
}

// GenerateHCL generates HCL from a struct using reflection
func (g *HclGenerator) GenerateHCL(config interface{}) (string, error) {
	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()
	moduleBlock := rootBody.AppendNewBlock(g.blockType, g.blockLabels)
	moduleBody := moduleBlock.Body()

	// Use reflection to iterate over struct fields
	val := reflect.ValueOf(config)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)
		hclTag := fieldType.Tag.Get("hcl")
		if hclTag == "" {
			continue
		}

		// Split the hcl tag to get the field name (remove any options like ",optional")
		fieldName := strings.Split(hclTag, ",")[0]

		// Handle HclField type
		if !field.IsNil() {
			config := field.Interface().(*HclField)
			if err := handleHclField(moduleBody, fieldName, config); err != nil {
				return "", err
			}
		}
	}

	return string(hclwrite.Format(f.Bytes())), nil
}

func handleHclField(
	moduleBody *hclwrite.Body,
	attrName string,
	config *HclField,
) error {
	if config == nil {
		return nil
	}

	if config.Type == ConfigTypeDynamic {
		tokens, diags := hclsyntax.LexExpression([]byte(config.Dynamic), "", hcl.Pos{Line: 1, Column: 1})
		if diags.HasErrors() {
			return fmt.Errorf("failed to lex dynamic %s expression: %v", attrName, diags)
		}
		moduleBody.SetAttributeRaw(attrName, ConvertHCLSyntaxToHCLWrite(tokens))
	} else {
		value, err := convertToAttributeValue(config.Static, config.ValueType)
		if err != nil {
			return err
		}
		moduleBody.SetAttributeValue(attrName, value)
	}
	return nil
}

func convertToAttributeValue(value interface{}, valueType ValueType) (cty.Value, error) {
	switch valueType {
	case ValueTypeString:
		return cty.StringVal(value.(string)), nil
	case ValueTypeBool:
		return cty.BoolVal(value.(bool)), nil
	case ValueTypeStringMap:
		m := value.(map[string]string)
		return cty.MapVal(stringMapToValues(m)), nil
	case ValueTypeStringList:
		list := value.([]string)
		return cty.ListVal(stringsToValues(list)), nil
	default:
		return cty.NilVal, fmt.Errorf("unsupported value type: %s", valueType)
	}
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
