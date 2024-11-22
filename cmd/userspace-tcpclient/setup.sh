#! /bin/bash

ip tuntap add user viveksb007 mode tun tun0
ip link set tun0 up
ip addr add 192.168.1.1/24 dev tun0

## To route traffic from TUN to Internet
sysctl -w net.ipv4.ip_forward=1
iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE
