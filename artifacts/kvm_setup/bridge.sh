#!/bin/bash


ip link add name br0 type bridge
ip addr add 172.20.0.1/16 brd + dev br0
ip link set br0 up

dnsmasq --interface=br0 --bind-interfaces --dhcp-range=172.20.0.2,172.20.255.254

# disable what docker does
sysctl net.bridge.bridge-nf-call-iptables=0

