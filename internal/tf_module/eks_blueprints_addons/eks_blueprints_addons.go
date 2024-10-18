package eks_blueprints_addons

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

// EKSBlueprintsAddonsConfig represents the configuration for EKS Blueprints Addons
type EKSBlueprintsAddonsConfig struct {
	ClusterName     string              `hcl:"cluster_name"`
	ClusterEndpoint string              `hcl:"cluster_endpoint"`
	ClusterVersion  string              `hcl:"cluster_version"`
	OIDCProviderARN string              `hcl:"oidc_provider_arn"`
	EKSAddons       EKSAddons           `hcl:"eks_addons,block"`
	EnableKarpenter bool                `hcl:"enable_karpenter"`
	Karpenter       KarpenterConfig     `hcl:"karpenter,block"`
	KarpenterNode   KarpenterNodeConfig `hcl:"karpenter_node,block"`
	Tags            map[string]string   `hcl:"tags,optional"`
}

// EKSAddons represents the configuration for EKS addons
type EKSAddons struct {
	CoreDNS   *EKSAddon `hcl:"coredns,block"`
	VPCCni    *EKSAddon `hcl:"vpc-cni,block"`
	KubeProxy *EKSAddon `hcl:"kube-proxy,block"`
}

// EKSAddon represents the configuration for an EKS addon
type EKSAddon struct {
	ConfigurationValues string `hcl:"configuration_values,optional"`
}

// KarpenterConfig represents the configuration for Karpenter
type KarpenterConfig struct {
	HelmConfig KarpenterHelmConfig `hcl:"helm_config,block"`
}

// KarpenterHelmConfig represents the Helm configuration for Karpenter
type KarpenterHelmConfig struct {
	CacheDir string `hcl:"cache_dir"`
}

// KarpenterNodeConfig represents the configuration for Karpenter node
type KarpenterNodeConfig struct {
	IAMRoleUseNamePrefix bool `hcl:"iam_role_use_name_prefix"`
}

// EKSBlueprintsAddonsConfigBuilder is a builder for EKSBlueprintsAddonsConfig
type EKSBlueprintsAddonsConfigBuilder struct {
	config *EKSBlueprintsAddonsConfig
}

// NewEKSBlueprintsAddonsConfig creates a new EKSBlueprintsAddonsConfigBuilder with default values
func NewEKSBlueprintsAddonsConfig() *EKSBlueprintsAddonsConfigBuilder {
	return &EKSBlueprintsAddonsConfigBuilder{
		config: &EKSBlueprintsAddonsConfig{
			ClusterName:     "${module.eks.cluster_name}",
			ClusterEndpoint: "${module.eks.cluster_endpoint}",
			ClusterVersion:  "${module.eks.cluster_version}",
			OIDCProviderARN: "${module.eks.oidc_provider_arn}",
			EKSAddons: EKSAddons{
				CoreDNS:   &EKSAddon{ConfigurationValues: getDefaultCoreDNSConfig()},
				VPCCni:    &EKSAddon{},
				KubeProxy: &EKSAddon{},
			},
			EnableKarpenter: true,
			Karpenter: KarpenterConfig{
				HelmConfig: KarpenterHelmConfig{
					CacheDir: "/tmp/.helmcache",
				},
			},
			KarpenterNode: KarpenterNodeConfig{
				IAMRoleUseNamePrefix: false,
			},
			Tags: map[string]string{},
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
	b.config.ClusterName = name
	return b
}

// SetClusterEndpoint sets the cluster endpoint
func (b *EKSBlueprintsAddonsConfigBuilder) SetClusterEndpoint(endpoint string) *EKSBlueprintsAddonsConfigBuilder {
	b.config.ClusterEndpoint = endpoint
	return b
}

// SetClusterVersion sets the cluster version
func (b *EKSBlueprintsAddonsConfigBuilder) SetClusterVersion(version string) *EKSBlueprintsAddonsConfigBuilder {
	b.config.ClusterVersion = version
	return b
}

// SetOIDCProviderARN sets the OIDC provider ARN
func (b *EKSBlueprintsAddonsConfigBuilder) SetOIDCProviderARN(arn string) *EKSBlueprintsAddonsConfigBuilder {
	b.config.OIDCProviderARN = arn
	return b
}

// SetEKSAddon sets or updates an EKS addon
func (b *EKSBlueprintsAddonsConfigBuilder) SetEKSAddon(name string, addon *EKSAddon) *EKSBlueprintsAddonsConfigBuilder {
	switch name {
	case "coredns":
		b.config.EKSAddons.CoreDNS = addon
	case "vpc-cni":
		b.config.EKSAddons.VPCCni = addon
	case "kube-proxy":
		b.config.EKSAddons.KubeProxy = addon
	}
	return b
}

// SetEnableKarpenter sets whether Karpenter is enabled
func (b *EKSBlueprintsAddonsConfigBuilder) SetEnableKarpenter(enable bool) *EKSBlueprintsAddonsConfigBuilder {
	b.config.EnableKarpenter = enable
	return b
}

// SetKarpenterHelmCacheDir sets the Karpenter Helm cache directory
func (b *EKSBlueprintsAddonsConfigBuilder) SetKarpenterHelmCacheDir(cacheDir string) *EKSBlueprintsAddonsConfigBuilder {
	b.config.Karpenter.HelmConfig.CacheDir = cacheDir
	return b
}

// SetKarpenterNodeIAMRoleUseNamePrefix sets whether to use name prefix for Karpenter node IAM role
func (b *EKSBlueprintsAddonsConfigBuilder) SetKarpenterNodeIAMRoleUseNamePrefix(useNamePrefix bool) *EKSBlueprintsAddonsConfigBuilder {
	b.config.KarpenterNode.IAMRoleUseNamePrefix = useNamePrefix
	return b
}

// AddTag adds a tag to the configuration
func (b *EKSBlueprintsAddonsConfigBuilder) AddTag(key, value string) *EKSBlueprintsAddonsConfigBuilder {
	if b.config.Tags == nil {
		b.config.Tags = make(map[string]string)
	}
	b.config.Tags[key] = value
	return b
}

// Build creates the final EKSBlueprintsAddonsConfig
func (b *EKSBlueprintsAddonsConfigBuilder) Build() (*EKSBlueprintsAddonsConfig, error) {
	if err := b.config.Validate(); err != nil {
		return nil, err
	}
	return b.config, nil
}

// Validate checks if the EKSBlueprintsAddonsConfig is valid
func (c *EKSBlueprintsAddonsConfig) Validate() error {
	if c.ClusterName == "" {
		return fmt.Errorf("cluster name cannot be empty")
	}
	if c.ClusterEndpoint == "" {
		return fmt.Errorf("cluster endpoint cannot be empty")
	}
	if c.ClusterVersion == "" {
		return fmt.Errorf("cluster version cannot be empty")
	}
	if c.OIDCProviderARN == "" {
		return fmt.Errorf("OIDC provider ARN cannot be empty")
	}
	return nil
}

// GenerateHCL generates HCL configuration for EKSBlueprintsAddonsConfig
func (c *EKSBlueprintsAddonsConfig) GenerateHCL() (string, error) {
	f := hclwrite.NewEmptyFile()
	gohcl.EncodeIntoBody(c, f.Body())
	return string(f.Bytes()), nil
}
