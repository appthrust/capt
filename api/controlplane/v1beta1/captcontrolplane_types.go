package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

// Condition Types
const (
	// ControlPlaneReadyCondition indicates the control plane is ready
	ControlPlaneReadyCondition = "Ready"

	// ControlPlaneInitializedCondition indicates the control plane has been initialized
	ControlPlaneInitializedCondition = "Initialized"

	// ControlPlaneFailedCondition indicates the control plane has failed
	ControlPlaneFailedCondition = "Failed"

	// ControlPlaneCreatingCondition indicates the control plane is being created
	ControlPlaneCreatingCondition = "Creating"

	// FargateProfileReadyCondition indicates a Fargate profile is ready
	FargateProfileReadyCondition = "FargateProfileReady"

	// FargateProfileCreatingCondition indicates a Fargate profile is being created
	FargateProfileCreatingCondition = "FargateProfileCreating"

	// FargateProfileFailedCondition indicates a Fargate profile creation has failed
	FargateProfileFailedCondition = "FargateProfileFailed"
)

// Condition Reasons
const (
	// ReasonCreating indicates the control plane is being created
	ReasonCreating = "Creating"

	// ReasonReady indicates the control plane is ready
	ReasonReady = "Ready"

	// ReasonFailed indicates the control plane has failed
	ReasonFailed = "Failed"

	// ReasonWaitingForVPC indicates waiting for VPC to be ready
	ReasonWaitingForVPC = "WaitingForVPC"

	// ReasonVPCReadyTimeout indicates VPC ready check timed out
	ReasonVPCReadyTimeout = "VPCReadyTimeout"

	// ReasonControlPlaneTimeout indicates control plane creation timed out
	ReasonControlPlaneTimeout = "ControlPlaneTimeout"

	// ReasonWorkspaceError indicates an error with the workspace
	ReasonWorkspaceError = "WorkspaceError"

	// ReasonFargateProfileCreating indicates a Fargate profile is being created
	ReasonFargateProfileCreating = "FargateProfileCreating"

	// ReasonFargateProfileReady indicates a Fargate profile is ready
	ReasonFargateProfileReady = "FargateProfileReady"

	// ReasonFargateProfileFailed indicates a Fargate profile creation has failed
	ReasonFargateProfileFailed = "FargateProfileFailed"
)

// CAPTControlPlaneSpec defines the desired state of CAPTControlPlane
type CAPTControlPlaneSpec struct {
	// Version defines the desired Kubernetes version.
	// +kubebuilder:validation:Required
	Version string `json:"version"`

	// WorkspaceTemplateRef is a reference to the WorkspaceTemplate used for creating the control plane.
	// +kubebuilder:validation:Required
	WorkspaceTemplateRef WorkspaceTemplateReference `json:"workspaceTemplateRef"`

	// ControlPlaneConfig contains additional configuration for the EKS control plane.
	// +optional
	ControlPlaneConfig *ControlPlaneConfig `json:"controlPlaneConfig,omitempty"`

	// AdditionalTags is an optional set of tags to add to AWS resources managed by the AWS provider.
	// +optional
	AdditionalTags map[string]string `json:"additionalTags,omitempty"`

	// ControlPlaneEndpoint represents the endpoint used to communicate with the control plane.
	// +optional
	ControlPlaneEndpoint clusterv1.APIEndpoint `json:"controlPlaneEndpoint,omitempty"`

	// AdditionalFargateProfiles defines additional Fargate profiles to be created
	// +optional
	AdditionalFargateProfiles []AdditionalFargateProfile `json:"additionalFargateProfiles,omitempty"`
}

// WorkspaceTemplateReference contains the reference to WorkspaceTemplate
type WorkspaceTemplateReference struct {
	// Name is the name of the WorkspaceTemplate.
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Namespace is the namespace of the WorkspaceTemplate.
	// +optional
	Namespace string `json:"namespace,omitempty"`
}

// ControlPlaneConfig contains EKS-specific configuration
type ControlPlaneConfig struct {
	// EndpointAccess defines the access configuration for the API server endpoint
	// +optional
	EndpointAccess *EndpointAccess `json:"endpointAccess,omitempty"`

	// Addons defines the EKS addons to be installed
	// +optional
	Addons []Addon `json:"addons,omitempty"`
}

// EndpointAccess defines the access configuration for the API server endpoint
type EndpointAccess struct {
	// Public controls whether the API server has public access
	// +optional
	Public bool `json:"public,omitempty"`

	// Private controls whether the API server has private access
	// +optional
	Private bool `json:"private,omitempty"`

	// PublicCIDRs is a list of CIDR blocks that can access the API server
	// +optional
	PublicCIDRs []string `json:"publicCIDRs,omitempty"`
}

// Addon represents an EKS addon
type Addon struct {
	// Name is the name of the addon
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Version is the version of the addon
	// +optional
	Version string `json:"version,omitempty"`

	// ConfigurationValues is a string containing configuration values
	// +optional
	ConfigurationValues string `json:"configurationValues,omitempty"`
}

// AdditionalFargateProfile defines an additional Fargate profile to be created
type AdditionalFargateProfile struct {
	// Name is the name of the Fargate profile
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Selectors is a list of label selectors to use for pods
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=1
	Selectors []FargateSelector `json:"selectors"`

	// WorkspaceTemplateRef is a reference to the WorkspaceTemplate used for creating the Fargate profile
	// +kubebuilder:validation:Required
	WorkspaceTemplateRef WorkspaceTemplateReference `json:"workspaceTemplateRef"`
}

// FargateSelector defines the selectors for a Fargate profile
type FargateSelector struct {
	// Namespace is the Kubernetes namespace to select
	// +kubebuilder:validation:Required
	Namespace string `json:"namespace"`

	// Labels is a map of Kubernetes labels to match
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
}

// CAPTControlPlaneStatus defines the observed state of CAPTControlPlane
type CAPTControlPlaneStatus struct {
	// Ready denotes that the control plane is ready
	// +optional
	Ready bool `json:"ready"`

	// Initialized denotes if the control plane has been initialized
	// +optional
	Initialized bool `json:"initialized"`

	// WorkspaceTemplateStatus contains the status of the WorkspaceTemplate
	// +optional
	WorkspaceTemplateStatus *WorkspaceTemplateStatus `json:"workspaceTemplateStatus,omitempty"`

	// FargateProfileStatuses contains the status of additional Fargate profiles
	// +optional
	FargateProfileStatuses []FargateProfileStatus `json:"fargateProfileStatuses,omitempty"`

	// FailureReason indicates that there is a terminal problem reconciling the
	// state, and will be set to a token value suitable for programmatic
	// interpretation.
	// +optional
	FailureReason *string `json:"failureReason,omitempty"`

	// FailureMessage indicates that there is a terminal problem reconciling the
	// state, and will be set to a descriptive error message.
	// +optional
	FailureMessage *string `json:"failureMessage,omitempty"`

	// Phase represents the current phase of the control plane
	// Valid values are: "Creating", "Ready", "Failed"
	// +optional
	// +kubebuilder:validation:Enum=Creating;Ready;Failed
	Phase string `json:"phase,omitempty"`

	// Conditions defines current service state of the CAPTControlPlane.
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// FargateProfileStatus contains the status of a Fargate profile
type FargateProfileStatus struct {
	// Name is the name of the Fargate profile
	Name string `json:"name"`

	// Ready indicates if the Fargate profile is ready
	Ready bool `json:"ready"`

	// WorkspaceTemplateApplyName is the name of the WorkspaceTemplateApply resource
	WorkspaceTemplateApplyName string `json:"workspaceTemplateApplyName"`

	// FailureReason indicates that there is a problem with the Fargate profile
	// +optional
	FailureReason *string `json:"failureReason,omitempty"`

	// FailureMessage provides more detail about the failure
	// +optional
	FailureMessage *string `json:"failureMessage,omitempty"`
}

// WorkspaceTemplateStatus contains the status of the WorkspaceTemplate
type WorkspaceTemplateStatus struct {
	// Ready indicates if the WorkspaceTemplate is ready
	// +optional
	Ready bool `json:"ready"`

	// State represents the current state of the WorkspaceTemplate
	// +optional
	State string `json:"state,omitempty"`

	// LastAppliedRevision is the revision of the WorkspaceTemplate that was last applied
	// +optional
	LastAppliedRevision string `json:"lastAppliedRevision,omitempty"`

	// Outputs contains the outputs from the WorkspaceTemplate
	// +optional
	Outputs map[string]string `json:"outputs,omitempty"`

	// LastFailedRevision is the revision of the WorkspaceTemplate that last failed
	// +optional
	LastFailedRevision string `json:"lastFailedRevision,omitempty"`

	// LastFailureMessage contains the error message from the last failure
	// +optional
	LastFailureMessage string `json:"lastFailureMessage,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.ready",description="Control Plane Ready status"
//+kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase",description="Control Plane Phase"
//+kubebuilder:printcolumn:name="Version",type="string",JSONPath=".spec.version",description="Kubernetes version"
//+kubebuilder:printcolumn:name="Endpoint",type="string",JSONPath=".spec.controlPlaneEndpoint.host",description="API Server Endpoint"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// CAPTControlPlane is the Schema for the captcontrolplanes API
type CAPTControlPlane struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CAPTControlPlaneSpec   `json:"spec,omitempty"`
	Status CAPTControlPlaneStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CAPTControlPlaneList contains a list of CAPTControlPlane
type CAPTControlPlaneList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CAPTControlPlane `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CAPTControlPlane{}, &CAPTControlPlaneList{})
}
