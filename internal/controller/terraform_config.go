package controller

import (
	"encoding/json"
	"fmt"

	infrastructurev1beta1 "github.com/appthrust/capt/api/v1beta1"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

type TerraformConfig struct {
	Terraform TerraformBlock `json:"terraform"`
	Provider  ProviderBlock  `json:"provider"`
	Resource  ResourceBlock  `json:"resource"`
}

type TerraformBlock struct {
	RequiredVersion   string              `json:"required_version"`
	RequiredProviders map[string]Provider `json:"required_providers"`
}

type Provider struct {
	Source  string `json:"source"`
	Version string `json:"version"`
}

type ProviderBlock struct {
	AWS AWSProvider `json:"aws"`
}

type AWSProvider struct {
	Region string `json:"region"`
}

type ResourceBlock struct {
	AWSVPC          AWSVPC            `json:"aws_vpc"`
	AWSEKSCluster   AWSEKSCluster     `json:"aws_eks_cluster"`
	AWSIAMRole      []AWSIAMRole      `json:"aws_iam_role"`
	AWSEKSNodeGroup []AWSEKSNodeGroup `json:"aws_eks_node_group"`
}

type AWSVPC struct {
	Main VPCConfig `json:"main"`
}

type VPCConfig struct {
	CIDRBlock          string `json:"cidr_block"`
	EnableDNSHostnames bool   `json:"enable_dns_hostnames"`
	EnableDNSSupport   bool   `json:"enable_dns_support"`
}

type AWSEKSCluster struct {
	Main EKSClusterConfig `json:"main"`
}

type EKSClusterConfig struct {
	Name      string    `json:"name"`
	Version   string    `json:"version"`
	RoleARN   string    `json:"role_arn"`
	VPCConfig VPCConfig `json:"vpc_config"`
}

type AWSIAMRole struct {
	Name             string `json:"name"`
	AssumeRolePolicy string `json:"assume_role_policy"`
}

type AWSEKSNodeGroup struct {
	Name          string        `json:"name"`
	ClusterName   string        `json:"cluster_name"`
	NodeGroupName string        `json:"node_group_name"`
	NodeRoleARN   string        `json:"node_role_arn"`
	ScalingConfig ScalingConfig `json:"scaling_config"`
	InstanceTypes []string      `json:"instance_types"`
}

type ScalingConfig struct {
	DesiredSize int `json:"desired_size"`
	MaxSize     int `json:"max_size"`
	MinSize     int `json:"min_size"`
}

func generateTerraformConfig(cluster *infrastructurev1beta1.CAPTCluster) TerraformConfig {
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
				Region: cluster.Spec.Region,
			},
		},
		Resource: ResourceBlock{
			AWSVPC: AWSVPC{
				Main: VPCConfig{
					CIDRBlock:          cluster.Spec.VPC.CIDR,
					EnableDNSHostnames: true,
					EnableDNSSupport:   true,
				},
			},
			AWSEKSCluster: AWSEKSCluster{
				Main: EKSClusterConfig{
					Name:    cluster.Name,
					Version: cluster.Spec.EKS.Version,
					RoleARN: "${aws_iam_role.eks_cluster_role.arn}",
					VPCConfig: VPCConfig{
						EnableDNSHostnames: cluster.Spec.EKS.PrivateAccess,
						EnableDNSSupport:   cluster.Spec.EKS.PublicAccess,
					},
				},
			},
			AWSIAMRole: []AWSIAMRole{
				{
					Name: fmt.Sprintf("%s-eks-cluster-role", cluster.Name),
					AssumeRolePolicy: `{
						"Version": "2012-10-17",
						"Statement": [
							{
								"Effect": "Allow",
								"Principal": {
									"Service": "eks.amazonaws.com"
								},
								"Action": "sts:AssumeRole"
							}
						]
					}`,
				},
				{
					Name: fmt.Sprintf("%s-eks-node-role", cluster.Name),
					AssumeRolePolicy: `{
						"Version": "2012-10-17",
						"Statement": [
							{
								"Effect": "Allow",
								"Principal": {
									"Service": "ec2.amazonaws.com"
								},
								"Action": "sts:AssumeRole"
							}
						]
					}`,
				},
			},
		},
	}

	for _, ng := range cluster.Spec.EKS.NodeGroups {
		config.Resource.AWSEKSNodeGroup = append(config.Resource.AWSEKSNodeGroup, AWSEKSNodeGroup{
			Name:          ng.Name,
			ClusterName:   cluster.Name,
			NodeGroupName: ng.Name,
			NodeRoleARN:   "${aws_iam_role.eks_node_role.arn}",
			ScalingConfig: ScalingConfig{
				DesiredSize: int(ng.DesiredSize),
				MaxSize:     int(ng.MaxSize),
				MinSize:     int(ng.MinSize),
			},
			InstanceTypes: []string{ng.InstanceType},
		})
	}

	return config
}

func convertToJSON(config TerraformConfig) ([]byte, error) {
	return json.Marshal(config)
}

func convertJSONToHCL(jsonData []byte) ([]byte, error) {
	var tfConfig map[string]interface{}
	err := json.Unmarshal(jsonData, &tfConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()

	for key, value := range tfConfig {
		switch key {
		case "terraform":
			writeTerraformBlock(rootBody, value.(map[string]interface{}))
		case "provider":
			writeProviderBlock(rootBody, value.(map[string]interface{}))
		case "resource":
			writeResourceBlocks(rootBody, value.(map[string]interface{}))
		}
	}

	return f.Bytes(), nil
}

func writeTerraformBlock(body *hclwrite.Body, terraform map[string]interface{}) {
	tfBlock := body.AppendNewBlock("terraform", nil)
	tfBody := tfBlock.Body()

	if version, ok := terraform["required_version"].(string); ok {
		tfBody.SetAttributeValue("required_version", cty.StringVal(version))
	}

	if providers, ok := terraform["required_providers"].(map[string]interface{}); ok {
		providersBlock := tfBody.AppendNewBlock("required_providers", nil)
		providersBody := providersBlock.Body()

		for name, config := range providers {
			providerConfig := config.(map[string]interface{})
			providerBlock := providersBody.AppendNewBlock(name, nil)
			providerBody := providerBlock.Body()

			if source, ok := providerConfig["source"].(string); ok {
				providerBody.SetAttributeValue("source", cty.StringVal(source))
			}
			if version, ok := providerConfig["version"].(string); ok {
				providerBody.SetAttributeValue("version", cty.StringVal(version))
			}
		}
	}
}

func writeProviderBlock(body *hclwrite.Body, provider map[string]interface{}) {
	for name, config := range provider {
		providerBlock := body.AppendNewBlock("provider", []string{name})
		providerBody := providerBlock.Body()

		configMap := config.(map[string]interface{})
		for key, value := range configMap {
			providerBody.SetAttributeValue(key, convertToCtyValue(value))
		}
	}
}

func writeResourceBlocks(body *hclwrite.Body, resources map[string]interface{}) {
	for resourceType, resourceConfigs := range resources {
		configs := resourceConfigs.(map[string]interface{})
		for resourceName, config := range configs {
			resourceBlock := body.AppendNewBlock("resource", []string{resourceType, resourceName})
			resourceBody := resourceBlock.Body()

			configMap := config.(map[string]interface{})
			for key, value := range configMap {
				writeAttributeOrBlock(resourceBody, key, value)
			}
		}
	}
}

func writeAttributeOrBlock(body *hclwrite.Body, key string, value interface{}) {
	switch v := value.(type) {
	case map[string]interface{}:
		block := body.AppendNewBlock(key, nil)
		blockBody := block.Body()
		for subKey, subValue := range v {
			writeAttributeOrBlock(blockBody, subKey, subValue)
		}
	case []interface{}:
		for _, item := range v {
			if itemMap, ok := item.(map[string]interface{}); ok {
				block := body.AppendNewBlock(key, nil)
				blockBody := block.Body()
				for subKey, subValue := range itemMap {
					writeAttributeOrBlock(blockBody, subKey, subValue)
				}
			} else {
				body.SetAttributeValue(key, convertToCtyValue(v))
				break
			}
		}
	default:
		body.SetAttributeValue(key, convertToCtyValue(v))
	}
}

func convertToCtyValue(v interface{}) cty.Value {
	switch value := v.(type) {
	case string:
		return cty.StringVal(value)
	case bool:
		return cty.BoolVal(value)
	case float64:
		return cty.NumberFloatVal(value)
	case int:
		return cty.NumberIntVal(int64(value))
	case []interface{}:
		values := make([]cty.Value, len(value))
		for i, v := range value {
			values[i] = convertToCtyValue(v)
		}
		return cty.ListVal(values)
	case map[string]interface{}:
		return cty.ObjectVal(convertMapToCtyValueMap(value))
	default:
		return cty.NullVal(cty.DynamicPseudoType)
	}
}

func convertMapToCtyValueMap(m map[string]interface{}) map[string]cty.Value {
	result := make(map[string]cty.Value)
	for k, v := range m {
		result[k] = convertToCtyValue(v)
	}
	return result
}
