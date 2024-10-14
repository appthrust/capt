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
)

// CAPTMachineTemplateSpec defines the desired state of CAPTMachineTemplate
type CAPTMachineTemplateSpec struct {
	Template CAPTMachineTemplateResource `json:"template"`
}

// CAPTMachineTemplateResource describes the data needed to create a CAPTMachine from a template
type CAPTMachineTemplateResource struct {
	Spec CAPTMachineSpec `json:"spec"`
}

// CAPTMachineSpec defines the desired state of CAPTMachine
type CAPTMachineSpec struct {
	FargateProfile []FargateProfileConfig `json:"fargateProfile"`
}

// FargateProfileConfig defines the desired state of a Fargate profile
type FargateProfileConfig struct {
	Name      string           `json:"name"`
	Selectors []SelectorConfig `json:"selectors"`
}

// SelectorConfig defines the selectors for a Fargate profile
type SelectorConfig struct {
	Namespace string `json:"namespace"`
}

// CAPTMachineTemplateStatus defines the observed state of CAPTMachineTemplate
type CAPTMachineTemplateStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// CAPTMachineTemplate is the Schema for the captmachinetemplates API
type CAPTMachineTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CAPTMachineTemplateSpec   `json:"spec,omitempty"`
	Status CAPTMachineTemplateStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CAPTMachineTemplateList contains a list of CAPTMachineTemplate
type CAPTMachineTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CAPTMachineTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CAPTMachineTemplate{}, &CAPTMachineTemplateList{})
}
