package apps

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/gin-gonic/gin"
	"iCloud/commons"
	"iCloud/log"
	"net/http"
	"sync"
	"time"
)

func HostList(ctx *gin.Context) {
	var (
		rsp = make(gin.H)
		wg  = sync.WaitGroup{}
	)
	wg.Add(1)

	etcdHandlerCtx, etcdHandlerCancel := context.WithTimeout(context.TODO(), time.Second*2)
	defer etcdHandlerCancel()

	go hostListHandler(etcdHandlerCtx, &wg, rsp)
	wg.Wait()

	ctx.JSON(http.StatusOK, rsp)
}

func hostListHandler(ctx context.Context, wg *sync.WaitGroup, rsp gin.H) {
	var (
		getRsp *clientv3.GetResponse
		err    error
		m      = "apps.hosts.hostListHandler()"
		rspCh  = make(chan struct{}, 1)
		host = new(commons.Host)
		data = make([]*commons.Host, 0)
	)
	defer func() {
		close(rspCh)
		wg.Done()
		fmt.Println("*****************************************")
	}()

	go func() {
		if getRsp, err = commons.EtcdCli.Get(context.TODO(), commons.ETCD_KEY_PRE, clientv3.WithPrefix()); err != nil {
			log.Logger.Errorf("%s error, get all host from etcd error: %v", m, err)
			rsp["ErrorCode"], rsp["Data"] = 1, "get host list error"
		} else {
			rsp["ErrorCode"] = 0
			for _, v := range getRsp.Kvs {
				if err = json.Unmarshal(v.Value, host); err != nil {
					log.Logger.Error("%s error, %s json unmarshal error: %v", m, v.Key, err)
				} else {
					data = append(data, host)
				}
			}
			rsp["Data"] = data
		}
		rspCh <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		rsp["ErrorCode"], rsp["Data"] = commons.ETCD_TIMEOUT, "etcd time out, check etcd and restart iCloud server"
	case <-rspCh:
	}

	return
}