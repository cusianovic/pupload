package step

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/pupload/pupload/internal/logging"
	"github.com/pupload/pupload/internal/models"
	"github.com/pupload/pupload/internal/syncplane"
	"github.com/pupload/pupload/internal/telemetry"

	cont "github.com/pupload/pupload/internal/worker/container"

	"github.com/moby/moby/api/types/container"
	"go.opentelemetry.io/otel/attribute"
	"golang.org/x/sync/errgroup"
)

func (ss *StepService) StepExecute(ctx context.Context, payload syncplane.StepExecutePayload, resource container.Resources) error {

	l := logging.LoggerFromCtx(ctx)
	ctx, span := telemetry.Tracer("pupload.worker").Start(ctx, "StepExecute")
	defer span.End()
	span.SetAttributes(
		attribute.String("step_id", payload.Step.ID),
		attribute.String("container_image", payload.Task.Image),
		attribute.String("tier", payload.Task.Tier),
	)
	l.With("span_id", span.SpanContext().SpanID().String())

	// handle worker capabiliites

	in, out, err := ss.prepareIO(payload.InputURLs, payload.OutputURLs, payload.Task, "/tmp")
	if err != nil {
		return err
	}

	command, err := ss.generateCommand(payload.Step, payload.Task, in, out)
	if err != nil {
		return err
	}

	l.Info("validating container image")
	ok, err := ss.CS.IM.Validate(ctx, payload.Task.Image)
	if err != nil {
		l.Error("error validating image", "err", err)
		return err
	}
	if !ok {
		l.Warn("image not found, attempting pull")
		err := ss.CS.IM.Pull(ctx, payload.Task.Image)
		if err != nil {
			l.Error("error pulling image", "err", err)
			return err
		}
	}

	containerID, err := ss.CS.RT.CreateContainer(ctx, cont.ContainerConfig{
		Image: payload.Task.Image,
		Name:  fmt.Sprintf("pupload-%s-%s", payload.RunID, payload.Step.ID),
		Cmd:   command,

		HostConfig: &container.HostConfig{
			AutoRemove: false,
			Resources:  resource,
		},
	})

	if err != nil {
		return err
	}

	l.With("container_id", containerID)
	l.Info("container created")
	span.AddEvent("container created")

	defer ss.CS.RT.RemoveContainer(ctx, containerID)

	if err := ss.downloadAllInputsToContainer(ctx, containerID, in); err != nil {
		return err
	}

	l.Info("files downloaded to container")

	if err := ss.CS.RT.StartContainer(ctx, containerID); err != nil {
		return err
	}

	l.Info("container started")
	span.AddEvent("container started")

	res, err := ss.CS.RT.WaitContainer(ctx, containerID)
	if err != nil {
		return err
	}

	l.With(
		"exit_code", res.ExitCode,
		"exit_message", res.Error,
	)
	l.Info("container finished")
	span.AddEvent("container finished")

	logs, err := ss.CS.RT.GetLogs(ctx, containerID)
	if err != nil {
		return err
	}

	if res.ExitCode != 0 {
		l.Warn("container logs", "logs", logs)

		return fmt.Errorf("contained exited with non-0 exit code")
	}

	l.Info("container logs", "logs", logs)

	if err := ss.uploadAllOutputsFromContainer(ctx, containerID, out); err != nil {
		return err
	}

	l.Info("files uploaded from container")

	return nil
}

func (ss *StepService) downloadAllInputsToContainer(ctx context.Context, containerID string, inputs []preparedIO) error {
	inGroup, errCtx := errgroup.WithContext(ctx)
	for _, i := range inputs {
		i := i
		inGroup.Go(func() error {
			return ss.CS.IO.DownloadIntoContainer(errCtx, containerID, i.url, i.base_path, i.filename)
		})
	}

	return inGroup.Wait()
}

func (ss *StepService) uploadAllOutputsFromContainer(ctx context.Context, containerID string, outputs []preparedIO) error {
	outGroup, errCtx := errgroup.WithContext(ctx)
	for _, o := range outputs {
		o := o
		outGroup.Go(func() error {
			return ss.CS.IO.UploadFromContainer(errCtx, containerID, o.url, o.path, o.filename)
		})
	}

	return outGroup.Wait()
}

func (ss *StepService) generateCommand(step models.Step, task models.Task, in, out []preparedIO) ([]string, error) {
	envMap := make(map[string]string)

	if err := ss.addEnvFlagMap(envMap, task, step); err != nil {
		return nil, err
	}

	// prep inputs
	ss.addIOToEnvMap(envMap, in)
	ss.addIOToEnvMap(envMap, out)

	expand := os.Expand(task.Command.Exec, func(s string) string {
		return envMap[s]
	})

	command := strings.Fields(expand)
	return command, nil
}
