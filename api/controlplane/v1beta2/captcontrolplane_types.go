package v1beta2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

// Condition/Reason names are shared with v1beta1 for compatibility
const (
	ControlPlaneReadyCondition       = "Ready"
	ControlPlaneInitializedCondition = "Initialized"
	ControlPlaneFailedCondition      = "Failed"
	ControlPlaneCreatingCondition    = "Creating"

	ReasonCreating            = "Creating"
	ReasonReady               = "Ready"
	ReasonFailed              = "Failed"
	ReasonWaitingForVPC       = "WaitingForVPC"
	ReasonVPCReadyTimeout     = "VPCReadyTimeout"
	ReasonControlPlaneTimeout = "ControlPlaneTimeout"
	ReasonWorkspaceError      = "WorkspaceError"
)

type CAPTControlPlaneSpec struct {
	Version                    string                     `json:"version"`
	WorkspaceTemplateRef       WorkspaceTemplateReference `json:"workspaceTemplateRef"`
	ControlPlaneConfig         *ControlPlaneConfig        `json:"controlPlaneConfig,omitempty"`
	AdditionalTags             map[string]string          `json:"additionalTags,omitempty"`
	ControlPlaneEndpoint       clusterv1.APIEndpoint      `json:"controlPlaneEndpoint,omitempty"`
	WorkspaceTemplateApplyName string                     `json:"workspaceTemplateApplyName,omitempty"`
}

type WorkspaceTemplateReference struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
}

type TimeoutConfig struct {
	ControlPlaneTimeout *int `json:"controlPlaneTimeout,omitempty"`
	VPCReadyTimeout     *int `json:"vpcReadyTimeout,omitempty"`
}

type ControlPlaneConfig struct {
	Region         string          `json:"region"`
	EndpointAccess *EndpointAccess `json:"endpointAccess,omitempty"`
	Addons         []Addon         `json:"addons,omitempty"`
	Timeouts       *TimeoutConfig  `json:"timeouts,omitempty"`
}

type EndpointAccess struct {
	Public      bool     `json:"public,omitempty"`
	Private     bool     `json:"private,omitempty"`
	PublicCIDRs []string `json:"publicCIDRs,omitempty"`
}

type Addon struct {
	Name                string `json:"name"`
	Version             string `json:"version,omitempty"`
	ConfigurationValues string `json:"configurationValues,omitempty"`
}

type WorkspaceTemplateStatus struct {
	Ready               bool              `json:"ready"`
	State               string            `json:"state,omitempty"`
	LastAppliedRevision string            `json:"lastAppliedRevision,omitempty"`
	Outputs             map[string]string `json:"outputs,omitempty"`
	LastFailedRevision  string            `json:"lastFailedRevision,omitempty"`
	LastFailureMessage  string            `json:"lastFailureMessage,omitempty"`
	// WorkspaceName is the name of the associated Terraform Workspace
	// +optional
	WorkspaceName string `json:"workspaceName,omitempty"`
}

type WorkspaceStatus struct {
	Ready bool   `json:"ready"`
	State string `json:"state,omitempty"`
	// AtProvider contains provider-specific observed state
	// +optional
	AtProvider *runtime.RawExtension `json:"atProvider,omitempty"`
}

type CAPTControlPlaneStatus struct {
	Ready                   bool                     `json:"ready"`
	Initialized             bool                     `json:"initialized"`
	SecretsReady            bool                     `json:"secretsReady"`
	WorkspaceTemplateStatus *WorkspaceTemplateStatus `json:"workspaceTemplateStatus,omitempty"`
	WorkspaceStatus         *WorkspaceStatus         `json:"workspaceStatus,omitempty"`
	FailureReason           *string                  `json:"failureReason,omitempty"`
	FailureMessage          *string                  `json:"failureMessage,omitempty"`
	Phase                   string                   `json:"phase,omitempty"`
	Conditions              []metav1.Condition       `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.ready",description="Control Plane Ready status"
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase",description="Control Plane Phase"
// +kubebuilder:printcolumn:name="Version",type="string",JSONPath=".spec.version",description="Kubernetes version"
// +kubebuilder:printcolumn:name="Endpoint",type="string",JSONPath=".spec.controlPlaneEndpoint.host",description="API Server Endpoint"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:path=captcontrolplanes,scope=Namespaced,categories=cluster-api
// +kubebuilder:storageversion

type CAPTControlPlane struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CAPTControlPlaneSpec   `json:"spec,omitempty"`
	Status CAPTControlPlaneStatus `json:"status,omitempty"`
}

// Hub marker
func (*CAPTControlPlane) Hub() {}

// +kubebuilder:object:root=true

type CAPTControlPlaneList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CAPTControlPlane `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CAPTControlPlane{}, &CAPTControlPlaneList{})
}
