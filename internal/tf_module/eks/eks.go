package eks

import (
	"fmt"
)

// EKSConfig represents the configuration for an EKS cluster and its associated resources
type EKSConfig struct {
	ClusterName    string            `hcl:"cluster_name"`
	ClusterVersion string            `hcl:"cluster_version"`
	Region         string            `hcl:"region"`
	VPC            VPCConfig         `hcl:"vpc,block"`
	NodeGroups     []NodeGroupConfig `hcl:"node_groups,block"`
	AddOns         []AddOnConfig     `hcl:"cluster_addons,block"`
}

// VPCConfig represents the VPC configuration for the EKS cluster
type VPCConfig struct {
	CIDR             string   `hcl:"cidr"`
	PrivateSubnets   []string `hcl:"private_subnets"`
	PublicSubnets    []string `hcl:"public_subnets"`
	EnableNATGateway bool     `hcl:"enable_nat_gateway"`
	SingleNATGateway bool     `hcl:"single_nat_gateway"`
}

// NodeGroupConfig represents the configuration for an EKS node group
type NodeGroupConfig struct {
	Name         string `hcl:"name,label"`
	InstanceType string `hcl:"instance_types"`
	DesiredSize  int    `hcl:"desired_size"`
	MinSize      int    `hcl:"min_size"`
	MaxSize      int    `hcl:"max_size"`
	DiskSize     int    `hcl:"disk_size"`
}

// AddOnConfig represents the configuration for an EKS add-on
type AddOnConfig struct {
	Name    string `hcl:"name,label"`
	Version string `hcl:"version"`
}

// EKSConfigBuilder is a builder for EKSConfig
type EKSConfigBuilder struct {
	config *EKSConfig
}

// NewEKSConfigBuilder creates a new EKSConfigBuilder with an empty configuration
func NewEKSConfigBuilder() *EKSConfigBuilder {
	return &EKSConfigBuilder{
		config: &EKSConfig{},
	}
}

// SetDefault sets the default values for the EKSConfig
func (b *EKSConfigBuilder) SetDefault() *EKSConfigBuilder {
	b.config = &EKSConfig{
		ClusterName:    "eks-cluster",
		ClusterVersion: "1.31",
		Region:         "us-west-2",
		VPC: VPCConfig{
			CIDR:             "10.0.0.0/16",
			PrivateSubnets:   []string{"10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"},
			PublicSubnets:    []string{"10.0.4.0/24", "10.0.5.0/24", "10.0.6.0/24"},
			EnableNATGateway: true,
			SingleNATGateway: true,
		},
		NodeGroups: []NodeGroupConfig{
			{
				Name:         "ng-1",
				InstanceType: "t3.medium",
				DesiredSize:  2,
				MinSize:      1,
				MaxSize:      3,
				DiskSize:     20,
			},
		},
		AddOns: []AddOnConfig{
			{
				Name:    "vpc-cni",
				Version: "latest",
			},
			{
				Name:    "coredns",
				Version: "latest",
			},
			{
				Name:    "kube-proxy",
				Version: "latest",
			},
		},
	}
	return b
}

// Reset resets the configuration to an empty state
func (b *EKSConfigBuilder) Reset() *EKSConfigBuilder {
	b.config = &EKSConfig{}
	return b
}

// SetClusterName sets the cluster name
func (b *EKSConfigBuilder) SetClusterName(name string) *EKSConfigBuilder {
	b.config.ClusterName = name
	return b
}

// SetClusterVersion sets the cluster version
func (b *EKSConfigBuilder) SetClusterVersion(version string) *EKSConfigBuilder {
	b.config.ClusterVersion = version
	return b
}

// SetRegion sets the AWS region
func (b *EKSConfigBuilder) SetRegion(region string) *EKSConfigBuilder {
	b.config.Region = region
	return b
}

// SetVPCConfig sets the VPC configuration
func (b *EKSConfigBuilder) SetVPCConfig(vpc VPCConfig) *EKSConfigBuilder {
	b.config.VPC = vpc
	return b
}

// SetNodeGroups sets the node groups
func (b *EKSConfigBuilder) SetNodeGroups(nodeGroups []NodeGroupConfig) *EKSConfigBuilder {
	b.config.NodeGroups = nodeGroups
	return b
}

// AddNodeGroup adds a new node group to the configuration
func (b *EKSConfigBuilder) AddNodeGroup(ng NodeGroupConfig) *EKSConfigBuilder {
	b.config.NodeGroups = append(b.config.NodeGroups, ng)
	return b
}

// SetAddOns sets the add-ons
func (b *EKSConfigBuilder) SetAddOns(addOns []AddOnConfig) *EKSConfigBuilder {
	b.config.AddOns = addOns
	return b
}

// AddAddOn adds a new add-on to the configuration
func (b *EKSConfigBuilder) AddAddOn(addon AddOnConfig) *EKSConfigBuilder {
	b.config.AddOns = append(b.config.AddOns, addon)
	return b
}

// Build creates the final EKSConfig
func (b *EKSConfigBuilder) Build() (*EKSConfig, error) {
	if err := b.config.Validate(); err != nil {
		return nil, err
	}
	return b.config, nil
}

// Validate checks if the EKSConfig is valid
func (c *EKSConfig) Validate() error {
	if c.ClusterName == "" {
		return fmt.Errorf("cluster name cannot be empty")
	}
	if c.ClusterVersion == "" {
		return fmt.Errorf("cluster version cannot be empty")
	}
	if c.Region == "" {
		return fmt.Errorf("region cannot be empty")
	}
	if err := c.VPC.Validate(); err != nil {
		return fmt.Errorf("invalid VPC configuration: %w", err)
	}
	for i, ng := range c.NodeGroups {
		if err := ng.Validate(); err != nil {
			return fmt.Errorf("invalid node group configuration at index %d: %w", i, err)
		}
	}
	return nil
}

// Validate checks if the VPCConfig is valid
func (v *VPCConfig) Validate() error {
	if v.CIDR == "" {
		return fmt.Errorf("VPC CIDR cannot be empty")
	}
	if len(v.PrivateSubnets) == 0 {
		return fmt.Errorf("at least one private subnet must be specified")
	}
	if len(v.PublicSubnets) == 0 {
		return fmt.Errorf("at least one public subnet must be specified")
	}
	return nil
}

// Validate checks if the NodeGroupConfig is valid
func (n *NodeGroupConfig) Validate() error {
	if n.Name == "" {
		return fmt.Errorf("node group name cannot be empty")
	}
	if n.InstanceType == "" {
		return fmt.Errorf("instance type cannot be empty")
	}
	if n.MinSize > n.MaxSize {
		return fmt.Errorf("min size cannot be greater than max size")
	}
	if n.DesiredSize < n.MinSize || n.DesiredSize > n.MaxSize {
		return fmt.Errorf("desired size must be between min and max size")
	}
	return nil
}

// ToMap converts the EKSConfig to a map representation
func (c *EKSConfig) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"cluster_name":    c.ClusterName,
		"cluster_version": c.ClusterVersion,
		"region":          c.Region,
		"vpc":             c.VPC.ToMap(),
		"node_groups":     c.nodeGroupsToMap(),
		"add_ons":         c.addOnsToMap(),
	}
}

// ToMap converts the VPCConfig to a map representation
func (v *VPCConfig) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"cidr":               v.CIDR,
		"private_subnets":    v.PrivateSubnets,
		"public_subnets":     v.PublicSubnets,
		"enable_nat_gateway": v.EnableNATGateway,
		"single_nat_gateway": v.SingleNATGateway,
	}
}

// nodeGroupsToMap converts the NodeGroups to a map representation
func (c *EKSConfig) nodeGroupsToMap() []map[string]interface{} {
	result := make([]map[string]interface{}, len(c.NodeGroups))
	for i, ng := range c.NodeGroups {
		result[i] = map[string]interface{}{
			"name":          ng.Name,
			"instance_type": ng.InstanceType,
			"desired_size":  ng.DesiredSize,
			"min_size":      ng.MinSize,
			"max_size":      ng.MaxSize,
			"disk_size":     ng.DiskSize,
		}
	}
	return result
}

// addOnsToMap converts the AddOns to a map representation
func (c *EKSConfig) addOnsToMap() []map[string]interface{} {
	result := make([]map[string]interface{}, len(c.AddOns))
	for i, addon := range c.AddOns {
		result[i] = map[string]interface{}{
			"name":    addon.Name,
			"version": addon.Version,
		}
	}
	return result
}
