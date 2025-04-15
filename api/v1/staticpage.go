package v1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

/*
TypeMeta describes an individual object in an API response or request with strings representing the type of the object and its API schema version.
Structures that are versioned or persisted should inline TypeMeta.

ListMeta describes metadata that synthetic resources must have, including lists and various status objects.
*/

type StaticPageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []StaticPage `json:"items"`
}

type StaticPage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec StaticPageSpec `json:"spec"`
}

type StaticPageSpec struct {
	Contents string `json:"contents"`
	Image    string `json:"image"`
	Replicas int    `json:"replicas"`
}
