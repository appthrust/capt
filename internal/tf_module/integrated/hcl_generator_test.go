package integrated

import (
	"strings"
	"testing"

	"github.com/appthrust/capt/internal/tf_module/eks"
	"github.com/appthrust/capt/internal/tf_module/eks_blueprints_addons"
	"github.com/appthrust/capt/internal/tf_module/vpc"
)

func TestGenerateHCL(t *testing.T) {
	config := &IntegratedConfig{
		Name:   "test-cluster",
		Region: "us-west-2",
		VPCConfig: &vpc.VPCConfig{
			Name:           "test-vpc",
			CIDR:           "10.0.0.0/16",
			PrivateSubnets: []string{"10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"},
			PublicSubnets:  []string{"10.0.101.0/24", "10.0.102.0/24", "10.0.103.0/24"},
			AZs:            []string{"us-west-2a", "us-west-2b", "us-west-2c"},
		},
		EKSConfig: &eks.EKSConfig{
			ClusterName:    "test-cluster",
			ClusterVersion: "1.21",
			Region:         "us-west-2",
			VPC: eks.VPCConfig{
				CIDR:             "10.0.0.0/16",
				PrivateSubnets:   []string{"10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"},
				PublicSubnets:    []string{"10.0.101.0/24", "10.0.102.0/24", "10.0.103.0/24"},
				EnableNATGateway: true,
				SingleNATGateway: true,
			},
			NodeGroups: []eks.NodeGroupConfig{
				{
					Name:         "ng-1",
					InstanceType: "t3.medium",
					DesiredSize:  2,
					MinSize:      1,
					MaxSize:      3,
					DiskSize:     20,
				},
			},
			AddOns: []eks.AddOnConfig{
				{
					Name:    "vpc-cni",
					Version: "latest",
				},
			},
		},
		AddonsConfig: &eks_blueprints_addons.EKSBlueprintsAddonsConfig{
			ClusterName: "test-cluster",
			EKSAddons: map[string]eks_blueprints_addons.EKSAddon{
				"vpc-cni": {ConfigurationValues: "{}"},
			},
		},
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
			Name: "test-cluster",
			Tags: map[string]string{
				"Environment": "test",
			},
		},
		Variables: Variables{
			Name: Variable{
				Type:        "string",
				Description: "Name of the EKS cluster",
				Default:     "test-cluster",
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
	}

	hcl, err := config.GenerateHCL()
	if err != nil {
		t.Fatalf("Failed to generate HCL: %v", err)
	}

	expectedContent := []string{
		"data \"aws_availability_zones\" \"available\"",
		"locals {",
		"module {",
		"variable \"name\"",
		"variable \"vpc_cidr\"",
		"variable \"region\"",
		"10.0.1.0/24",
		"10.0.101.0/24",
		"us-west-2a",
		"test-cluster",
		"1.21",
		"ng-1",
		"t3.medium",
		"vpc-cni",
	}

	for _, content := range expectedContent {
		if !strings.Contains(hcl, content) {
			t.Errorf("Generated HCL does not contain expected content: %s", content)
		}
	}

	t.Logf("Generated HCL:\n%s", hcl)
}
