

## Prepare

```
apt-get install cloud-image-utils qemu-kvm 
```

## Setup Bridge

```
./bridge.sh
```

## Setup VM

```
./prepare.sh rancher-1
./run-kvm --hostname rancher-1

./prepare.sh rancher-2
./run-kvm --hostname rancher-2
```

## Setup Rancher

Please run the rancher-server on host, and run agent on VM.

You should update the server-url of rancher-server to `https://172.20.0.1`.

When you add a cluster, please specify the `Public Address` and `Internal Address` with the host-only IP address, such as 172.20.115.71.
And the flannel iface should be `ens4`, you can try this link: `https://raw.githubusercontent.com/niusmallnan/rancher2-poc-demo/master/macvlan/flannel-daemonset-gcpkvm.yml`

As rancher-server runs on docker, docker can drop all bridge packets, so we need to disable iptables on bridge.
```
sysctl net.bridge.bridge-nf-call-iptables=0
```

It is recommanded to install docker via `curl https://releases.rancher.com/install-docker/18.06.sh | sh`.
