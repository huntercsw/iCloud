package commons

import (
	"github.com/coreos/etcd/clientv3"
	"github.com/docker/docker/client"
	"time"
)

const (
	GB                     uint64 = 1024 * 1024 * 1024
	CONTAINER_STOP_TIMEOUT        = time.Second * 5
	ETCD_KEY_PRE                  = "/iCloud/host_info/"
	ETCD_TIMEOUT                  = 100
)

var (
	DockerApiCli *client.Client
	EtcdCli      *clientv3.Client
)
