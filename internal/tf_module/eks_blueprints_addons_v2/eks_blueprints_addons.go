package eks_blueprints_addons_v2

import (
	"encoding/json"

	"github.com/appthrust/capt/internal/hcl"
)

// EKSBlueprintsAddonsConfig represents the configuration for EKS Blueprints Addons
type EKSBlueprintsAddonsConfig struct {
	Source                  *hcl.HclField          `hcl:"source"`
	Version                 *hcl.HclField          `hcl:"version"`
	ClusterName             *hcl.HclField          `hcl:"cluster_name"`
	ClusterEndpoint         *hcl.HclField          `hcl:"cluster_endpoint"`
	ClusterVersion          *hcl.HclField          `hcl:"cluster_version"`
	OIDCProviderARN         *hcl.HclField          `hcl:"oidc_provider_arn"`
	CreateDelayDependencies *hcl.HclField          `hcl:"create_delay_dependencies"`
	EKSAddons               map[string]interface{} `hcl:"eks_addons,block"`
	EnableKarpenter         *hcl.HclField          `hcl:"enable_karpenter"`
	Karpenter               map[string]interface{} `hcl:"karpenter,block"`
	KarpenterNode           map[string]interface{} `hcl:"karpenter_node,block"`
	Tags                    *hcl.HclField          `hcl:"tags"`
}

// EKSBlueprintsAddonsConfigBuilder is a builder for EKSBlueprintsAddonsConfig
type EKSBlueprintsAddonsConfigBuilder struct {
	config *EKSBlueprintsAddonsConfig
}

// NewEKSBlueprintsAddonsConfig creates a new EKSBlueprintsAddonsConfigBuilder with default values
func NewEKSBlueprintsAddonsConfig() *EKSBlueprintsAddonsConfigBuilder {
	karpenterBlock := map[string]interface{}{
		"helm_config": map[string]interface{}{
			"cacheDir": "/tmp/.helmcache",
		},
	}

	karpenterNodeBlock := map[string]interface{}{
		"iam_role_use_name_prefix": false,
	}

	eksAddonsBlock := map[string]interface{}{
		"coredns": map[string]interface{}{
			"configuration_values": getDefaultCoreDNSConfig(),
		},
		"vpc-cni":    map[string]interface{}{},
		"kube-proxy": map[string]interface{}{},
	}

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
			EKSAddons: eksAddonsBlock,
			EnableKarpenter: &hcl.HclField{
				Type:      hcl.ConfigTypeStatic,
				Static:    true,
				ValueType: hcl.ValueTypeBool,
			},
			Karpenter:     karpenterBlock,
			KarpenterNode: karpenterNodeBlock,
			Tags: &hcl.HclField{
				Type:      hcl.ConfigTypeDynamic,
				Dynamic:   "local.tags",
				ValueType: hcl.ValueTypeString,
			},
		},
	}
}

// getDefaultCoreDNSConfig returns the default configuration for CoreDNS
func getDefaultCoreDNSConfig() string {
	config := map[string]interface{}{
		"computeType": "Fargate",
		"resources": map[string]interface{}{
			"limits": map[string]string{
				"cpu":    "0.25",
				"memory": "256M",
			},
			"requests": map[string]string{
				"cpu":    "0.25",
				"memory": "256M",
			},
		},
	}

	jsonBytes, _ := json.MarshalIndent(config, "", "  ")
	return string(jsonBytes)
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
	b.config.Karpenter = map[string]interface{}{
		"helm_config": map[string]interface{}{
			"cacheDir": cacheDir,
		},
	}
	return b
}

// SetKarpenterNodeConfig sets the Karpenter node configuration
func (b *EKSBlueprintsAddonsConfigBuilder) SetKarpenterNodeConfig(useNamePrefix bool) *EKSBlueprintsAddonsConfigBuilder {
	b.config.KarpenterNode = map[string]interface{}{
		"iam_role_use_name_prefix": useNamePrefix,
	}
	return b
}

// SetCoreDNSConfig sets the CoreDNS configuration
func (b *EKSBlueprintsAddonsConfigBuilder) SetCoreDNSConfig(configValues string) *EKSBlueprintsAddonsConfigBuilder {
	eksAddons := b.config.EKSAddons
	if eksAddons == nil {
		eksAddons = make(map[string]interface{})
	}
	eksAddons["coredns"] = map[string]interface{}{
		"configuration_values": configValues,
	}
	b.config.EKSAddons = eksAddons
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
