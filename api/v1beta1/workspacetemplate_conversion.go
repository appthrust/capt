package v1beta1

import (
	v1beta2 "github.com/appthrust/capt/api/v1beta2"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *WorkspaceTemplate) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1beta2.WorkspaceTemplate)
	dst.ObjectMeta = src.ObjectMeta

	dst.Spec.Template = v1beta2.WorkspaceTemplateDefinition{}
	if src.Spec.Template.Metadata != nil {
		dst.Spec.Template.Metadata = &v1beta2.WorkspaceTemplateMetadata{
			Description: src.Spec.Template.Metadata.Description,
			Version:     src.Spec.Template.Metadata.Version,
			Tags:        src.Spec.Template.Metadata.Tags,
		}
	}
	dst.Spec.Template.Spec = src.Spec.Template.Spec
	dst.Spec.WriteConnectionSecretToRef = src.Spec.WriteConnectionSecretToRef

	dst.Status.WorkspaceName = src.Status.WorkspaceName
	dst.Status.Conditions = src.Status.Conditions
	return nil
}

func (dst *WorkspaceTemplate) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1beta2.WorkspaceTemplate)
	dst.ObjectMeta = src.ObjectMeta

	dst.Spec.Template = WorkspaceTemplateDefinition{}
	if src.Spec.Template.Metadata != nil {
		dst.Spec.Template.Metadata = &WorkspaceTemplateMetadata{
			Description: src.Spec.Template.Metadata.Description,
			Version:     src.Spec.Template.Metadata.Version,
			Tags:        src.Spec.Template.Metadata.Tags,
		}
	}
	dst.Spec.Template.Spec = src.Spec.Template.Spec
	dst.Spec.WriteConnectionSecretToRef = src.Spec.WriteConnectionSecretToRef

	dst.Status.WorkspaceName = src.Status.WorkspaceName
	dst.Status.Conditions = src.Status.Conditions
	return nil
}
