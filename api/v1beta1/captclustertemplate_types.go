package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CAPTClusterTemplateSpec defines the desired state of CAPTClusterTemplate
type CAPTClusterTemplateSpec struct {
	Template CAPTClusterTemplateResource `json:"template"`
}

// CAPTClusterTemplateResource describes the data needed to create a CAPTCluster from a template
type CAPTClusterTemplateResource struct {
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	ObjectMeta metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec is the specification for the template. All fields are optional to
	// allow ClusterClass/topology patches to populate values.
	// +optional
	Spec CAPTClusterTemplateResourceSpec `json:"spec,omitempty"`
}

// TemplateWorkspaceTemplateReference contains the reference to a WorkspaceTemplate (all optional for templates)
type TemplateWorkspaceTemplateReference struct {
	// +optional
	Name string `json:"name,omitempty"`
	// +optional
	Namespace string `json:"namespace,omitempty"`
}

// CAPTClusterTemplateResourceSpec defines the desired state of CAPTCluster for templates (all optional)
type CAPTClusterTemplateResourceSpec struct {
	// +optional
	Region string `json:"region,omitempty"`

	// VPCTemplateRef is a reference to a WorkspaceTemplate resource for VPC configuration
	// +optional
	VPCTemplateRef *TemplateWorkspaceTemplateReference `json:"vpcTemplateRef,omitempty"`

	// ExistingVPCID is the ID of an existing VPC to use
	// +optional
	ExistingVPCID string `json:"existingVpcId,omitempty"`

	// RetainVPCOnDelete specifies whether to retain the VPC when the parent cluster is deleted
	// +optional
	RetainVPCOnDelete bool `json:"retainVpcOnDelete,omitempty"`

	// VPCConfig contains VPC-specific configuration
	// +optional
	VPCConfig *VPCConfig `json:"vpcConfig,omitempty"`

	// WorkspaceTemplateApplyName is the name of the WorkspaceTemplateApply used for this cluster.
	// +optional
	WorkspaceTemplateApplyName string `json:"workspaceTemplateApplyName,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=captclustertemplates,scope=Namespaced,categories=cluster-api,shortName=captct
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// CAPTClusterTemplate is the Schema for the captclustertemplates API
type CAPTClusterTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec CAPTClusterTemplateSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

// CAPTClusterTemplateList contains a list of CAPTClusterTemplate
type CAPTClusterTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CAPTClusterTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CAPTClusterTemplate{}, &CAPTClusterTemplateList{})
}
