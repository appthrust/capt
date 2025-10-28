package v1beta1

import (
	v1beta2 "github.com/appthrust/capt/api/v1beta2"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

// ConvertTo converts this v1beta1 CaptMachineTemplate to the hub version (v1beta2).
func (src *CaptMachineTemplate) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1beta2.CaptMachineTemplate)

	dst.ObjectMeta = src.ObjectMeta

	// Map nested machine spec
	dst.Spec.Template.Spec.NodeGroupRef = v1beta2.NodeGroupReference{
		Name:      src.Spec.Template.Spec.NodeGroupRef.Name,
		Namespace: src.Spec.Template.Spec.NodeGroupRef.Namespace,
	}
	dst.Spec.Template.Spec.WorkspaceTemplateRef = v1beta2.WorkspaceTemplateReference{
		Name:      src.Spec.Template.Spec.WorkspaceTemplateRef.Name,
		Namespace: src.Spec.Template.Spec.WorkspaceTemplateRef.Namespace,
	}
	dst.Spec.Template.Spec.InstanceType = src.Spec.Template.Spec.InstanceType
	dst.Spec.Template.Spec.Labels = src.Spec.Template.Spec.Labels
	dst.Spec.Template.Spec.Tags = src.Spec.Template.Spec.Tags

	return nil
}

// ConvertFrom converts from the hub version (v1beta2) to this version.
func (dst *CaptMachineTemplate) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1beta2.CaptMachineTemplate)

	dst.ObjectMeta = src.ObjectMeta

	// Map nested machine spec
	dst.Spec.Template.Spec.NodeGroupRef = NodeGroupReference{
		Name:      src.Spec.Template.Spec.NodeGroupRef.Name,
		Namespace: src.Spec.Template.Spec.NodeGroupRef.Namespace,
	}
	dst.Spec.Template.Spec.WorkspaceTemplateRef = WorkspaceTemplateReference{
		Name:      src.Spec.Template.Spec.WorkspaceTemplateRef.Name,
		Namespace: src.Spec.Template.Spec.WorkspaceTemplateRef.Namespace,
	}
	dst.Spec.Template.Spec.InstanceType = src.Spec.Template.Spec.InstanceType
	dst.Spec.Template.Spec.Labels = src.Spec.Template.Spec.Labels
	dst.Spec.Template.Spec.Tags = src.Spec.Template.Spec.Tags

	return nil
}
