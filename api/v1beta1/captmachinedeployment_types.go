/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// CaptMachineDeploymentSpec defines the desired state of CaptMachineDeployment
type CaptMachineDeploymentSpec struct {
	// Replicas is the number of desired replicas.
	// This is a pointer to distinguish between explicit zero and unspecified.
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// Selector is a label query over machines that should match the replica count.
	// It must match the machine template's labels.
	// +optional
	Selector *metav1.LabelSelector `json:"selector,omitempty"`

	// Template is the object that describes the machine that will be created if
	// insufficient replicas are detected.
	Template CaptMachineTemplateSpec `json:"template"`

	// Strategy describes how to replace existing machines with new ones.
	// +optional
	Strategy *MachineDeploymentStrategy `json:"strategy,omitempty"`

	// MinReadySeconds is the minimum number of seconds for which a newly created machine should
	// be ready without any of its container crashing, for it to be considered available.
	// Defaults to 0 (machine will be considered available as soon as it is ready)
	// +optional
	MinReadySeconds int32 `json:"minReadySeconds,omitempty"`

	// RevisionHistoryLimit is the maximum number of old MachineSets to retain to allow rollback.
	// This is a pointer to distinguish between explicit zero and not specified.
	// Defaults to 10.
	// +optional
	RevisionHistoryLimit *int32 `json:"revisionHistoryLimit,omitempty"`

	// Paused indicates that the deployment is paused.
	// +optional
	Paused bool `json:"paused,omitempty"`

	// ProgressDeadlineSeconds is the maximum time in seconds for a deployment to
	// make progress before it is considered to be failed. The deployment controller will
	// continue to process failed deployments and a condition with a ProgressDeadlineExceeded
	// reason will be surfaced in the deployment status. Note that progress will not be
	// estimated during the time a deployment is paused. Defaults to 600s.
	// +optional
	ProgressDeadlineSeconds *int32 `json:"progressDeadlineSeconds,omitempty"`
}

// MachineDeploymentStrategy describes how to replace existing machines with new ones.
type MachineDeploymentStrategy struct {
	// Type of deployment. Can be "Recreate" or "RollingUpdate". Default is RollingUpdate.
	// +optional
	Type string `json:"type,omitempty"`

	// Rolling update config params. Present only if DeploymentStrategyType =
	// RollingUpdate.
	// +optional
	RollingUpdate *MachineRollingUpdateDeployment `json:"rollingUpdate,omitempty"`
}

// MachineRollingUpdateDeployment is the configuration for a rolling update.
type MachineRollingUpdateDeployment struct {
	// The maximum number of machines that can be unavailable during the update.
	// Value can be an absolute number (ex: 5) or a percentage of desired machines (ex: 10%).
	// Absolute number is calculated from percentage by rounding down.
	// Defaults to 0.
	// +optional
	MaxUnavailable *intstr.IntOrString `json:"maxUnavailable,omitempty"`

	// The maximum number of machines that can be scheduled above the desired number of
	// machines.
	// Value can be an absolute number (ex: 5) or a percentage of desired machines (ex: 10%).
	// Absolute number is calculated from percentage by rounding up.
	// Defaults to 1.
	// +optional
	MaxSurge *intstr.IntOrString `json:"maxSurge,omitempty"`
}

// CaptMachineDeploymentStatus defines the observed state of CaptMachineDeployment
type CaptMachineDeploymentStatus struct {
	// ObservedGeneration reflects the generation of the most recently observed MachineDeployment.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// The generation observed by the deployment controller.
	// +optional
	UpdatedReplicas int32 `json:"updatedReplicas,omitempty"`

	// Total number of non-terminated machines targeted by this deployment (their labels match the selector).
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// Total number of available machines (ready for at least minReadySeconds).
	// +optional
	AvailableReplicas int32 `json:"availableReplicas,omitempty"`

	// Total number of unavailable machines targeted by this deployment. This is the total number of
	// machines that are still required for the deployment to have 100% available capacity. They may
	// either be machines that are running but not yet available or machines that still have not been created.
	// +optional
	UnavailableReplicas int32 `json:"unavailableReplicas,omitempty"`

	// Represents the latest available observations of a deployment's current state.
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// Count of hash collisions for the MachineDeployment. The MachineDeployment controller
	// uses this field as a collision avoidance mechanism when it needs to create the name for the
	// newest MachineSet.
	// +optional
	CollisionCount *int32 `json:"collisionCount,omitempty"`

	// FailureReason indicates that there is a terminal problem reconciling the
	// state, and will be set to a token value suitable for programmatic
	// interpretation.
	// +optional
	FailureReason *string `json:"failureReason,omitempty"`

	// FailureMessage indicates that there is a terminal problem reconciling the
	// state, and will be set to a descriptive error message.
	// +optional
	FailureMessage *string `json:"failureMessage,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.replicas
//+kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
//+kubebuilder:printcolumn:name="Replicas",type="integer",JSONPath=".status.replicas"
//+kubebuilder:printcolumn:name="Updated",type="integer",JSONPath=".status.updatedReplicas"
//+kubebuilder:printcolumn:name="Available",type="integer",JSONPath=".status.availableReplicas"

// CaptMachineDeployment is the Schema for the captmachinedeployments API
type CaptMachineDeployment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CaptMachineDeploymentSpec   `json:"spec,omitempty"`
	Status CaptMachineDeploymentStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CaptMachineDeploymentList contains a list of CaptMachineDeployment
type CaptMachineDeploymentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CaptMachineDeployment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CaptMachineDeployment{}, &CaptMachineDeploymentList{})
}
