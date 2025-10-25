package v1beta1

import (
	v1beta2 "github.com/appthrust/capt/api/v1beta2"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

// ConvertTo converts this v1beta1 WorkspaceTemplate to the hub version (v1beta2).
func (src *WorkspaceTemplate) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1beta2.WorkspaceTemplate)

	dst.ObjectMeta = src.ObjectMeta

	// Spec
	dst.Spec.Template = v1beta2.WorkspaceTemplateDefinition{
		Metadata: nil,
		Spec:     src.Spec.Template.Spec,
	}
	if src.Spec.Template.Metadata != nil {
		dst.Spec.Template.Metadata = &v1beta2.WorkspaceTemplateMetadata{
			Description: src.Spec.Template.Metadata.Description,
			Version:     src.Spec.Template.Metadata.Version,
			Tags:        src.Spec.Template.Metadata.Tags,
		}
	}
	dst.Spec.WriteConnectionSecretToRef = src.Spec.WriteConnectionSecretToRef

	// Status
	dst.Status.WorkspaceName = src.Status.WorkspaceName
	// Conditions type differs (xpv1.Condition in both), direct copy
	dst.Status.Conditions = src.Status.Conditions
	return nil
}

// ConvertFrom converts from the hub version (v1beta2) to this version.
func (dst *WorkspaceTemplate) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1beta2.WorkspaceTemplate)

	dst.ObjectMeta = src.ObjectMeta

	// Spec
	dst.Spec.Template = WorkspaceTemplateDefinition{
		Spec: src.Spec.Template.Spec,
	}
	if src.Spec.Template.Metadata != nil {
		dst.Spec.Template.Metadata = &WorkspaceTemplateMetadata{
			Description: src.Spec.Template.Metadata.Description,
			Version:     src.Spec.Template.Metadata.Version,
			Tags:        src.Spec.Template.Metadata.Tags,
		}
	} else {
		dst.Spec.Template.Metadata = nil
	}
	dst.Spec.WriteConnectionSecretToRef = src.Spec.WriteConnectionSecretToRef

	// Status
	dst.Status.WorkspaceName = src.Status.WorkspaceName
	dst.Status.Conditions = src.Status.Conditions
	return nil
}
