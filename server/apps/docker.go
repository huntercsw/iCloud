package apps

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"iCloud/log"
	"time"
)

const RemoteHost = "tcp://192.168.0.100:7777"

var (
	DockerApiCliMap = make(map[string]*client.Client)
)

func DockerApiCliPoolAdd(ip string) (err error) {
	var (
		cli *client.Client
		m = "apps.docker.DockerApiCliAdd()"
	)
	if _, exist := DockerApiCliMap[ip]; exist {
		return
	}

	// host: it need remote docker server open remote API, edit configuration "/usr/lib/systemd/system/docker.service"
	if cli, err = client.NewClient("tcp://" + ip, client.DefaultVersion, nil, nil); err != nil {
		log.Logger.Error("%s error, create docker api client to %s error: %v", m, ip, err)
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
		m = "apps.docker.DockerApiCliMapReconnect()"
		cli *client.Client
	)
	if _, exist := DockerApiCliMap[ip]; exist {
		DockerApiCliMap[ip].Close()
	}

	if cli, err = client.NewClient("tcp://" + ip, client.DefaultVersion, nil, nil); err != nil {
		log.Logger.Error("%s error, create docker api client to %s error: %v", m, ip, err)
		return
	}

	DockerApiCliMap[ip] = cli
	return
}

func GetDockerApiClient(ip string) (cli *client.Client) {
	if _, exist := DockerApiCliMap[ip]; !exist {
		return
	}

	return DockerApiCliMap[ip]
}

func DockerApiCliPoolClose() {
	for _, cli := range DockerApiCliMap {
		cli.Close()
	}
}

func RemoteDockerApiInit() (cli *client.Client, ok bool){
	var (
		err error
	)
	// host: it need remote docker server open remote API, edit configuration "/usr/lib/systemd/system/docker.service"
	if cli, err = client.NewClient(RemoteHost, client.DefaultVersion, nil, nil); err != nil {
		fmt.Println("connect to remote docker api error:", err)
		return nil, false
	}
	return cli, true
}

func LsContainers(cli *client.Client) {
	var (
		containers []types.Container
		err        error
	)

	if containers, err = cli.ContainerList(context.Background(), types.ContainerListOptions{All:true}); err != nil {
		fmt.Println("get running containers error:", err)
		return
	}
	for _, container := range containers {
		fmt.Printf("%s %s\n", container.ID, container.Image)
	}
}

func CreateRunContainer(cli *client.Client) (containerId string, ok bool) {
	var (
		exportPortMap = make(nat.PortSet, 10)
		exportPort    nat.Port
		err           error
		config        *container.Config
		portBind      nat.PortBinding
		bindPortMap   = make(nat.PortMap, 0)
		hostConfig    *container.HostConfig
		_container    container.ContainerCreateCreatedBody
	)
	// create port(port of container)
	if exportPort, err = nat.NewPort("tcp", "8888"); err != nil {
		fmt.Println("create export port error:", err)
		return "", false
	}
	exportPortMap[exportPort] = struct{}{}

	// add port to containers configuration
	config = &container.Config{Image: "nginx", ExposedPorts: exportPortMap}

	// redirect to local port
	portBind = nat.PortBinding{HostPort: "8090"}

	tmp := make([]nat.PortBinding, 0, 1)
	tmp = append(tmp, portBind)

	bindPortMap[exportPort] = tmp

	hostConfig = &container.HostConfig{PortBindings: bindPortMap}

	containerName := "nginx_test"

	// mount local dir to container: local_dir:container_dir
	hostConfig.Binds = []string{"/root/docker_api_test/nginx_conf:/etc/nginx/conf.d", "/crontab:/crontab", "/root/docker_api_test/nginx_log:/var/log/nginx"}
	if _container, err = cli.ContainerCreate(context.Background(), config, hostConfig, nil, containerName); err != nil {
		fmt.Println("create container error:", err)
		return "", false
	}

	return _container.ID, true
}

func StartContainer(containerId string, cli *client.Client) {
	if err := cli.ContainerStart(context.Background(), containerId, types.ContainerStartOptions{}); err != nil {
		fmt.Println("start container error id:", containerId, err)
	}

	fmt.Println(containerId, "started")
}

func StopContainer(containerID string, cli *client.Client) {
	timeout := time.Second * 10
	err := cli.ContainerStop(context.Background(), containerID, &timeout)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("容器%s已经被停止\n", containerID)
	}
}

func RemoveContainer(containerID string, cli *client.Client) (string, error) {
	err := cli.ContainerRemove(context.Background(), containerID, types.ContainerRemoveOptions{})
	fmt.Println(err)
	return containerID, err
}

