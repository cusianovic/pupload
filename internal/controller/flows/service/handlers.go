package service

import (
	"context"
	"fmt"
	"time"

	"github.com/pupload/pupload/internal/syncplane"
)

func (f *FlowService) FlowStepHandler(ctx context.Context, payload syncplane.FlowStepPayload) error {
	key := fmt.Sprintf("runtimelock:%s", payload.RunID)
	m := f.syncLayer.NewMutex(key, 10*time.Second)
	err := m.Lock(ctx)

	if err != nil {
		f.log.Error("runtime lock already in use", "run_id", payload.RunID)
		return err
	}
	defer m.Unlock(ctx)

	runtime, err := f.runtimeRepo.LoadRuntime(payload.RunID)
	if err != nil {
		f.log.Error("unable to get runtime flow from runtimeRepo", "runID", payload.RunID)
	}

	runtime.RebuildRuntimeFlow()
	runtime.Step(f.syncLayer)
	if runtime.IsComplete() || runtime.IsError() {
		f.HandleFlowComplete(runtime)
		return nil
	}

	f.runtimeRepo.SaveRuntime(runtime)

	return nil
}

func (f *FlowService) StepFinishedHandler(ctx context.Context, payload syncplane.StepFinishedPayload) error {
	f.log.Info("StepFinishedHandler: starting step finished task", "run_id", payload.RunID)

	key := fmt.Sprintf("runtimelock:%s", payload.RunID)
	m := f.syncLayer.NewMutex(key, 10*time.Second)
	err := m.Lock(ctx)
	if err != nil {
		f.log.Error("StepFinishedHandler: runtime lock already in use", "run_id", payload.RunID, "err", err)
		return err
	}
	defer m.Unlock(ctx)

	runtime, err := f.runtimeRepo.LoadRuntime(payload.RunID)
	if err != nil {
		f.log.Error("StepFinishedHandler: error loading runtime", "run_id", payload.RunID, "err", err)
		return err
	}

	runtime.RebuildRuntimeFlow()

	if err := runtime.HandleStepFinished(payload.StepID, payload.Logs); err != nil {
		f.log.Error("StepFinishedHandler: error handling step finished", "run_id", payload.RunID, "step_id", payload.StepID, "err", err)
		return err
	}

	runtime.Step(f.syncLayer)
	if err := f.runtimeRepo.SaveRuntime(runtime); err != nil {
		f.log.Error("StepFinishedHandler: error saving runtime", "run_id", payload.RunID, "step_id", payload.StepID, "err", err)
		return err
	}

	return nil
}

func (f *FlowService) StepFailedHandler(ctx context.Context, payload syncplane.StepFailedPayload) error {

	isFinalFailure := payload.Attempt >= payload.MaxAttempts
	key := fmt.Sprintf("runtimelock:%s", payload.RunID)
	m := f.syncLayer.NewMutex(key, 10*time.Second)
	err := m.Lock(ctx)
	if err != nil {
		f.log.Error("StepFailedHandler: runtime lock already in use", "run_id", payload.RunID, "err", err)
		return err
	}
	defer m.Unlock(ctx)

	runtime, err := f.runtimeRepo.LoadRuntime(payload.RunID)
	if err != nil {
		f.log.Error("StepFailedHandler: error loading runtime", "run_id", payload.RunID, "err", err)
		return err
	}

	runtime.RebuildRuntimeFlow()

	if err := runtime.HandleStepFailed(payload.StepID, payload.Logs, payload.Error, payload.Attempt, payload.MaxAttempts, isFinalFailure); err != nil {
		f.log.Error("StepFailedHandler: error handling step failed", "run_id", payload.RunID, "step_id", payload.StepID, "err", err)
		return err
	}

	runtime.Step(f.syncLayer)
	if err := f.runtimeRepo.SaveRuntime(runtime); err != nil {
		f.log.Error("StepFailedHandler: error saving runtime", "run_id", payload.RunID, "step_id", payload.StepID, "err", err)
		return err
	}

	return nil
}
