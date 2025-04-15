package v1

import "k8s.io/apimachinery/pkg/runtime"

// DeepCopyInto copies all properties from the receiver to the given object of the same type.
func (in *StaticPage) DeepCopyInto(out *StaticPage) {
	out.TypeMeta = in.TypeMeta
	out.ObjectMeta = in.ObjectMeta
	out.Spec = StaticPageSpec{
		Replicas: in.Spec.Replicas,
		Image:    in.Spec.Image,
		Contents: in.Spec.Contents,
	}
}

// DeepCopyObject returns a generically typed copy of the receiver.
func (in *StaticPage) DeepCopyObject() runtime.Object {
	out := StaticPage{}
	in.DeepCopyInto(&out)

	return &out
}

// DeepCopyObject returns a generically typed copy of the receiver.
func (in *StaticPageList) DeepCopyObject() runtime.Object {
	out := StaticPageList{}
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta

	if in.Items != nil {
		out.Items = make([]StaticPage, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}

	return &out
}
