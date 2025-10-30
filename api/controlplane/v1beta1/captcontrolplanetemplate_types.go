package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CAPTControlPlaneTemplateSpec defines the desired state of CAPTControlPlaneTemplate
type CAPTControlPlaneTemplateSpec struct {
	// Template is the template for the CaptControlPlane
	Template CAPTControlPlaneTemplateResource `json:"template"`
}

// CAPTControlPlaneTemplateResource describes the data needed to create a CAPTControlPlane from a template
type CAPTControlPlaneTemplateResource struct {
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	ObjectMeta metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec is the specification for the template. All fields are optional to
	// allow ClusterClass/topology patches to populate values.
	// +optional
	Spec CAPTControlPlaneTemplateResourceSpec `json:"spec,omitempty"`
}

// APIEndpointTemplate represents the endpoint used to communicate with the control plane (all optional for templates)
type APIEndpointTemplate struct {
	// The hostname on which the API server is serving.
	// +optional
	Host string `json:"host,omitempty"`

	// The port on which the API server is serving.
	// +optional
	Port int32 `json:"port,omitempty"`
}

// WorkspaceTemplateReferenceTemplate contains the reference to WorkspaceTemplate (all optional for templates)
type WorkspaceTemplateReferenceTemplate struct {
	// Name is the name of the WorkspaceTemplate.
	// +optional
	Name string `json:"name,omitempty"`

	// Namespace is the namespace of the WorkspaceTemplate.
	// +optional
	Namespace string `json:"namespace,omitempty"`
}

// TimeoutConfigTemplate defines timeout settings (kept optional)
type TimeoutConfigTemplate struct {
	// ControlPlaneTimeout is the timeout in minutes for control plane creation
	// +optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=30
	ControlPlaneTimeout *int `json:"controlPlaneTimeout,omitempty"`

	// VPCReadyTimeout is the timeout in minutes for VPC ready check
	// +optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=15
	VPCReadyTimeout *int `json:"vpcReadyTimeout,omitempty"`
}

// EndpointAccessTemplate defines API server endpoint access (all optional)
type EndpointAccessTemplate struct {
	// +optional
	Public bool `json:"public,omitempty"`
	// +optional
	Private bool `json:"private,omitempty"`
	// +optional
	PublicCIDRs []string `json:"publicCIDRs,omitempty"`
}

// AddonTemplate represents an EKS addon (all optional)
type AddonTemplate struct {
	// +optional
	Name string `json:"name,omitempty"`
	// +optional
	Version string `json:"version,omitempty"`
	// +optional
	ConfigurationValues string `json:"configurationValues,omitempty"`
}

// ControlPlaneConfigTemplate contains EKS-specific configuration for templates (all optional)
type ControlPlaneConfigTemplate struct {
	// AWS region where the control plane will be created
	// +optional
	// +kubebuilder:validation:Pattern="^[a-z]{2}-[a-z]+-[0-9]$"
	Region string `json:"region,omitempty"`

	// +optional
	EndpointAccess *EndpointAccessTemplate `json:"endpointAccess,omitempty"`
	// +optional
	Addons []AddonTemplate `json:"addons,omitempty"`
	// +optional
	Timeouts *TimeoutConfigTemplate `json:"timeouts,omitempty"`
}

// CAPTControlPlaneTemplateResourceSpec is the template spec with all fields optional
type CAPTControlPlaneTemplateResourceSpec struct {
	// +optional
	Version string `json:"version,omitempty"`
	// +optional
	WorkspaceTemplateRef *WorkspaceTemplateReferenceTemplate `json:"workspaceTemplateRef,omitempty"`
	// +optional
	ControlPlaneConfig *ControlPlaneConfigTemplate `json:"controlPlaneConfig,omitempty"`
	// +optional
	AdditionalTags map[string]string `json:"additionalTags,omitempty"`
	// +optional
	ControlPlaneEndpoint *APIEndpointTemplate `json:"controlPlaneEndpoint,omitempty"`
	// +optional
	WorkspaceTemplateApplyName string `json:"workspaceTemplateApplyName,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=captcontrolplanetemplates,scope=Namespaced,categories=cluster-api
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// CAPTControlPlaneTemplate is the Schema for the captcontrolplanetemplates API
type CAPTControlPlaneTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec CAPTControlPlaneTemplateSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

// CAPTControlPlaneTemplateList contains a list of CAPTControlPlaneTemplate
type CAPTControlPlaneTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CAPTControlPlaneTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CAPTControlPlaneTemplate{}, &CAPTControlPlaneTemplateList{})
}
