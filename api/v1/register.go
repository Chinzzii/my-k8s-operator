package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const GroupName = "kubernetes.chinzzii.com"
const GroupVersion = "v1"

// GroupVersion contains the "group" and the "version", which uniquely identifies the API.
var SchemaGroupVersion = schema.GroupVersion{Group: GroupName, Version: GroupVersion}

var (
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes) // NewSchemeBuilder calls Register for you.
	AddToScheme   = SchemeBuilder.AddToScheme // AddToScheme applies all the stored functions to the scheme.
)

func addKnownTypes(scheme *runtime.Scheme) error {
	// AddKnownTypes registers all types passed in 'types' as being members of version 'version'.
	// All objects passed to types should be pointers to structs.
	scheme.AddKnownTypes(SchemaGroupVersion,
		&StaticPage{},
		&StaticPageList{},
	)

	// AddToGroupVersion registers common meta types into schemas.
	metav1.AddToGroupVersion(scheme, SchemaGroupVersion)
	return nil
}
