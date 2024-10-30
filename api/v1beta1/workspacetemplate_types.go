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
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ModuleSourceType defines the type of module source
type ModuleSourceType string

const (
	// ModuleSourceInline indicates the module is defined inline
	ModuleSourceInline ModuleSourceType = "Inline"
	// ModuleSourceGit indicates the module is sourced from a git repository
	ModuleSourceGit ModuleSourceType = "Git"
)

// Variable defines a variable to be passed to the Terraform module
type Variable struct {
	// Key is the name of the variable
	Key string `json:"key"`

	// Value is the value of the variable
	// +optional
	Value string `json:"value,omitempty"`

	// ValueFrom allows the value to be sourced from another resource
	// +optional
	ValueFrom *VariableSource `json:"valueFrom,omitempty"`
}

// VariableSource defines the source of a variable value
type VariableSource struct {
	// SecretKeyRef selects a key of a Secret
	// +optional
	SecretKeyRef *SecretKeySelector `json:"secretKeyRef,omitempty"`

	// ConfigMapKeyRef selects a key of a ConfigMap
	// +optional
	ConfigMapKeyRef *ConfigMapKeySelector `json:"configMapKeyRef,omitempty"`
}

// SecretKeySelector selects a key of a Secret
type SecretKeySelector struct {
	// Name of the secret
	Name string `json:"name"`

	// Namespace of the secret
	// +optional
	Namespace string `json:"namespace,omitempty"`

	// Key is the key in the secret
	Key string `json:"key"`
}

// ConfigMapKeySelector selects a key of a ConfigMap
type ConfigMapKeySelector struct {
	// Name of the configmap
	Name string `json:"name"`

	// Namespace of the configmap
	// +optional
	Namespace string `json:"namespace,omitempty"`

	// Key is the key in the configmap
	Key string `json:"key"`
}

// WorkspaceTemplateSpec defines the desired state of WorkspaceTemplate
type WorkspaceTemplateSpec struct {
	// Module is the HCL format Terraform module
	// +kubebuilder:validation:Required
	Module string `json:"module"`

	// Source specifies how the module is sourced
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=Inline;Git
	Source ModuleSourceType `json:"source"`

	// Variables is a list of variables to pass to the Terraform module
	// +optional
	Variables []Variable `json:"variables,omitempty"`

	// WriteConnectionSecretToRef specifies the namespace and name of a
	// Secret to which any connection details for this managed resource should
	// be written.
	// +optional
	WriteConnectionSecretToRef *xpv1.SecretReference `json:"writeConnectionSecretToRef,omitempty"`
}

// WorkspaceTemplateStatus defines the observed state of WorkspaceTemplate
type WorkspaceTemplateStatus struct {
	// WorkspaceName is the name of the created Terraform Workspace
	// +optional
	WorkspaceName string `json:"workspaceName,omitempty"`

	// Conditions of the resource.
	// +optional
	Conditions []xpv1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="WORKSPACE",type="string",JSONPath=".status.workspaceName"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"

// WorkspaceTemplate is the Schema for the workspacetemplates API
type WorkspaceTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WorkspaceTemplateSpec   `json:"spec,omitempty"`
	Status WorkspaceTemplateStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// WorkspaceTemplateList contains a list of WorkspaceTemplate
type WorkspaceTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WorkspaceTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&WorkspaceTemplate{}, &WorkspaceTemplateList{})
}
