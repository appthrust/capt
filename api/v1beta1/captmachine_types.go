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
	// WorkspaceTemplateRef is a reference to the WorkspaceTemplate used for creating the node group
	// +kubebuilder:validation:Required
	WorkspaceTemplateRef WorkspaceTemplateReference `json:"workspaceTemplateRef"`

	// NodeGroupConfig contains the configuration for the managed node group
	// +kubebuilder:validation:Required
	NodeGroupConfig NodeGroupConfig `json:"nodeGroupConfig"`
}

// NodeGroupConfig defines the configuration for the managed node group
type NodeGroupConfig struct {
	// Name is the name of the managed node group
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// InstanceType is the EC2 instance type to use for the nodes
	// +kubebuilder:validation:Required
	InstanceType string `json:"instanceType"`

	// Scaling defines the scaling configuration for the node group
	// +kubebuilder:validation:Required
	Scaling ScalingConfig `json:"scaling"`

	// UpdateConfig defines the update configuration for the node group
	// +optional
	UpdateConfig *UpdateConfig `json:"updateConfig,omitempty"`

	// Labels is a map of kubernetes labels to apply to the nodes
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// Tags is a map of tags to apply to the node group
	// +optional
	Tags map[string]string `json:"tags,omitempty"`
}

// ScalingConfig defines the scaling configuration for the node group
type ScalingConfig struct {
	// MinSize is the minimum size of the node group
	// +kubebuilder:validation:Minimum=0
	MinSize int32 `json:"minSize"`

	// MaxSize is the maximum size of the node group
	// +kubebuilder:validation:Minimum=1
	MaxSize int32 `json:"maxSize"`

	// DesiredSize is the desired size of the node group
	// +kubebuilder:validation:Minimum=0
	DesiredSize int32 `json:"desiredSize"`
}

// UpdateConfig defines the update configuration for the node group
type UpdateConfig struct {
	// MaxUnavailablePercentage is the maximum percentage of nodes that can be unavailable during an update
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	MaxUnavailablePercentage *int32 `json:"maxUnavailablePercentage,omitempty"`
}

// CaptMachineStatus defines the observed state of CaptMachine
type CaptMachineStatus struct {
	// Ready denotes that the node group is ready
	// +optional
	Ready bool `json:"ready"`

	// CurrentSize is the current size of the node group
	// +optional
	CurrentSize *int32 `json:"currentSize,omitempty"`

	// LastScalingTime is the last time the node group was scaled
	// +optional
	LastScalingTime *metav1.Time `json:"lastScalingTime,omitempty"`

	// LastUpdateTime is the last time the node group was updated
	// +optional
	LastUpdateTime *metav1.Time `json:"lastUpdateTime,omitempty"`

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
//+kubebuilder:printcolumn:name="Current Size",type="integer",JSONPath=".status.currentSize",description="Current number of nodes"

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
