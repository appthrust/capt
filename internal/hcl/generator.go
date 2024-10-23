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
		tagParts := strings.Split(hclTag, ",")
		fieldName := tagParts[0]
		isBlock := len(tagParts) > 1 && tagParts[1] == "block"

		if field.IsNil() {
			continue
		}

		// Handle map type for blocks
		if field.Type().Kind() == reflect.Map {
			if err := handleMapField(moduleBody, fieldName, field.Interface()); err != nil {
				return "", err
			}
			continue
		}

		// Handle HclField type
		config := field.Interface().(*HclField)
		if err := handleHclField(moduleBody, fieldName, config, isBlock); err != nil {
			return "", err
		}
	}

	return string(hclwrite.Format(f.Bytes())), nil
}

func handleMapField(body *hclwrite.Body, blockName string, value interface{}) error {
	m, ok := value.(map[string]interface{})
	if !ok {
		return fmt.Errorf("expected map[string]interface{}, got %T", value)
	}

	block := body.AppendNewBlock(blockName, nil)
	blockBody := block.Body()

	for k, v := range m {
		switch vt := v.(type) {
		case map[string]interface{}:
			if err := handleMapField(blockBody, k, vt); err != nil {
				return err
			}
		case string:
			blockBody.SetAttributeValue(k, cty.StringVal(vt))
		case bool:
			blockBody.SetAttributeValue(k, cty.BoolVal(vt))
		case []string:
			values := make([]cty.Value, len(vt))
			for i, s := range vt {
				values[i] = cty.StringVal(s)
			}
			blockBody.SetAttributeValue(k, cty.ListVal(values))
		default:
			return fmt.Errorf("unsupported type for key %s: %T", k, v)
		}
	}

	return nil
}

func handleHclField(
	moduleBody *hclwrite.Body,
	attrName string,
	config *HclField,
	isBlock bool,
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
		if isBlock {
			return handleMapField(moduleBody, attrName, config.Static)
		} else {
			value, err := convertToAttributeValue(config.Static, config.ValueType)
			if err != nil {
				return err
			}
			moduleBody.SetAttributeValue(attrName, value)
		}
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
	case ValueTypeBlock:
		// Blocks are handled separately by handleMapField
		return cty.NilVal, nil
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
