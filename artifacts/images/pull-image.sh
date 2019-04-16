#!/bin/bash
set -e
images=(
    k8s.gcr.io/kube-apiserver:v1.13.4
    k8s.gcr.io/kube-controller-manager:v1.13.4
    k8s.gcr.io/kube-scheduler:v1.13.4
    k8s.gcr.io/kube-proxy:v1.13.4
    k8s.gcr.io/pause:3.1
    k8s.gcr.io/etcd:3.2.24
    k8s.gcr.io/coredns:1.2.6
    wardenlym/static-pod-controller:v0.1.0
    nfvpe/multus:v3.2
    quay.io/coreos/flannel:v0.10.0-amd64
)

for imageName in ${images[@]} ; do
    docker pull $imageName
done

docker save -o images.tar ${images[@]}
