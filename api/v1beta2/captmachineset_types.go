package v1beta2

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type CaptMachineSetSpec struct {
	Replicas *int32                  `json:"replicas,omitempty"`
	Selector *metav1.LabelSelector   `json:"selector,omitempty"`
	Template CaptMachineTemplateSpec `json:"template"`
}

type CaptMachineTemplateSpec struct {
	ObjectMeta metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec       CaptMachineSpec   `json:"spec,omitempty"`
}

type CaptMachineSetStatus struct {
	Replicas             int32              `json:"replicas"`
	FullyLabeledReplicas int32              `json:"fullyLabeledReplicas,omitempty"`
	ReadyReplicas        int32              `json:"readyReplicas,omitempty"`
	AvailableReplicas    int32              `json:"availableReplicas,omitempty"`
	ObservedGeneration   int64              `json:"observedGeneration,omitempty"`
	Conditions           []metav1.Condition `json:"conditions,omitempty"`
	FailureReason        *string            `json:"failureReason,omitempty"`
	FailureMessage       *string            `json:"failureMessage,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.replicas
// +kubebuilder:printcolumn:name="Desired",type="integer",JSONPath=".spec.replicas",description="Number of desired machines"
// +kubebuilder:printcolumn:name="Current",type="integer",JSONPath=".status.replicas",description="Current number of machines"
// +kubebuilder:printcolumn:name="Ready",type="integer",JSONPath=".status.readyReplicas",description="Number of ready machines"
// +kubebuilder:printcolumn:name="Available",type="integer",JSONPath=".status.availableReplicas",description="Number of available machines"
// +kubebuilder:resource:path=captmachinesets,scope=Namespaced,categories=cluster-api
// +kubebuilder:storageversion

type CaptMachineSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CaptMachineSetSpec   `json:"spec,omitempty"`
	Status CaptMachineSetStatus `json:"status,omitempty"`
}

// Hub marker
func (*CaptMachineSet) Hub() {}

// +kubebuilder:object:root=true

type CaptMachineSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CaptMachineSet `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CaptMachineSet{}, &CaptMachineSetList{})
}
