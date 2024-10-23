package eks_blueprints_addons_v2

import (
	"encoding/json"

	"github.com/appthrust/capt/internal/hcl"
)

// EKSBlueprintsAddonsConfig represents the configuration for EKS Blueprints Addons
type EKSBlueprintsAddonsConfig struct {
	Source                 *hcl.HclField `hcl:"source"`
	Version                *hcl.HclField `hcl:"version"`
	ClusterName            *hcl.HclField `hcl:"cluster_name"`
	ClusterEndpoint        *hcl.HclField `hcl:"cluster_endpoint"`
	ClusterVersion         *hcl.HclField `hcl:"cluster_version"`
	OIDCProviderARN        *hcl.HclField `hcl:"oidc_provider_arn"`
	EnableKarpenter        *hcl.HclField `hcl:"enable_karpenter"`
	KarpenterHelmCacheDir  *hcl.HclField `hcl:"karpenter_helm_cache_dir"`
	KarpenterUseNamePrefix *hcl.HclField `hcl:"karpenter_use_name_prefix"`
	CoreDNSConfigValues    *hcl.HclField `hcl:"coredns_config_values,optional"`
	VPCCNIConfigValues     *hcl.HclField `hcl:"vpc_cni_config_values,optional"`
	KubeProxyConfigValues  *hcl.HclField `hcl:"kube_proxy_config_values,optional"`
	Tags                   *hcl.HclField `hcl:"tags,optional"`
}

// EKSBlueprintsAddonsConfigBuilder is a builder for EKSBlueprintsAddonsConfig
type EKSBlueprintsAddonsConfigBuilder struct {
	config *EKSBlueprintsAddonsConfig
}

// NewEKSBlueprintsAddonsConfig creates a new EKSBlueprintsAddonsConfigBuilder with default values
func NewEKSBlueprintsAddonsConfig() *EKSBlueprintsAddonsConfigBuilder {
	return &EKSBlueprintsAddonsConfigBuilder{
		config: &EKSBlueprintsAddonsConfig{
			Source: &hcl.HclField{
				Type:      hcl.ConfigTypeStatic,
				Static:    "aws-ia/eks-blueprints-addons/aws",
				ValueType: hcl.ValueTypeString,
			},
			Version: &hcl.HclField{
				Type:      hcl.ConfigTypeStatic,
				Static:    "4.32.1",
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
			EnableKarpenter: &hcl.HclField{
				Type:      hcl.ConfigTypeStatic,
				Static:    true,
				ValueType: hcl.ValueTypeBool,
			},
			KarpenterHelmCacheDir: &hcl.HclField{
				Type:      hcl.ConfigTypeStatic,
				Static:    "/tmp/.helmcache",
				ValueType: hcl.ValueTypeString,
			},
			KarpenterUseNamePrefix: &hcl.HclField{
				Type:      hcl.ConfigTypeStatic,
				Static:    false,
				ValueType: hcl.ValueTypeBool,
			},
			CoreDNSConfigValues: &hcl.HclField{
				Type:      hcl.ConfigTypeStatic,
				Static:    getDefaultCoreDNSConfig(),
				ValueType: hcl.ValueTypeString,
			},
			Tags: &hcl.HclField{
				Type:      hcl.ConfigTypeStatic,
				Static:    map[string]string{},
				ValueType: hcl.ValueTypeStringMap,
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
	jsonConfig, _ := json.Marshal(config)
	return string(jsonConfig)
}

// SetClusterName sets the cluster name
func (b *EKSBlueprintsAddonsConfigBuilder) SetClusterName(name string) *EKSBlueprintsAddonsConfigBuilder {
	b.config.ClusterName = &hcl.HclField{
		Type:      hcl.ConfigTypeStatic,
		Static:    name,
		ValueType: hcl.ValueTypeString,
	}
	return b
}

// SetClusterEndpoint sets the cluster endpoint
func (b *EKSBlueprintsAddonsConfigBuilder) SetClusterEndpoint(endpoint string) *EKSBlueprintsAddonsConfigBuilder {
	b.config.ClusterEndpoint = &hcl.HclField{
		Type:      hcl.ConfigTypeStatic,
		Static:    endpoint,
		ValueType: hcl.ValueTypeString,
	}
	return b
}

// SetClusterVersion sets the cluster version
func (b *EKSBlueprintsAddonsConfigBuilder) SetClusterVersion(version string) *EKSBlueprintsAddonsConfigBuilder {
	b.config.ClusterVersion = &hcl.HclField{
		Type:      hcl.ConfigTypeStatic,
		Static:    version,
		ValueType: hcl.ValueTypeString,
	}
	return b
}

// SetOIDCProviderARN sets the OIDC provider ARN
func (b *EKSBlueprintsAddonsConfigBuilder) SetOIDCProviderARN(arn string) *EKSBlueprintsAddonsConfigBuilder {
	b.config.OIDCProviderARN = &hcl.HclField{
		Type:      hcl.ConfigTypeStatic,
		Static:    arn,
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
	b.config.KarpenterHelmCacheDir = &hcl.HclField{
		Type:      hcl.ConfigTypeStatic,
		Static:    cacheDir,
		ValueType: hcl.ValueTypeString,
	}
	return b
}

// SetKarpenterUseNamePrefix sets whether to use name prefix for Karpenter
func (b *EKSBlueprintsAddonsConfigBuilder) SetKarpenterUseNamePrefix(useNamePrefix bool) *EKSBlueprintsAddonsConfigBuilder {
	b.config.KarpenterUseNamePrefix = &hcl.HclField{
		Type:      hcl.ConfigTypeStatic,
		Static:    useNamePrefix,
		ValueType: hcl.ValueTypeBool,
	}
	return b
}

// SetCoreDNSConfigValues sets the CoreDNS configuration values
func (b *EKSBlueprintsAddonsConfigBuilder) SetCoreDNSConfigValues(configValues string) *EKSBlueprintsAddonsConfigBuilder {
	b.config.CoreDNSConfigValues = &hcl.HclField{
		Type:      hcl.ConfigTypeStatic,
		Static:    configValues,
		ValueType: hcl.ValueTypeString,
	}
	return b
}

// SetVPCCNIConfigValues sets the VPC CNI configuration values
func (b *EKSBlueprintsAddonsConfigBuilder) SetVPCCNIConfigValues(configValues string) *EKSBlueprintsAddonsConfigBuilder {
	b.config.VPCCNIConfigValues = &hcl.HclField{
		Type:      hcl.ConfigTypeStatic,
		Static:    configValues,
		ValueType: hcl.ValueTypeString,
	}
	return b
}

// SetKubeProxyConfigValues sets the Kube Proxy configuration values
func (b *EKSBlueprintsAddonsConfigBuilder) SetKubeProxyConfigValues(configValues string) *EKSBlueprintsAddonsConfigBuilder {
	b.config.KubeProxyConfigValues = &hcl.HclField{
		Type:      hcl.ConfigTypeStatic,
		Static:    configValues,
		ValueType: hcl.ValueTypeString,
	}
	return b
}

// AddTag adds a tag to the configuration
func (b *EKSBlueprintsAddonsConfigBuilder) AddTag(key, value string) *EKSBlueprintsAddonsConfigBuilder {
	tags := make(map[string]string)
	if b.config.Tags != nil && b.config.Tags.Type == hcl.ConfigTypeStatic {
		tags = b.config.Tags.Static.(map[string]string)
	}
	tags[key] = value
	b.config.Tags = &hcl.HclField{
		Type:      hcl.ConfigTypeStatic,
		Static:    tags,
		ValueType: hcl.ValueTypeStringMap,
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
