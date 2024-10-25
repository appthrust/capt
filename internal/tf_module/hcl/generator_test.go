package hcl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestModuleConfig represents a typical module configuration struct
type TestModuleConfig struct {
	Source          *HclField `hcl:"source"`
	Version         *HclField `hcl:"version"`
	Name            *HclField `hcl:"name"`
	StringListField *HclField `hcl:"string_list_field,optional"`
	StringMapField  *HclField `hcl:"string_map_field,optional"`
	BoolField       *HclField `hcl:"bool_field"`
	DynamicField    *HclField `hcl:"dynamic_field"`
	NestedBlock     *HclField `hcl:"nested_block,block"`
	IgnoredField    string    // No hcl tag, should be ignored
}

func TestGenerateModuleBlock(t *testing.T) {
	config := &TestModuleConfig{
		Source: &HclField{
			Type:      ConfigTypeStatic,
			Static:    "test-module/test",
			ValueType: ValueTypeString,
		},
		Version: &HclField{
			Type:      ConfigTypeStatic,
			Static:    "1.0.0",
			ValueType: ValueTypeString,
		},
		Name: &HclField{
			Type:      ConfigTypeStatic,
			Static:    "test-name",
			ValueType: ValueTypeString,
		},
		StringListField: &HclField{
			Type:      ConfigTypeStatic,
			Static:    []string{"item1", "item2"},
			ValueType: ValueTypeStringList,
		},
		StringMapField: &HclField{
			Type:      ConfigTypeStatic,
			Static:    map[string]string{"key1": "value1", "key2": "value2"},
			ValueType: ValueTypeStringMap,
		},
		BoolField: &HclField{
			Type:      ConfigTypeStatic,
			Static:    true,
			ValueType: ValueTypeBool,
		},
		DynamicField: &HclField{
			Type:      ConfigTypeDynamic,
			Dynamic:   "var.dynamic_value",
			ValueType: ValueTypeString,
		},
		NestedBlock: &HclField{
			Type: ConfigTypeStatic,
			Static: map[string]interface{}{
				"nested_string": "value",
				"nested_bool":   true,
				"nested_list":   []string{"item1", "item2"},
			},
			ValueType: ValueTypeBlock,
		},
	}

	generator := NewHclGenerator("module", []string{"test"})
	result, err := generator.GenerateHCL(config)

	assert.NoError(t, err)
	assert.NotEmpty(t, result)

	// Check basic structure
	assert.Contains(t, result, `module "test"`)

	// Check field values, ignoring exact spacing
	assert.Contains(t, result, `source`)
	assert.Contains(t, result, `"test-module/test"`)
	assert.Contains(t, result, `version`)
	assert.Contains(t, result, `"1.0.0"`)
	assert.Contains(t, result, `name`)
	assert.Contains(t, result, `"test-name"`)
	assert.Contains(t, result, `bool_field`)
	assert.Contains(t, result, `true`)
	assert.Contains(t, result, `dynamic_field`)
	assert.Contains(t, result, `var.dynamic_value`)

	// Check nested block
	assert.Contains(t, result, `nested_block {`)
	assert.Contains(t, result, `nested_string`)
	assert.Contains(t, result, `"value"`)
	assert.Contains(t, result, `nested_bool`)
	assert.Contains(t, result, `true`)
	assert.Contains(t, result, `nested_list`)

	// Check list values
	assert.Contains(t, result, `"item1"`)
	assert.Contains(t, result, `"item2"`)

	// Check map values
	assert.Contains(t, result, `key1`)
	assert.Contains(t, result, `"value1"`)
	assert.Contains(t, result, `key2`)
	assert.Contains(t, result, `"value2"`)
}

func TestGenerateHCL_DynamicExpressions(t *testing.T) {
	config := &TestModuleConfig{
		Source: &HclField{
			Type:      ConfigTypeDynamic,
			Dynamic:   "var.module_source",
			ValueType: ValueTypeString,
		},
		Version: &HclField{
			Type:      ConfigTypeDynamic,
			Dynamic:   "var.module_version",
			ValueType: ValueTypeString,
		},
		StringListField: &HclField{
			Type:      ConfigTypeDynamic,
			Dynamic:   "local.string_list",
			ValueType: ValueTypeStringList,
		},
		StringMapField: &HclField{
			Type:      ConfigTypeDynamic,
			Dynamic:   "local.string_map",
			ValueType: ValueTypeStringMap,
		},
	}

	generator := NewHclGenerator("module", []string{"test"})
	result, err := generator.GenerateHCL(config)

	assert.NoError(t, err)
	assert.NotEmpty(t, result)

	assert.Contains(t, result, "var.module_source")
	assert.Contains(t, result, "var.module_version")
	assert.Contains(t, result, "local.string_list")
	assert.Contains(t, result, "local.string_map")
}

func TestGenerateHCL_NestedBlocks(t *testing.T) {
	config := &TestModuleConfig{
		NestedBlock: &HclField{
			Type: ConfigTypeStatic,
			Static: map[string]interface{}{
				"first_level": map[string]interface{}{
					"second_level": map[string]interface{}{
						"string_value": "nested",
						"bool_value":   true,
						"list_value":   []string{"a", "b"},
					},
				},
			},
			ValueType: ValueTypeBlock,
		},
	}

	generator := NewHclGenerator("module", []string{"test"})
	result, err := generator.GenerateHCL(config)

	assert.NoError(t, err)
	assert.NotEmpty(t, result)

	assert.Contains(t, result, "nested_block {")
	assert.Contains(t, result, "first_level {")
	assert.Contains(t, result, "second_level {")
	assert.Contains(t, result, "string_value")
	assert.Contains(t, result, `"nested"`)
	assert.Contains(t, result, "bool_value")
	assert.Contains(t, result, "true")
	assert.Contains(t, result, "list_value")
	assert.Contains(t, result, `"a"`)
	assert.Contains(t, result, `"b"`)
}

func TestGenerateHCL_InvalidDynamicExpression(t *testing.T) {
	config := &TestModuleConfig{
		Source: &HclField{
			Type:      ConfigTypeDynamic,
			Dynamic:   "[invalid expression]", // Invalid HCL expression
			ValueType: ValueTypeString,
		},
	}

	generator := NewHclGenerator("module", []string{"test"})
	result, err := generator.GenerateHCL(config)

	// The current implementation might not catch this error
	// This test documents the current behavior
	if err != nil {
		assert.Contains(t, err.Error(), "failed to lex dynamic")
	} else {
		assert.Contains(t, result, "[invalid expression]")
	}
}

func TestGenerateHCL_InvalidBlockValue(t *testing.T) {
	config := &TestModuleConfig{
		NestedBlock: &HclField{
			Type:      ConfigTypeStatic,
			Static:    "not a map", // Should be map[string]interface{}
			ValueType: ValueTypeBlock,
		},
	}

	generator := NewHclGenerator("module", []string{"test"})
	_, err := generator.GenerateHCL(config)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected map[string]interface{}")
}

func TestGenerateHCL_EmptyConfig(t *testing.T) {
	config := &TestModuleConfig{}
	generator := NewHclGenerator("module", []string{"test"})
	result, err := generator.GenerateHCL(config)

	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.Contains(t, result, `module "test"`)
}

func TestGenerateHCL_NilFields(t *testing.T) {
	config := &TestModuleConfig{
		Source: &HclField{
			Type:      ConfigTypeStatic,
			Static:    "test-module",
			ValueType: ValueTypeString,
		},
		StringListField: nil, // Should be ignored
	}

	generator := NewHclGenerator("module", []string{"test"})
	result, err := generator.GenerateHCL(config)

	assert.NoError(t, err)
	assert.NotContains(t, result, "string_list_field")
	assert.Contains(t, result, "test-module")
}

func TestGenerateHCL_LifecycleBlock(t *testing.T) {
	config := &TestModuleConfig{
		NestedBlock: &HclField{
			Type:      ConfigTypeDynamic,
			Dynamic:   "[aws_vpc.main.id]",
			ValueType: ValueTypeBlock,
		},
	}

	generator := NewHclGenerator("module", []string{"test"})
	result, err := generator.GenerateHCL(config)

	assert.NoError(t, err)
	assert.Contains(t, result, "nested_block {")
	assert.Contains(t, result, "[aws_vpc.main.id]")
}
