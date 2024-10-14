package eks_blueprints_addons

import (
	"encoding/json"
)

// EKSBlueprintsAddonsConfig represents the configuration for EKS Blueprints Addons
type EKSBlueprintsAddonsConfig struct {
	ClusterName     string
	ClusterEndpoint string
	ClusterVersion  string
	OIDCProviderARN string
	EKSAddons       map[string]EKSAddon
	EnableKarpenter bool
	Karpenter       KarpenterConfig
	KarpenterNode   KarpenterNodeConfig
	Tags            map[string]string
}

// EKSAddon represents the configuration for an EKS addon
type EKSAddon struct {
	ConfigurationValues string
}

// KarpenterConfig represents the configuration for Karpenter
type KarpenterConfig struct {
	HelmConfig KarpenterHelmConfig
}

// KarpenterHelmConfig represents the Helm configuration for Karpenter
type KarpenterHelmConfig struct {
	CacheDir string
}

// KarpenterNodeConfig represents the configuration for Karpenter node
type KarpenterNodeConfig struct {
	IAMRoleUseNamePrefix bool
}

// NewEKSBlueprintsAddonsConfig creates a new EKSBlueprintsAddonsConfig with default values
func NewEKSBlueprintsAddonsConfig() *EKSBlueprintsAddonsConfig {
	return &EKSBlueprintsAddonsConfig{
		ClusterName:     "${module.eks.cluster_name}",
		ClusterEndpoint: "${module.eks.cluster_endpoint}",
		ClusterVersion:  "${module.eks.cluster_version}",
		OIDCProviderARN: "${module.eks.oidc_provider_arn}",
		EKSAddons: map[string]EKSAddon{
			"coredns": {
				ConfigurationValues: getDefaultCoreDNSConfig(),
			},
			"vpc-cni":    {},
			"kube-proxy": {},
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
func (c *EKSBlueprintsAddonsConfig) SetClusterName(name string) {
	c.ClusterName = name
}

// SetClusterEndpoint sets the cluster endpoint
func (c *EKSBlueprintsAddonsConfig) SetClusterEndpoint(endpoint string) {
	c.ClusterEndpoint = endpoint
}

// SetClusterVersion sets the cluster version
func (c *EKSBlueprintsAddonsConfig) SetClusterVersion(version string) {
	c.ClusterVersion = version
}

// SetOIDCProviderARN sets the OIDC provider ARN
func (c *EKSBlueprintsAddonsConfig) SetOIDCProviderARN(arn string) {
	c.OIDCProviderARN = arn
}

// AddEKSAddon adds or updates an EKS addon
func (c *EKSBlueprintsAddonsConfig) AddEKSAddon(name string, addon EKSAddon) {
	c.EKSAddons[name] = addon
}

// SetEnableKarpenter sets whether Karpenter is enabled
func (c *EKSBlueprintsAddonsConfig) SetEnableKarpenter(enable bool) {
	c.EnableKarpenter = enable
}

// SetKarpenterHelmCacheDir sets the Karpenter Helm cache directory
func (c *EKSBlueprintsAddonsConfig) SetKarpenterHelmCacheDir(cacheDir string) {
	c.Karpenter.HelmConfig.CacheDir = cacheDir
}

// SetKarpenterNodeIAMRoleUseNamePrefix sets whether to use name prefix for Karpenter node IAM role
func (c *EKSBlueprintsAddonsConfig) SetKarpenterNodeIAMRoleUseNamePrefix(useNamePrefix bool) {
	c.KarpenterNode.IAMRoleUseNamePrefix = useNamePrefix
}

// AddTag adds a tag to the configuration
func (c *EKSBlueprintsAddonsConfig) AddTag(key, value string) {
	c.Tags[key] = value
}
