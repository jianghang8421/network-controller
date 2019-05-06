## macvlan network 功能说明文档

- 主机设置

关闭swap

为支持macvlan的网卡设备开启混杂模式

```
swapoff -a
ip link set eth0 promisc on
```

- 下载镜像

```
docker pull cnrancher/rancher:v2.2.2-macvlan
```

- 创建集群

选择 添加集群 - Custom

选择v1.14.1版本的k8s

配置好其他选项后，点选"编辑YAML"，将其中的 network/plugin 字段修改为none，并添加addons如下：

```
network: 
  plugin: "none"

addons_include:
  - https://raw.githubusercontent.com/cnrancher/network-controller/v0.2.0/artifacts/multus-daemonset.yml
  - https://raw.githubusercontent.com/cnrancher/network-controller/v0.2.0/artifacts/network-cni-daemonset.yml
  - https://raw.githubusercontent.com/cnrancher/network-controller/v0.2.0/artifacts/flannel-daemonset.yml
  - https://raw.githubusercontent.com/cnrancher/network-controller/v0.2.0/artifacts/network-controller.yml

```

下一步，之后等待集群创建完成。

- 创建MacvlanSubnet资源

在集群-扁平网络-创建MacvlanSubnet

配置子网属性

- 创建workload

在创建工作负载-高级-启用扁平网络

中设定静态ip或者mac，当不指定时，为自动分配模式

- 测试

测试同一vlan的连通性，测试不通vlan pod的连通性