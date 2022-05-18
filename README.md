## 运行方法

1. 首先运行 `go run mydocker.go daemon`：创建一个网桥
2. 然后对仓库中resource中的busybox.tar进行解压，解压到resource中busybox文件夹中
3. 运行`go run mydocker.go run /bin/sh`

## 注意
以上方法都需要在root模式运行，同时对于config中的sh脚本，需要赋予执行权限：`sudo chmod +x *.sh`

