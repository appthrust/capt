package v1beta1

import (
	v1beta2 "github.com/appthrust/capt/api/v1beta2"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

// ConvertTo converts this v1beta1 WorkspaceTemplateApply to the hub version (v1beta2).
func (src *WorkspaceTemplateApply) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1beta2.WorkspaceTemplateApply)

	dst.ObjectMeta = src.ObjectMeta

	// Spec
	dst.Spec.TemplateRef = v1beta2.WorkspaceTemplateReference{
		Name:      src.Spec.TemplateRef.Name,
		Namespace: src.Spec.TemplateRef.Namespace,
	}
	dst.Spec.WriteConnectionSecretToRef = src.Spec.WriteConnectionSecretToRef
	dst.Spec.Variables = src.Spec.Variables
	dst.Spec.WaitForSecrets = src.Spec.WaitForSecrets
	dst.Spec.WaitForWorkspaces = make([]v1beta2.WorkspaceReference, len(src.Spec.WaitForWorkspaces))
	for i, w := range src.Spec.WaitForWorkspaces {
		dst.Spec.WaitForWorkspaces[i] = v1beta2.WorkspaceReference{Name: w.Name, Namespace: w.Namespace}
	}
	dst.Spec.RetainWorkspaceOnDelete = src.Spec.RetainWorkspaceOnDelete

	// Status
	dst.Status.WorkspaceName = src.Status.WorkspaceName
	dst.Status.Applied = src.Status.Applied
	dst.Status.LastAppliedTime = src.Status.LastAppliedTime
	dst.Status.Conditions = src.Status.Conditions
	return nil
}

// ConvertFrom converts from the hub version (v1beta2) to this version.
func (dst *WorkspaceTemplateApply) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1beta2.WorkspaceTemplateApply)

	dst.ObjectMeta = src.ObjectMeta

	// Spec
	dst.Spec.TemplateRef = WorkspaceTemplateReference{
		Name:      src.Spec.TemplateRef.Name,
		Namespace: src.Spec.TemplateRef.Namespace,
	}
	dst.Spec.WriteConnectionSecretToRef = src.Spec.WriteConnectionSecretToRef
	dst.Spec.Variables = src.Spec.Variables
	dst.Spec.WaitForSecrets = src.Spec.WaitForSecrets
	dst.Spec.WaitForWorkspaces = make([]WorkspaceReference, len(src.Spec.WaitForWorkspaces))
	for i, w := range src.Spec.WaitForWorkspaces {
		dst.Spec.WaitForWorkspaces[i] = WorkspaceReference{Name: w.Name, Namespace: w.Namespace}
	}
	dst.Spec.RetainWorkspaceOnDelete = src.Spec.RetainWorkspaceOnDelete

	// Status
	dst.Status.WorkspaceName = src.Status.WorkspaceName
	dst.Status.Applied = src.Status.Applied
	dst.Status.LastAppliedTime = src.Status.LastAppliedTime
	dst.Status.Conditions = src.Status.Conditions
	return nil
}

