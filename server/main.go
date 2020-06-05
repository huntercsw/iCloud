package main

import (
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/docker/docker/client"
	"github.com/gin-gonic/gin"
	"iCloud/commons"
	"iCloud/conf"
	"iCloud/docker"
	"iCloud/log"
	"io"
	"os"
	"strconv"
)

var (
	DockerApiCli *client.Client
	EtcdCli *clientv3.Client
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

	if EtcdCli, err = commons.EtcdInit(); err != nil {
		log.Logger.Errorf("etcd init error: %v", err)
	}

	var ok bool
	if DockerApiCli, ok = docker.RemoteDockerApiInit(); !ok {
		os.Exit(1)
	}
}

func main() {
	log.Logger.Info("iCloud server started")

	defer func() {
		DockerApiCli.Close()
		log.Logger.Info("iCloud server closed")
		log.Logger.Sync()
	}()

	f, _ := os.Create(conf.Iconf.Log.WebLogName)
	defer f.Close()
	gin.DefaultWriter = io.MultiWriter(f)

	// 如果需要同时将日志写入文件和控制台，请使用以下代码。
	gin.DefaultWriter = io.MultiWriter(f, os.Stdout)

	ginEngine := gin.Default()

	hostRouters := ginEngine.Group("/hosts")
	{
		hostRouters.GET("/list")
	}
	dockerConfigRouters := ginEngine.Group("/containers")
	{
		dockerConfigRouters.GET("/list")
		dockerConfigRouters.POST("/add")
	}

	ginEngine.Run(conf.Iconf.Ip+":"+strconv.Itoa(conf.Iconf.Port))
}


