## macvlan network 功能说明文档

### 主机设置

关闭swap

为支持macvlan的网卡设备开启混杂模式，请替换eth0：

```
swapoff -a
ip link set eth0 promisc on
```

### 下载镜像

server请下载：

```
docker pull cnrancher/rancher:v2.2.2-macvlan
```

agent请下载：

```
docker pull rancher/rancher-agent:v2.2.2
docker tag rancher/rancher-agent:v2.2.2 cnrancher/rancher-agent:v2.2.2-macvlan
```

### 创建集群

选择 **添加集群 - Custom**

选择任意版本的k8s。

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

请注意，如果主机存在多个网卡，有时需要指定flannel vxlan网络的网卡链路，请调整flannel-daemonset.yml，参考如下：

```
...
...
      containers:
      - name: kube-flannel
        image: quay.io/coreos/flannel:v0.10.0-amd64
        command:
        - /opt/bin/flanneld
        args:
        - --ip-masq
        - --kube-subnet-mgr
        - --iface=ens4 #here
...
```

如果用户的网络提供web服务比较麻烦，可以把addons相关yaml文件拷贝到rancher-server中，RKE可以读取rancher-server容器的本地路径，比如：

```
mkdir -p network-addons
# download addons yml to network-addons dir
curl .....

docker cp network-addons <rancher-server>:/var/lib/rancher/


# then we can try this cluster yml files
network:
  plugin: "none"

addons_include:
  - /var/lib/rancher/network-addons/multus-daemonset.yml
  - /var/lib/rancher/network-addons/network-cni-daemonset.yml
  - /var/lib/rancher/network-addons/flannel-daemonset.yml
  - /var/lib/rancher/network-addons/network-controller.yml
```

下一步，之后等待集群创建完成。

### 创建MacvlanSubnet资源

在集群-扁平网络-创建MacvlanSubnet

配置子网属性

### 创建workload

在 **工作负载-高级-启用扁平网络** 中设定静态ip或者mac，当不指定时，为自动分配模式

### 测试

测试同一vlan的连通性，测试不同vlan pod的连通性

## FAQ

### 普通用户如何使用扁平网络

这里涉及到两个CRD的RBAC配置，分别是：macvlanips 和 macvlansubnets。如果一个用户作为cluster member和project member来管理扁平网络资源，可以参考如下：

1. Clone Cluster Member role，添加macvlanips 和 macvlansubnets访问权限，命名为Cluster Member Macvlan
2. Clone Project Member role，添加macvlanips 和 macvlansubnets访问权限，命名为Project Member Macvlan
3. 在某个Cluster下添加该用户为member，并设置访问权限为Cluster Member Macvlan
4. 某个Project下权限设置同上

### 如何确认Flannel vxlan的链路正常

在Host上查看vtep设备的fdb信息，只要dst目标地址是我们期待的网卡链路，那么就可以确保基础链路是没问题的，参考：

```
bridge fdb show dev flannel.1
7e:18:48:60:3b:97 dst 172.20.115.72 self permanent
```

如果dst有问题，则需要修改Flannel启动参数，一般调整iface即可。

