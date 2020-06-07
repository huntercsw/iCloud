package main

import (
	"github.com/gin-gonic/gin"
	"iCloud/apps"
)

func ICloudRouter(r *gin.Engine) {
	HostRouters := r.Group("/hosts")
	{
		HostRouters.GET("/list", apps.HostList)
	}
	DockerConfigRouters := r.Group("/containers")
	{
		DockerConfigRouters.GET("/list")
		DockerConfigRouters.POST("/add")
	}
}

