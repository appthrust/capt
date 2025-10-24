package v1beta2

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type WorkspaceTemplateApplySpec struct {
	TemplateRef                WorkspaceTemplateReference `json:"templateRef"`
	WriteConnectionSecretToRef *xpv1.SecretReference      `json:"writeConnectionSecretToRef,omitempty"`
	Variables                  map[string]string          `json:"variables,omitempty"`
	WaitForSecrets             []xpv1.SecretReference     `json:"waitForSecrets,omitempty"`
	WaitForWorkspaces          []WorkspaceReference       `json:"waitForWorkspaces,omitempty"`
	RetainWorkspaceOnDelete    bool                       `json:"retainWorkspaceOnDelete,omitempty"`
}

func (s *WorkspaceTemplateApplySpec) ValidateConfiguration() error { return nil }

// moved to common_types.go

type WorkspaceTemplateApplyStatus struct {
	WorkspaceName   string           `json:"workspaceName,omitempty"`
	Applied         bool             `json:"applied,omitempty"`
	LastAppliedTime *metav1.Time     `json:"lastAppliedTime,omitempty"`
	Conditions      []xpv1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="WORKSPACE",type="string",JSONPath=".status.workspaceName"
// +kubebuilder:printcolumn:name="APPLIED",type="boolean",JSONPath=".status.applied"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:categories={capt,terraform},shortName=wtapply,scope=Namespaced,path=workspacetemplateapplies,singular=workspacetemplateapply
// +kubebuilder:storageversion

type WorkspaceTemplateApply struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WorkspaceTemplateApplySpec   `json:"spec,omitempty"`
	Status WorkspaceTemplateApplyStatus `json:"status,omitempty"`
}

// Hub marker
func (*WorkspaceTemplateApply) Hub() {}

// +kubebuilder:object:root=true

type WorkspaceTemplateApplyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WorkspaceTemplateApply `json:"items"`
}

func init() {
	SchemeBuilder.Register(&WorkspaceTemplateApply{}, &WorkspaceTemplateApplyList{})
}
