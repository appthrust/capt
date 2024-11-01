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
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CAPTClusterSpec defines the desired state of CAPTCluster
type CAPTClusterSpec struct {
	// Region is the AWS region where the cluster will be created
	Region string `json:"region"`

	// VPCTemplateRef is a reference to a WorkspaceTemplate resource for VPC configuration
	// If specified, a new VPC will be created using this template
	// +optional
	VPCTemplateRef *WorkspaceTemplateReference `json:"vpcTemplateRef,omitempty"`

	// ExistingVPCID is the ID of an existing VPC to use
	// If specified, VPCTemplateRef must not be set
	// +optional
	ExistingVPCID string `json:"existingVpcId,omitempty"`
}

// CAPTClusterStatus defines the observed state of CAPTCluster
type CAPTClusterStatus struct {
	// VPCWorkspaceName is the name of the associated VPC Terraform Workspace
	// +optional
	VPCWorkspaceName string `json:"vpcWorkspaceName,omitempty"`

	// VPCID is the ID of the VPC being used
	// This could be either a newly created VPC or an existing one
	VPCID string `json:"vpcId,omitempty"`

	// Ready denotes that the cluster infrastructure is ready
	// +optional
	Ready bool `json:"ready,omitempty"`

	// Conditions defines current service state of the CAPTCluster
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

const (
	// VPCReadyCondition indicates that the VPC is ready
	VPCReadyCondition = "VPCReady"

	// VPCCreatingCondition indicates that the VPC is being created
	VPCCreatingCondition = "VPCCreating"

	// VPCFailedCondition indicates that VPC creation failed
	VPCFailedCondition = "VPCFailed"

	// ReasonExistingVPCUsed represents that an existing VPC is being used
	ReasonExistingVPCUsed = "ExistingVPCUsed"

	// ReasonVPCCreated represents that a new VPC has been created
	ReasonVPCCreated = "VPCCreated"

	// ReasonVPCCreating represents that a VPC is being created
	ReasonVPCCreating = "VPCCreating"

	// ReasonVPCCreationFailed represents that VPC creation failed
	ReasonVPCCreationFailed = "VPCCreationFailed"
)

// ValidateVPCConfiguration validates that the VPC configuration is valid
func (s *CAPTClusterSpec) ValidateVPCConfiguration() error {
	if s.VPCTemplateRef != nil && s.ExistingVPCID != "" {
		return fmt.Errorf("cannot specify both VPCTemplateRef and ExistingVPCID")
	}
	if s.VPCTemplateRef == nil && s.ExistingVPCID == "" {
		return fmt.Errorf("must specify either VPCTemplateRef or ExistingVPCID")
	}
	return nil
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="VPC-ID",type="string",JSONPath=".status.vpcId"
// +kubebuilder:printcolumn:name="READY",type="boolean",JSONPath=".status.ready"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"

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
