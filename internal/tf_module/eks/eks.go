package eks_v2

import (
	"fmt"

	"github.com/appthrust/capt/internal/tf_module/hcl"
)

// EKSConfig represents the configuration for an EKS cluster
type EKSConfig struct {
	// Module settings
	Source  *hcl.HclField `hcl:"source"`
	Version *hcl.HclField `hcl:"version"`

	// Cluster settings
	ClusterName           *hcl.HclField `hcl:"cluster_name"`
	ClusterVersion        *hcl.HclField `hcl:"cluster_version"`
	ClusterEndpointPublic *hcl.HclField `hcl:"cluster_endpoint_public_access"`

	// Network settings
	VPCId     *hcl.HclField `hcl:"vpc_id"`
	SubnetIds *hcl.HclField `hcl:"subnet_ids"`

	// Security settings
	CreateClusterSecurityGroup           *hcl.HclField `hcl:"create_cluster_security_group"`
	CreateNodeSecurityGroup              *hcl.HclField `hcl:"create_node_security_group"`
	EnableClusterCreatorAdminPermissions *hcl.HclField `hcl:"enable_cluster_creator_admin_permissions"`

	// Fargate settings
	FargateProfiles *hcl.HclField `hcl:"fargate_profiles"`

	// Tags
	Tags *hcl.HclField `hcl:"tags"`
}

// EKSConfigBuilder is a builder for EKSConfig
type EKSConfigBuilder struct {
	config *EKSConfig
}

// NewEKSConfig creates a new EKSConfigBuilder with default values
func NewEKSConfig() *EKSConfigBuilder {
	defaultFargateProfiles := `{
		karpenter = {
			selectors = [
				{ namespace = "karpenter" }
			]
		}
		kube_system = {
			name = "kube-system"
			selectors = [
				{ namespace = "kube-system" }
			]
		}
	}`

	return &EKSConfigBuilder{
		config: &EKSConfig{
			Source: &hcl.HclField{
				Type:      hcl.ConfigTypeStatic,
				Static:    "terraform-aws-modules/eks/aws",
				ValueType: hcl.ValueTypeString,
			},
			Version: &hcl.HclField{
				Type:      hcl.ConfigTypeStatic,
				Static:    "~> 20.11",
				ValueType: hcl.ValueTypeString,
			},
			ClusterName: &hcl.HclField{
				Type:      hcl.ConfigTypeDynamic,
				Dynamic:   "local.name",
				ValueType: hcl.ValueTypeString,
			},
			ClusterVersion: &hcl.HclField{
				Type:      hcl.ConfigTypeStatic,
				Static:    "1.31",
				ValueType: hcl.ValueTypeString,
			},
			ClusterEndpointPublic: &hcl.HclField{
				Type:      hcl.ConfigTypeStatic,
				Static:    true,
				ValueType: hcl.ValueTypeBool,
			},
			VPCId: &hcl.HclField{
				Type:      hcl.ConfigTypeDynamic,
				Dynamic:   "module.vpc.vpc_id",
				ValueType: hcl.ValueTypeString,
			},
			SubnetIds: &hcl.HclField{
				Type:      hcl.ConfigTypeDynamic,
				Dynamic:   "module.vpc.private_subnets",
				ValueType: hcl.ValueTypeStringList,
			},
			CreateClusterSecurityGroup: &hcl.HclField{
				Type:      hcl.ConfigTypeStatic,
				Static:    false,
				ValueType: hcl.ValueTypeBool,
			},
			CreateNodeSecurityGroup: &hcl.HclField{
				Type:      hcl.ConfigTypeStatic,
				Static:    false,
				ValueType: hcl.ValueTypeBool,
			},
			EnableClusterCreatorAdminPermissions: &hcl.HclField{
				Type:      hcl.ConfigTypeStatic,
				Static:    true,
				ValueType: hcl.ValueTypeBool,
			},
			FargateProfiles: &hcl.HclField{
				Type:      hcl.ConfigTypeDynamic,
				Dynamic:   defaultFargateProfiles,
				ValueType: hcl.ValueTypeString,
			},
			Tags: &hcl.HclField{
				Type:      hcl.ConfigTypeDynamic,
				Dynamic:   `merge(local.tags, { "karpenter.sh/discovery" = local.name })`,
				ValueType: hcl.ValueTypeStringMap,
			},
		},
	}
}

// Builder methods

func (b *EKSConfigBuilder) SetSource(source string) *EKSConfigBuilder {
	b.config.Source = &hcl.HclField{
		Type:      hcl.ConfigTypeStatic,
		Static:    source,
		ValueType: hcl.ValueTypeString,
	}
	return b
}

func (b *EKSConfigBuilder) SetVersion(version string) *EKSConfigBuilder {
	b.config.Version = &hcl.HclField{
		Type:      hcl.ConfigTypeStatic,
		Static:    version,
		ValueType: hcl.ValueTypeString,
	}
	return b
}

func (b *EKSConfigBuilder) SetClusterName(expr string) *EKSConfigBuilder {
	b.config.ClusterName = &hcl.HclField{
		Type:      hcl.ConfigTypeDynamic,
		Dynamic:   expr,
		ValueType: hcl.ValueTypeString,
	}
	return b
}

func (b *EKSConfigBuilder) SetClusterVersion(version string) *EKSConfigBuilder {
	b.config.ClusterVersion = &hcl.HclField{
		Type:      hcl.ConfigTypeStatic,
		Static:    version,
		ValueType: hcl.ValueTypeString,
	}
	return b
}

func (b *EKSConfigBuilder) SetVPCId(expr string) *EKSConfigBuilder {
	b.config.VPCId = &hcl.HclField{
		Type:      hcl.ConfigTypeDynamic,
		Dynamic:   expr,
		ValueType: hcl.ValueTypeString,
	}
	return b
}

func (b *EKSConfigBuilder) SetSubnetIds(expr string) *EKSConfigBuilder {
	b.config.SubnetIds = &hcl.HclField{
		Type:      hcl.ConfigTypeDynamic,
		Dynamic:   expr,
		ValueType: hcl.ValueTypeStringList,
	}
	return b
}

func (b *EKSConfigBuilder) SetClusterEndpointPublicAccess(enable bool) *EKSConfigBuilder {
	b.config.ClusterEndpointPublic = &hcl.HclField{
		Type:      hcl.ConfigTypeStatic,
		Static:    enable,
		ValueType: hcl.ValueTypeBool,
	}
	return b
}

func (b *EKSConfigBuilder) SetCreateSecurityGroups(cluster, node bool) *EKSConfigBuilder {
	b.config.CreateClusterSecurityGroup = &hcl.HclField{
		Type:      hcl.ConfigTypeStatic,
		Static:    cluster,
		ValueType: hcl.ValueTypeBool,
	}
	b.config.CreateNodeSecurityGroup = &hcl.HclField{
		Type:      hcl.ConfigTypeStatic,
		Static:    node,
		ValueType: hcl.ValueTypeBool,
	}
	return b
}

func (b *EKSConfigBuilder) SetEnableClusterCreatorAdminPermissions(enable bool) *EKSConfigBuilder {
	b.config.EnableClusterCreatorAdminPermissions = &hcl.HclField{
		Type:      hcl.ConfigTypeStatic,
		Static:    enable,
		ValueType: hcl.ValueTypeBool,
	}
	return b
}

func (b *EKSConfigBuilder) SetFargateProfiles(expr string) *EKSConfigBuilder {
	b.config.FargateProfiles = &hcl.HclField{
		Type:      hcl.ConfigTypeDynamic,
		Dynamic:   expr,
		ValueType: hcl.ValueTypeString,
	}
	return b
}

func (b *EKSConfigBuilder) SetTags(expr string) *EKSConfigBuilder {
	b.config.Tags = &hcl.HclField{
		Type:      hcl.ConfigTypeDynamic,
		Dynamic:   expr,
		ValueType: hcl.ValueTypeStringMap,
	}
	return b
}

func (b *EKSConfigBuilder) Build() (*EKSConfig, error) {
	if err := b.config.Validate(); err != nil {
		return nil, err
	}
	return b.config, nil
}

func (c *EKSConfig) Validate() error {
	if c.Source == nil || c.Source.Type == hcl.ConfigTypeStatic && c.Source.Static.(string) == "" {
		return fmt.Errorf("EKS module source cannot be empty")
	}
	if c.Version == nil || c.Version.Type == hcl.ConfigTypeStatic && c.Version.Static.(string) == "" {
		return fmt.Errorf("EKS module version cannot be empty")
	}
	if c.ClusterName == nil {
		return fmt.Errorf("EKS cluster name cannot be empty")
	}
	if c.ClusterVersion == nil || c.ClusterVersion.Type == hcl.ConfigTypeStatic && c.ClusterVersion.Static.(string) == "" {
		return fmt.Errorf("EKS cluster version cannot be empty")
	}
	return nil
}

func (c *EKSConfig) GenerateHCL() (string, error) {
	generator := hcl.NewHclGenerator("module", []string{"eks"})
	return generator.GenerateHCL(c)
}
