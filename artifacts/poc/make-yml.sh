#!/bin/bash -x
set -e

cd $(dirname $0)
cat ../multus-daemonset.yml \
    ../flannel-daemonset.yml \
    ../rbac.yml \
    ../crd.yml \
    ../static-macvlan-cni.yml \
    ../static-pod-controller.yml \
        > macvlan-network.yml