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
	"k8s.io/apimachinery/pkg/runtime"
)

// CAPTControlPlaneSpec defines the desired state of CAPTControlPlane
type CAPTControlPlaneSpec struct {
	// Version is the Kubernetes version of the control plane
	Version string `json:"version"`

	// MachineTemplate is a reference to the CAPTMachineTemplate that should be used
	// to create control plane instances
	MachineTemplate CAPTControlPlaneMachineTemplate `json:"machineTemplate"`
}

// CAPTControlPlaneMachineTemplate defines the template for creating control plane instances
type CAPTControlPlaneMachineTemplate struct {
	// InfrastructureRef is a reference to a provider-specific resource that holds the details
	// for provisioning the Control Plane for a Cluster.
	InfrastructureRef runtime.RawExtension `json:"infrastructureRef"`
}

// CAPTControlPlaneStatus defines the observed state of CAPTControlPlane
type CAPTControlPlaneStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// CAPTControlPlane is the Schema for the captcontrolplanes API
type CAPTControlPlane struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CAPTControlPlaneSpec   `json:"spec,omitempty"`
	Status CAPTControlPlaneStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CAPTControlPlaneList contains a list of CAPTControlPlane
type CAPTControlPlaneList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CAPTControlPlane `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CAPTControlPlane{}, &CAPTControlPlaneList{})
}
