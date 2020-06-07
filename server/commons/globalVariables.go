package commons

import (
	"github.com/coreos/etcd/clientv3"
	"github.com/docker/docker/client"
)

const (
	GB           uint64 = 1024 * 1024 * 1024
	ETCD_KEY_PRE        = "/iCloud/host_info/"
	ETCD_TIMEOUT        = 100
)

var (
	DockerApiCli *client.Client
	EtcdCli      *clientv3.Client
)
