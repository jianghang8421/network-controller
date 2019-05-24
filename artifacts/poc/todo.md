- 创建workload ，去掉podid
- 逗号分隔，pod 多个ip 多个mac
- docs flannel默认出口 eth0


# 前端

## 扁平网络部分

- gateway 校验，必须是ip，不能是cidr形式 192.168.1.1

- 重复macvlansubnet前端校验规则

macvlansubnet的list接口，可以传递使用lableSelector 筛选master && vlan，当已经存在vlan与master都同名的item时，不允许创建

labelSelector=master=eth0&vlan=10

- 创建subnet中增加ip range

UI加入一个可选的ip段标签栏，允许输入两个值start-end，发送给后端请求时，转换进新增的ranges字段中：

```
"spec": {}
        "cidr": "192.168.56.0/24",
        "master": "enp0s9",
        "mode": "bridge",
        "ranges": [
            {
                "rangeEnd": "192.168.56.20",
                "rangeStart": "192.168.56.10"
            },
            {
                "rangeEnd": "192.168.56.40",
                "rangeStart": "192.168.56.30"
            }
        ]
    }

- 
```



## 创建workload部分

- 查看deployment的信息显示mac为 n/a 可以为auto

mac字段为空字符串时，前端可以补auto，跟ip保持一致。（后端现在兼容了”“和”auto“两种）

- ”CIDR“字段变更

CIDR 字段 变更为IP，支持三种情况：

```
1. 空，不填，这时候跟之前一样，前端补充为”auto“
2. 192.168.1.1 格式的单个ip
3. 多个ip，后端需要的形式为 ”192.168.1.1-192.168.1.2-192.168.1.3“ ，用”-“分隔的多个ip
多个ip在ui上的交互展现待考虑

```

- 修改spec template中的annotation key

加入namespace前缀了

```
	cidr         macvlan.pandaria.cattle.io/ip
	subnet       macvlan.pandaria.cattle.io/subnet
	mac          macvlan.pandaria.cattle.io/mac
```

- todo

```
x iprange 
x multiple ip
x event log
x gateway ip
x static-macvlan-cni set promisc on
x own ref delete crd
x code refactoring
x offline install: images.list 
x timer delete crd
x v0.2.1
x namespace bug
```

v0.3.0

* macvlan svc discovy
* ui select multus-macvlan plugin
* ingress ip select macvlan net1

```
vlan范围：0~4095
    0，4095 保留 仅限系统使用 用户不能查看和使用这些VLAN
    1 正常 Cisco默认VLAN 用户能够使用该VLAN，但不能删除它
    2-1001 正常 用于以太网的VLAN 用户可以创建、使用和删除这些VLAN
    1002-1005 正常 用于FDDI和令牌环的Cisco默认VLAN 用户不能删除这些VLAN
    1006-1024 保留 仅限系统使用 用户不能查看和使用这些VLAN
    1025-4094 扩展 仅用于以太网VLAN
```

* ui select multus-flannel

<添加集群> - <Custom> 界面中 <网络组件> 新增一项, 显示为：Multus+Flannel

windows support 和 项目网络隔离 两个单选框disable，置灰

模式选项：

当选择这种模式时，设置

```
network:
  plugin: "multus-flannel"
```

本版本提供一个参数：flannel_iface, 默认时不需要传。

当用户输入时（比如输入“ens4”），将值作为字符串设置在option中：

```
network:
  plugin: "multus-flannel"
  options:
    flannel_iface: "ens4"
```