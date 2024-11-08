package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CaptControlPlaneTemplateSpec defines the desired state of CaptControlPlaneTemplate
type CaptControlPlaneTemplateSpec struct {
	// Template is the template for the CaptControlPlane
	Template CaptControlPlaneTemplateResource `json:"template"`
}

// CaptControlPlaneTemplateResource describes the data needed to create a CaptControlPlane from a template
type CaptControlPlaneTemplateResource struct {
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	ObjectMeta metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec is the specification of the desired behavior of the CaptControlPlane.
	// This spec allows for all the same configuration as CaptControlPlane.
	// +optional
	Spec CAPTControlPlaneSpec `json:"spec"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:path=captcontrolplanetemplates,scope=Namespaced,categories=cluster-api
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// CaptControlPlaneTemplate is the Schema for the captcontrolplanetemplates API
type CaptControlPlaneTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec CaptControlPlaneTemplateSpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true

// CaptControlPlaneTemplateList contains a list of CaptControlPlaneTemplate
type CaptControlPlaneTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CaptControlPlaneTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CaptControlPlaneTemplate{}, &CaptControlPlaneTemplateList{})
}
