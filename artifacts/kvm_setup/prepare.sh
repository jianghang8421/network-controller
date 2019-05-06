#!/bin/bash

HOSTNAME=$1
CLOUD_IMG="https://cloud-images.ubuntu.com/bionic/current/bionic-server-cloudimg-amd64.img"
STATE_DIR=$(dirname $0)/state
IMAGES_DIR=$(dirname $0)/images
LOCAL_IMG=${IMAGES_DIR}/$(basename ${CLOUD_IMG})
VM_IMG=${STATE_DIR}/${HOSTNAME}/hd.img
USERDATA_IMG=${STATE_DIR}/${HOSTNAME}/user-data.img

if [ "$#" -ne 1 ]; then
    echo "You should specify a device,  like: $0 rancher-1"
    exit 1
fi

mkdir -p ${STATE_DIR}/${HOSTNAME} ${IMAGES_DIR}

if [ ! -e ${LOCAL_IMG} ]; then
    curl -fL -o ${IMAGES_DIR}/$(basename ${CLOUD_IMG}) ${CLOUD_IMG}
fi

if [ ! -e ${VM_IMG} ]; then
    cp -v ${LOCAL_IMG} ${VM_IMG}
    qemu-img resize ${VM_IMG} +20G
fi

#---------- TAP DEVICE------------
ip tuntap add dev ${HOSTNAME} mode tap || true
ip link set ${HOSTNAME} up promisc on || true
ip link set ${HOSTNAME} master br0 || true

#---------- USER DATA ------------------

if [ -e ${USERDATA_IMG} ]; then
    exit 0
fi

USERDATA_FILE=$(dirname $0)/user-data

cat > ${USERDATA_FILE} <<EOF
#cloud-config
password: asdfgh
chpasswd: { expire: False }
ssh_pwauth: True
hostname: $HOSTNAME
EOF

cloud-localds ${USERDATA_IMG} ${USERDATA_FILE}

rm -f ${USERDATA_FILE}
