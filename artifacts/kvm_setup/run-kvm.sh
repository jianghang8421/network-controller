#!/bin/bash
set -ex

QEMUARCH=x86_64
MEMORY=4096
CPU="-cpu host"
MACHINE="-enable-kvm"
HOSTNAME=
TAP_DEVICE=

while [ "$#" -gt 0 ]; do
    case $1 in
        --memory)
            shift 1
            MEMORY="$1"
            ;;
        --hostname)
            shift 1
            HOSTNAME="$1"
            ;;
        *)
            break
            ;;
    esac
    shift 1
done

if [ -z "${HOSTNAME}" ]; then
    echo "You must specify a hostname"
    exit 1
fi

MAC="DE:AD:BE:EF:96:3"$(echo ${HOSTNAME} | awk -F'-' '{print $2}')
NETWORK="-net nic,vlan=0,model=virtio -net user,vlan=0,hostname=${HOSTNAME}"
NETWORK="${NETWORK} -netdev tap,id=t0,ifname=${HOSTNAME},script=no,downscript=no -device e1000,netdev=t0,id=nic0,mac=${MAC}"

HD=$(dirname $0)/state/${HOSTNAME}/hd.img
if [ ! -e ${HD} ]; then
    echo "You must prepare a hd file"
    exit 1
fi

#HD_OPTS="-drive if=virtio,file=${HD},format=qcow2"
HD_OPTS="-hda ${HD}"

#USER_DATA_IMG="-drive file=$(dirname $0)/state/${HOSTNAME}/user-data.img,format=raw"
USER_DATA_IMG="-hdb $(dirname $0)/state/${HOSTNAME}/user-data.img"

exec qemu-system-${QEMUARCH} \
        -boot c \
        -nographic \
        -serial mon:stdio \
        -display none \
        -rtc base=utc,clock=host \
        ${MACHINE} \
        ${CPU} \
        -m ${MEMORY} \
        ${NETWORK} \
        ${HD_OPTS} \
        ${USER_DATA_IMG} \
        -smp 1 \
        -device virtio-rng-pci

