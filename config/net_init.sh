#!/bin/sh
echo "初始化网络，新建转发和网桥"

iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE

echo "net.ipv4.ip_forward = 1" >> /etc/sysctl.conf

sysctl -p

#新增网桥
brctl addbr container_br


#配置网桥ip
ip addr add 10.0.0.1/24 dev container_br

#启动网桥
ip link set container_br up