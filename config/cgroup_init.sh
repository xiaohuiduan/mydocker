#! /bin/sh

CFS_PERIOD_US=100000
CFS_QUOTA_US=50000
CPU_SHARE=4096    # default 1024

mkdir /sys/fs/cgroup/cpu/$1 # $1 = __demo__
cd /sys/fs/cgroup/cpu/$1

echo $CFS_PERIOD_US > cpu.cfs_period_us
echo $CFS_QUOTA_US > cpu.cfs_quota_us
echo $CPU_SHARE > cpu.shares
