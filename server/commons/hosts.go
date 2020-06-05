package commons

import (
	"github.com/coreos/etcd/clientv3"
)

type Host struct {
	HostName  string `json:"hostName"`
	OS        string `json:"os"`
	Ip        string `json:"ip"`
	CpuCores  int    `json:"cpuCores"`
	CpuUsage  string `json:"cpuUsage"`
	TotalMem  string    `json:"totalMem"`
	FreeMem   string    `json:"freeMen"`
	TotalDisk int    `json:"totalDisk"`
	FreeDisk  int    `json:"freeDisk"`
	Heartbeat int64  `json:"heartbeat"`		// current  timestamp
}

func HostHeartbeat(cli *clientv3.Client) {

}
