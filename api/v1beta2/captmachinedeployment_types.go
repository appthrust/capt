package v1beta2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type CaptMachineDeploymentSpec struct {
	Replicas                *int32                     `json:"replicas,omitempty"`
	Selector                *metav1.LabelSelector      `json:"selector,omitempty"`
	Template                CaptMachineTemplateSpec    `json:"template"`
	Strategy                *MachineDeploymentStrategy `json:"strategy,omitempty"`
	MinReadySeconds         int32                      `json:"minReadySeconds,omitempty"`
	RevisionHistoryLimit    *int32                     `json:"revisionHistoryLimit,omitempty"`
	Paused                  bool                       `json:"paused,omitempty"`
	ProgressDeadlineSeconds *int32                     `json:"progressDeadlineSeconds,omitempty"`
}

type MachineDeploymentStrategy struct {
	Type          string                          `json:"type,omitempty"`
	RollingUpdate *MachineRollingUpdateDeployment `json:"rollingUpdate,omitempty"`
}

type MachineRollingUpdateDeployment struct {
	MaxUnavailable *intstr.IntOrString `json:"maxUnavailable,omitempty"`
	MaxSurge       *intstr.IntOrString `json:"maxSurge,omitempty"`
}

type CaptMachineDeploymentStatus struct {
	ObservedGeneration  int64              `json:"observedGeneration,omitempty"`
	UpdatedReplicas     int32              `json:"updatedReplicas,omitempty"`
	Replicas            int32              `json:"replicas,omitempty"`
	AvailableReplicas   int32              `json:"availableReplicas,omitempty"`
	UnavailableReplicas int32              `json:"unavailableReplicas,omitempty"`
	Conditions          []metav1.Condition `json:"conditions,omitempty"`
	CollisionCount      *int32             `json:"collisionCount,omitempty"`
	FailureReason       *string            `json:"failureReason,omitempty"`
	FailureMessage      *string            `json:"failureMessage,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.replicas
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="Replicas",type="integer",JSONPath=".status.replicas"
// +kubebuilder:printcolumn:name="Updated",type="integer",JSONPath=".status.updatedReplicas"
// +kubebuilder:printcolumn:name="Available",type="integer",JSONPath=".status.availableReplicas"
// +kubebuilder:resource:path=captmachinedeployments,scope=Namespaced,categories=cluster-api
// +kubebuilder:storageversion

type CaptMachineDeployment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CaptMachineDeploymentSpec   `json:"spec,omitempty"`
	Status CaptMachineDeploymentStatus `json:"status,omitempty"`
}

// Hub marker
func (*CaptMachineDeployment) Hub() {}

// +kubebuilder:object:root=true

type CaptMachineDeploymentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CaptMachineDeployment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CaptMachineDeployment{}, &CaptMachineDeploymentList{})
}
