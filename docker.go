package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

func auth(ctx context.Context, username, password string) string {
	authStr := ""

	authConfig := types.AuthConfig{
		Username: username,
		Password: password,
	}
	encodedJSON, err := json.Marshal(authConfig)
	if err != nil {
		panic(err)
	}
	authStr = base64.URLEncoding.EncodeToString(encodedJSON)

	return authStr
	// out, err := cli.ImagePull(ctx, "alpine", types.ImagePullOptions{RegistryAuth: authStr})
}

func create(ctx context.Context, name, image_url string, ports map[string]string, env []string, labels map[string]string, cmd []string, username, password string) error {

	portsMapping := make(nat.PortMap)
	for hostas, vm := range ports {
		host, port, err := net.SplitHostPort(hostas)
		hostBinding := nat.PortBinding{}
		if err == nil {
			hostBinding.HostIP = host
			hostBinding.HostPort = port
		}
		parts := strings.Split(vm, "/")
		var containerPort nat.Port
		if len(parts) == 2 {
			containerPort, err = nat.NewPort(parts[1], parts[0])
			if err != nil {
				panic(err)
			}
		} else {
			containerPort, err = nat.NewPort("tcp", parts[0])
			if err != nil {
				panic(err)
			}
		}
		portsMapping[containerPort] = append(portsMapping[containerPort], hostBinding)
	}

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return (err)
	}

	reader, err := cli.ImagePull(ctx, image_url, types.ImagePullOptions{RegistryAuth: auth(ctx, username, password)})
	if err != nil {
		return (err)
	}
	io.Copy(io.Discard, reader)

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:  image_url,
		Cmd:    cmd,
		Env:    env,
		Labels: labels,
	}, &container.HostConfig{
		PortBindings: portsMapping,
	}, nil, nil, name)
	if err != nil {
		return (err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return (err)
	}

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)

	select {
	case err := <-errCh:
		if err != nil {
			return (err)
		}
	case <-statusCh:
	default:
	}

	return nil
}

func delete(ctx context.Context, name string) error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	dur, _ := time.ParseDuration("3s")
	cli.ContainerStop(ctx, name, &dur)
	if err := cli.ContainerRemove(ctx, name, types.ContainerRemoveOptions{}); err != nil {
		return err
	}
	return nil
}

func inspect(ctx context.Context, name string) (types.ContainerJSON, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	js, err := cli.ContainerInspect(ctx, name)
	if err != nil {
		return types.ContainerJSON{}, err
	}
	return js, nil
}

func listContainers(ctx context.Context) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{
		All: true,
	})
	if err != nil {
		panic(err)
	}

	for _, container := range containers {
		js, err := cli.ContainerInspect(ctx, container.ID)
		if err != nil {
			log.Printf("error :%v", err)
		}
		log.Printf("%s %s", js.Name, js.Config.Env)
	}
}
