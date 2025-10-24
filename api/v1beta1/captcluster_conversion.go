package v1beta1

import (
	v1beta2 "github.com/appthrust/capt/api/v1beta2"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

// ConvertTo converts this v1beta1 CAPTCluster to the hub version (v1beta2).
func (src *CAPTCluster) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1beta2.CAPTCluster)

	dst.ObjectMeta = src.ObjectMeta

	// Spec
	dst.Spec.Region = src.Spec.Region
	if src.Spec.VPCTemplateRef != nil {
		dst.Spec.VPCTemplateRef = &v1beta2.WorkspaceTemplateReference{
			Name:      src.Spec.VPCTemplateRef.Name,
			Namespace: src.Spec.VPCTemplateRef.Namespace,
		}
	} else {
		dst.Spec.VPCTemplateRef = nil
	}
	dst.Spec.ExistingVPCID = src.Spec.ExistingVPCID
	dst.Spec.RetainVPCOnDelete = src.Spec.RetainVPCOnDelete
	if src.Spec.VPCConfig != nil {
		dst.Spec.VPCConfig = &v1beta2.VPCConfig{Name: src.Spec.VPCConfig.Name}
	} else {
		dst.Spec.VPCConfig = nil
	}
	dst.Spec.WorkspaceTemplateApplyName = src.Spec.WorkspaceTemplateApplyName

	// Status
	dst.Status.VPCWorkspaceName = src.Status.VPCWorkspaceName
	dst.Status.VPCID = src.Status.VPCID
	dst.Status.Ready = src.Status.Ready
	dst.Status.FailureReason = src.Status.FailureReason
	dst.Status.FailureMessage = src.Status.FailureMessage
	dst.Status.FailureDomains = src.Status.FailureDomains
	dst.Status.Conditions = src.Status.Conditions
	if src.Status.WorkspaceTemplateStatus != nil {
		dst.Status.WorkspaceTemplateStatus = &v1beta2.CAPTClusterWorkspaceStatus{
			Ready:               src.Status.WorkspaceTemplateStatus.Ready,
			LastAppliedRevision: src.Status.WorkspaceTemplateStatus.LastAppliedRevision,
			LastFailedRevision:  src.Status.WorkspaceTemplateStatus.LastFailedRevision,
			LastFailureMessage:  src.Status.WorkspaceTemplateStatus.LastFailureMessage,
			WorkspaceName:       src.Status.WorkspaceTemplateStatus.WorkspaceName,
			LastAppliedTime:     src.Status.WorkspaceTemplateStatus.LastAppliedTime,
		}
	} else {
		dst.Status.WorkspaceTemplateStatus = nil
	}

	return nil
}

// ConvertFrom converts from the hub version (v1beta2) to this version.
func (dst *CAPTCluster) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1beta2.CAPTCluster)

	dst.ObjectMeta = src.ObjectMeta

	// Spec
	dst.Spec.Region = src.Spec.Region
	if src.Spec.VPCTemplateRef != nil {
		dst.Spec.VPCTemplateRef = &WorkspaceTemplateReference{
			Name:      src.Spec.VPCTemplateRef.Name,
			Namespace: src.Spec.VPCTemplateRef.Namespace,
		}
	} else {
		dst.Spec.VPCTemplateRef = nil
	}
	dst.Spec.ExistingVPCID = src.Spec.ExistingVPCID
	dst.Spec.RetainVPCOnDelete = src.Spec.RetainVPCOnDelete
	if src.Spec.VPCConfig != nil {
		dst.Spec.VPCConfig = &VPCConfig{Name: src.Spec.VPCConfig.Name}
	} else {
		dst.Spec.VPCConfig = nil
	}
	dst.Spec.WorkspaceTemplateApplyName = src.Spec.WorkspaceTemplateApplyName

	// Status
	dst.Status.VPCWorkspaceName = src.Status.VPCWorkspaceName
	dst.Status.VPCID = src.Status.VPCID
	dst.Status.Ready = src.Status.Ready
	dst.Status.FailureReason = src.Status.FailureReason
	dst.Status.FailureMessage = src.Status.FailureMessage
	dst.Status.FailureDomains = src.Status.FailureDomains
	dst.Status.Conditions = src.Status.Conditions
	if src.Status.WorkspaceTemplateStatus != nil {
		dst.Status.WorkspaceTemplateStatus = &CAPTClusterWorkspaceStatus{
			Ready:               src.Status.WorkspaceTemplateStatus.Ready,
			LastAppliedRevision: src.Status.WorkspaceTemplateStatus.LastAppliedRevision,
			LastFailedRevision:  src.Status.WorkspaceTemplateStatus.LastFailedRevision,
			LastFailureMessage:  src.Status.WorkspaceTemplateStatus.LastFailureMessage,
			WorkspaceName:       src.Status.WorkspaceTemplateStatus.WorkspaceName,
			LastAppliedTime:     src.Status.WorkspaceTemplateStatus.LastAppliedTime,
		}
	} else {
		dst.Status.WorkspaceTemplateStatus = nil
	}

	return nil
}
