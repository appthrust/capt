package eks

import (
	"fmt"
)

// EKSConfig represents the configuration for an EKS cluster and its associated resources
type EKSConfig struct {
	ClusterName    string
	ClusterVersion string
	Region         string
	VPC            VPCConfig
	NodeGroups     []NodeGroupConfig
	AddOns         []AddOnConfig
}

// VPCConfig represents the VPC configuration for the EKS cluster
type VPCConfig struct {
	CIDR             string
	PrivateSubnets   []string
	PublicSubnets    []string
	EnableNATGateway bool
	SingleNATGateway bool
}

// NodeGroupConfig represents the configuration for an EKS node group
type NodeGroupConfig struct {
	Name         string
	InstanceType string
	DesiredSize  int
	MinSize      int
	MaxSize      int
	DiskSize     int
}

// AddOnConfig represents the configuration for an EKS add-on
type AddOnConfig struct {
	Name    string
	Version string
}

// NewEKSConfig creates a new EKSConfig with default values
func NewEKSConfig() *EKSConfig {
	return &EKSConfig{
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
}

// SetClusterName sets the cluster name
func (c *EKSConfig) SetClusterName(name string) {
	c.ClusterName = name
}

// SetRegion sets the AWS region
func (c *EKSConfig) SetRegion(region string) {
	c.Region = region
}

// AddNodeGroup adds a new node group to the configuration
func (c *EKSConfig) AddNodeGroup(ng NodeGroupConfig) {
	c.NodeGroups = append(c.NodeGroups, ng)
}

// SetVPCCIDR sets the VPC CIDR
func (c *EKSConfig) SetVPCCIDR(cidr string) {
	c.VPC.CIDR = cidr
}

// AddAddOn adds a new add-on to the configuration
func (c *EKSConfig) AddAddOn(addon AddOnConfig) {
	c.AddOns = append(c.AddOns, addon)
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
