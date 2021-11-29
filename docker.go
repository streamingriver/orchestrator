package main

import (
	"context"
	"io"
	"log"
	"net"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

func create(ctx context.Context, name, image_url string, ports map[string]string, env []string, labels map[string]string, cmd []string) error {

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

	// portBinding := nat.PortMap{containerPort: []nat.PortBinding{hostBinding}}

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return (err)
	}

	reader, err := cli.ImagePull(ctx, image_url, types.ImagePullOptions{})
	if err != nil {
		return (err)
	}
	io.Copy(io.Discard, reader)

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: image_url,
		Cmd:   cmd,
		// ExposedPorts: portsMapping,
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
	}

	out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		return (err)
	}
	defer out.Close()

	// stdcopy.StdCopy(os.Stdout, os.Stderr, out)
	io.Copy(io.Discard, out)
	return nil
}

func delete(ctx context.Context, name string) error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
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
