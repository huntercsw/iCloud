package commons

type Host struct {
	HostName  string `json:"hostName"`
	OS        string `json:"os"`
	Ip        string `json:"ip"`
	ApiPort   string `json:"apiPort"`
	CpuCores  int    `json:"cpuCores"`
	CpuUsage  string `json:"cpuUsage"`
	TotalMem  string    `json:"totalMem"`
	FreeMem   string    `json:"freeMem"`
	TotalDisk int    `json:"totalDisk"`
	FreeDisk  int    `json:"freeDisk"`
	Heartbeat int64  `json:"heartbeat"`		// current  timestamp
	GrpcPort  string `json:"grpcPort"`
}

