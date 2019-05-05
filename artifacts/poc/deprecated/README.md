static-macvlan功能说明文档
========
## 主机环境需求

需要主机支持macvlan(不能使用aws)

可以使用物理机或本地虚拟机的桥接网卡/host-only网卡（部分无线网卡的桥接模式也无法支持macvlan，建议使用有线）

## 版本信息

k8s version >= stable-1.13

## 测试机需要设置的配置

关闭swap

将用于macvlan网络的网卡混杂模式开启

```
sudo swapoff -a 
sudo ip link set eth0 promisc on
```

## 需要的第三方镜像

    quay.io/coreos/flannel:v0.10.0-amd64
    nfvpe/multus:v3.2
    cnrancher/network-controller:v0.1.0

## 验证流程

- 在所有node上安装cni插件

    ./artifacts/poc/install-cni.sh

运行此脚本会download下列cni-plugin到host的 /opt/cni/bin/

    cni-plugins-amd64-v0.7.5
    multus-cni_3.2_linux_amd64
    static-macvlan-cni-amd64-v0.1.0
    static-ipam-amd64-v0.1.0

```
$ ls /opt/cni/bin 
bridge             host-device        loopback           portmap            static-ipam        vlan
dhcp               host-local         macvlan            ptp                static-macvlan-cni
flannel            ipvlan             multus-cni         sample             tuning
```

- 使用kubeadm创建cluster

init适用于flannel网络的集群

```
sudo kubeadm init --apiserver-advertise-address=<YOUR-APISERVER-ADDR> --pod-network-cidr=10.244.0.0/16
```

单节点测试时，配置允许master创建pod

    kubectl taint nodes --all node-role.kubernetes.io/master-

- 初始化k8s网络

本项目需要的资源对象已经整合在macvlan-network.yml中：

    kubectl apply -f artifacts/poc/macvlan-network.yml

其中已经包含了下列组件：

```
multus-daemonset.yml           # 配置multus
flannel-daemonset.yml          # 针对multus修改过的flannel ds
rbac.yml                       # 本项目需要的rbac
crd.yml                        # 本项目相关crd: staticpods vlansubnets
static-macvlan-cni.yml         # multus cni config
network-controller.yml      # crd controller
```

- 查看crd成功创建，controller正常运行

```
$ kubectl apply -f artifacts/poc/macvlan-network.yml
customresourcedefinition.apiextensions.k8s.io/network-attachment-definitions.k8s.cni.cncf.io created
clusterrole.rbac.authorization.k8s.io/multus created
clusterrolebinding.rbac.authorization.k8s.io/multus created
serviceaccount/multus created
configmap/multus-cni-config created
daemonset.extensions/kube-multus-ds-amd64 created
clusterrole.rbac.authorization.k8s.io/flannel created
clusterrolebinding.rbac.authorization.k8s.io/flannel created
serviceaccount/flannel created
configmap/kube-flannel-cfg created
daemonset.extensions/kube-flannel-ds-amd64 created
daemonset.extensions/kube-flannel-ds-arm64 created
daemonset.extensions/kube-flannel-ds-arm created
daemonset.extensions/kube-flannel-ds-ppc64le created
daemonset.extensions/kube-flannel-ds-s390x created
clusterrole.rbac.authorization.k8s.io/multus configured
customresourcedefinition.apiextensions.k8s.io/staticpods.staticmacvlan.rancher.com created
customresourcedefinition.apiextensions.k8s.io/vlansubnets.staticmacvlan.rancher.com created
networkattachmentdefinition.k8s.cni.cncf.io/static-macvlan-cni created
pod/network-controller created

$ kubectl get pods -o wide --all-namespaces
NAMESPACE     NAME                                     READY   STATUS    RESTARTS   AGE     IP             NODE             NOMINATED NODE   READINESS GATES
kube-system   coredns-fb8b8dccf-2hq9m                  1/1     Running   0          3m32s   10.244.0.99    ip-172-31-42-3   <none>           <none>
kube-system   coredns-fb8b8dccf-qghvf                  1/1     Running   0          3m32s   10.244.0.98    ip-172-31-42-3   <none>           <none>
kube-system   etcd-ip-172-31-42-3                      1/1     Running   0          2m50s   172.31.42.3    ip-172-31-42-3   <none>           <none>
kube-system   kube-apiserver-ip-172-31-42-3            1/1     Running   0          2m39s   172.31.42.3    ip-172-31-42-3   <none>           <none>
kube-system   kube-controller-manager-ip-172-31-42-3   1/1     Running   0          2m33s   172.31.42.3    ip-172-31-42-3   <none>           <none>
kube-system   kube-flannel-ds-amd64-ndtj6              1/1     Running   0          92s     172.31.42.3    ip-172-31-42-3   <none>           <none>
kube-system   kube-multus-ds-amd64-jk7vv               1/1     Running   0          93s     172.31.42.3    ip-172-31-42-3   <none>           <none>
kube-system   kube-proxy-n2rjj                         1/1     Running   0          3m32s   172.31.42.3    ip-172-31-42-3   <none>           <none>
kube-system   kube-scheduler-ip-172-31-42-3            1/1     Running   0          2m54s   172.31.42.3    ip-172-31-42-3   <none>           <none>
kube-system   network-controller                    1/1     Running   0          10s     10.244.0.100   ip-172-31-42-3   <none>           <none>
```

- 创建vlansubnet

```
$ kubectl apply -f artifacts/sample/vlan-subnet.yml
vlansubnet.staticmacvlan.rancher.com/mv1.100 created

$ kubectl get vlansubnets
NAME      MASTER   CIDR
mv1.100   eth0     192.168.1.0/24
```

- 测试创建带有静态指定的ip/mac地址的pod

pod-1中的自定义标识

```
  annotations:
    k8s.v1.cni.cncf.io/networks: static-macvlan-cni
    static-ip: 192.168.1.100/24
    static-mac: 0a:00:27:00:00:00
    vlan: mv1.100
```

```
$ kubectl apply -f artifacts/sample/pod-1.yml
pod/samplepod1 created

$ kubectl exec samplepod1 -it -- ip a
1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    inet 127.0.0.1/8 scope host lo
       valid_lft forever preferred_lft forever
3: eth0@if9: <BROADCAST,MULTICAST,UP,LOWER_UP,M-DOWN> mtu 8951 qdisc noqueue
    link/ether 3a:73:cb:08:2b:ce brd ff:ff:ff:ff:ff:ff
    inet 10.244.0.101/24 scope global eth0
       valid_lft forever preferred_lft forever
4: net1@if2: <BROADCAST,MULTICAST,UP,LOWER_UP,M-DOWN> mtu 9001 qdisc noqueue
    link/ether 0a:00:27:00:00:00 brd ff:ff:ff:ff:ff:ff
    inet 192.168.1.100/24 brd 192.168.1.255 scope global net1
       valid_lft forever preferred_lft forever
```

- 测试创建指定子网自动分配ip的pod

pod-2中的自定义标识

```
  annotations:
    k8s.v1.cni.cncf.io/networks: static-macvlan-cni
    static-ip: auto
    vlan: mv1.100
```

```
$ kubectl apply -f artifacts/sample/pod-2.yml
pod/samplepod2 created

# ubuntu @ ip-172-31-42-3 in ~/go/src/github.com/cnrancher/network-controller on git:dev o [16:15:38]
$ kubectl exec samplepod2 -it -- ip a
1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    inet 127.0.0.1/8 scope host lo
       valid_lft forever preferred_lft forever
3: eth0@if11: <BROADCAST,MULTICAST,UP,LOWER_UP,M-DOWN> mtu 8951 qdisc noqueue
    link/ether 16:d1:f6:b8:29:d5 brd ff:ff:ff:ff:ff:ff
    inet 10.244.0.102/24 scope global eth0
       valid_lft forever preferred_lft forever
4: net1@if2: <BROADCAST,MULTICAST,UP,LOWER_UP,M-DOWN> mtu 9001 qdisc noqueue
    link/ether b2:76:1e:4a:88:70 brd ff:ff:ff:ff:ff:ff
    inet 192.168.1.133/24 brd 192.168.1.255 scope global net1
       valid_lft forever preferred_lft forever
```

- 查看host创建了相应的设备

```
$ ip link ls mv1.100
10: mv1.100@eth0: <BROADCAST,MULTICAST> mtu 9001 qdisc noop state DOWN mode DEFAULT group default
    link/ether b2:fc:cd:02:53:65 brd ff:ff:ff:ff:ff:ff
```

- 查看crd staticpods

```
$ kubectl get staticpods
NAME         VLAN      POD-ID                                 IP                 MAC
samplepod1   mv1.100   4ae53a2a-5f99-11e9-97f6-063e2068ed12   192.168.1.100/24   0a:00:27:00:00:00
samplepod2   mv1.100   b4ec67a5-5f99-11e9-97f6-063e2068ed12   192.168.1.133/24
```
- 测试pod的互通性

** 这里使用了我本机vm上测试的真实log，所以ip跟之前不同

```
$ kubectl exec samplepod2 -it -- ping 192.168.56.100
PING 192.168.56.100 (192.168.56.100): 56 data bytes
64 bytes from 192.168.56.100: seq=0 ttl=64 time=0.093 ms
64 bytes from 192.168.56.100: seq=1 ttl=64 time=0.099 ms

$ kubectl exec samplepod1 -it -- ping 192.168.56.230
PING 192.168.56.230 (192.168.56.230): 56 data bytes
64 bytes from 192.168.56.230: seq=0 ttl=64 time=0.051 ms
64 bytes from 192.168.56.230: seq=1 ttl=64 time=0.072 ms
```

## 已知待改进的问题

目前没有实现支持一个pod中设置多个static macvlan接口(之后可以设计在cni config中配置多接口)

默认使用了bridge mode创建vlan（之后可以配置在vlansubnet中）

macvlan设计上，由于安全原因，子namespace中创建的接口不能与被继承自的物理接口直接互通(可以将主机ip设置在其他的vlan接口上访问)

host上创建的vlan设备只会检测当不存在时创建，没有处理删除逻辑(可以手动删除，之后可以做一个daemon处理这些)

crd使用的namesapce待确定(使用cattle/cluster)