package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	MacvlanSubnetNamespace = "kube-system"

	MacvlanAnnotationPrefix = "macvlan.pandaria.cattle.io/"
	AnnotationIP            = MacvlanAnnotationPrefix + "ip"
	AnnotationSubnet        = MacvlanAnnotationPrefix + "subnet"
	AnnotationMac           = MacvlanAnnotationPrefix + "mac"

	LabelSelectedIP     = MacvlanAnnotationPrefix + "selectedIp"
	LabelMultipleIPHash = MacvlanAnnotationPrefix + "multipleIpHash"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MacvlanIP is a specification for a MacvlanIP resource
type MacvlanIP struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec MacvlanIPSpec `json:"spec"`
}

// MacvlanIPSpec is the spec for a MacvlanIP resource
type MacvlanIPSpec struct {
	Subnet string `json:"subnet"`
	PodID  string `json:"podId"`
	CIDR   string `json:"cidr"`
	MAC    string `json:"mac"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MacvlanIPList is a list of MacvlanIP resources
type MacvlanIPList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []MacvlanIP `json:"items"`
}

////////////////////

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MacvlanSubnet is a specification for a MacvlanSubnet resource
type MacvlanSubnet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec MacvlanSubnetSpec `json:"spec"`
}

// MacvlanSubnetSpec is the spec for a MacvlanSubnet resource
type MacvlanSubnetSpec struct {
	Master  string    `json:"master"`
	VLAN    int       `json:"vlan"`
	CIDR    string    `json:"cidr"`
	Mode    string    `json:"mode"`
	Gateway string    `json:"gateway"`
	Ranges  []IPRange `json:"ranges"`
}

type IPRange struct {
	RangeStart string `json:"rangeStart"`
	RangeEnd   string `json:"rangeEnd"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MacvlanSubnetList is a list of MacvlanSubnet resources
type MacvlanSubnetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []MacvlanSubnet `json:"items"`
}
