/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CaptMachineSpec defines the desired state of CaptMachine
type CaptMachineSpec struct {
	// NodeGroupRef is a reference to the NodeGroup this machine belongs to
	// +kubebuilder:validation:Required
	NodeGroupRef NodeGroupReference `json:"nodeGroupRef"`

	// WorkspaceTemplateRef is a reference to the WorkspaceTemplate used for creating the machine
	// +kubebuilder:validation:Required
	WorkspaceTemplateRef WorkspaceTemplateReference `json:"workspaceTemplateRef"`

	// InstanceType is the EC2 instance type to use for the node
	// +kubebuilder:validation:Required
	InstanceType string `json:"instanceType"`

	// Labels is a map of kubernetes labels to apply to the node
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// Tags is a map of tags to apply to the node
	// +optional
	Tags map[string]string `json:"tags,omitempty"`
}

// NodeGroupReference contains the information necessary to let you specify a NodeGroup
type NodeGroupReference struct {
	// Name is the name of the NodeGroup
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Namespace is the namespace of the NodeGroup
	// +kubebuilder:validation:Required
	Namespace string `json:"namespace"`
}

// CaptMachineStatus defines the observed state of CaptMachine
type CaptMachineStatus struct {
	// Ready denotes that the machine is ready and joined to the node group
	// +optional
	Ready bool `json:"ready"`

	// InstanceID is the ID of the EC2 instance
	// +optional
	InstanceID *string `json:"instanceId,omitempty"`

	// PrivateIP is the private IP address of the machine
	// +optional
	PrivateIP *string `json:"privateIp,omitempty"`

	// LastTransitionTime is the last time the Ready condition changed
	// +optional
	LastTransitionTime *metav1.Time `json:"lastTransitionTime,omitempty"`

	// Conditions defines current service state of the CaptMachine
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// FailureReason indicates that there is a terminal problem reconciling the
	// state, and will be set to a token value suitable for programmatic
	// interpretation.
	// +optional
	FailureReason *string `json:"failureReason,omitempty"`

	// FailureMessage indicates that there is a terminal problem reconciling the
	// state, and will be set to a descriptive error message.
	// +optional
	FailureMessage *string `json:"failureMessage,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Ready",type="boolean",JSONPath=".status.ready",description="Machine Ready status"
//+kubebuilder:printcolumn:name="Instance ID",type="string",JSONPath=".status.instanceId",description="EC2 Instance ID"
//+kubebuilder:printcolumn:name="Node Group",type="string",JSONPath=".spec.nodeGroupRef.name",description="Node Group name"

// CaptMachine is the Schema for the captmachines API
type CaptMachine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CaptMachineSpec   `json:"spec,omitempty"`
	Status CaptMachineStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CaptMachineList contains a list of CaptMachine
type CaptMachineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CaptMachine `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CaptMachine{}, &CaptMachineList{})
}
