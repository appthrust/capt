package v1beta2

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type CAPTControlPlaneTemplateSpec struct {
	Template CAPTControlPlaneTemplateResource `json:"template"`
}

type CAPTControlPlaneTemplateResource struct {
	ObjectMeta metav1.ObjectMeta    `json:"metadata,omitempty"`
	Spec       CAPTControlPlaneSpec `json:"spec"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=captcontrolplanetemplates,scope=Namespaced,categories=cluster-api
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:storageversion

type CAPTControlPlaneTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec CAPTControlPlaneTemplateSpec `json:"spec,omitempty"`
}

// Hub marker
func (*CAPTControlPlaneTemplate) Hub() {}

// +kubebuilder:object:root=true

type CAPTControlPlaneTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CAPTControlPlaneTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CAPTControlPlaneTemplate{}, &CAPTControlPlaneTemplateList{})
}
