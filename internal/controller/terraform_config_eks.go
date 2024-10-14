package controller

import (
	infrastructurev1beta1 "github.com/appthrust/capt/api/v1beta1"
)

type EKSTerraformConfig struct {
	Data     EKSDataBlock     `json:"data"`
	Locals   EKSLocalsBlock   `json:"locals"`
	Module   EKSModuleBlock   `json:"module"`
	Resource EKSResourceBlock `json:"resource"`
	Output   EKSOutputBlock   `json:"output"`
}

type EKSDataBlock struct {
	AWSAvailabilityZones AWSAvailabilityZones `json:"aws_availability_zones"`
}

type AWSAvailabilityZones struct {
	Available AvailabilityZonesFilter `json:"available"`
}

type AvailabilityZonesFilter struct {
	Filter []Filter `json:"filter"`
}

type Filter struct {
	Name   string   `json:"name"`
	Values []string `json:"values"`
}

type EKSLocalsBlock struct {
	AZs  string            `json:"azs"`
	Name string            `json:"name"`
	Tags map[string]string `json:"tags"`
}

type EKSModuleBlock struct {
	EKS                 EKSModule                 `json:"eks"`
	EKSBlueprintsAddons EKSBlueprintsAddonsModule `json:"eks_blueprints_addons"`
	VPC                 VPCModule                 `json:"vpc"`
}

type EKSModule struct {
	Source                               string                    `json:"source"`
	Version                              string                    `json:"version"`
	ClusterName                          string                    `json:"cluster_name"`
	ClusterVersion                       string                    `json:"cluster_version"`
	ClusterEndpointPublicAccess          bool                      `json:"cluster_endpoint_public_access"`
	VPCID                                string                    `json:"vpc_id"`
	SubnetIDs                            []string                  `json:"subnet_ids"`
	CreateClusterSecurityGroup           bool                      `json:"create_cluster_security_group"`
	CreateNodeSecurityGroup              bool                      `json:"create_node_security_group"`
	EnableClusterCreatorAdminPermissions bool                      `json:"enable_cluster_creator_admin_permissions"`
	FargateProfiles                      map[string]FargateProfile `json:"fargate_profiles"`
	Tags                                 map[string]string         `json:"tags"`
}

type FargateProfile struct {
	Name      string     `json:"name,omitempty"`
	Selectors []Selector `json:"selectors"`
}

type Selector struct {
	Namespace string `json:"namespace"`
}

type EKSBlueprintsAddonsModule struct {
	Source                  string              `json:"source"`
	Version                 string              `json:"version"`
	ClusterName             string              `json:"cluster_name"`
	ClusterEndpoint         string              `json:"cluster_endpoint"`
	ClusterVersion          string              `json:"cluster_version"`
	OIDCProviderARN         string              `json:"oidc_provider_arn"`
	CreateDelayDependencies []string            `json:"create_delay_dependencies"`
	EKSAddons               map[string]EKSAddon `json:"eks_addons"`
	EnableKarpenter         bool                `json:"enable_karpenter"`
	Karpenter               KarpenterConfig     `json:"karpenter"`
	KarpenterNode           KarpenterNodeConfig `json:"karpenter_node"`
	Tags                    map[string]string   `json:"tags"`
}

type EKSAddon struct {
	ConfigurationValues string `json:"configuration_values,omitempty"`
}

type KarpenterConfig struct {
	HelmConfig KarpenterHelmConfig `json:"helm_config"`
}

type KarpenterHelmConfig struct {
	CacheDir string `json:"cacheDir"`
}

type KarpenterNodeConfig struct {
	IAMRoleUseNamePrefix bool `json:"iam_role_use_name_prefix"`
}

type VPCModule struct {
	Source            string            `json:"source"`
	Version           string            `json:"version"`
	Name              string            `json:"name"`
	CIDR              string            `json:"cidr"`
	AZs               []string          `json:"azs"`
	PrivateSubnets    []string          `json:"private_subnets"`
	PublicSubnets     []string          `json:"public_subnets"`
	EnableNATGateway  bool              `json:"enable_nat_gateway"`
	SingleNATGateway  bool              `json:"single_nat_gateway"`
	PublicSubnetTags  map[string]string `json:"public_subnet_tags"`
	PrivateSubnetTags map[string]string `json:"private_subnet_tags"`
	Tags              map[string]string `json:"tags"`
}

type EKSResourceBlock struct {
	AWSEKSAccessEntry AWSEKSAccessEntry `json:"aws_eks_access_entry"`
}

type AWSEKSAccessEntry struct {
	KarpenterNodeAccessEntry KarpenterNodeAccessEntry `json:"karpenter_node_access_entry"`
}

type KarpenterNodeAccessEntry struct {
	ClusterName      string    `json:"cluster_name"`
	PrincipalARN     string    `json:"principal_arn"`
	KubernetesGroups []string  `json:"kubernetes_groups"`
	Type             string    `json:"type"`
	Lifecycle        Lifecycle `json:"lifecycle"`
}

type Lifecycle struct {
	IgnoreChanges []string `json:"ignore_changes"`
}

type EKSOutputBlock struct {
	ClusterEndpoint        EKSOutput `json:"cluster_endpoint"`
	ClusterSecurityGroupID EKSOutput `json:"cluster_security_group_id"`
	ClusterName            EKSOutput `json:"cluster_name"`
}

type EKSOutput struct {
	Description string `json:"description"`
	Value       string `json:"value"`
}

func generateEKSTerraformConfig(cluster *infrastructurev1beta1.CAPTCluster) EKSTerraformConfig {
	config := EKSTerraformConfig{
		Data: EKSDataBlock{
			AWSAvailabilityZones: AWSAvailabilityZones{
				Available: AvailabilityZonesFilter{
					Filter: []Filter{
						{
							Name:   "opt-in-status",
							Values: []string{"opt-in-not-required"},
						},
					},
				},
			},
		},
		Locals: EKSLocalsBlock{
			AZs:  "slice(data.aws_availability_zones.available.names, 0, 3)",
			Name: cluster.Name,
			Tags: map[string]string{
				"Module":     "capt-generated",
				"GithubRepo": "github.com/appthrust/capt",
			},
		},
		Module: EKSModuleBlock{
			EKS: EKSModule{
				Source:                               "terraform-aws-modules/eks/aws",
				Version:                              "~> 20.11",
				ClusterName:                          "local.name",
				ClusterVersion:                       cluster.Spec.EKS.Version,
				ClusterEndpointPublicAccess:          cluster.Spec.EKS.PublicAccess,
				VPCID:                                "module.vpc.vpc_id",
				SubnetIDs:                            []string{"module.vpc.private_subnets"},
				CreateClusterSecurityGroup:           false,
				CreateNodeSecurityGroup:              false,
				EnableClusterCreatorAdminPermissions: true,
				FargateProfiles: map[string]FargateProfile{
					"karpenter": {
						Selectors: []Selector{{Namespace: "karpenter"}},
					},
					"kube_system": {
						Name:      "kube-system",
						Selectors: []Selector{{Namespace: "kube-system"}},
					},
				},
				Tags: map[string]string{
					"karpenter.sh/discovery": "local.name",
				},
			},
			EKSBlueprintsAddons: EKSBlueprintsAddonsModule{
				Source:                  "aws-ia/eks-blueprints-addons/aws",
				Version:                 "~> 1.16",
				ClusterName:             "module.eks.cluster_name",
				ClusterEndpoint:         "module.eks.cluster_endpoint",
				ClusterVersion:          "module.eks.cluster_version",
				OIDCProviderARN:         "module.eks.oidc_provider_arn",
				CreateDelayDependencies: []string{"for prof in module.eks.fargate_profiles : prof.fargate_profile_arn"},
				EKSAddons: map[string]EKSAddon{
					"coredns": {
						ConfigurationValues: `jsonencode({
							computeType = "Fargate"
							resources = {
								limits = {
									cpu = "0.25"
									memory = "256M"
								}
								requests = {
									cpu = "0.25"
									memory = "256M"
								}
							}
						})`,
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
			},
			VPC: VPCModule{
				Source:           "terraform-aws-modules/vpc/aws",
				Version:          "~> 5.0",
				Name:             "local.name",
				CIDR:             cluster.Spec.VPC.CIDR,
				AZs:              []string{"local.azs"},
				PrivateSubnets:   []string{"for k, v in local.azs : cidrsubnet(var.vpc_cidr, 4, k)"},
				PublicSubnets:    []string{"for k, v in local.azs : cidrsubnet(var.vpc_cidr, 8, k + 48)"},
				EnableNATGateway: cluster.Spec.VPC.EnableNatGateway,
				SingleNATGateway: cluster.Spec.VPC.SingleNatGateway,
				PublicSubnetTags: map[string]string{
					"kubernetes.io/role/elb": "1",
				},
				PrivateSubnetTags: map[string]string{
					"kubernetes.io/role/internal-elb": "1",
					"karpenter.sh/discovery":          "local.name",
				},
				Tags: map[string]string{},
			},
		},
		Resource: EKSResourceBlock{
			AWSEKSAccessEntry: AWSEKSAccessEntry{
				KarpenterNodeAccessEntry: KarpenterNodeAccessEntry{
					ClusterName:      "module.eks.cluster_name",
					PrincipalARN:     "module.eks_blueprints_addons.karpenter.node_iam_role_arn",
					KubernetesGroups: []string{},
					Type:             "EC2_LINUX",
					Lifecycle: Lifecycle{
						IgnoreChanges: []string{"kubernetes_groups"},
					},
				},
			},
		},
		Output: EKSOutputBlock{
			ClusterEndpoint: EKSOutput{
				Description: "Endpoint for EKS control plane",
				Value:       "module.eks.cluster_endpoint",
			},
			ClusterSecurityGroupID: EKSOutput{
				Description: "Security group ids attached to the cluster control plane",
				Value:       "module.eks.cluster_security_group_id",
			},
			ClusterName: EKSOutput{
				Description: "Kubernetes Cluster Name",
				Value:       "module.eks.cluster_name",
			},
		},
	}

	return config
}
