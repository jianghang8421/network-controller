
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