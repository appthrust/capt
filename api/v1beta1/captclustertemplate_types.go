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

	// Spec is the specification of the desired behavior of the CAPTCluster.
	// This spec allows for all the same configuration as CAPTCluster.
	// +optional
	Spec CAPTClusterSpec `json:"spec,omitempty"`
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
