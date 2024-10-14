//go:build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1beta1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AddonConfig) DeepCopyInto(out *AddonConfig) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AddonConfig.
func (in *AddonConfig) DeepCopy() *AddonConfig {
	if in == nil {
		return nil
	}
	out := new(AddonConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CAPTCluster) DeepCopyInto(out *CAPTCluster) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CAPTCluster.
func (in *CAPTCluster) DeepCopy() *CAPTCluster {
	if in == nil {
		return nil
	}
	out := new(CAPTCluster)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CAPTCluster) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CAPTClusterAddons) DeepCopyInto(out *CAPTClusterAddons) {
	*out = *in
	out.CoreDNS = in.CoreDNS
	out.VpcCni = in.VpcCni
	out.KubeProxy = in.KubeProxy
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CAPTClusterAddons.
func (in *CAPTClusterAddons) DeepCopy() *CAPTClusterAddons {
	if in == nil {
		return nil
	}
	out := new(CAPTClusterAddons)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CAPTClusterKarpenter) DeepCopyInto(out *CAPTClusterKarpenter) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CAPTClusterKarpenter.
func (in *CAPTClusterKarpenter) DeepCopy() *CAPTClusterKarpenter {
	if in == nil {
		return nil
	}
	out := new(CAPTClusterKarpenter)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CAPTClusterList) DeepCopyInto(out *CAPTClusterList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]CAPTCluster, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CAPTClusterList.
func (in *CAPTClusterList) DeepCopy() *CAPTClusterList {
	if in == nil {
		return nil
	}
	out := new(CAPTClusterList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CAPTClusterList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CAPTClusterSpec) DeepCopyInto(out *CAPTClusterSpec) {
	*out = *in
	out.Addons = in.Addons
	out.Karpenter = in.Karpenter
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CAPTClusterSpec.
func (in *CAPTClusterSpec) DeepCopy() *CAPTClusterSpec {
	if in == nil {
		return nil
	}
	out := new(CAPTClusterSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CAPTClusterStatus) DeepCopyInto(out *CAPTClusterStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CAPTClusterStatus.
func (in *CAPTClusterStatus) DeepCopy() *CAPTClusterStatus {
	if in == nil {
		return nil
	}
	out := new(CAPTClusterStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CAPTMachineTemplate) DeepCopyInto(out *CAPTMachineTemplate) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CAPTMachineTemplate.
func (in *CAPTMachineTemplate) DeepCopy() *CAPTMachineTemplate {
	if in == nil {
		return nil
	}
	out := new(CAPTMachineTemplate)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CAPTMachineTemplate) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CAPTMachineTemplateList) DeepCopyInto(out *CAPTMachineTemplateList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]CAPTMachineTemplate, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CAPTMachineTemplateList.
func (in *CAPTMachineTemplateList) DeepCopy() *CAPTMachineTemplateList {
	if in == nil {
		return nil
	}
	out := new(CAPTMachineTemplateList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CAPTMachineTemplateList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CAPTMachineTemplateSpec) DeepCopyInto(out *CAPTMachineTemplateSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CAPTMachineTemplateSpec.
func (in *CAPTMachineTemplateSpec) DeepCopy() *CAPTMachineTemplateSpec {
	if in == nil {
		return nil
	}
	out := new(CAPTMachineTemplateSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CAPTMachineTemplateStatus) DeepCopyInto(out *CAPTMachineTemplateStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CAPTMachineTemplateStatus.
func (in *CAPTMachineTemplateStatus) DeepCopy() *CAPTMachineTemplateStatus {
	if in == nil {
		return nil
	}
	out := new(CAPTMachineTemplateStatus)
	in.DeepCopyInto(out)
	return out
}
