package v1beta1

import (
	v1beta2 "github.com/appthrust/capt/api/controlplane/v1beta2"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

// ConvertTo converts this v1beta1 CAPTControlPlaneTemplate to the hub version (v1beta2).
func (src *CAPTControlPlaneTemplate) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1beta2.CAPTControlPlaneTemplate)

	dst.ObjectMeta = src.ObjectMeta

	// Template metadata
	dst.Spec.Template.ObjectMeta = src.Spec.Template.ObjectMeta

	// Spec mapping (mirror CAPTControlPlane conversion logic for nested Spec)
	dst.Spec.Template.Spec.Version = src.Spec.Template.Spec.Version
	dst.Spec.Template.Spec.WorkspaceTemplateRef = v1beta2.WorkspaceTemplateReference{
		Name:      src.Spec.Template.Spec.WorkspaceTemplateRef.Name,
		Namespace: src.Spec.Template.Spec.WorkspaceTemplateRef.Namespace,
	}
	if src.Spec.Template.Spec.ControlPlaneConfig != nil {
		dst.Spec.Template.Spec.ControlPlaneConfig = &v1beta2.ControlPlaneConfig{}
		dst.Spec.Template.Spec.ControlPlaneConfig.Region = src.Spec.Template.Spec.ControlPlaneConfig.Region
		if src.Spec.Template.Spec.ControlPlaneConfig.EndpointAccess != nil {
			dst.Spec.Template.Spec.ControlPlaneConfig.EndpointAccess = &v1beta2.EndpointAccess{
				Public:      src.Spec.Template.Spec.ControlPlaneConfig.EndpointAccess.Public,
				Private:     src.Spec.Template.Spec.ControlPlaneConfig.EndpointAccess.Private,
				PublicCIDRs: append([]string(nil), src.Spec.Template.Spec.ControlPlaneConfig.EndpointAccess.PublicCIDRs...),
			}
		}
		if len(src.Spec.Template.Spec.ControlPlaneConfig.Addons) > 0 {
			dst.Spec.Template.Spec.ControlPlaneConfig.Addons = make([]v1beta2.Addon, len(src.Spec.Template.Spec.ControlPlaneConfig.Addons))
			for i, a := range src.Spec.Template.Spec.ControlPlaneConfig.Addons {
				dst.Spec.Template.Spec.ControlPlaneConfig.Addons[i] = v1beta2.Addon{
					Name:                a.Name,
					Version:             a.Version,
					ConfigurationValues: a.ConfigurationValues,
				}
			}
		}
		if src.Spec.Template.Spec.ControlPlaneConfig.Timeouts != nil {
			dst.Spec.Template.Spec.ControlPlaneConfig.Timeouts = &v1beta2.TimeoutConfig{
				ControlPlaneTimeout: src.Spec.Template.Spec.ControlPlaneConfig.Timeouts.ControlPlaneTimeout,
				VPCReadyTimeout:     src.Spec.Template.Spec.ControlPlaneConfig.Timeouts.VPCReadyTimeout,
			}
		}
	}
	dst.Spec.Template.Spec.AdditionalTags = src.Spec.Template.Spec.AdditionalTags
	dst.Spec.Template.Spec.ControlPlaneEndpoint = src.Spec.Template.Spec.ControlPlaneEndpoint
	dst.Spec.Template.Spec.WorkspaceTemplateApplyName = src.Spec.Template.Spec.WorkspaceTemplateApplyName

	return nil
}

// ConvertFrom converts from the hub version (v1beta2) to this version.
func (dst *CAPTControlPlaneTemplate) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1beta2.CAPTControlPlaneTemplate)

	dst.ObjectMeta = src.ObjectMeta

	// Template metadata
	dst.Spec.Template.ObjectMeta = src.Spec.Template.ObjectMeta

	// Spec mapping (mirror CAPTControlPlane conversion logic for nested Spec)
	dst.Spec.Template.Spec.Version = src.Spec.Template.Spec.Version
	dst.Spec.Template.Spec.WorkspaceTemplateRef = WorkspaceTemplateReference{
		Name:      src.Spec.Template.Spec.WorkspaceTemplateRef.Name,
		Namespace: src.Spec.Template.Spec.WorkspaceTemplateRef.Namespace,
	}
	if src.Spec.Template.Spec.ControlPlaneConfig != nil {
		dst.Spec.Template.Spec.ControlPlaneConfig = &ControlPlaneConfig{}
		dst.Spec.Template.Spec.ControlPlaneConfig.Region = src.Spec.Template.Spec.ControlPlaneConfig.Region
		if src.Spec.Template.Spec.ControlPlaneConfig.EndpointAccess != nil {
			dst.Spec.Template.Spec.ControlPlaneConfig.EndpointAccess = &EndpointAccess{
				Public:      src.Spec.Template.Spec.ControlPlaneConfig.EndpointAccess.Public,
				Private:     src.Spec.Template.Spec.ControlPlaneConfig.EndpointAccess.Private,
				PublicCIDRs: append([]string(nil), src.Spec.Template.Spec.ControlPlaneConfig.EndpointAccess.PublicCIDRs...),
			}
		}
		if len(src.Spec.Template.Spec.ControlPlaneConfig.Addons) > 0 {
			dst.Spec.Template.Spec.ControlPlaneConfig.Addons = make([]Addon, len(src.Spec.Template.Spec.ControlPlaneConfig.Addons))
			for i, a := range src.Spec.Template.Spec.ControlPlaneConfig.Addons {
				dst.Spec.Template.Spec.ControlPlaneConfig.Addons[i] = Addon{
					Name:                a.Name,
					Version:             a.Version,
					ConfigurationValues: a.ConfigurationValues,
				}
			}
		}
		if src.Spec.Template.Spec.ControlPlaneConfig.Timeouts != nil {
			dst.Spec.Template.Spec.ControlPlaneConfig.Timeouts = &TimeoutConfig{
				ControlPlaneTimeout: src.Spec.Template.Spec.ControlPlaneConfig.Timeouts.ControlPlaneTimeout,
				VPCReadyTimeout:     src.Spec.Template.Spec.ControlPlaneConfig.Timeouts.VPCReadyTimeout,
			}
		}
	}
	dst.Spec.Template.Spec.AdditionalTags = src.Spec.Template.Spec.AdditionalTags
	dst.Spec.Template.Spec.ControlPlaneEndpoint = src.Spec.Template.Spec.ControlPlaneEndpoint
	dst.Spec.Template.Spec.WorkspaceTemplateApplyName = src.Spec.Template.Spec.WorkspaceTemplateApplyName

	return nil
}
