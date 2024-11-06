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

// WorkspaceTemplateApplySpec defines the desired state of WorkspaceTemplateApply
type WorkspaceTemplateApplySpec struct {
	// TemplateRef references the WorkspaceTemplate to be applied
	// +kubebuilder:validation:Required
	TemplateRef WorkspaceTemplateReference `json:"templateRef"`

	// WriteConnectionSecretToRef specifies the namespace and name of a
	// Secret to which any connection details for this managed resource should
	// be written.
	// +optional
	WriteConnectionSecretToRef *xpv1.SecretReference `json:"writeConnectionSecretToRef,omitempty"`

	// Variables are used to override or provide additional variables to the workspace
	// +optional
	Variables map[string]string `json:"variables,omitempty"`

	// WaitForSecrets specifies a list of secrets that must exist before creating the workspace
	// +optional
	WaitForSecrets []xpv1.SecretReference `json:"waitForSecrets,omitempty"`

	// WaitForWorkspaces specifies a list of workspaces that must be ready before creating this workspace
	// +optional
	WaitForWorkspaces []WorkspaceReference `json:"waitForWorkspaces,omitempty"`

	// RetainWorkspaceOnDelete specifies whether to retain the Workspace when this WorkspaceTemplateApply is deleted
	// This is useful when the Workspace manages shared resources that should outlive this WorkspaceTemplateApply
	// +optional
	RetainWorkspaceOnDelete bool `json:"retainWorkspaceOnDelete,omitempty"`
}

// WorkspaceTemplateReference contains the reference to a WorkspaceTemplate
type WorkspaceTemplateReference struct {
	// Name of the referenced WorkspaceTemplate
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Namespace of the referenced WorkspaceTemplate
	// +optional
	Namespace string `json:"namespace,omitempty"`
}

// WorkspaceReference defines a reference to a Workspace
type WorkspaceReference struct {
	// Name of the referenced Workspace
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Namespace of the referenced Workspace
	// +optional
	Namespace string `json:"namespace,omitempty"`
}

// WorkspaceTemplateApplyStatus defines the observed state of WorkspaceTemplateApply
type WorkspaceTemplateApplyStatus struct {
	// WorkspaceName is the name of the created Terraform Workspace
	// +optional
	WorkspaceName string `json:"workspaceName,omitempty"`

	// Applied indicates whether the template has been successfully applied
	// +optional
	Applied bool `json:"applied,omitempty"`

	// LastAppliedTime is the last time this template was applied
	// +optional
	LastAppliedTime *metav1.Time `json:"lastAppliedTime,omitempty"`

	// Conditions of the resource.
	// +optional
	Conditions []xpv1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="WORKSPACE",type="string",JSONPath=".status.workspaceName"
//+kubebuilder:printcolumn:name="APPLIED",type="boolean",JSONPath=".status.applied"
//+kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:resource:categories={capt,terraform},shortName=wtapply,scope=Namespaced,path=workspacetemplateapplies,singular=workspacetemplateapply
//+groupName=infrastructure.cluster.x-k8s.io

// WorkspaceTemplateApply is the Schema for the workspacetemplateapplies API
type WorkspaceTemplateApply struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WorkspaceTemplateApplySpec   `json:"spec,omitempty"`
	Status WorkspaceTemplateApplyStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// WorkspaceTemplateApplyList contains a list of WorkspaceTemplateApply
type WorkspaceTemplateApplyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WorkspaceTemplateApply `json:"items"`
}

func init() {
	SchemeBuilder.Register(&WorkspaceTemplateApply{}, &WorkspaceTemplateApplyList{})
}
