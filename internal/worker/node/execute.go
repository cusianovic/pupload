package node

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"pupload/internal/logging"
	"pupload/internal/syncplane"
	"strings"

	"github.com/hibiken/asynq"
	"github.com/moby/moby/api/pkg/stdcopy"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

func (n NodeService) HandleNodeExecuteTask(ctx context.Context, t *asynq.Task) error {

	l := logging.LoggerFromCtx(ctx)

	var p syncplane.NodeExecutePayload

	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return err
	}

	capable := n.CanWorkerRunContainer(p.NodeDef)
	if !capable {

	}

	env, err := n.GetEnvArray(p.NodeDef, p.Node)
	if err != nil {
		l.Error("error getting env values", "err", err)
		return err
	}

	envMap := make(map[string]string)

	err = n.AddEnvFlagMap(envMap, p.NodeDef, p.Node)
	if err != nil {
		l.Error("error getting env values", "err", err)
		return err
	}

	ins := n.GetInputStreams(p.InputURLs, "/tmp")
	for val, key := range ins {
		v := fmt.Sprintf("%s=%s", val, key.path)
		l.Info("adding input path to env", "name", val, "path", key.path)
		env = append(env, v)
		envMap[val] = key.path

	}

	outs, err := n.GetOutputPaths(p.OutputURLs, "/tmp")
	if err != nil {
		l.Error("can't get output paths", "err", err)
		return err
	}

	for val, key := range outs {
		path := fmt.Sprintf("%s=%s", val, key)
		l.Info("adding output to env", "name", val, "path", path)
		env = append(env, path)
		envMap[val] = key
	}

	expand := os.Expand(p.NodeDef.Command.Exec, func(s string) string {
		return envMap[s]
	})

	command := strings.Fields(expand)

	l.Info("command to run", "command", command)

	res, err := n.ContainerService.DockerClient.ContainerCreate(ctx, client.ContainerCreateOptions{
		Config: &container.Config{
			Env: env,
			Cmd: command,
		},
		HostConfig: &container.HostConfig{
			AutoRemove: false,
			Resources: container.Resources{
				DeviceRequests: []container.DeviceRequest{{
					Driver:       "nvidia",
					Count:        1,
					Capabilities: [][]string{{"compute", "video"}},
				}},
			},
		},
		Image: p.NodeDef.Image,
		Name:  "test2",
	})

	container_id := res.ID

	if err != nil {
		l.Error("error creating container", "err", err)
		return err
	}

	for key, val := range ins {
		fmt.Println(key)
		_, err := n.ContainerService.DockerClient.CopyToContainer(ctx, container_id, client.CopyToContainerOptions{
			DestinationPath: val.base_path,
			Content:         val.reader,
		})

		if err != nil {
			l.Error("error copying file to container", "err", err)
			return err
		}

		val.reader.Close()
	}

	if _, err := n.ContainerService.DockerClient.ContainerStart(ctx, container_id, client.ContainerStartOptions{}); err != nil {
		l.Error("error starting container", "err", err)
		return err
	}

	defer n.ContainerService.DockerClient.ContainerRemove(ctx, container_id, client.ContainerRemoveOptions{
		Force: true,
	})

	wait := n.ContainerService.DockerClient.ContainerWait(context.TODO(), container_id, client.ContainerWaitOptions{
		Condition: container.WaitConditionNextExit,
	})

	select {
	case res := <-wait.Result:
		if res.Error != nil {
			l.Error("container wait error", "err", res.Error.Message)
			return errors.New(res.Error.Message)
		}
	case err := <-wait.Error:
		if err != nil {
			l.Error("container wait error", "err", err)
			return err
		}
	}

	// check error codes and tty
	inspect, err := n.ContainerService.DockerClient.ContainerInspect(ctx, container_id, client.ContainerInspectOptions{})
	if err != nil {
		l.Error("error inspecting container", "err", err)
	}

	hasTTY := inspect.Container.Config.Tty
	exitCode := inspect.Container.State.ExitCode

	// CHECK LOGS

	logs, err := n.ContainerService.DockerClient.ContainerLogs(ctx, container_id, client.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	})

	if err != nil {
		l.Error("failed to get container logs", "err", err)
	}

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	if hasTTY {
		l.Info("container has tty")
		_, err := io.Copy(stdout, logs)
		if err != nil {
			l.Error("unable to copy container stdout")
		}

	} else {
		l.Info("container has no tty")
		_, copyErr := stdcopy.StdCopy(stdout, stderr, logs)
		if copyErr != nil {
			l.Error("unable to split stdout and stderr streams", "err", copyErr)
		}
	}

	stdoutString := stdout.String()
	stderrString := stderr.String()
	if stdoutString != "" {
		l.Info("container stdout", "message", stdoutString)
	}

	if stderrString != "" {
		l.Warn("container stderr", "message", stderrString)
	}

	if exitCode != 0 {
		l.Error("container exited with error", "exit_code", inspect.Container.State.ExitCode)
		return fmt.Errorf("container exited with error")
	}

	// Copy from container and upload
	for key, val := range outs {
		l.Info("copying from container", "output_name", key, "path", val)
		res, err := n.ContainerService.DockerClient.CopyFromContainer(ctx, container_id, client.CopyFromContainerOptions{
			SourcePath: val,
		})

		if err != nil {
			l.Error("could not copy file from container", "err", err)
			return err
		}

		tarStream := res.Content
		url, ok := p.OutputURLs[key]
		if !ok {
			l.Error("invalid key name. i don't know how this could happen")
			return fmt.Errorf("invalid key name")
		}
		uploadErr := n.UploadTarReaderToS3(ctx, url, tarStream)
		if uploadErr != nil {
			l.Error("failed to upload to s3", "output_name", key)
			return err
		}
	}

	return nil
}
