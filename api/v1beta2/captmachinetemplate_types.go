package v1beta2

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// CAPTMachineTemplateSpec defines the desired state of CaptMachineTemplate
type CAPTMachineTemplateSpec struct {
	Template CAPTMachineTemplateResource `json:"template"`
}

// CAPTMachineTemplateResource describes the data needed to create a CaptMachine from a template
type CAPTMachineTemplateResource struct {
	Spec CaptMachineSpec `json:"spec"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=captmachinetemplates,scope=Namespaced,categories=cluster-api
// +kubebuilder:storageversion

// CaptMachineTemplate is the Schema for the captmachinetemplates API
type CaptMachineTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec CAPTMachineTemplateSpec `json:"spec,omitempty"`
}

// Hub marker
func (*CaptMachineTemplate) Hub() {}

// +kubebuilder:object:root=true

// CaptMachineTemplateList contains a list of CaptMachineTemplate
type CaptMachineTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CaptMachineTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CaptMachineTemplate{}, &CaptMachineTemplateList{})
}
