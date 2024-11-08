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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NodeType defines the type of node group
// +kubebuilder:validation:Enum=ManagedNodeGroup;Fargate
type NodeType string

const (
	// ManagedNodeGroup represents an EKS managed node group
	ManagedNodeGroup NodeType = "ManagedNodeGroup"
	// Fargate represents a Fargate profile
	Fargate NodeType = "Fargate"
)

// CaptInfraMachineTemplateSpec defines the desired state of CaptMachineTemplate
type CaptInfraMachineTemplateSpec struct {
	// Template is the template for creating a CaptMachine
	Template CaptInfraMachineTemplateResource `json:"template"`
}

// CaptInfraMachineTemplateResource describes the data needed to create a CaptMachine from a template
type CaptInfraMachineTemplateResource struct {
	// Spec is the specification of the desired behavior of the machine.
	Spec CaptInfraMachineTemplateResourceSpec `json:"spec"`
}

// CaptInfraMachineTemplateResourceSpec defines the desired state of CaptMachineTemplateResource
type CaptInfraMachineTemplateResourceSpec struct {
	// WorkspaceTemplateRef is a reference to the WorkspaceTemplate used for creating the machine
	// +kubebuilder:validation:Required
	WorkspaceTemplateRef WorkspaceTemplateReference `json:"workspaceTemplateRef"`

	// NodeType specifies the type of node group (ManagedNodeGroup or Fargate)
	// +kubebuilder:validation:Required
	NodeType NodeType `json:"nodeType"`

	// InstanceType is the EC2 instance type to use for the node
	// +optional
	InstanceType string `json:"instanceType,omitempty"`

	// Scaling defines the scaling configuration for the node group
	// +optional
	Scaling *ScalingConfig `json:"scaling,omitempty"`

	// Labels is a map of kubernetes labels to apply to the node
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// Taints specifies the taints to apply to the nodes
	// +optional
	Taints []corev1.Taint `json:"taints,omitempty"`

	// AdditionalTags is a map of additional AWS tags to apply to the node group
	// +optional
	AdditionalTags map[string]string `json:"additionalTags,omitempty"`
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

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:path=captmachinetemplates,scope=Namespaced,categories=cluster-api

// CaptMachineTemplate is the Schema for the captmachinetemplates API
type CaptMachineTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec CaptInfraMachineTemplateSpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true

// CaptMachineTemplateList contains a list of CaptMachineTemplate
type CaptMachineTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CaptMachineTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CaptMachineTemplate{}, &CaptMachineTemplateList{})
}
