package eks_blueprints_addons_v2

import (
	"github.com/appthrust/capt/internal/hcl"
)

// EKSBlueprintsAddonsConfig represents the configuration for EKS Blueprints Addons
type EKSBlueprintsAddonsConfig struct {
	Source                  *hcl.HclField `hcl:"source"`
	Version                 *hcl.HclField `hcl:"version"`
	ClusterName             *hcl.HclField `hcl:"cluster_name"`
	ClusterEndpoint         *hcl.HclField `hcl:"cluster_endpoint"`
	ClusterVersion          *hcl.HclField `hcl:"cluster_version"`
	OIDCProviderARN         *hcl.HclField `hcl:"oidc_provider_arn"`
	CreateDelayDependencies *hcl.HclField `hcl:"create_delay_dependencies"`
	EKSAddons               *hcl.HclField `hcl:"eks_addons"`
	EnableKarpenter         *hcl.HclField `hcl:"enable_karpenter"`
	Karpenter               *hcl.HclField `hcl:"karpenter"`
	KarpenterNode           *hcl.HclField `hcl:"karpenter_node"`
	Tags                    *hcl.HclField `hcl:"tags"`
}

// EKSBlueprintsAddonsConfigBuilder is a builder for EKSBlueprintsAddonsConfig
type EKSBlueprintsAddonsConfigBuilder struct {
	config *EKSBlueprintsAddonsConfig
}

// NewEKSBlueprintsAddonsConfig creates a new EKSBlueprintsAddonsConfigBuilder with default values
func NewEKSBlueprintsAddonsConfig() *EKSBlueprintsAddonsConfigBuilder {
	eksAddonsExpr := `{
		coredns = {
			configuration_values = jsonencode({
				computeType = "Fargate"
				resources = {
					limits = {
						cpu = "0.25"
						memory = "256M"
					}
					requests = {
						cpu = "0.25"
						memory = "256M"
					}
				}
			})
		}
		vpc-cni = {}
		kube-proxy = {}
	}`

	karpenterExpr := `{
		helm_config = {
			cacheDir = "/tmp/.helmcache"
		}
	}`

	karpenterNodeExpr := `{
		iam_role_use_name_prefix = false
	}`

	return &EKSBlueprintsAddonsConfigBuilder{
		config: &EKSBlueprintsAddonsConfig{
			Source: &hcl.HclField{
				Type:      hcl.ConfigTypeStatic,
				Static:    "aws-ia/eks-blueprints-addons/aws",
				ValueType: hcl.ValueTypeString,
			},
			Version: &hcl.HclField{
				Type:      hcl.ConfigTypeStatic,
				Static:    "~> 1.16",
				ValueType: hcl.ValueTypeString,
			},
			ClusterName: &hcl.HclField{
				Type:      hcl.ConfigTypeDynamic,
				Dynamic:   "module.eks.cluster_name",
				ValueType: hcl.ValueTypeString,
			},
			ClusterEndpoint: &hcl.HclField{
				Type:      hcl.ConfigTypeDynamic,
				Dynamic:   "module.eks.cluster_endpoint",
				ValueType: hcl.ValueTypeString,
			},
			ClusterVersion: &hcl.HclField{
				Type:      hcl.ConfigTypeDynamic,
				Dynamic:   "module.eks.cluster_version",
				ValueType: hcl.ValueTypeString,
			},
			OIDCProviderARN: &hcl.HclField{
				Type:      hcl.ConfigTypeDynamic,
				Dynamic:   "module.eks.oidc_provider_arn",
				ValueType: hcl.ValueTypeString,
			},
			CreateDelayDependencies: &hcl.HclField{
				Type:      hcl.ConfigTypeDynamic,
				Dynamic:   "[for prof in module.eks.fargate_profiles : prof.fargate_profile_arn]",
				ValueType: hcl.ValueTypeString,
			},
			EKSAddons: &hcl.HclField{
				Type:      hcl.ConfigTypeDynamic,
				Dynamic:   eksAddonsExpr,
				ValueType: hcl.ValueTypeString,
			},
			EnableKarpenter: &hcl.HclField{
				Type:      hcl.ConfigTypeStatic,
				Static:    true,
				ValueType: hcl.ValueTypeBool,
			},
			Karpenter: &hcl.HclField{
				Type:      hcl.ConfigTypeDynamic,
				Dynamic:   karpenterExpr,
				ValueType: hcl.ValueTypeString,
			},
			KarpenterNode: &hcl.HclField{
				Type:      hcl.ConfigTypeDynamic,
				Dynamic:   karpenterNodeExpr,
				ValueType: hcl.ValueTypeString,
			},
			Tags: &hcl.HclField{
				Type:      hcl.ConfigTypeDynamic,
				Dynamic:   "local.tags",
				ValueType: hcl.ValueTypeString,
			},
		},
	}
}

// SetClusterName sets the cluster name
func (b *EKSBlueprintsAddonsConfigBuilder) SetClusterName(name string) *EKSBlueprintsAddonsConfigBuilder {
	b.config.ClusterName = &hcl.HclField{
		Type:      hcl.ConfigTypeDynamic,
		Dynamic:   name,
		ValueType: hcl.ValueTypeString,
	}
	return b
}

// SetClusterEndpoint sets the cluster endpoint
func (b *EKSBlueprintsAddonsConfigBuilder) SetClusterEndpoint(endpoint string) *EKSBlueprintsAddonsConfigBuilder {
	b.config.ClusterEndpoint = &hcl.HclField{
		Type:      hcl.ConfigTypeDynamic,
		Dynamic:   endpoint,
		ValueType: hcl.ValueTypeString,
	}
	return b
}

// SetClusterVersion sets the cluster version
func (b *EKSBlueprintsAddonsConfigBuilder) SetClusterVersion(version string) *EKSBlueprintsAddonsConfigBuilder {
	b.config.ClusterVersion = &hcl.HclField{
		Type:      hcl.ConfigTypeDynamic,
		Dynamic:   version,
		ValueType: hcl.ValueTypeString,
	}
	return b
}

// SetOIDCProviderARN sets the OIDC provider ARN
func (b *EKSBlueprintsAddonsConfigBuilder) SetOIDCProviderARN(arn string) *EKSBlueprintsAddonsConfigBuilder {
	b.config.OIDCProviderARN = &hcl.HclField{
		Type:      hcl.ConfigTypeDynamic,
		Dynamic:   arn,
		ValueType: hcl.ValueTypeString,
	}
	return b
}

// SetEnableKarpenter sets whether Karpenter is enabled
func (b *EKSBlueprintsAddonsConfigBuilder) SetEnableKarpenter(enable bool) *EKSBlueprintsAddonsConfigBuilder {
	b.config.EnableKarpenter = &hcl.HclField{
		Type:      hcl.ConfigTypeStatic,
		Static:    enable,
		ValueType: hcl.ValueTypeBool,
	}
	return b
}

// SetKarpenterHelmCacheDir sets the Karpenter Helm cache directory
func (b *EKSBlueprintsAddonsConfigBuilder) SetKarpenterHelmCacheDir(cacheDir string) *EKSBlueprintsAddonsConfigBuilder {
	expr := `{
		helm_config = {
			cacheDir = "` + cacheDir + `"
		}
	}`
	b.config.Karpenter = &hcl.HclField{
		Type:      hcl.ConfigTypeDynamic,
		Dynamic:   expr,
		ValueType: hcl.ValueTypeString,
	}
	return b
}

// SetKarpenterNodeConfig sets the Karpenter node configuration
func (b *EKSBlueprintsAddonsConfigBuilder) SetKarpenterNodeConfig(useNamePrefix bool) *EKSBlueprintsAddonsConfigBuilder {
	expr := `{
		iam_role_use_name_prefix = ` + boolToString(useNamePrefix) + `
	}`
	b.config.KarpenterNode = &hcl.HclField{
		Type:      hcl.ConfigTypeDynamic,
		Dynamic:   expr,
		ValueType: hcl.ValueTypeString,
	}
	return b
}

// SetCoreDNSConfig sets the CoreDNS configuration
func (b *EKSBlueprintsAddonsConfigBuilder) SetCoreDNSConfig(configValues string) *EKSBlueprintsAddonsConfigBuilder {
	expr := `{
		coredns = {
			configuration_values = ` + configValues + `
		}
		vpc-cni = {}
		kube-proxy = {}
	}`
	b.config.EKSAddons = &hcl.HclField{
		Type:      hcl.ConfigTypeDynamic,
		Dynamic:   expr,
		ValueType: hcl.ValueTypeString,
	}
	return b
}

// Build creates the final EKSBlueprintsAddonsConfig
func (b *EKSBlueprintsAddonsConfigBuilder) Build() (*EKSBlueprintsAddonsConfig, error) {
	return b.config, nil
}

// GenerateHCL generates HCL configuration for EKSBlueprintsAddonsConfig
func (c *EKSBlueprintsAddonsConfig) GenerateHCL() (string, error) {
	generator := hcl.NewHclGenerator("module", []string{"eks_blueprints_addons"})
	return generator.GenerateHCL(c)
}

func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
