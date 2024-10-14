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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CAPTClusterSpec defines the desired state of CAPTCluster
type CAPTClusterSpec struct {
	// Region is the AWS region where the EKS cluster will be created
	Region string `json:"region"`

	// VpcCIDR is the CIDR block for the VPC
	VpcCIDR string `json:"vpcCIDR"`

	// PublicAccess determines if the EKS cluster has public access enabled
	PublicAccess bool `json:"publicAccess"`

	// Version is the Kubernetes version for the EKS cluster
	Version string `json:"version"`

	// Addons specifies the EKS add-ons to be enabled
	Addons CAPTClusterAddons `json:"addons"`

	// Karpenter specifies the Karpenter configuration
	Karpenter CAPTClusterKarpenter `json:"karpenter"`

	// PublicSubnetTags are the tags to apply to public subnets
	PublicSubnetTags map[string]string `json:"publicSubnetTags,omitempty"`

	// PrivateSubnetTags are the tags to apply to private subnets
	PrivateSubnetTags map[string]string `json:"privateSubnetTags,omitempty"`
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
