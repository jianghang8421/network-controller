
# api test

```
#!/bin/bash

# list vlansubnets
curl --insecure -H 'Cookie:CSRF=092c2c6668; R_USERNAME=admin; R_SESS=token-9ln7p:z96c4cb4cm244xvhs522xtghw2x8mhv2x4j5npwx4r9kkhjpzn2cl2' \
    https://52.195.0.164:8082/k8s/clusters/c-76l8k/apis/staticmacvlan.rancher.com/v1/namespaces/default/vlansubnets

# create vlansubnet
curl --insecure -XPOST -H "Content-Type: application/json" -H 'Cookie:CSRF=092c2c6668; R_USERNAME=admin; R_SESS=token-9ln7p:z96c4cb4cm244xvhs522xtghw2x8mhv2x4j5npwx4r9kkhjpzn2cl2' \
    https://52.195.0.164:8082/k8s/clusters/c-76l8k/apis/staticmacvlan.rancher.com/v1/namespaces/default/vlansubnets \
    -d '{
    "apiVersion": "staticmacvlan.rancher.com/v1",
    "kind": "VLANSubnet",
    "metadata": {
        "name": "testvlan",
        "namespace": "default"
    },
    "spec": {
        "master": "eth0",
        "cidr": "192.168.2.0/24"
    }
}'

# delete vlansubnet
curl --insecure -XDELETE -H "Content-Type: application/json" -H 'Cookie:CSRF=092c2c6668; R_USERNAME=admin; R_SESS=token-9ln7p:z96c4cb4cm244xvhs522xtghw2x8mhv2x4j5npwx4r9kkhjpzn2cl2' \
    https://52.195.0.164:8082/k8s/clusters/c-76l8k/apis/staticmacvlan.rancher.com/v1/namespaces/default/vlansubnets/testvlan


# list staticpods
curl --insecure -H 'Cookie:CSRF=092c2c6668; R_USERNAME=admin; R_SESS=token-9ln7p:z96c4cb4cm244xvhs522xtghw2x8mhv2x4j5npwx4r9kkhjpzn2cl2' \
    https://52.195.0.164:8082/k8s/clusters/c-76l8k/apis/staticmacvlan.rancher.com/v1/staticpods

```

# staticpod crd annotation

```
  annotations:
    k8s.v1.cni.cncf.io/networks: static-macvlan-cni
    static-ip: 192.168.1.100/24
    static-mac: 0a:00:27:00:00:00
    vlan: mv1.100
```

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