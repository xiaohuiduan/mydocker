#!/bin/sh
echo "新建veth pair for ip："$1,$2 # $1代表container进程的ip

# 新建pair
ip link add $1-c type veth peer name $1-br

#启动 veth0-br 网卡
ip link set $1-br up
#将 veth0-br(网卡) 接入到 网桥(交换机) container_br 上
brctl addif container_br $1-br

#bind to container pid
ip link set $1-c netns $2