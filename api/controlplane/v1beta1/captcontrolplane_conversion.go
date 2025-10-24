package v1beta1

import (
	v1beta2 "github.com/appthrust/capt/api/controlplane/v1beta2"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

// ConvertTo converts this v1beta1 CAPTControlPlane to the hub version (v1beta2).
func (src *CAPTControlPlane) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1beta2.CAPTControlPlane)
	dst.ObjectMeta = src.ObjectMeta

	// Spec
	dst.Spec.Version = src.Spec.Version
	dst.Spec.WorkspaceTemplateRef = v1beta2.WorkspaceTemplateReference{
		Name:      src.Spec.WorkspaceTemplateRef.Name,
		Namespace: src.Spec.WorkspaceTemplateRef.Namespace,
	}
	if src.Spec.ControlPlaneConfig != nil {
		dst.Spec.ControlPlaneConfig = &v1beta2.ControlPlaneConfig{}
		dst.Spec.ControlPlaneConfig.Region = src.Spec.ControlPlaneConfig.Region
		if src.Spec.ControlPlaneConfig.EndpointAccess != nil {
			dst.Spec.ControlPlaneConfig.EndpointAccess = &v1beta2.EndpointAccess{
				Public:      src.Spec.ControlPlaneConfig.EndpointAccess.Public,
				Private:     src.Spec.ControlPlaneConfig.EndpointAccess.Private,
				PublicCIDRs: append([]string(nil), src.Spec.ControlPlaneConfig.EndpointAccess.PublicCIDRs...),
			}
		}
		if len(src.Spec.ControlPlaneConfig.Addons) > 0 {
			dst.Spec.ControlPlaneConfig.Addons = make([]v1beta2.Addon, len(src.Spec.ControlPlaneConfig.Addons))
			for i, a := range src.Spec.ControlPlaneConfig.Addons {
				dst.Spec.ControlPlaneConfig.Addons[i] = v1beta2.Addon{
					Name:                a.Name,
					Version:             a.Version,
					ConfigurationValues: a.ConfigurationValues,
				}
			}
		}
		if src.Spec.ControlPlaneConfig.Timeouts != nil {
			dst.Spec.ControlPlaneConfig.Timeouts = &v1beta2.TimeoutConfig{
				ControlPlaneTimeout: src.Spec.ControlPlaneConfig.Timeouts.ControlPlaneTimeout,
				VPCReadyTimeout:     src.Spec.ControlPlaneConfig.Timeouts.VPCReadyTimeout,
			}
		}
	}
	dst.Spec.AdditionalTags = src.Spec.AdditionalTags
	dst.Spec.ControlPlaneEndpoint = src.Spec.ControlPlaneEndpoint
	dst.Spec.WorkspaceTemplateApplyName = src.Spec.WorkspaceTemplateApplyName

	// Status
	dst.Status.Ready = src.Status.Ready
	dst.Status.Initialized = src.Status.Initialized
	dst.Status.SecretsReady = src.Status.SecretsReady
	if src.Status.WorkspaceTemplateStatus != nil {
		dst.Status.WorkspaceTemplateStatus = &v1beta2.WorkspaceTemplateStatus{
			Ready:               src.Status.WorkspaceTemplateStatus.Ready,
			State:               src.Status.WorkspaceTemplateStatus.State,
			LastAppliedRevision: src.Status.WorkspaceTemplateStatus.LastAppliedRevision,
			Outputs:             src.Status.WorkspaceTemplateStatus.Outputs,
			LastFailedRevision:  src.Status.WorkspaceTemplateStatus.LastFailedRevision,
			LastFailureMessage:  src.Status.WorkspaceTemplateStatus.LastFailureMessage,
		}
	}
	if src.Status.WorkspaceStatus != nil {
		dst.Status.WorkspaceStatus = &v1beta2.WorkspaceStatus{
			Ready: src.Status.WorkspaceStatus.Ready,
			State: src.Status.WorkspaceStatus.State,
		}
		if src.Status.WorkspaceStatus.AtProvider != nil {
			dst.Status.WorkspaceStatus.AtProvider = &runtime.RawExtension{Raw: src.Status.WorkspaceStatus.AtProvider.Raw}
		}
	}
	dst.Status.FailureReason = src.Status.FailureReason
	dst.Status.FailureMessage = src.Status.FailureMessage
	dst.Status.Phase = src.Status.Phase
	dst.Status.Conditions = src.Status.Conditions
	return nil
}

// ConvertFrom converts from the hub version (v1beta2) to this version.
func (dst *CAPTControlPlane) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1beta2.CAPTControlPlane)
	dst.ObjectMeta = src.ObjectMeta

	// Spec
	dst.Spec.Version = src.Spec.Version
	dst.Spec.WorkspaceTemplateRef = WorkspaceTemplateReference{
		Name:      src.Spec.WorkspaceTemplateRef.Name,
		Namespace: src.Spec.WorkspaceTemplateRef.Namespace,
	}
	if src.Spec.ControlPlaneConfig != nil {
		dst.Spec.ControlPlaneConfig = &ControlPlaneConfig{}
		dst.Spec.ControlPlaneConfig.Region = src.Spec.ControlPlaneConfig.Region
		if src.Spec.ControlPlaneConfig.EndpointAccess != nil {
			dst.Spec.ControlPlaneConfig.EndpointAccess = &EndpointAccess{
				Public:      src.Spec.ControlPlaneConfig.EndpointAccess.Public,
				Private:     src.Spec.ControlPlaneConfig.EndpointAccess.Private,
				PublicCIDRs: append([]string(nil), src.Spec.ControlPlaneConfig.EndpointAccess.PublicCIDRs...),
			}
		}
		if len(src.Spec.ControlPlaneConfig.Addons) > 0 {
			dst.Spec.ControlPlaneConfig.Addons = make([]Addon, len(src.Spec.ControlPlaneConfig.Addons))
			for i, a := range src.Spec.ControlPlaneConfig.Addons {
				dst.Spec.ControlPlaneConfig.Addons[i] = Addon{
					Name:                a.Name,
					Version:             a.Version,
					ConfigurationValues: a.ConfigurationValues,
				}
			}
		}
		if src.Spec.ControlPlaneConfig.Timeouts != nil {
			dst.Spec.ControlPlaneConfig.Timeouts = &TimeoutConfig{
				ControlPlaneTimeout: src.Spec.ControlPlaneConfig.Timeouts.ControlPlaneTimeout,
				VPCReadyTimeout:     src.Spec.ControlPlaneConfig.Timeouts.VPCReadyTimeout,
			}
		}
	}
	dst.Spec.AdditionalTags = src.Spec.AdditionalTags
	dst.Spec.ControlPlaneEndpoint = src.Spec.ControlPlaneEndpoint
	dst.Spec.WorkspaceTemplateApplyName = src.Spec.WorkspaceTemplateApplyName

	// Status
	dst.Status.Ready = src.Status.Ready
	dst.Status.Initialized = src.Status.Initialized
	dst.Status.SecretsReady = src.Status.SecretsReady
	if src.Status.WorkspaceTemplateStatus != nil {
		dst.Status.WorkspaceTemplateStatus = &WorkspaceTemplateStatus{
			Ready:               src.Status.WorkspaceTemplateStatus.Ready,
			State:               src.Status.WorkspaceTemplateStatus.State,
			LastAppliedRevision: src.Status.WorkspaceTemplateStatus.LastAppliedRevision,
			Outputs:             src.Status.WorkspaceTemplateStatus.Outputs,
			LastFailedRevision:  src.Status.WorkspaceTemplateStatus.LastFailedRevision,
			LastFailureMessage:  src.Status.WorkspaceTemplateStatus.LastFailureMessage,
		}
	}
	if src.Status.WorkspaceStatus != nil {
		dst.Status.WorkspaceStatus = &WorkspaceStatus{
			Ready: src.Status.WorkspaceStatus.Ready,
			State: src.Status.WorkspaceStatus.State,
		}
		if src.Status.WorkspaceStatus.AtProvider != nil {
			dst.Status.WorkspaceStatus.AtProvider = &runtime.RawExtension{Raw: src.Status.WorkspaceStatus.AtProvider.Raw}
		}
	}
	dst.Status.FailureReason = src.Status.FailureReason
	dst.Status.FailureMessage = src.Status.FailureMessage
	dst.Status.Phase = src.Status.Phase
	dst.Status.Conditions = src.Status.Conditions
	return nil
}
