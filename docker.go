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
	"github.com/docker/docker/api/types/mount"
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
}

func create(ctx context.Context, eec EventCreated, volumes map[string]string) error {

	portsMapping := make(nat.PortMap)
	for hostas, vm := range eec.Ports {
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

	reader, err := cli.ImagePull(ctx, eec.Image, types.ImagePullOptions{RegistryAuth: auth(ctx, eec.Auth.Username, eec.Auth.Password)})
	if err != nil {
		return (err)
	}
	io.Copy(io.Discard, reader)

	mounts := []mount.Mount{}

	for src, dst := range eec.Volumes {
		mounts = append(mounts, mount.Mount{
			Type:   "volume",
			Source: src,
			Target: dst,
		})
	}
	for src, dst := range eec.Binds {
		mounts = append(mounts, mount.Mount{
			Type:   "bind",
			Source: src,
			Target: dst,
		})
	}

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:  eec.Image,
		Cmd:    eec.Cmd,
		Env:    eec.Env,
		Labels: eec.Labels,
	}, &container.HostConfig{
		PortBindings: portsMapping,
		Mounts:       mounts,
		ExtraHosts:   []string{"gateway:host-gateway"},
	}, nil, nil, eec.Name)
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
