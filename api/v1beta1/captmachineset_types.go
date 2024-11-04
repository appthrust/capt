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

// CaptMachineSetSpec defines the desired state of CaptMachineSet
type CaptMachineSetSpec struct {
	// Replicas is the number of desired replicas.
	// This is a pointer to distinguish between explicit zero and unspecified.
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// Selector is a label query over machines that should match the replica count.
	// It must match the machine template's labels.
	// +optional
	Selector *metav1.LabelSelector `json:"selector,omitempty"`

	// Template is the object that describes the machine that will be created if
	// insufficient replicas are detected.
	Template CaptMachineTemplateSpec `json:"template"`
}

// CaptMachineTemplateSpec describes the data needed to create a CaptMachine from a template
type CaptMachineTemplateSpec struct {
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	ObjectMeta metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the desired behavior of the machine.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	// +optional
	Spec CaptMachineSpec `json:"spec,omitempty"`
}

// CaptMachineSetStatus defines the observed state of CaptMachineSet
type CaptMachineSetStatus struct {
	// Replicas is the most recently observed number of replicas.
	Replicas int32 `json:"replicas"`

	// The number of replicas that have labels matching the labels of the machine template of the MachineSet.
	// +optional
	FullyLabeledReplicas int32 `json:"fullyLabeledReplicas,omitempty"`

	// The number of ready replicas for this machine set.
	// +optional
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`

	// The number of available replicas (ready for at least minReadySeconds) for this machine set.
	// +optional
	AvailableReplicas int32 `json:"availableReplicas,omitempty"`

	// ObservedGeneration reflects the generation of the most recently observed MachineSet.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Conditions defines current service state of the CaptMachineSet
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
//+kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.replicas
//+kubebuilder:printcolumn:name="Desired",type="integer",JSONPath=".spec.replicas",description="Number of desired machines"
//+kubebuilder:printcolumn:name="Current",type="integer",JSONPath=".status.replicas",description="Current number of machines"
//+kubebuilder:printcolumn:name="Ready",type="integer",JSONPath=".status.readyReplicas",description="Number of ready machines"
//+kubebuilder:printcolumn:name="Available",type="integer",JSONPath=".status.availableReplicas",description="Number of available machines"

// CaptMachineSet is the Schema for the captmachinesets API
type CaptMachineSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CaptMachineSetSpec   `json:"spec,omitempty"`
	Status CaptMachineSetStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CaptMachineSetList contains a list of CaptMachineSet
type CaptMachineSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CaptMachineSet `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CaptMachineSet{}, &CaptMachineSetList{})
}
