// +build !ignore_autogenerated

// Code generated by deepcopy-gen. DO NOT EDIT.

package v1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceExport) DeepCopyInto(out *ServiceExport) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceExport.
func (in *ServiceExport) DeepCopy() *ServiceExport {
	if in == nil {
		return nil
	}
	out := new(ServiceExport)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ServiceExport) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceExportConditions) DeepCopyInto(out *ServiceExportConditions) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceExportConditions.
func (in *ServiceExportConditions) DeepCopy() *ServiceExportConditions {
	if in == nil {
		return nil
	}
	out := new(ServiceExportConditions)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceExportList) DeepCopyInto(out *ServiceExportList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ServiceExport, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceExportList.
func (in *ServiceExportList) DeepCopy() *ServiceExportList {
	if in == nil {
		return nil
	}
	out := new(ServiceExportList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ServiceExportList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceExportSpec) DeepCopyInto(out *ServiceExportSpec) {
	*out = *in
	if in.FromLabels != nil {
		in, out := &in.FromLabels, &out.FromLabels
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.FromAnnoatations != nil {
		in, out := &in.FromAnnoatations, &out.FromAnnoatations
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.FromNamespaces != nil {
		in, out := &in.FromNamespaces, &out.FromNamespaces
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.ToCluster != nil {
		in, out := &in.ToCluster, &out.ToCluster
		*out = new(ToCluster)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceExportSpec.
func (in *ServiceExportSpec) DeepCopy() *ServiceExportSpec {
	if in == nil {
		return nil
	}
	out := new(ServiceExportSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceExportStatus) DeepCopyInto(out *ServiceExportStatus) {
	*out = *in
	if in.LastScheduleTime != nil {
		in, out := &in.LastScheduleTime, &out.LastScheduleTime
		*out = (*in).DeepCopy()
	}
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]ServiceExportConditions, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceExportStatus.
func (in *ServiceExportStatus) DeepCopy() *ServiceExportStatus {
	if in == nil {
		return nil
	}
	out := new(ServiceExportStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ToCluster) DeepCopyInto(out *ToCluster) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ToCluster.
func (in *ToCluster) DeepCopy() *ToCluster {
	if in == nil {
		return nil
	}
	out := new(ToCluster)
	in.DeepCopyInto(out)
	return out
}
