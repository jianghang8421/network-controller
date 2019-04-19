#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

./vendor/k8s.io/code-generator/generate-groups.sh "deepcopy,client,informer,lister" \
  github.com/cnrancher/static-pod-controller/pkg/generated github.com/cnrancher/static-pod-controller/types/apis \
  staticmacvlan:v1