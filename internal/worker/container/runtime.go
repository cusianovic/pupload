package container

import (
	"bytes"
	"context"
	"strings"

	"github.com/moby/moby/api/pkg/stdcopy"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

type IContainerRuntime interface {
	CreateContainer()
	StartContainer()
	WaitContainer()
	RemoveContainer()
}

type ContainerRuntime struct {
	client  *client.Client
	runtime string
}

type ContainerConfig struct {
	Image string
	Name  string
	Cmd   []string
	Env   []string

	HostConfig *container.HostConfig
}

func (c *ContainerRuntime) CreateContainer(ctx context.Context, cfg ContainerConfig) (string, error) {

	if cfg.HostConfig == nil {
		cfg.HostConfig = &container.HostConfig{}
	}

	cfg.HostConfig.Runtime = c.runtime

	res, err := c.client.ContainerCreate(ctx, client.ContainerCreateOptions{
		Config: &container.Config{
			Env: cfg.Env,
			Cmd: cfg.Cmd,
		},

		HostConfig: cfg.HostConfig,
		Image:      cfg.Image,
		Name:       cfg.Name,
	})

	return res.ID, err
}

func (c *ContainerRuntime) StartContainer(ctx context.Context, containerID string) error {
	_, err := c.client.ContainerStart(ctx, containerID, client.ContainerStartOptions{})
	return err
}

type ContainerResult struct {
	ExitCode int
	Error    string
}

func (c *ContainerRuntime) WaitContainer(ctx context.Context, containerID string) (ContainerResult, error) {
	wait := c.client.ContainerWait(ctx, containerID, client.ContainerWaitOptions{
		Condition: container.WaitConditionNextExit,
	})

	select {
	case <-ctx.Done():
		return ContainerResult{}, ctx.Err()
	case err := <-wait.Error:
		return ContainerResult{}, err
	case res := <-wait.Result:
		return ContainerResult{ExitCode: int(res.StatusCode), Error: ""}, nil
	}
}

func (c *ContainerRuntime) RemoveContainer(ctx context.Context, containerID string) error {
	_, err := c.client.ContainerRemove(ctx, containerID, client.ContainerRemoveOptions{
		Force: true,
	})

	return err
}

func (c *ContainerRuntime) KillContainer(ctx context.Context, containerID string) error {
	_, err := c.client.ContainerKill(ctx, containerID, client.ContainerKillOptions{})
	return err
}

func (c *ContainerRuntime) GetLogs(ctx context.Context, containerID string) ([]string, error) {
	stream, err := c.client.ContainerLogs(ctx, containerID, client.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	})
	if err != nil {
		return nil, err
	}
	defer stream.Close()

	var logBuf bytes.Buffer
	if _, err := stdcopy.StdCopy(&logBuf, &logBuf, stream); err != nil {
		return nil, err
	}

	logString := logBuf.String()
	logArray := strings.Split(strings.TrimSpace(logString), "\n")

	return logArray, nil
}
