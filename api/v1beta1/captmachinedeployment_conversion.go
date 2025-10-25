package v1beta1

import (
	v1beta2 "github.com/appthrust/capt/api/v1beta2"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

// ConvertTo converts this v1beta1 CaptMachineDeployment to the hub version (v1beta2).
func (src *CaptMachineDeployment) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1beta2.CaptMachineDeployment)

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
	dst.Spec.Template = v1beta2.CaptMachineTemplateSpec{
		ObjectMeta: src.Spec.Template.ObjectMeta,
		Spec: v1beta2.CaptMachineSpec{
			NodeGroupRef: v1beta2.NodeGroupReference{
				Name:      src.Spec.Template.Spec.NodeGroupRef.Name,
				Namespace: src.Spec.Template.Spec.NodeGroupRef.Namespace,
			},
			WorkspaceTemplateRef: v1beta2.WorkspaceTemplateReference{
				Name:      src.Spec.Template.Spec.WorkspaceTemplateRef.Name,
				Namespace: src.Spec.Template.Spec.WorkspaceTemplateRef.Namespace,
			},
			InstanceType: src.Spec.Template.Spec.InstanceType,
			Labels:       src.Spec.Template.Spec.Labels,
			Tags:         src.Spec.Template.Spec.Tags,
		},
	}
	if src.Spec.Strategy != nil {
		dst.Spec.Strategy = &v1beta2.MachineDeploymentStrategy{Type: src.Spec.Strategy.Type}
		if src.Spec.Strategy.RollingUpdate != nil {
			dst.Spec.Strategy.RollingUpdate = &v1beta2.MachineRollingUpdateDeployment{
				MaxUnavailable: src.Spec.Strategy.RollingUpdate.MaxUnavailable,
				MaxSurge:       src.Spec.Strategy.RollingUpdate.MaxSurge,
			}
		}
	} else {
		dst.Spec.Strategy = nil
	}
	dst.Spec.MinReadySeconds = src.Spec.MinReadySeconds
	dst.Spec.RevisionHistoryLimit = src.Spec.RevisionHistoryLimit
	dst.Spec.Paused = src.Spec.Paused
	dst.Spec.ProgressDeadlineSeconds = src.Spec.ProgressDeadlineSeconds

	// Status
	dst.Status.ObservedGeneration = src.Status.ObservedGeneration
	dst.Status.UpdatedReplicas = src.Status.UpdatedReplicas
	dst.Status.Replicas = src.Status.Replicas
	dst.Status.AvailableReplicas = src.Status.AvailableReplicas
	dst.Status.UnavailableReplicas = src.Status.UnavailableReplicas
	dst.Status.Conditions = src.Status.Conditions
	dst.Status.CollisionCount = src.Status.CollisionCount
	dst.Status.FailureReason = src.Status.FailureReason
	dst.Status.FailureMessage = src.Status.FailureMessage
	return nil
}

// ConvertFrom converts from the hub version (v1beta2) to this version.
func (dst *CaptMachineDeployment) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1beta2.CaptMachineDeployment)

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
	dst.Spec.Template = CaptMachineTemplateSpec{
		ObjectMeta: src.Spec.Template.ObjectMeta,
		Spec: CaptMachineSpec{
			NodeGroupRef: NodeGroupReference{
				Name:      src.Spec.Template.Spec.NodeGroupRef.Name,
				Namespace: src.Spec.Template.Spec.NodeGroupRef.Namespace,
			},
			WorkspaceTemplateRef: WorkspaceTemplateReference{
				Name:      src.Spec.Template.Spec.WorkspaceTemplateRef.Name,
				Namespace: src.Spec.Template.Spec.WorkspaceTemplateRef.Namespace,
			},
			InstanceType: src.Spec.Template.Spec.InstanceType,
			Labels:       src.Spec.Template.Spec.Labels,
			Tags:         src.Spec.Template.Spec.Tags,
		},
	}
	if src.Spec.Strategy != nil {
		dst.Spec.Strategy = &MachineDeploymentStrategy{Type: src.Spec.Strategy.Type}
		if src.Spec.Strategy.RollingUpdate != nil {
			dst.Spec.Strategy.RollingUpdate = &MachineRollingUpdateDeployment{
				MaxUnavailable: src.Spec.Strategy.RollingUpdate.MaxUnavailable,
				MaxSurge:       src.Spec.Strategy.RollingUpdate.MaxSurge,
			}
		}
	} else {
		dst.Spec.Strategy = nil
	}
	dst.Spec.MinReadySeconds = src.Spec.MinReadySeconds
	dst.Spec.RevisionHistoryLimit = src.Spec.RevisionHistoryLimit
	dst.Spec.Paused = src.Spec.Paused
	dst.Spec.ProgressDeadlineSeconds = src.Spec.ProgressDeadlineSeconds

	// Status
	dst.Status.ObservedGeneration = src.Status.ObservedGeneration
	dst.Status.UpdatedReplicas = src.Status.UpdatedReplicas
	dst.Status.Replicas = src.Status.Replicas
	dst.Status.AvailableReplicas = src.Status.AvailableReplicas
	dst.Status.UnavailableReplicas = src.Status.UnavailableReplicas
	dst.Status.Conditions = src.Status.Conditions
	dst.Status.CollisionCount = src.Status.CollisionCount
	dst.Status.FailureReason = src.Status.FailureReason
	dst.Status.FailureMessage = src.Status.FailureMessage
	return nil
}

