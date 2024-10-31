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
	tfv1beta1 "github.com/upbound/provider-terraform/apis/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// WorkspaceTemplateSpec defines the desired state of WorkspaceTemplate
type WorkspaceTemplateSpec struct {
	// Template defines the workspace template
	// +kubebuilder:validation:Required
	Template WorkspaceTemplateDefinition `json:"template"`

	// WriteConnectionSecretToRef specifies the namespace and name of a
	// Secret to which any connection details for this managed resource should
	// be written.
	// +optional
	WriteConnectionSecretToRef *xpv1.SecretReference `json:"writeConnectionSecretToRef,omitempty"`
}

// WorkspaceTemplateDefinition defines the template for creating workspaces
type WorkspaceTemplateDefinition struct {
	// Metadata contains template-specific metadata
	// +optional
	Metadata *WorkspaceTemplateMetadata `json:"metadata,omitempty"`

	// Spec defines the desired state of the workspace
	// +kubebuilder:validation:Required
	Spec tfv1beta1.WorkspaceSpec `json:"spec"`
}

// WorkspaceTemplateMetadata contains metadata specific to the workspace template
type WorkspaceTemplateMetadata struct {
	// Description provides a human-readable description of the template
	// +optional
	Description string `json:"description,omitempty"`

	// Version specifies the version of this template
	// +optional
	Version string `json:"version,omitempty"`

	// Tags are key-value pairs that can be used to organize and categorize templates
	// +optional
	Tags map[string]string `json:"tags,omitempty"`
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

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="WORKSPACE",type="string",JSONPath=".status.workspaceName"
//+kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"

// WorkspaceTemplate is the Schema for the workspacetemplates API
type WorkspaceTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WorkspaceTemplateSpec   `json:"spec,omitempty"`
	Status WorkspaceTemplateStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// WorkspaceTemplateList contains a list of WorkspaceTemplate
type WorkspaceTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WorkspaceTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&WorkspaceTemplate{}, &WorkspaceTemplateList{})
}
