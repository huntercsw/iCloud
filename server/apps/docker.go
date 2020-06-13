package apps

import (
	"bytes"
	"context"
	"errors"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/gin-gonic/gin"
	"iCloud/commons"
	"iCloud/log"
	"io"
	"net/http"
	"strconv"
	"strings"
)

const RemoteDockerPort = ":7777"

var (
	DockerApiCliMap map[string]*client.Client
)

type ContainerConfiguration struct {
	ContainerName string   `json:"container_name"`
	SourceDir     []string `json:"source_dir"`
	ContainerPort []string `json:"container_port"`
	HostPort      []string `json:"host_port"`
	ImageName     string   `json:"container_image"`
	MaxCpu        string   `json:"maxCpu"`
	MaxMem        string   `json:"maxMem"`
}

func (conf *ContainerConfiguration) confCheck() (err error) {
	if conf.ContainerName == "" {
		err = errors.New("container name is null")
		return
	}

	if len(conf.ContainerPort) != len(conf.HostPort) {
		err = errors.New("count of port exported in container is not equal to it redirected to host")
		return
	}

	if conf.ImageName == "" {
		err = errors.New("image name is null")
		return
	}

	if c, err1 := strconv.ParseFloat(conf.MaxCpu, 64); err1 != nil {
		log.Logger.Errorf("MaxCpu from string to int error: %v", err1)
		err = errors.New("type of MaxCpu is not number")
		return
	} else {
		if c <= 0 {
			err = errors.New("MaxCpu is not bigger than 0")
			return
		}
	}

	if c, err1 := strconv.ParseFloat(conf.MaxMem, 64); err1 != nil {
		log.Logger.Errorf("MaxMem from string to int error: %v", err1)
		err = errors.New("type of MaxMem is not number")
		return
	} else {
		if c <= 0 {
			err = errors.New("MaxMem is not bigger than 0")
			return
		}
	}

	return
}

func DockerApiCliMapInit() {
	DockerApiCliMap = make(map[string]*client.Client)
}

func DockerApiCliPoolAdd(ip, port string) (err error) {
	var (
		cli *client.Client
		m   = "apps.docker.DockerApiCliAdd()"
	)
	if _, exist := DockerApiCliMap[ip]; exist {
		return
	}

	// host: it need remote docker server open remote API, edit configuration "/usr/lib/systemd/system/docker.service"
	if cli, err = client.NewClient("tcp://"+ip+":"+port, client.DefaultVersion, nil, nil); err != nil {
		log.Logger.Errorf("%s error, create docker api client to %s:%s error: %v", m, ip, port, err)
		return
	}

	DockerApiCliMap[ip] = cli
	return
}

func DockerApiCliMapDelete(ip string) {
	if _, exist := DockerApiCliMap[ip]; exist {
		DockerApiCliMap[ip].Close()
		delete(DockerApiCliMap, ip)
	}
}

func DockerApiCliMapReconnect(ip string) (err error) {
	var (
		m   = "apps.docker.DockerApiCliMapReconnect()"
		cli *client.Client
	)
	if _, exist := DockerApiCliMap[ip]; exist {
		DockerApiCliMap[ip].Close()
	}

	if cli, err = client.NewClient("tcp://"+ip, client.DefaultVersion, nil, nil); err != nil {
		log.Logger.Errorf("%s error, create docker api client to %s error: %v", m, ip, err)
		return
	}

	DockerApiCliMap[ip] = cli
	return
}

func DockerApiCliPoolClose() {
	for _, cli := range DockerApiCliMap {
		cli.Close()
	}
}

func containerConfInit(conf *ContainerConfiguration) (containerConf *container.Config, err error) {
	var (
		port             nat.Port
		containerPortSet = make(nat.PortSet)
		m                = "apps.docker.containerConfInit()"
	)

	// ports container exported
	if len(conf.ContainerPort) > 0 {
		for _, p := range conf.ContainerPort {
			if port, err = nat.NewPort("tcp", p); err != nil {
				log.Logger.Errorf("%s error, export port in container error: %v", m, err)
				return
			}
			containerPortSet[port] = struct{}{}
		}
	}

	return &container.Config{Image: conf.ImageName, ExposedPorts: containerPortSet}, nil
}

func hostConfInit(conf *ContainerConfiguration) (hostConf *container.HostConfig, err error) {
	var (
		portCount    = len(conf.ContainerPort)
		exportPort   nat.Port
		portBind     nat.PortBinding
		bindPortMap  = make(nat.PortMap)
		mountConf    = make([]string, 0)
		resourceConf *container.Resources
		m            = "apps.docker.HostConfInit()"
	)

	// container port redirect to host port
	if portCount > 0 {
		for i := 0; i < portCount; i++ {
			if exportPort, err = nat.NewPort("tcp", conf.ContainerPort[i]); err != nil {
				log.Logger.Errorf("%s error, export port in container error: %v", m, err)
				return
			}
			portBind = nat.PortBinding{HostPort: conf.HostPort[i]}
			bindPortMap[exportPort] = make([]nat.PortBinding, 0, 1)
			bindPortMap[exportPort] = append(bindPortMap[exportPort], portBind)
		}
	}

	// mount host dir to container: local_dir:container_dir
	if len(conf.SourceDir) > 0 {
		for _, dir := range conf.SourceDir {
			mountConf = append(mountConf, dir+":"+dir)
		}
	}
	// mount local /etc/localtime to container /etc/localtime, to make time in container is equal to host
	mountConf = append(mountConf, "/etc/localtime")

	if resourceConf, err = resourceConfInit(conf); err != nil {
		err = errors.New("resource config init error")
		return
	}

	return &container.HostConfig{
		PortBindings: bindPortMap,
		Binds:        mountConf,
		Resources:    *resourceConf,
	}, nil
}

func resourceConfInit(conf *ContainerConfiguration) (resourceConf *container.Resources, err error) {
	var (
		mem       float64
		cpuPeriod = float64(100000)
		cpuQuota  float64
		m         = "apps.docker.resourceConfInit()"
	)

	if mem, err = strconv.ParseFloat(conf.MaxMem, 64); err != nil {
		log.Logger.Errorf("%s error, mem to int error: %v", m, err)
		return
	}

	if cpuQuota, err = strconv.ParseFloat(conf.MaxCpu, 64); err != nil {
		log.Logger.Errorf("%s error, cpuQuota to int error: %v", m, err)
		return
	}

	return &container.Resources{
		Memory:    int64(mem * float64(commons.GB)),
		CPUPeriod: int64(cpuPeriod),
		CPUQuota:  int64(cpuQuota * cpuPeriod),
	}, nil
}

func createContainer(cli *client.Client, conf *ContainerConfiguration) (containerId string, err error) {
	var (
		configObj  *container.Config
		hostConfig *container.HostConfig
		_container container.ContainerCreateCreatedBody
		m          = "apps.docker.createRunContainer()"
	)

	if configObj, err = containerConfInit(conf); err != nil {
		err = errors.New("container config object init error")
		return
	}

	if hostConfig, err = hostConfInit(conf); err != nil {
		err = errors.New("host config object init error")
		return
	}

	if _container, err = cli.ContainerCreate(context.Background(), configObj, hostConfig, nil, conf.ContainerName); err != nil {
		log.Logger.Errorf("%s error, create container error: %v", m, err)
		err = errors.New("create container error")
		return
	}

	return _container.ID, nil
}

func startContainer(containerId string, cli *client.Client) (err error) {
	var (
		m = "apps.docker.startContainer()"
	)

	if err = cli.ContainerStart(context.Background(), containerId, types.ContainerStartOptions{}); err != nil {
		log.Logger.Errorf("%s error, start container %s error: %v", m, containerId, err)
		return errors.New("start container error")
	}

	return
}

func stopContainer(containerID string, cli *client.Client) (err error) {
	var (
		m       = "apps.docker.stopContainer()"
		timeout = commons.CONTAINER_STOP_TIMEOUT
	)

	if err = cli.ContainerStop(context.Background(), containerID, &timeout); err != nil {
		log.Logger.Errorf("%s error, stop container %s error: %v", m, containerID, err)
		return errors.New("stop container error")
	}

	return
}

func removeContainer(containerID string, cli *client.Client) (err error) {
	var (
		m = "apps.docker.removeContainer()"
	)

	if err = stopContainer(containerID, cli); err != nil {
		return
	}

	if err = cli.ContainerRemove(context.Background(), containerID, types.ContainerRemoveOptions{}); err != nil {
		log.Logger.Errorf("%s error, remove container %s error: %v", m, containerID, err)
		return errors.New("remove container error")
	}

	return
}

func ContainerList(ctx *gin.Context) {
	var (
		rsp        = make(gin.H)
		err        error
		containers []types.Container
		m          = "apps.docker.ContainerList()"
		exist      bool
	)
	hostIp, remotePort := ctx.Query("ip"), ctx.Query("port")
	if hostIp == "" || remotePort == "" {
		rsp["ErrorCode"], rsp["Data"] = 1, "param error, host ip and port are required in url"
		goto RESPONSE
	}

	if _, exist = DockerApiCliMap[hostIp]; !exist {
		if err = DockerApiCliPoolAdd(hostIp, remotePort); err != nil {
			rsp["ErrorCode"], rsp["Data"] = 1, "create connection to docker api on "+hostIp+":"+remotePort+" error"+err.Error()
			goto RESPONSE
		}
	}

	if containers, err = DockerApiCliMap[hostIp].ContainerList(context.Background(), types.ContainerListOptions{All: true}); err != nil {
		log.Logger.Errorf("%s error, list all containers on host[%s] error: %v", m, hostIp, err)
		rsp["ErrorCode"], rsp["Data"] = 1, "list containers error"
		goto RESPONSE
	}

	rsp["ErrorCode"], rsp["Data"] = 0, containers

RESPONSE:
	ctx.JSON(http.StatusOK, rsp)
}

func ImageList(ctx *gin.Context) {
	var (
		rsp    = make(gin.H)
		err    error
		exist  bool
		images []types.ImageSummary
		m      = "apps.docker.ImageList()"
	)
	hostIp, remotePort := ctx.Query("ip"), ctx.Query("port")
	if hostIp == "" || remotePort == "" {
		rsp["ErrorCode"], rsp["Data"] = 1, "param error, host ip and port are required in url"
		goto RESPONSE
	}

	if _, exist = DockerApiCliMap[hostIp]; !exist {
		if err = DockerApiCliPoolAdd(hostIp, remotePort); err != nil {
			rsp["ErrorCode"], rsp["Data"] = 1, "create connection to docker api on "+hostIp+":"+remotePort+" error"+err.Error()
			goto RESPONSE
		}
	}

	if images, err = DockerApiCliMap[hostIp].ImageList(context.TODO(), types.ImageListOptions{All: true}); err != nil {
		log.Logger.Errorf("%s error, get images from %s error: %v", m, hostIp, err)
		rsp["ErrorCode"], rsp["Data"] = 1, "get all images error"
		goto RESPONSE
	}

	rsp["ErrorCode"], rsp["Data"] = 0, images

RESPONSE:
	ctx.JSON(http.StatusOK, rsp)
}

func ContainerCreate(ctx *gin.Context) {
	var (
		err           error
		m             = "apps.docker.ContainerCreateAndRun()"
		containerConf = new(ContainerConfiguration)
		rsp           = make(gin.H)
		cli           *client.Client
		exist         bool
	)
	hostIp, remotePort := ctx.Param("ip"), ctx.Param("port")
	if hostIp == "" || remotePort == "" {
		rsp["ErrorCode"], rsp["Data"] = 1, "param error, host ip and port are required in url"
		goto RESPONSE
	}

	if err = ctx.BindJSON(containerConf); err != nil {
		log.Logger.Errorf("%s error, read request data error: %v", m, err)
		rsp["ErrorCode"], rsp["Data"] = 1, "request data error"
		goto RESPONSE
	}

	if err = containerConf.confCheck(); err != nil {
		rsp["ErrorCode"], rsp["Data"] = 1, err.Error()
		goto RESPONSE
	}

	if cli, exist = DockerApiCliMap[hostIp]; !exist {
		if err = DockerApiCliPoolAdd(hostIp, remotePort); err != nil {
			rsp["ErrorCode"], rsp["Data"] = 1, "connect to remote docker api error"
			goto RESPONSE
		}
		cli, _ = DockerApiCliMap[hostIp]
	}

	if _, err = createContainer(cli, containerConf); err != nil {
		rsp["ErrorCode"], rsp["Data"] = 1, err.Error()
		goto RESPONSE
	}

	rsp["ErrorCode"], rsp["Data"] = 0, ""

RESPONSE:
	ctx.JSON(http.StatusOK, rsp)
}

func ContainerStart(ctx *gin.Context) {
	var (
		rsp   = make(gin.H)
		cli   *client.Client
		exist bool
		err   error
	)
	ip, id := ctx.Param("ip"), ctx.Param("id")
	if cli, exist = DockerApiCliMap[ip]; !exist {
		rsp["ErrorCode"], rsp["Data"] = 1, "ip error"
		goto RESPONSE
	}

	if err = startContainer(id, cli); err != nil {
		rsp["ErrorCode"], rsp["Data"] = 1, err.Error()
		goto RESPONSE
	}

	rsp["ErrorCode"], rsp["Data"] = 0, ""
RESPONSE:
	ctx.JSON(http.StatusOK, rsp)
}

func ContainerStop(ctx *gin.Context) {
	var (
		rsp   = make(gin.H)
		cli   *client.Client
		exist bool
		err   error
	)
	ip, id := ctx.Param("ip"), ctx.Param("id")
	if cli, exist = DockerApiCliMap[ip]; !exist {
		rsp["ErrorCode"], rsp["Data"] = 1, "ip error"
		goto RESPONSE
	}

	if err = stopContainer(id, cli); err != nil {
		rsp["ErrorCode"], rsp["Data"] = 1, err.Error()
		goto RESPONSE
	}

	rsp["ErrorCode"], rsp["Data"] = 0, ""
RESPONSE:
	ctx.JSON(http.StatusOK, rsp)
}

func ContainerRemove(ctx *gin.Context) {
	var (
		rsp   = make(gin.H)
		cli   *client.Client
		exist bool
		err   error
	)
	ip, id := ctx.Param("ip"), ctx.Param("id")
	if cli, exist = DockerApiCliMap[ip]; !exist {
		rsp["ErrorCode"], rsp["Data"] = 1, "ip error"
		goto RESPONSE
	}

	if err = stopContainer(id, cli); err != nil {
		rsp["ErrorCode"], rsp["Data"] = 1, err.Error()
		goto RESPONSE
	}

	if err = removeContainer(id, cli); err != nil {
		rsp["ErrorCode"], rsp["Data"] = 1, err.Error()
		goto RESPONSE
	}

	rsp["ErrorCode"], rsp["Data"] = 0, ""
RESPONSE:
	ctx.JSON(http.StatusOK, rsp)
}

func ContainerDetail(ctx *gin.Context) {
	var (
		rsp    = make(gin.H)
		cli    *client.Client
		exist  bool
		err    error
		detail types.ContainerJSON
	)
	ip, id := ctx.Param("ip"), ctx.Param("id")
	if cli, exist = DockerApiCliMap[ip]; !exist {
		rsp["ErrorCode"], rsp["Data"] = 1, "ip error"
		goto RESPONSE
	}
	if detail, err = cli.ContainerInspect(context.TODO(), id); err != nil {
		log.Logger.Errorf("show container[%s] inspect error: %v", id, err)
		rsp["ErrorCode"], rsp["Data"] = 1, "get container inspect error"
		goto RESPONSE
	}

	rsp["ErrorCode"], rsp["Data"] = 0, detail
RESPONSE:
	ctx.JSON(http.StatusOK, rsp)
}

func ContainerLogs(ctx *gin.Context) {
	var (
		rsp          = make(gin.H)
		cli          *client.Client
		exist        bool
		err          error
		containerLog io.ReadCloser
		logOption    = new(types.ContainerLogsOptions)
		m            = "apps.docker.ContainerLogs"
		ip, id       = ctx.Param("ip"), ctx.Param("id")
		buf = new(bytes.Buffer)
	)

	if err = ctx.BindJSON(logOption); err != nil {
		log.Logger.Errorf("%s error, read container log options data error: %v", m, err)
		rsp["ErrorCode"], rsp["Data"] = 1, "request container log options error"
		goto RESPONSE
	}

	if cli, exist = DockerApiCliMap[ip]; !exist {
		rsp["ErrorCode"], rsp["Data"] = 1, "ip error"
		goto RESPONSE
	}

	if containerLog, err = cli.ContainerLogs(context.TODO(), id, *logOption); err != nil {
		log.Logger.Errorf("%s error, get container[%s] log error: %v", m, id, err)
		goto RESPONSE
	}

	buf.ReadFrom(containerLog)
	rsp["ErrorCode"], rsp["Data"] = 0, strings.Split(buf.String(), "\r\n")

RESPONSE:
	ctx.JSON(http.StatusOK, rsp)
}

