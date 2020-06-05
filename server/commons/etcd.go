package commons

import (
	"github.com/coreos/etcd/clientv3"
	"iCloud/conf"
	"time"
)

func EtcdInit() (cli *clientv3.Client, err error) {
	if cli, err = clientv3.New(clientv3.Config{
		Endpoints: conf.Iconf.Etcd,
		DialTimeout: 2 * time.Second,
	}); err != nil {
		return
	}
	return
}


