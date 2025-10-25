package v1beta1

import (
	v1beta2 "github.com/appthrust/capt/api/v1beta2"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

// ConvertTo converts this v1beta1 CaptMachineSet to the hub version (v1beta2).
func (src *CaptMachineSet) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1beta2.CaptMachineSet)

	// ObjectMeta
	dst.ObjectMeta = src.ObjectMeta

	// Spec
	if src.Spec.Replicas != nil {
		dst.Spec.Replicas = new(int32)
		*dst.Spec.Replicas = *src.Spec.Replicas
	} else {
		dst.Spec.Replicas = nil
	}
	if src.Spec.Selector != nil {
		cpy := *src.Spec.Selector
		dst.Spec.Selector = &cpy
	} else {
		dst.Spec.Selector = nil
	}
	dst.Spec.Template.ObjectMeta = src.Spec.Template.ObjectMeta
	// Machine spec mapping
	dst.Spec.Template.Spec.InstanceType = src.Spec.Template.Spec.InstanceType
	dst.Spec.Template.Spec.Labels = src.Spec.Template.Spec.Labels
	dst.Spec.Template.Spec.Tags = src.Spec.Template.Spec.Tags
	dst.Spec.Template.Spec.NodeGroupRef = v1beta2.NodeGroupReference{
		Name:      src.Spec.Template.Spec.NodeGroupRef.Name,
		Namespace: src.Spec.Template.Spec.NodeGroupRef.Namespace,
	}
	dst.Spec.Template.Spec.WorkspaceTemplateRef = v1beta2.WorkspaceTemplateReference{
		Name:      src.Spec.Template.Spec.WorkspaceTemplateRef.Name,
		Namespace: src.Spec.Template.Spec.WorkspaceTemplateRef.Namespace,
	}

	// Status
	dst.Status.Replicas = src.Status.Replicas
	dst.Status.FullyLabeledReplicas = src.Status.FullyLabeledReplicas
	dst.Status.ReadyReplicas = src.Status.ReadyReplicas
	dst.Status.AvailableReplicas = src.Status.AvailableReplicas
	dst.Status.ObservedGeneration = src.Status.ObservedGeneration
	dst.Status.Conditions = src.Status.Conditions
	dst.Status.FailureReason = src.Status.FailureReason
	dst.Status.FailureMessage = src.Status.FailureMessage
	return nil
}

// ConvertFrom converts from the hub version (v1beta2) to this version.
func (dst *CaptMachineSet) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1beta2.CaptMachineSet)

	// ObjectMeta
	dst.ObjectMeta = src.ObjectMeta

	// Spec
	if src.Spec.Replicas != nil {
		dst.Spec.Replicas = new(int32)
		*dst.Spec.Replicas = *src.Spec.Replicas
	} else {
		dst.Spec.Replicas = nil
	}
	if src.Spec.Selector != nil {
		cpy := *src.Spec.Selector
		dst.Spec.Selector = &cpy
	} else {
		dst.Spec.Selector = nil
	}
	dst.Spec.Template.ObjectMeta = src.Spec.Template.ObjectMeta
	// Machine spec mapping
	dst.Spec.Template.Spec.InstanceType = src.Spec.Template.Spec.InstanceType
	dst.Spec.Template.Spec.Labels = src.Spec.Template.Spec.Labels
	dst.Spec.Template.Spec.Tags = src.Spec.Template.Spec.Tags
	dst.Spec.Template.Spec.NodeGroupRef = NodeGroupReference{
		Name:      src.Spec.Template.Spec.NodeGroupRef.Name,
		Namespace: src.Spec.Template.Spec.NodeGroupRef.Namespace,
	}
	dst.Spec.Template.Spec.WorkspaceTemplateRef = WorkspaceTemplateReference{
		Name:      src.Spec.Template.Spec.WorkspaceTemplateRef.Name,
		Namespace: src.Spec.Template.Spec.WorkspaceTemplateRef.Namespace,
	}

	// Status
	dst.Status.Replicas = src.Status.Replicas
	dst.Status.FullyLabeledReplicas = src.Status.FullyLabeledReplicas
	dst.Status.ReadyReplicas = src.Status.ReadyReplicas
	dst.Status.AvailableReplicas = src.Status.AvailableReplicas
	dst.Status.ObservedGeneration = src.Status.ObservedGeneration
	dst.Status.Conditions = src.Status.Conditions
	dst.Status.FailureReason = src.Status.FailureReason
	dst.Status.FailureMessage = src.Status.FailureMessage
	return nil
}

