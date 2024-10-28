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
	commonv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"k8s.io/api/core/v1"
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
	in.Spec.DeepCopyInto(&out.Spec)
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
	if in.NetworkRef != nil {
		in, out := &in.NetworkRef, &out.NetworkRef
		*out = new(v1.ObjectReference)
		**out = **in
	}
	in.VPC.DeepCopyInto(&out.VPC)
	in.EKS.DeepCopyInto(&out.EKS)
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
func (in *CAPTMachineSpec) DeepCopyInto(out *CAPTMachineSpec) {
	*out = *in
	if in.FargateProfile != nil {
		in, out := &in.FargateProfile, &out.FargateProfile
		*out = make([]FargateProfileConfig, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CAPTMachineSpec.
func (in *CAPTMachineSpec) DeepCopy() *CAPTMachineSpec {
	if in == nil {
		return nil
	}
	out := new(CAPTMachineSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CAPTMachineTemplate) DeepCopyInto(out *CAPTMachineTemplate) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
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
func (in *CAPTMachineTemplateResource) DeepCopyInto(out *CAPTMachineTemplateResource) {
	*out = *in
	in.Spec.DeepCopyInto(&out.Spec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CAPTMachineTemplateResource.
func (in *CAPTMachineTemplateResource) DeepCopy() *CAPTMachineTemplateResource {
	if in == nil {
		return nil
	}
	out := new(CAPTMachineTemplateResource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CAPTMachineTemplateSpec) DeepCopyInto(out *CAPTMachineTemplateSpec) {
	*out = *in
	in.Template.DeepCopyInto(&out.Template)
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

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CAPTVPCTemplate) DeepCopyInto(out *CAPTVPCTemplate) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CAPTVPCTemplate.
func (in *CAPTVPCTemplate) DeepCopy() *CAPTVPCTemplate {
	if in == nil {
		return nil
	}
	out := new(CAPTVPCTemplate)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CAPTVPCTemplate) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CAPTVPCTemplateList) DeepCopyInto(out *CAPTVPCTemplateList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]CAPTVPCTemplate, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CAPTVPCTemplateList.
func (in *CAPTVPCTemplateList) DeepCopy() *CAPTVPCTemplateList {
	if in == nil {
		return nil
	}
	out := new(CAPTVPCTemplateList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CAPTVPCTemplateList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CAPTVPCTemplateSpec) DeepCopyInto(out *CAPTVPCTemplateSpec) {
	*out = *in
	if in.PublicSubnetTags != nil {
		in, out := &in.PublicSubnetTags, &out.PublicSubnetTags
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.PrivateSubnetTags != nil {
		in, out := &in.PrivateSubnetTags, &out.PrivateSubnetTags
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Tags != nil {
		in, out := &in.Tags, &out.Tags
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.WriteConnectionSecretToRef != nil {
		in, out := &in.WriteConnectionSecretToRef, &out.WriteConnectionSecretToRef
		*out = new(commonv1.SecretReference)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CAPTVPCTemplateSpec.
func (in *CAPTVPCTemplateSpec) DeepCopy() *CAPTVPCTemplateSpec {
	if in == nil {
		return nil
	}
	out := new(CAPTVPCTemplateSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CAPTVPCTemplateStatus) DeepCopyInto(out *CAPTVPCTemplateStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CAPTVPCTemplateStatus.
func (in *CAPTVPCTemplateStatus) DeepCopy() *CAPTVPCTemplateStatus {
	if in == nil {
		return nil
	}
	out := new(CAPTVPCTemplateStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EKSConfig) DeepCopyInto(out *EKSConfig) {
	*out = *in
	if in.NodeGroups != nil {
		in, out := &in.NodeGroups, &out.NodeGroups
		*out = make([]NodeGroupConfig, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EKSConfig.
func (in *EKSConfig) DeepCopy() *EKSConfig {
	if in == nil {
		return nil
	}
	out := new(EKSConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FargateProfileConfig) DeepCopyInto(out *FargateProfileConfig) {
	*out = *in
	if in.Selectors != nil {
		in, out := &in.Selectors, &out.Selectors
		*out = make([]SelectorConfig, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FargateProfileConfig.
func (in *FargateProfileConfig) DeepCopy() *FargateProfileConfig {
	if in == nil {
		return nil
	}
	out := new(FargateProfileConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodeGroupConfig) DeepCopyInto(out *NodeGroupConfig) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodeGroupConfig.
func (in *NodeGroupConfig) DeepCopy() *NodeGroupConfig {
	if in == nil {
		return nil
	}
	out := new(NodeGroupConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SelectorConfig) DeepCopyInto(out *SelectorConfig) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SelectorConfig.
func (in *SelectorConfig) DeepCopy() *SelectorConfig {
	if in == nil {
		return nil
	}
	out := new(SelectorConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VPCConfig) DeepCopyInto(out *VPCConfig) {
	*out = *in
	if in.PublicSubnetTags != nil {
		in, out := &in.PublicSubnetTags, &out.PublicSubnetTags
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.PrivateSubnetTags != nil {
		in, out := &in.PrivateSubnetTags, &out.PrivateSubnetTags
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VPCConfig.
func (in *VPCConfig) DeepCopy() *VPCConfig {
	if in == nil {
		return nil
	}
	out := new(VPCConfig)
	in.DeepCopyInto(out)
	return out
}
