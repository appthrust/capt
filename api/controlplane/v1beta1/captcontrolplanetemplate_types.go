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

	// Spec is the specification of the desired behavior of the CAPTControlPlane.
	// This spec allows for all the same configuration as CAPTControlPlane.
	// +optional
	Spec CAPTControlPlaneSpec `json:"spec,omitempty"`
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
