package v1beta2

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// CaptMachineSpec defines the desired state of CaptMachine
type CaptMachineSpec struct {
	NodeGroupRef         NodeGroupReference         `json:"nodeGroupRef"`
	WorkspaceTemplateRef WorkspaceTemplateReference `json:"workspaceTemplateRef"`
	InstanceType         string                     `json:"instanceType"`
	Labels               map[string]string          `json:"labels,omitempty"`
	Tags                 map[string]string          `json:"tags,omitempty"`
}

type NodeGroupReference struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// CaptMachineStatus defines the observed state of CaptMachine
type CaptMachineStatus struct {
	Ready              bool               `json:"ready"`
	InstanceID         *string            `json:"instanceId,omitempty"`
	PrivateIP          *string            `json:"privateIp,omitempty"`
	LastTransitionTime *metav1.Time       `json:"lastTransitionTime,omitempty"`
	Conditions         []metav1.Condition `json:"conditions,omitempty"`
	FailureReason      *string            `json:"failureReason,omitempty"`
	FailureMessage     *string            `json:"failureMessage,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Ready",type="boolean",JSONPath=".status.ready",description="Machine Ready status"
// +kubebuilder:printcolumn:name="Instance ID",type="string",JSONPath=".status.instanceId",description="EC2 Instance ID"
// +kubebuilder:printcolumn:name="Node Group",type="string",JSONPath=".spec.nodeGroupRef.name",description="Node Group name"
// +kubebuilder:resource:path=captmachines,scope=Namespaced,categories=cluster-api
// +kubebuilder:storageversion

// CaptMachine is the Schema for the captmachines API
type CaptMachine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CaptMachineSpec   `json:"spec,omitempty"`
	Status CaptMachineStatus `json:"status,omitempty"`
}

// Hub marker
func (*CaptMachine) Hub() {}

// +kubebuilder:object:root=true

// CaptMachineList contains a list of CaptMachine
type CaptMachineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CaptMachine `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CaptMachine{}, &CaptMachineList{})
}
