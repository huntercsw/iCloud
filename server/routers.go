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
		DockerConfigRouters.POST("/createAndRun/:ip/:port/:rPort", apps.ContainerCreate)
		DockerConfigRouters.PUT("/start/:id/:ip/:port/:rPort", apps.ContainerStart)
		DockerConfigRouters.PUT("/stop/:id/:ip/:port/:rPort", apps.ContainerStop)
		DockerConfigRouters.DELETE("/remove/:id/:ip/:port/:rPort", apps.ContainerRemove)
		DockerConfigRouters.GET("/detail/:id/:ip/:port/:rPort", apps.ContainerDetail)
	}

	DockerLogRouters := r.Group("/iCloudApi/logs")
	{
		DockerLogRouters.POST("/:id/:ip/:port", apps.ContainerLogs)
	}
}

