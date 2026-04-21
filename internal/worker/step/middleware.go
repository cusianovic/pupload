package step

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/pupload/pupload/internal/logging"
	"github.com/pupload/pupload/internal/models"
	"github.com/pupload/pupload/internal/syncplane"
	"github.com/pupload/pupload/internal/telemetry"
)

func (ss *StepService) FinishedMiddleware(ctx context.Context, payload syncplane.StepExecutePayload) error {

	ctx = telemetry.ExtractContext(ctx, payload.TraceParent)

	logs := make([]models.LogRecord, 0, 64)
	ch := &logging.CollectHandler{
		Inner:   logging.Root().Handler(),
		Records: &logs,
	}

	jobLog := slog.New(ch)
	jobLog.With(
		"run_id", payload.RunID,
		"step_id", payload.Step.ID,
		"task_publisher", payload.Task.Publisher,
		"task_name", payload.Task.Name,
		"container_image", payload.Task.Image,
	)
	jobLog.Debug("inside worker")

	ctx = logging.CtxWithLogger(ctx, jobLog)

	if err := ss.tryReserve(payload.Task.Tier); err != nil {
		jobLog.Error("error attempting to reserve resources", "err", err)
		return err
	}

	res, genErr := ss.ResourceManger.GenerateContainerResource(payload.Task.Tier)
	if genErr != nil {
		jobLog.Error("error attempting to generate ContainerResource", "err", genErr)
		return genErr
	}

	err := ss.StepExecute(ctx, payload, res)
	if err == nil {
		if err := ss.SyncLayer.EnqueueStepFinished(syncplane.StepFinishedPayload{
			RunID:  payload.RunID,
			StepID: payload.Step.ID,
			Logs:   logs,
		}); err != nil {
			jobLog.Error("error send step finished message", "err", err)

		}
	} else {
		if enqueueErr := ss.SyncLayer.EnqueueStepFailed(syncplane.StepFailedPayload{
			RunID:       payload.RunID,
			StepID:      payload.Step.ID,
			Logs:        logs,
			Attempt:     payload.Attempt,
			MaxAttempts: payload.MaxAttempts,
			Error:       err.Error(),
			TraceParent: payload.TraceParent,
		}); enqueueErr != nil {
			jobLog.Error("error send step error message", "err", err)
		}

	}

	if err != nil {
		jobLog.Error(err.Error())
	}

	if err := ss.tryRelease(payload.Task.Tier); err != nil {
		jobLog.Error("error attempting to release resources", "err", err)

	}

	return err
}

func (ss *StepService) tryReserve(s string) error {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	if err := ss.ResourceManger.Reserve(s); err != nil {
		return fmt.Errorf("ExecuteStepHandler: Could not reserve resource %s: %w", s, err)
	}

	queueMap := ss.ResourceManger.GetValidTierMap()

	ss.SyncLayer.UpdateSubscribedQueues(queueMap)

	return nil
}

func (ss *StepService) tryRelease(s string) error {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	if err := ss.ResourceManger.Release(s); err != nil {
		return fmt.Errorf("ExecuteStepHandler: Could not release resource %s: %w", s, err)
	}

	queueMap := ss.ResourceManger.GetValidTierMap()

	ss.SyncLayer.UpdateSubscribedQueues(queueMap)

	return nil
}
