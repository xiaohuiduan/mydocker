package util

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net"
)

type IPAllocation struct {
	Subnet string         `json:"subnet"`
	Ip     map[string]int `json:"ip"`
	GetWay string         `json:"getway"`
}

func _must(err error) {
	if err != nil {
		panic(err)
	}
}

func AllocationIp(allocation *IPAllocation) string {
	subnet := InetAtoN(allocation.Subnet)
	startIp := InetAtoN(allocation.GetWay)
	// ip可用数量
	availableNum := 1<<32 - subnet - 1 //减去255
	var i int64
	for i = 2; i < availableNum; i++ {
		newIpInt := startIp + i

		newIp := InetNtoA(newIpInt)
		if allocation.Ip[newIp] == 0 {
			return newIp
		}
	}
	return ""
}

func InetNtoA(ip int64) string {
	// int转ip
	return fmt.Sprintf("%d.%d.%d.%d",
		byte(ip>>24), byte(ip>>16), byte(ip>>8), byte(ip))
}

func InetAtoN(ip string) int64 {
	// ip转成int
	ret := big.NewInt(0)
	ret.SetBytes(net.ParseIP(ip).To4())
	return ret.Int64()
}

func NewIPAllocation(jsonpath string) *IPAllocation {
	// 从文件中读取数据
	dataEncoded, _ := ioutil.ReadFile(jsonpath)
	var ipAllocation IPAllocation
	err := json.Unmarshal(dataEncoded, &ipAllocation)
	_must(err)
	return &ipAllocation
}

func WriteIPAllocationToFile(ipAllocation *IPAllocation, jsonpath string) {
	// 将数据保存到json文件中
	data, _ := json.Marshal(ipAllocation)

	// 将 JSON 格式数据写入当前目录下的d books 文件（文件不存在会自动创建）
	err := ioutil.WriteFile(jsonpath, data, 0644)
	if err != nil {
		panic(err)
	}
}

//func test() {
//	iPAllocation := NewIPAllocation("ip.json")
//	a := allocationIp(iPAllocation)
//	fmt.Println(a)
//	iPAllocation.Ip[a] = 1
//	WriteIPAllocationToFile(iPAllocation, "ip.json")
//
//}
