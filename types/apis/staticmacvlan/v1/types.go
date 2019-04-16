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
	VLAN  string `json:"vlan"`
	PodID string `json:"pod-id"`
	IP    string `json:"ip"`
	MAC   string `json:"mac"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// StaticPodList is a list of StaticPod resources
type StaticPodList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []StaticPod `json:"items"`
}

////////////////////

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VLANSubnet is a specification for a VLANSubnet resource
type VLANSubnet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec VLANSubnetSpec `json:"spec"`
}

// VLANSubnetSpec is the spec for a VLANSubnet resource
type VLANSubnetSpec struct {
	Master string `json:"master"`
	CIDR   string `json:"cidr"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VLANSubnetList is a list of VLANSubnet resources
type VLANSubnetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []VLANSubnet `json:"items"`
}
