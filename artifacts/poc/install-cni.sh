#!/bin/bash -x

VERSION="v0.1.0"

sudo rm -rf /opt/cni/bin
sudo mkdir -p /opt/cni/bin

wget https://github.com/containernetworking/plugins/releases/download/v0.7.5/cni-plugins-amd64-v0.7.5.tgz
sudo tar zxvf cni-plugins-amd64-v0.7.5.tgz -C /opt/cni/bin/

wget https://github.com/intel/multus-cni/releases/download/v3.2/multus-cni_3.2_linux_amd64.tar.gz
tar zxvf multus-cni_3.2_linux_amd64.tar.gz 
sudo cp multus-cni_3.2_linux_amd64/multus-cni /opt/cni/bin/

wget https://github.com/wardenlym/static-macvlan-cni/releases/download/$VERSION/static-macvlan-cni-amd64-$VERSION.tgz
sudo tar zxvf static-macvlan-cni-amd64-$VERSION.tgz -C /opt/cni/bin/

wget https://github.com/wardenlym/static-ipam/releases/download/$VERSION/static-ipam-amd64-$VERSION.tgz
sudo tar zxvf static-ipam-amd64-$VERSION.tgz -C /opt/cni/bin/

rm cni-plugins-amd64-v0.7.5.tgz
rm -rf multus-cni_3.2_linux_amd64
rm multus-cni_3.2_linux_amd64.tar.gz 
rm static-macvlan-cni-amd64-$VERSION.tgz
rm static-ipam-amd64-$VERSION.tgz