package v1beta1

import (
	v1beta2 "github.com/appthrust/capt/api/v1beta2"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

// ConvertTo converts this v1beta1 CaptMachine to the hub version (v1beta2).
func (src *CaptMachine) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1beta2.CaptMachine)

	dst.ObjectMeta = src.ObjectMeta

	// Spec
	dst.Spec.NodeGroupRef = v1beta2.NodeGroupReference{
		Name:      src.Spec.NodeGroupRef.Name,
		Namespace: src.Spec.NodeGroupRef.Namespace,
	}
	dst.Spec.WorkspaceTemplateRef = v1beta2.WorkspaceTemplateReference{
		Name:      src.Spec.WorkspaceTemplateRef.Name,
		Namespace: src.Spec.WorkspaceTemplateRef.Namespace,
	}
	dst.Spec.InstanceType = src.Spec.InstanceType
	dst.Spec.Labels = src.Spec.Labels
	dst.Spec.Tags = src.Spec.Tags

	// Status
	dst.Status.Ready = src.Status.Ready
	dst.Status.InstanceID = src.Status.InstanceID
	dst.Status.PrivateIP = src.Status.PrivateIP
	dst.Status.LastTransitionTime = src.Status.LastTransitionTime
	dst.Status.Conditions = src.Status.Conditions
	dst.Status.FailureReason = src.Status.FailureReason
	dst.Status.FailureMessage = src.Status.FailureMessage

	return nil
}

// ConvertFrom converts from the hub version (v1beta2) to this version.
func (dst *CaptMachine) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1beta2.CaptMachine)

	dst.ObjectMeta = src.ObjectMeta

	// Spec
	dst.Spec.NodeGroupRef = NodeGroupReference{
		Name:      src.Spec.NodeGroupRef.Name,
		Namespace: src.Spec.NodeGroupRef.Namespace,
	}
	dst.Spec.WorkspaceTemplateRef = WorkspaceTemplateReference{
		Name:      src.Spec.WorkspaceTemplateRef.Name,
		Namespace: src.Spec.WorkspaceTemplateRef.Namespace,
	}
	dst.Spec.InstanceType = src.Spec.InstanceType
	dst.Spec.Labels = src.Spec.Labels
	dst.Spec.Tags = src.Spec.Tags

	// Status
	dst.Status.Ready = src.Status.Ready
	dst.Status.InstanceID = src.Status.InstanceID
	dst.Status.PrivateIP = src.Status.PrivateIP
	dst.Status.LastTransitionTime = src.Status.LastTransitionTime
	dst.Status.Conditions = src.Status.Conditions
	dst.Status.FailureReason = src.Status.FailureReason
	dst.Status.FailureMessage = src.Status.FailureMessage

	return nil
}
