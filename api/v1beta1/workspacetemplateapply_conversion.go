package v1beta1

import (
	v1beta2 "github.com/appthrust/capt/api/v1beta2"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *WorkspaceTemplateApply) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1beta2.WorkspaceTemplateApply)
	dst.ObjectMeta = src.ObjectMeta

	dst.Spec.TemplateRef = v1beta2.WorkspaceTemplateReference{
		Name:      src.Spec.TemplateRef.Name,
		Namespace: src.Spec.TemplateRef.Namespace,
	}
	dst.Spec.WriteConnectionSecretToRef = src.Spec.WriteConnectionSecretToRef
	dst.Spec.Variables = src.Spec.Variables
	dst.Spec.WaitForSecrets = src.Spec.WaitForSecrets
	if len(src.Spec.WaitForWorkspaces) > 0 {
		dst.Spec.WaitForWorkspaces = make([]v1beta2.WorkspaceReference, len(src.Spec.WaitForWorkspaces))
		for i, w := range src.Spec.WaitForWorkspaces {
			dst.Spec.WaitForWorkspaces[i] = v1beta2.WorkspaceReference{Name: w.Name, Namespace: w.Namespace}
		}
	}
	dst.Spec.RetainWorkspaceOnDelete = src.Spec.RetainWorkspaceOnDelete

	dst.Status.WorkspaceName = src.Status.WorkspaceName
	dst.Status.Applied = src.Status.Applied
	dst.Status.LastAppliedTime = src.Status.LastAppliedTime
	dst.Status.Conditions = src.Status.Conditions
	return nil
}

func (dst *WorkspaceTemplateApply) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1beta2.WorkspaceTemplateApply)
	dst.ObjectMeta = src.ObjectMeta

	dst.Spec.TemplateRef = WorkspaceTemplateReference{
		Name:      src.Spec.TemplateRef.Name,
		Namespace: src.Spec.TemplateRef.Namespace,
	}
	dst.Spec.WriteConnectionSecretToRef = src.Spec.WriteConnectionSecretToRef
	dst.Spec.Variables = src.Spec.Variables
	dst.Spec.WaitForSecrets = src.Spec.WaitForSecrets
	if len(src.Spec.WaitForWorkspaces) > 0 {
		dst.Spec.WaitForWorkspaces = make([]WorkspaceReference, len(src.Spec.WaitForWorkspaces))
		for i, w := range src.Spec.WaitForWorkspaces {
			dst.Spec.WaitForWorkspaces[i] = WorkspaceReference{Name: w.Name, Namespace: w.Namespace}
		}
	}
	dst.Spec.RetainWorkspaceOnDelete = src.Spec.RetainWorkspaceOnDelete

	dst.Status.WorkspaceName = src.Status.WorkspaceName
	dst.Status.Applied = src.Status.Applied
	dst.Status.LastAppliedTime = src.Status.LastAppliedTime
	dst.Status.Conditions = src.Status.Conditions
	return nil
}
