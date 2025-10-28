package v1beta1

import (
	v1beta2 "github.com/appthrust/capt/api/v1beta2"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *CaptMachineDeployment) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1beta2.CaptMachineDeployment)
	dst.ObjectMeta = src.ObjectMeta

	dst.Spec.Replicas = src.Spec.Replicas
	dst.Spec.Selector = src.Spec.Selector
	dst.Spec.MinReadySeconds = src.Spec.MinReadySeconds
	dst.Spec.RevisionHistoryLimit = src.Spec.RevisionHistoryLimit
	dst.Spec.Paused = src.Spec.Paused
	dst.Spec.ProgressDeadlineSeconds = src.Spec.ProgressDeadlineSeconds
	if src.Spec.Strategy != nil {
		dst.Spec.Strategy = &v1beta2.MachineDeploymentStrategy{Type: src.Spec.Strategy.Type}
		if src.Spec.Strategy.RollingUpdate != nil {
			dst.Spec.Strategy.RollingUpdate = &v1beta2.MachineRollingUpdateDeployment{
				MaxUnavailable: src.Spec.Strategy.RollingUpdate.MaxUnavailable,
				MaxSurge:       src.Spec.Strategy.RollingUpdate.MaxSurge,
			}
		}
	}
	dst.Spec.Template.ObjectMeta = src.Spec.Template.ObjectMeta
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

func (dst *CaptMachineDeployment) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1beta2.CaptMachineDeployment)
	dst.ObjectMeta = src.ObjectMeta

	dst.Spec.Replicas = src.Spec.Replicas
	dst.Spec.Selector = src.Spec.Selector
	dst.Spec.MinReadySeconds = src.Spec.MinReadySeconds
	dst.Spec.RevisionHistoryLimit = src.Spec.RevisionHistoryLimit
	dst.Spec.Paused = src.Spec.Paused
	dst.Spec.ProgressDeadlineSeconds = src.Spec.ProgressDeadlineSeconds
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
	dst.Spec.Template.ObjectMeta = src.Spec.Template.ObjectMeta
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
