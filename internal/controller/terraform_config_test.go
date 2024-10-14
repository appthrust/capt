package controller

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	infrastructurev1beta1 "github.com/appthrust/capt/api/v1beta1"
	"github.com/hashicorp/hcl/v2/hclparse"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGenerateTerraformConfig(t *testing.T) {
	cluster := &infrastructurev1beta1.CAPTCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-cluster",
		},
		Spec: infrastructurev1beta1.CAPTClusterSpec{
			Region: "us-west-2",
			VPC: infrastructurev1beta1.VPCConfig{
				CIDR: "10.0.0.0/16",
			},
			EKS: infrastructurev1beta1.EKSConfig{
				Version:       "1.21",
				PrivateAccess: true,
				PublicAccess:  true,
				NodeGroups: []infrastructurev1beta1.NodeGroupConfig{
					{
						Name:         "ng-1",
						InstanceType: "t3.medium",
						DesiredSize:  2,
						MinSize:      1,
						MaxSize:      3,
					},
				},
			},
		},
	}

	config := generateTerraformConfig(cluster)

	// Test Terraform block
	if config.Terraform.RequiredVersion != "~> 1.0" {
		t.Errorf("Expected Terraform version ~> 1.0, got %s", config.Terraform.RequiredVersion)
	}

	// Test Provider block
	if config.Provider.AWS.Region != "us-west-2" {
		t.Errorf("Expected AWS region us-west-2, got %s", config.Provider.AWS.Region)
	}

	// Test Resource block
	if config.Resource.AWSVPC.Main.CIDRBlock != "10.0.0.0/16" {
		t.Errorf("Expected VPC CIDR 10.0.0.0/16, got %s", config.Resource.AWSVPC.Main.CIDRBlock)
	}

	if config.Resource.AWSEKSCluster.Main.Name != "test-cluster" {
		t.Errorf("Expected EKS cluster name test-cluster, got %s", config.Resource.AWSEKSCluster.Main.Name)
	}

	if config.Resource.AWSEKSCluster.Main.Version != "1.21" {
		t.Errorf("Expected EKS version 1.21, got %s", config.Resource.AWSEKSCluster.Main.Version)
	}

	if len(config.Resource.AWSEKSNodeGroup) != 1 {
		t.Errorf("Expected 1 node group, got %d", len(config.Resource.AWSEKSNodeGroup))
	}

	ng := config.Resource.AWSEKSNodeGroup[0]
	if ng.Name != "ng-1" {
		t.Errorf("Expected node group name ng-1, got %s", ng.Name)
	}

	if ng.ScalingConfig.DesiredSize != 2 {
		t.Errorf("Expected desired size 2, got %d", ng.ScalingConfig.DesiredSize)
	}
}

func TestConvertToJSON(t *testing.T) {
	config := TerraformConfig{
		Terraform: TerraformBlock{
			RequiredVersion: "~> 1.0",
			RequiredProviders: map[string]Provider{
				"aws": {
					Source:  "hashicorp/aws",
					Version: "~> 4.0",
				},
			},
		},
		Provider: ProviderBlock{
			AWS: AWSProvider{
				Region: "us-west-2",
			},
		},
	}

	jsonData, err := convertToJSON(config)
	if err != nil {
		t.Fatalf("Failed to convert to JSON: %v", err)
	}

	var decodedConfig TerraformConfig
	err = json.Unmarshal(jsonData, &decodedConfig)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if !reflect.DeepEqual(config, decodedConfig) {
		t.Errorf("JSON conversion failed, expected %+v, got %+v", config, decodedConfig)
	}
}

func TestConvertJSONToHCL(t *testing.T) {
	jsonData := []byte(`{
		"terraform": {
			"required_version": "~> 1.0",
			"required_providers": {
				"aws": {
					"source": "hashicorp/aws",
					"version": "~> 4.0"
				}
			}
		},
		"provider": {
			"aws": {
				"region": "us-west-2"
			}
		},
		"resource": {
			"aws_vpc": {
				"main": {
					"cidr_block": "10.0.0.0/16",
					"enable_dns_hostnames": true,
					"enable_dns_support": true
				}
			}
		}
	}`)

	hclData, err := convertJSONToHCL(jsonData)
	if err != nil {
		t.Fatalf("Failed to convert JSON to HCL: %v", err)
	}

	// Print the generated HCL for debugging
	fmt.Printf("Generated HCL:\n%s\n", string(hclData))

	parser := hclparse.NewParser()
	_, diags := parser.ParseHCL(hclData, "test.tf")
	if diags.HasErrors() {
		t.Fatalf("Invalid HCL generated: %v", diags)
	}

	// Check for expected HCL content
	expectedStrings := []string{
		"terraform {",
		"required_version = \"~> 1.0\"",
		"required_providers {",
		"aws {",
		"source  = \"hashicorp/aws\"",
		"version = \"~> 4.0\"",
		"provider \"aws\" {",
		"region = \"us-west-2\"",
		"resource \"aws_vpc\" \"main\" {",
		"cidr_block           = \"10.0.0.0/16\"",
		"enable_dns_hostnames = true",
		"enable_dns_support   = true",
	}

	hclString := string(hclData)
	for _, str := range expectedStrings {
		if !strings.Contains(hclString, str) {
			t.Errorf("Expected HCL to contain '%s', but it didn't", str)
		}
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
