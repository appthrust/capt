package v1beta1

import (
	v1beta2 "github.com/appthrust/capt/api/v1beta2"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

// ConvertTo converts this v1beta1 CAPTClusterTemplate to the hub version (v1beta2).
func (src *CAPTClusterTemplate) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1beta2.CAPTClusterTemplate)

	dst.ObjectMeta = src.ObjectMeta

	// Template metadata
	dst.Spec.Template.ObjectMeta = src.Spec.Template.ObjectMeta

	// Spec mapping (mirror CAPTCluster conversion logic for nested Spec)
	dst.Spec.Template.Spec.Region = src.Spec.Template.Spec.Region
	if src.Spec.Template.Spec.VPCTemplateRef != nil {
		dst.Spec.Template.Spec.VPCTemplateRef = &v1beta2.WorkspaceTemplateReference{
			Name:      src.Spec.Template.Spec.VPCTemplateRef.Name,
			Namespace: src.Spec.Template.Spec.VPCTemplateRef.Namespace,
		}
	} else {
		dst.Spec.Template.Spec.VPCTemplateRef = nil
	}
	dst.Spec.Template.Spec.ExistingVPCID = src.Spec.Template.Spec.ExistingVPCID
	dst.Spec.Template.Spec.RetainVPCOnDelete = src.Spec.Template.Spec.RetainVPCOnDelete
	if src.Spec.Template.Spec.VPCConfig != nil {
		dst.Spec.Template.Spec.VPCConfig = &v1beta2.VPCConfig{Name: src.Spec.Template.Spec.VPCConfig.Name}
	} else {
		dst.Spec.Template.Spec.VPCConfig = nil
	}
	dst.Spec.Template.Spec.WorkspaceTemplateApplyName = src.Spec.Template.Spec.WorkspaceTemplateApplyName

	return nil
}

// ConvertFrom converts from the hub version (v1beta2) to this version.
func (dst *CAPTClusterTemplate) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1beta2.CAPTClusterTemplate)

	dst.ObjectMeta = src.ObjectMeta

	// Template metadata
	dst.Spec.Template.ObjectMeta = src.Spec.Template.ObjectMeta

	// Spec mapping (mirror CAPTCluster conversion logic for nested Spec)
	dst.Spec.Template.Spec.Region = src.Spec.Template.Spec.Region
	if src.Spec.Template.Spec.VPCTemplateRef != nil {
		dst.Spec.Template.Spec.VPCTemplateRef = &WorkspaceTemplateReference{
			Name:      src.Spec.Template.Spec.VPCTemplateRef.Name,
			Namespace: src.Spec.Template.Spec.VPCTemplateRef.Namespace,
		}
	} else {
		dst.Spec.Template.Spec.VPCTemplateRef = nil
	}
	dst.Spec.Template.Spec.ExistingVPCID = src.Spec.Template.Spec.ExistingVPCID
	dst.Spec.Template.Spec.RetainVPCOnDelete = src.Spec.Template.Spec.RetainVPCOnDelete
	if src.Spec.Template.Spec.VPCConfig != nil {
		dst.Spec.Template.Spec.VPCConfig = &VPCConfig{Name: src.Spec.Template.Spec.VPCConfig.Name}
	} else {
		dst.Spec.Template.Spec.VPCConfig = nil
	}
	dst.Spec.Template.Spec.WorkspaceTemplateApplyName = src.Spec.Template.Spec.WorkspaceTemplateApplyName

	return nil
}
