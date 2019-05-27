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
docker pull cnrancher/rancher:v2.2.3
```

其他需要的镜像(image.list):

```
nfvpe/multus:v3.2
quay.io/coreos/flannel:v0.10.0-amd64
cnrancher/static-macvlan-cni:v0.2.1
cnrancher/network-controller:v0.3.0
cnrancher/k8s-net-attach-def-controller:v0.1.0
```

### 创建集群

选择 **添加集群 - Custom**

选择任意版本的k8s。

配置好其他选项后，点选"编辑YAML"，将其中的 network/plugin 字段修改为none，并添加addons如下：

```
network:
  plugin: "none"

addons_include:
  - https://raw.githubusercontent.com/cnrancher/network-controller/v0.3.0/artifacts/templates/multus-flannel-macvlan.yml

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
  - /var/lib/rancher/network-addons/macvlan-network.yml
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

### 同一subnet不同namespace的POD连通性

同一个macvlan subnet下，POD可以部署在不同的namespace上，这时POD的IP在不同的namespace中不会重复，且POD之间的macvlan网络可以连接。

### 不同namespace POD的macvlan如何隔离

不同的namespace采用不同的subnet，利用vlan的隔离性来隔离namespace之间的网络。

### 同一个workload最多能够同时指定多少对IP和MAC

指定的IP和MAC会设置到label中，所以会受Kubernetes本身label value不能超过253字符长度的限制，所以同时指定的IP和MAC最多不能超过14对。

## 其他文档参考

- https://intel.github.io/multus-cni/doc/configuration.html


