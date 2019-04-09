package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// StaticPod is a specification for a StaticPod resource
type StaticPod struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec StaticPodSpec `json:"spec"`
}

// StaticPodSpec is the spec for a StaticPod resource
type StaticPodSpec struct {
	ContainerID string `json:"container-id"`
	IP          string `json:"ip"`
	PodID       string `json:"pod-id"`
	VLan        string `json:"vlan"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// StaticPodList is a list of StaticPod resources
type StaticPodList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []StaticPod `json:"items"`
}
