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
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

// VPCConfig contains configuration for the VPC
type VPCConfig struct {
	// Name is the name of the VPC
	// If not specified, defaults to {cluster-name}-vpc
	// +optional
	Name string `json:"name,omitempty"`
}

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

	// RetainVPCOnDelete specifies whether to retain the VPC when the parent cluster is deleted
	// This is useful when the VPC is shared among multiple projects
	// This field is only effective when VPCTemplateRef is set
	// +optional
	RetainVPCOnDelete bool `json:"retainVpcOnDelete,omitempty"`

	// VPCConfig contains VPC-specific configuration
	// +optional
	VPCConfig *VPCConfig `json:"vpcConfig,omitempty"`

	// WorkspaceTemplateApplyName is the name of the WorkspaceTemplateApply used for this cluster.
	// This field is managed by the controller and should not be modified manually.
	// +optional
	WorkspaceTemplateApplyName string `json:"workspaceTemplateApplyName,omitempty"`
}

// CAPTClusterWorkspaceStatus contains the status of the WorkspaceTemplate
type CAPTClusterWorkspaceStatus struct {
	// Ready indicates if the WorkspaceTemplate is ready
	// +optional
	Ready bool `json:"ready"`

	// LastAppliedRevision is the revision of the WorkspaceTemplate that was last applied
	// +optional
	LastAppliedRevision string `json:"lastAppliedRevision,omitempty"`

	// LastFailedRevision is the revision of the WorkspaceTemplate that last failed
	// +optional
	LastFailedRevision string `json:"lastFailedRevision,omitempty"`

	// LastFailureMessage contains the error message from the last failure
	// +optional
	LastFailureMessage string `json:"lastFailureMessage,omitempty"`

	// WorkspaceName is the name of the associated workspace
	// +optional
	WorkspaceName string `json:"workspaceName,omitempty"`

	// LastAppliedTime is the last time the template was applied
	// +optional
	LastAppliedTime *metav1.Time `json:"lastAppliedTime,omitempty"`
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

	// FailureReason indicates that there is a terminal problem reconciling the
	// state, and will be set to a token value suitable for programmatic
	// interpretation.
	// +optional
	FailureReason *string `json:"failureReason,omitempty"`

	// FailureMessage indicates that there is a terminal problem reconciling the
	// state, and will be set to a descriptive error message.
	// +optional
	FailureMessage *string `json:"failureMessage,omitempty"`

	// FailureDomains is a list of failure domain objects synced from the infrastructure provider.
	// +optional
	FailureDomains clusterv1.FailureDomains `json:"failureDomains,omitempty"`

	// Conditions defines current service state of the CAPTCluster
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// WorkspaceTemplateStatus contains the status of the WorkspaceTemplate
	// +optional
	WorkspaceTemplateStatus *CAPTClusterWorkspaceStatus `json:"workspaceTemplateStatus,omitempty"`
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
	if s.RetainVPCOnDelete && s.VPCTemplateRef == nil {
		return fmt.Errorf("retainVpcOnDelete can only be set when VPCTemplateRef is specified")
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
