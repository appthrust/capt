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

// CAPTMachineTemplateSpec defines the desired state of CaptMachineTemplate
type CAPTMachineTemplateSpec struct {
	// Template is the template for creating a CaptMachine
	Template CAPTMachineTemplateResource `json:"template"`
}

// CAPTMachineTemplateResource describes the data needed to create a CaptMachine from a template
type CAPTMachineTemplateResource struct {
	// Spec is the specification of the desired behavior of the machine.
	Spec CaptMachineSpec `json:"spec"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:path=captmachinetemplates,scope=Namespaced,categories=cluster-api

// CaptMachineTemplate is the Schema for the captmachinetemplates API
type CaptMachineTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec CAPTMachineTemplateSpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true

// CaptMachineTemplateList contains a list of CaptMachineTemplate
type CaptMachineTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CaptMachineTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CaptMachineTemplate{}, &CaptMachineTemplateList{})
}
