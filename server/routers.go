package main

import (
	"github.com/gin-gonic/gin"
	"iCloud/apps"
)

func ICloudRouter(r *gin.Engine) {
	HostRouters := r.Group("/iCloudApi/hosts")
	{
		HostRouters.GET("/list", apps.HostList)
	}

	DockerConfigRouters := r.Group("/iCloudApi/containers")
	{
		DockerConfigRouters.GET("/list", apps.ContainerList)
		DockerConfigRouters.GET("/lsImage", apps.ImageList)
		DockerConfigRouters.POST("/createAndRun/:ip/:port", apps.ContainerCreate)
		DockerConfigRouters.PUT("/start/:id/:ip/:port", apps.ContainerStart)
		DockerConfigRouters.PUT("/stop/:id/:ip/:port", apps.ContainerStop)
		DockerConfigRouters.DELETE("/remove/:id/:ip/:port", apps.ContainerRemove)
		DockerConfigRouters.GET("/detail/:id/:ip/:port", apps.ContainerDetail)
	}

	DockerLogRouters := r.Group("/iCloudApi/logs")
	{
		DockerLogRouters.POST("/:id/:ip/:port", apps.ContainerLogs)
	}
}

