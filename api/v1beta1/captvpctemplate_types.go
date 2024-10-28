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

// CAPTVPCTemplateSpec defines the desired state of CAPTVPCTemplate
type CAPTVPCTemplateSpec struct {
	// CIDR is the IP range to use for the VPC
	CIDR string `json:"cidr"`

	// EnableNatGateway determines if NAT Gateway should be created
	EnableNatGateway bool `json:"enableNatGateway"`

	// SingleNatGateway determines if a single NAT Gateway should be used for all AZs
	SingleNatGateway bool `json:"singleNatGateway"`

	// PublicSubnetTags are the tags to apply to public subnets
	PublicSubnetTags map[string]string `json:"publicSubnetTags,omitempty"`

	// PrivateSubnetTags are the tags to apply to private subnets
	PrivateSubnetTags map[string]string `json:"privateSubnetTags,omitempty"`

	// Tags are the tags to apply to the VPC and all its resources
	Tags map[string]string `json:"tags,omitempty"`
}

// CAPTVPCTemplateStatus defines the observed state of CAPTVPCTemplate
type CAPTVPCTemplateStatus struct {
	// WorkspaceName is the name of the associated Terraform Workspace
	WorkspaceName string `json:"workspaceName,omitempty"`

	// VPCID is the ID of the created VPC
	VPCID string `json:"vpcId,omitempty"`

	// VPCStatus represents the current status of the VPC
	VPCStatus string `json:"vpcStatus,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// CAPTVPCTemplate is the Schema for the captvpctemplates API
type CAPTVPCTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CAPTVPCTemplateSpec   `json:"spec,omitempty"`
	Status CAPTVPCTemplateStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CAPTVPCTemplateList contains a list of CAPTVPCTemplate
type CAPTVPCTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CAPTVPCTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CAPTVPCTemplate{}, &CAPTVPCTemplateList{})
}
