package v1beta2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
	Region                     string                      `json:"region"`
	VPCTemplateRef             *WorkspaceTemplateReference `json:"vpcTemplateRef,omitempty"`
	ExistingVPCID              string                      `json:"existingVpcId,omitempty"`
	RetainVPCOnDelete          bool                        `json:"retainVpcOnDelete,omitempty"`
	VPCConfig                  *VPCConfig                  `json:"vpcConfig,omitempty"`
	WorkspaceTemplateApplyName string                      `json:"workspaceTemplateApplyName,omitempty"`
}

// CAPTClusterWorkspaceStatus contains the status of the WorkspaceTemplate
type CAPTClusterWorkspaceStatus struct {
	Ready               bool         `json:"ready"`
	LastAppliedRevision string       `json:"lastAppliedRevision,omitempty"`
	LastFailedRevision  string       `json:"lastFailedRevision,omitempty"`
	LastFailureMessage  string       `json:"lastFailureMessage,omitempty"`
	WorkspaceName       string       `json:"workspaceName,omitempty"`
	LastAppliedTime     *metav1.Time `json:"lastAppliedTime,omitempty"`
}

// WorkspaceStatus contains the status of the associated Workspace
type WorkspaceStatus struct {
	Ready bool   `json:"ready"`
	State string `json:"state,omitempty"`
	// AtProvider contains provider-specific observed state
	// +optional
	AtProvider *runtime.RawExtension `json:"atProvider,omitempty"`
}

// CAPTClusterStatus defines the observed state of CAPTCluster
type CAPTClusterStatus struct {
	VPCWorkspaceName        string                      `json:"vpcWorkspaceName,omitempty"`
	VPCID                   string                      `json:"vpcId,omitempty"`
	Ready                   bool                        `json:"ready,omitempty"`
	FailureReason           *string                     `json:"failureReason,omitempty"`
	FailureMessage          *string                     `json:"failureMessage,omitempty"`
	FailureDomains          clusterv1.FailureDomains    `json:"failureDomains,omitempty"`
	Conditions              []metav1.Condition          `json:"conditions,omitempty"`
	WorkspaceTemplateStatus *CAPTClusterWorkspaceStatus `json:"workspaceTemplateStatus,omitempty"`
	WorkspaceStatus         *WorkspaceStatus            `json:"workspaceStatus,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="VPC-ID",type="string",JSONPath=".status.vpcId"
// +kubebuilder:printcolumn:name="READY",type="boolean",JSONPath=".status.ready"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:path=captclusters,scope=Namespaced,categories=cluster-api
// +kubebuilder:storageversion

// CAPTCluster is the Schema for the captclusters API
type CAPTCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CAPTClusterSpec   `json:"spec,omitempty"`
	Status CAPTClusterStatus `json:"status,omitempty"`
}

// Hub marker
func (*CAPTCluster) Hub() {}

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
