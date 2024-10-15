package integrated

import (
	"fmt"

	"github.com/appthrust/capt/internal/tf_module/eks"
	"github.com/appthrust/capt/internal/tf_module/eks_blueprints_addons"
	"github.com/appthrust/capt/internal/tf_module/vpc"
)

// IntegratedConfig represents the configuration for the integrated module
type IntegratedConfig struct {
	Name         string
	Region       string
	VPCConfig    *vpc.VPCConfig
	EKSConfig    *eks.EKSConfig
	AddonsConfig *eks_blueprints_addons.EKSBlueprintsAddonsConfig
	Tags         map[string]string
	DataSources  DataSources
	Locals       Locals
	Variables    Variables
}

// DataSources represents the data sources used in the integrated module
type DataSources struct {
	AvailabilityZones AvailabilityZones
}

// AvailabilityZones represents the aws_availability_zones data source
type AvailabilityZones struct {
	Filter Filter
}

// Filter represents the filter used in the aws_availability_zones data source
type Filter struct {
	Name   string
	Values []string
}

// Locals represents the local values used in the integrated module
type Locals struct {
	AZs  string
	Name string
	Tags map[string]string
}

// Variables represents the variables used in the integrated module
type Variables struct {
	Name    Variable
	VpcCIDR Variable
	Region  Variable
}

// Variable represents a single variable in the Terraform configuration
type Variable struct {
	Type        string
	Description string
	Default     string
}

// NewIntegratedConfig creates a new IntegratedConfig with default values
func NewIntegratedConfig() (*IntegratedConfig, error) {
	eksBuilder := eks.NewEKSConfigBuilder().SetDefault()
	eksConfig, err := eksBuilder.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build EKSConfig: %w", err)
	}

	return &IntegratedConfig{
		Name:         "eks-cluster",
		Region:       "us-west-2",
		VPCConfig:    vpc.NewVPCConfig(),
		EKSConfig:    eksConfig,
		AddonsConfig: eks_blueprints_addons.NewEKSBlueprintsAddonsConfig(),
		Tags:         make(map[string]string),
		DataSources: DataSources{
			AvailabilityZones: AvailabilityZones{
				Filter: Filter{
					Name:   "opt-in-status",
					Values: []string{"opt-in-not-required"},
				},
			},
		},
		Locals: Locals{
			AZs:  "slice(data.aws_availability_zones.available.names, 0, 3)",
			Name: "try(var.name, basename(path.cwd))",
			Tags: map[string]string{
				"Module":     "basename(path.cwd)",
				"GithubRepo": "github.com/labthrust/terraform-aws",
			},
		},
		Variables: Variables{
			Name: Variable{
				Type:        "string",
				Description: "Name of the EKS cluster",
				Default:     "eks-cluster",
			},
			VpcCIDR: Variable{
				Type:        "string",
				Description: "CIDR block for the VPC",
				Default:     "10.0.0.0/16",
			},
			Region: Variable{
				Type:        "string",
				Description: "AWS region",
				Default:     "us-west-2",
			},
		},
	}, nil
}

// SetName sets the name for the integrated configuration
func (c *IntegratedConfig) SetName(name string) error {
	c.Name = name
	c.VPCConfig.SetName(name + "-vpc")

	eksBuilder := eks.NewEKSConfigBuilder().SetDefault().SetClusterName(name)
	eksConfig, err := eksBuilder.Build()
	if err != nil {
		return fmt.Errorf("failed to build EKSConfig: %w", err)
	}
	c.EKSConfig = eksConfig

	c.Variables.Name.Default = name
	return nil
}

// SetRegion sets the region for the integrated configuration
func (c *IntegratedConfig) SetRegion(region string) error {
	c.Region = region

	eksBuilder := eks.NewEKSConfigBuilder().SetDefault().SetRegion(region)
	eksConfig, err := eksBuilder.Build()
	if err != nil {
		return fmt.Errorf("failed to build EKSConfig: %w", err)
	}
	c.EKSConfig = eksConfig

	c.Variables.Region.Default = region
	return nil
}

// AddTag adds a tag to the integrated configuration
func (c *IntegratedConfig) AddTag(key, value string) {
	c.Tags[key] = value
	c.VPCConfig.AddTag(key, value)
	c.AddonsConfig.AddTag(key, value)
}

// Validate checks if the IntegratedConfig is valid
func (c *IntegratedConfig) Validate() error {
	if err := c.VPCConfig.Validate(); err != nil {
		return err
	}
	if err := c.EKSConfig.Validate(); err != nil {
		return err
	}
	// Add validation for AddonsConfig if needed
	return nil
}
