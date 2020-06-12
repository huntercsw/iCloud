package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"
)

const (
	GB           uint64 = 1024 * 1024 * 1024
	ETCD_KEY_PRE        = "/iCloud/host_info/"
)

var (
	Conf     *ClientConf
	HostInfo *Host
	etcdCli  *clientv3.Client
)

type Host struct {
	HostName  string `json:"hostName"`
	OS        string `json:"os"`
	Ip        string `json:"ip"`
	ApiPort   string `json:"apiPort"`
	CpuCores  int    `json:"cpuCores"`
	CpuUsage  string `json:"cpuUsage"`
	TotalMem  string `json:"totalMem"`
	FreeMem   string `json:"freeMem"`
	TotalDisk int    `json:"totalDisk"`
	FreeDisk  int    `json:"freeDisk"`
	Heartbeat int64  `json:"heartbeat"` // current  timestamp
}

func init() {
	var (
		err error
	)

	Conf = new(ClientConf)
	Conf.newConf()

	InitLogger()

	if etcdCli, err = clientv3.New(clientv3.Config{
		Endpoints:   Conf.Etcd,
		DialTimeout: 2 * time.Second,
	}); err != nil {
		fmt.Println("etcd init error:", err)
		os.Exit(1)
	}
}

func (h *Host) Refresh() {
	var (
		m           string = "client.Refresh()"
		hostInfo    *host.InfoStat
		cpuTimeInfo []float64
		memInfo     *mem.VirtualMemoryStat
		diskInfo    []disk.PartitionStat
		err         error
	)

	h.empty()

	if hostInfo, err = host.Info(); err != nil {
		Logger.Errorf("%s error, %s[%s] get host info error: %v", m, h.HostName, h.Ip, err)
		return
	}
	if cpuTimeInfo, err = cpu.Percent(time.Second, false); err != nil {
		Logger.Errorf("%s error, %s[%s] get cpu times error: %v", m, h.HostName, h.Ip, err)
		return
	}
	if memInfo, err = mem.VirtualMemory(); err != nil {
		Logger.Errorf("%s error, %s[%s] get mem error: %v", m, h.HostName, h.Ip, err)
		return
	}

	h.HostName, h.OS = hostInfo.Hostname, hostInfo.OS
	h.Ip, h.ApiPort = Conf.ExportIp, Conf.ApiPort
	h.CpuCores = runtime.NumCPU()
	h.CpuUsage = strconv.FormatFloat(cpuTimeInfo[0], 'f', 2, 64)
	h.TotalMem = strconv.FormatFloat(float64(memInfo.Total)/float64(GB), 'f', 2, 64)
	h.FreeMem = strconv.FormatFloat(float64(memInfo.Available)/float64(GB), 'f', 2, 64)

	if h.OS == "windows" {
		if diskInfo, err = disk.Partitions(true); err != nil {
			Logger.Errorf("%s error, %s[%s] get disk error: %v", m, h.HostName, h.Ip, err)
			return
		}
		h.sumDisk(&diskInfo)
	} else {
		if diskStat, err1 := disk.Usage("/"); err1 != nil {
			Logger.Error("%s error, %s[%s] get disk '/' error: %v", m, h.HostName, h.Ip, err)
		} else {
			h.TotalDisk, h.FreeDisk = int(diskStat.Total/GB), int(diskStat.Free/GB)
		}
	}
	h.Heartbeat = time.Now().Unix()
}

func (h *Host) sumDisk(partitions *[]disk.PartitionStat) {
	var (
		total, free uint64
		err         error
		part        *disk.UsageStat
		m           = "client.sumDisk()"
	)
	for _, partition := range *partitions {
		if part, err = disk.Usage(partition.Device); err != nil {
			Logger.Errorf("%s error, %s get disk usage stat error: %v", m, partition.Device, err)
			return
		}
		total += part.Total
		free += part.Free
	}

	h.TotalDisk, h.FreeDisk = int(total/GB), int(free/GB)
}

func (h *Host) empty() {
	h.HostName = ""
	h.OS = ""
	h.Ip = ""
	h.ApiPort = ""
	h.CpuCores = 0
	h.CpuUsage = ""
	h.TotalMem = ""
	h.FreeMem = ""
	h.TotalDisk = 0
	h.FreeDisk = 0
}

func hostRegister() {
	var (
		err error
		h   []byte
		m   string = "client.HostRegister()"
	)

	HostInfo.Refresh()

	if h, err = json.Marshal(HostInfo); err != nil {
		Logger.Errorf("%s error, %s json marshal error: %v", m, HostInfo.Ip, err)
		return
	}

	if _, err = etcdCli.Put(context.TODO(), ETCD_KEY_PRE+HostInfo.Ip, string(h)); err != nil {
		Logger.Errorf("%s error, %s put to etcd error: %v", m, HostInfo.Ip, err)
	}
}

func main() {
	Logger.Info("iCloud Client started")
	defer func() {
		if err := recover(); err != nil {
			Logger.Errorf("client.main() panic: %v", err)
		}
		Logger.Sync()
	}()

	wg := sync.WaitGroup{}
	wg.Add(1)

	HostInfo = new(Host)

	go func() {
		for {
			hostRegister()
			time.Sleep(time.Second * 3)
		}
	}()

	wg.Wait()
}
