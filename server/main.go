package main

import (
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/gin-gonic/gin"
	"iCloud/apps"
	"iCloud/commons"
	"iCloud/conf"
	"iCloud/log"
	"io"
	"os"
	"os/signal"
	"strconv"
	"time"
)

func init() {
	var (
		err error
	)
	if err = conf.ICloudConfInit(); err != nil {
		fmt.Println("load configuration error:", err)
		os.Exit(1)
	}
	fmt.Println(conf.Iconf)

	log.InitLogger()

	if commons.EtcdCli, err = clientv3.New(clientv3.Config{
		Endpoints:   conf.Iconf.Etcd,
		DialTimeout: 2 * time.Second,
	}); err != nil {
		log.Logger.Error("etcd init error: %v", err)
	}

	apps.DockerApiCliMapInit()
}

func main() {
	log.Logger.Info("iCloud server started")

	defer func() {
		apps.DockerApiCliPoolClose()
		log.Logger.Info("iCloud server closed")
		log.Logger.Sync()
	}()

	f, _ := os.Create(conf.Iconf.Log.WebLogName)
	defer f.Close()
	gin.DefaultWriter = io.MultiWriter(f)

	// 如果需要同时将日志写入文件和控制台，请使用以下代码。
	gin.DefaultWriter = io.MultiWriter(f, os.Stdout)

	ginEngine := gin.Default()

	ICloudRouter(ginEngine)

	go ginEngine.Run(conf.Iconf.Ip + ":" + strconv.Itoa(conf.Iconf.Port))

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill)
	s := <-c
	fmt.Println(s)
}
