package v1beta2

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type CAPTClusterTemplateSpec struct {
	Template CAPTClusterTemplateResource `json:"template"`
}

type CAPTClusterTemplateResource struct {
	ObjectMeta metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec       CAPTClusterSpec   `json:"spec"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=captclustertemplates,scope=Namespaced,categories=cluster-api,shortName=captct
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:storageversion

type CAPTClusterTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec CAPTClusterTemplateSpec `json:"spec,omitempty"`
}

// Hub marker
func (*CAPTClusterTemplate) Hub() {}

// +kubebuilder:object:root=true

type CAPTClusterTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CAPTClusterTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CAPTClusterTemplate{}, &CAPTClusterTemplateList{})
}
