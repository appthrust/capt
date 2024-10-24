package aws_eks_access_entry

import (
	"fmt"

	"github.com/appthrust/capt/internal/hcl"
)

// AccessEntryConfig represents the configuration for an AWS EKS access entry
type AccessEntryConfig struct {
	ClusterName      *hcl.HclField `hcl:"cluster_name"`
	PrincipalARN     *hcl.HclField `hcl:"principal_arn"`
	KubernetesGroups *hcl.HclField `hcl:"kubernetes_groups"`
	Type             *hcl.HclField `hcl:"type"`
	Lifecycle        *hcl.HclField `hcl:"lifecycle,block"`
}

// AccessEntryConfigBuilder is a builder for AccessEntryConfig
type AccessEntryConfigBuilder struct {
	config *AccessEntryConfig
}

// NewAccessEntryConfig creates a new AccessEntryConfigBuilder with default values
func NewAccessEntryConfig() *AccessEntryConfigBuilder {
	return &AccessEntryConfigBuilder{
		config: &AccessEntryConfig{
			ClusterName: &hcl.HclField{
				Type:      hcl.ConfigTypeDynamic,
				Dynamic:   "module.eks.cluster_name",
				ValueType: hcl.ValueTypeString,
			},
			PrincipalARN: &hcl.HclField{
				Type:      hcl.ConfigTypeDynamic,
				Dynamic:   "module.eks_blueprints_addons.karpenter.node_iam_role_arn",
				ValueType: hcl.ValueTypeString,
			},
			KubernetesGroups: &hcl.HclField{
				Type:      hcl.ConfigTypeDynamic,
				Dynamic:   "[]",
				ValueType: hcl.ValueTypeStringList,
			},
			Type: &hcl.HclField{
				Type:      hcl.ConfigTypeStatic,
				Static:    "EC2_LINUX",
				ValueType: hcl.ValueTypeString,
			},
			Lifecycle: &hcl.HclField{
				Type:      hcl.ConfigTypeDynamic,
				Dynamic:   "[kubernetes_groups]",
				ValueType: hcl.ValueTypeBlock,
			},
		},
	}
}

// Builder methods

func (b *AccessEntryConfigBuilder) SetClusterName(expr string) *AccessEntryConfigBuilder {
	b.config.ClusterName = &hcl.HclField{
		Type:      hcl.ConfigTypeDynamic,
		Dynamic:   expr,
		ValueType: hcl.ValueTypeString,
	}
	return b
}

func (b *AccessEntryConfigBuilder) SetPrincipalARN(expr string) *AccessEntryConfigBuilder {
	b.config.PrincipalARN = &hcl.HclField{
		Type:      hcl.ConfigTypeDynamic,
		Dynamic:   expr,
		ValueType: hcl.ValueTypeString,
	}
	return b
}

func (b *AccessEntryConfigBuilder) SetKubernetesGroups(groups []string) *AccessEntryConfigBuilder {
	b.config.KubernetesGroups = &hcl.HclField{
		Type:      hcl.ConfigTypeStatic,
		Static:    groups,
		ValueType: hcl.ValueTypeStringList,
	}
	return b
}

func (b *AccessEntryConfigBuilder) SetKubernetesGroupsExpression(expr string) *AccessEntryConfigBuilder {
	b.config.KubernetesGroups = &hcl.HclField{
		Type:      hcl.ConfigTypeDynamic,
		Dynamic:   expr,
		ValueType: hcl.ValueTypeStringList,
	}
	return b
}

func (b *AccessEntryConfigBuilder) SetType(entryType string) *AccessEntryConfigBuilder {
	b.config.Type = &hcl.HclField{
		Type:      hcl.ConfigTypeStatic,
		Static:    entryType,
		ValueType: hcl.ValueTypeString,
	}
	return b
}

func (b *AccessEntryConfigBuilder) Build() (*AccessEntryConfig, error) {
	if err := b.config.Validate(); err != nil {
		return nil, err
	}
	return b.config, nil
}

func (c *AccessEntryConfig) Validate() error {
	if c.ClusterName == nil {
		return fmt.Errorf("cluster_name cannot be empty")
	}
	if c.PrincipalARN == nil {
		return fmt.Errorf("principal_arn cannot be empty")
	}
	if c.Type == nil || c.Type.Type == hcl.ConfigTypeStatic && c.Type.Static.(string) == "" {
		return fmt.Errorf("type cannot be empty")
	}
	return nil
}

func (c *AccessEntryConfig) GenerateHCL() (string, error) {
	generator := hcl.NewHclGenerator("resource", []string{"aws_eks_access_entry", "karpenter_node_access_entry"})
	return generator.GenerateHCL(c)
}
