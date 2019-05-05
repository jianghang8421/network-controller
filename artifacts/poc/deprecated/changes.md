API group 从staticmacvlan.rancher.com 变更为 macvlan.cluster.cattle.io

StaticPod更名为 MacvlanIP，字段中ip变更为cidr，vlan变更为subnet, subnet字段带有label

type MacvlanIP struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec MacvlanIPSpec `json:"spec"`
}

type MacvlanIPSpec struct {
	Subnet string `json:"subnet"`
	PodID  string `json:"podId"`
	CIDR   string `json:"cidr"`
	MAC    string `json:"mac"`
}

这一部分的规则，跟之前相同：

Subnet字段需要读取MacvlanSubnet的列表来取得名字
cidr设置“auto”：cidr用户输入为空时，前端至少要填充一个“auto”。mac可以不填


VLANSubnet更名为MacvlanSubnet，增加了Mode，VLAN（int型，取值为2-4095）gateway(optional) ，master vlan mode 三个字段带有label

type MacvlanSubnet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec MacvlanSubnetSpec `json:"spec"`
}

type MacvlanSubnetSpec struct {
	Master string `json:"master"`
	VLAN   int    `json:"vlan"`
	CIDR   string `json:"cidr"`
	Mode   string `json:"mode"`
    Gateway string `json:"gateway"`
}

VLAN 是一个数字（int型，取值为2-4095)

Mode字段字符串有4个可选值：

"bridge" "private" "vepa" "passthru"

默认 “bridge”

Gateway是形如 192.168.1.1 的网关地址