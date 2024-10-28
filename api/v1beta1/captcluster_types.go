/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CAPTClusterSpec defines the desired state of CAPTCluster
type CAPTClusterSpec struct {
	// Region is the AWS region where the EKS cluster will be created
	Region string `json:"region"`

	// NetworkRef is a reference to a CAPTVPCTemplate resource
	NetworkRef *corev1.ObjectReference `json:"networkRef,omitempty"`

	// VPC configuration
	VPC VPCConfig `json:"vpc"`

	// EKS configuration
	EKS EKSConfig `json:"eks"`

	// Addons specifies the EKS add-ons to be enabled
	Addons CAPTClusterAddons `json:"addons"`

	// Karpenter specifies the Karpenter configuration
	Karpenter CAPTClusterKarpenter `json:"karpenter"`
}

// VPCConfig defines the VPC configuration
type VPCConfig struct {
	// CIDR is the CIDR block for the VPC
	CIDR string `json:"cidr"`

	// EnableNatGateway determines if NAT Gateway should be created
	EnableNatGateway bool `json:"enableNatGateway"`

	// SingleNatGateway determines if a single NAT Gateway should be used for all AZs
	SingleNatGateway bool `json:"singleNatGateway"`

	// PublicSubnetTags are the tags to apply to public subnets
	PublicSubnetTags map[string]string `json:"publicSubnetTags,omitempty"`

	// PrivateSubnetTags are the tags to apply to private subnets
	PrivateSubnetTags map[string]string `json:"privateSubnetTags,omitempty"`
}

// EKSConfig defines the EKS cluster configuration
type EKSConfig struct {
	// Version is the Kubernetes version for the EKS cluster
	Version string `json:"version"`

	// PublicAccess determines if the EKS cluster has public access enabled
	PublicAccess bool `json:"publicAccess"`

	// PrivateAccess determines if the EKS cluster has private access enabled
	PrivateAccess bool `json:"privateAccess"`

	// NodeGroups defines the node groups for the EKS cluster
	NodeGroups []NodeGroupConfig `json:"nodeGroups,omitempty"`
}

// NodeGroupConfig defines the configuration for an EKS node group
type NodeGroupConfig struct {
	// Name of the node group
	Name string `json:"name"`

	// InstanceType is the EC2 instance type to use for the node group
	InstanceType string `json:"instanceType"`

	// DesiredSize is the desired number of nodes in the node group
	DesiredSize int `json:"desiredSize"`

	// MinSize is the minimum number of nodes in the node group
	MinSize int `json:"minSize"`

	// MaxSize is the maximum number of nodes in the node group
	MaxSize int `json:"maxSize"`
}

// CAPTClusterAddons defines the EKS add-ons configuration
type CAPTClusterAddons struct {
	// CoreDNS specifies whether the CoreDNS add-on is enabled
	CoreDNS AddonConfig `json:"coredns"`

	// VpcCni specifies whether the VPC CNI add-on is enabled
	VpcCni AddonConfig `json:"vpcCni"`

	// KubeProxy specifies whether the Kube-Proxy add-on is enabled
	KubeProxy AddonConfig `json:"kubeProxy"`
}

// AddonConfig defines the configuration for an EKS add-on
type AddonConfig struct {
	// Enabled specifies whether the add-on is enabled
	Enabled bool `json:"enabled"`
}

// CAPTClusterKarpenter defines the Karpenter configuration
type CAPTClusterKarpenter struct {
	// Enabled specifies whether Karpenter is enabled
	Enabled bool `json:"enabled"`
}

// CAPTClusterStatus defines the observed state of CAPTCluster
type CAPTClusterStatus struct {
	// WorkspaceName is the name of the associated Terraform Workspace
	WorkspaceName string `json:"workspaceName,omitempty"`

	// ClusterName is the name of the created EKS cluster
	ClusterName string `json:"clusterName,omitempty"`

	// ClusterEndpoint is the endpoint of the created EKS cluster
	ClusterEndpoint string `json:"clusterEndpoint,omitempty"`

	// ClusterStatus represents the current status of the EKS cluster
	ClusterStatus string `json:"clusterStatus,omitempty"`

	// NetworkWorkspaceName is the name of the associated Network Terraform Workspace
	NetworkWorkspaceName string `json:"networkWorkspaceName,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// CAPTCluster is the Schema for the captclusters API
type CAPTCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CAPTClusterSpec   `json:"spec,omitempty"`
	Status CAPTClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CAPTClusterList contains a list of CAPTCluster
type CAPTClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CAPTCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CAPTCluster{}, &CAPTClusterList{})
}
