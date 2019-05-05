## macvlan network 功能说明文档

# create a cluster in rancher-ui using macvlan via network addons

```
network: 
  plugin: "none"

addons_include:
  - https://raw.githubusercontent.com/cnrancher/network-controller/v0.2.0-dev/artifacts/multus-daemonset.yml
  - https://raw.githubusercontent.com/cnrancher/network-controller/v0.2.0-dev/artifacts/network-cni-daemonset.yml
  - https://raw.githubusercontent.com/cnrancher/network-controller/v0.2.0-dev/artifacts/flannel-daemonset.yml
  - https://raw.githubusercontent.com/cnrancher/network-controller/v0.2.0-dev/artifacts/network-controller.yml

```



```
addon_job_timeout: 30
authentication: 
  strategy: "x509"
ignore_docker_version: true

ingress: 
  provider: "nginx"
kubernetes_version: "v1.14.1-rancher1-1"
monitoring: 
  provider: "metrics-server"

network: 
  plugin: "none"

addons_include:
  - https://raw.githubusercontent.com/cnrancher/network-controller/v0.2.0-dev/artifacts/multus-daemonset.yml
  - https://raw.githubusercontent.com/cnrancher/network-controller/v0.2.0-dev/artifacts/network-cni-daemonset.yml
  - https://raw.githubusercontent.com/cnrancher/network-controller/v0.2.0-dev/artifacts/flannel-daemonset.yml
  - https://raw.githubusercontent.com/cnrancher/network-controller/v0.2.0-dev/artifacts/network-controller.yml

services: 
  etcd: 
    backup_config: 
      enabled: true
      interval_hours: 12
      retention: 6
    creation: "12h"
    extra_args: 
      heartbeat-interval: 500
      election-timeout: 5000
    retention: "72h"
    snapshot: false
  kube-api: 
    always_pull_images: false
    pod_security_policy: false
    service_node_port_range: "30000-32767"
ssh_agent_auth: false
# 
#   # Rancher Config
# 
docker_root_dir: "/var/lib/docker"
enable_cluster_alerting: false
enable_cluster_monitoring: false
enable_network_policy: false
local_cluster_auth_endpoint: 
  enabled: true
name: "test"
```