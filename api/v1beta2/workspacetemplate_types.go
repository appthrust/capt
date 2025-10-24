package v1beta2

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	tfv1beta1 "github.com/upbound/provider-terraform/apis/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type WorkspaceTemplateSpec struct {
	Template                   WorkspaceTemplateDefinition `json:"template"`
	WriteConnectionSecretToRef *xpv1.SecretReference       `json:"writeConnectionSecretToRef,omitempty"`
}

type WorkspaceTemplateDefinition struct {
	Metadata *WorkspaceTemplateMetadata `json:"metadata,omitempty"`
	Spec     tfv1beta1.WorkspaceSpec    `json:"spec"`
}

type WorkspaceTemplateMetadata struct {
	Description string            `json:"description,omitempty"`
	Version     string            `json:"version,omitempty"`
	Tags        map[string]string `json:"tags,omitempty"`
}

type WorkspaceTemplateStatus struct {
	WorkspaceName string           `json:"workspaceName,omitempty"`
	Conditions    []xpv1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="WORKSPACE",type="string",JSONPath=".status.workspaceName"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:path=workspacetemplates,scope=Namespaced
// +kubebuilder:storageversion

type WorkspaceTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WorkspaceTemplateSpec   `json:"spec,omitempty"`
	Status WorkspaceTemplateStatus `json:"status,omitempty"`
}

// Hub marker
func (*WorkspaceTemplate) Hub() {}

// +kubebuilder:object:root=true

type WorkspaceTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WorkspaceTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&WorkspaceTemplate{}, &WorkspaceTemplateList{})
}
