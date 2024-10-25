package hcl

// ConfigType represents the type of configuration (static or dynamic)
type ConfigType string

const (
	ConfigTypeStatic  ConfigType = "static"
	ConfigTypeDynamic ConfigType = "dynamic"
)

// ValueType represents the type of value stored in HclField
type ValueType string

const (
	ValueTypeString     ValueType = "string"
	ValueTypeBool       ValueType = "bool"
	ValueTypeStringMap  ValueType = "string_map"
	ValueTypeStringList ValueType = "string_list"
	ValueTypeBlock      ValueType = "block"
)

// HclField represents a field in HCL that can be either static or dynamic
type HclField struct {
	Type      ConfigType  `hcl:"type" json:"type"`
	Static    interface{} `hcl:"static,optional" json:"static,omitempty"`
	Dynamic   string      `hcl:"dynamic,optional" json:"dynamic,omitempty"`
	ValueType ValueType   `hcl:"value_type" json:"value_type"`
}
