package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/pupload/pupload/internal/controller/config"
	"github.com/pupload/pupload/internal/controller/flows/repo"
	"github.com/pupload/pupload/internal/controller/flows/runtime"
	"github.com/pupload/pupload/internal/syncplane"
	"github.com/pupload/pupload/internal/telemetry"
	"github.com/pupload/pupload/internal/validation"

	"github.com/pupload/pupload/internal/logging"
	"github.com/pupload/pupload/internal/models"

	"go.opentelemetry.io/otel/attribute"
)

type FlowService struct {
	runtimeRepo repo.RuntimeRepo

	syncLayer syncplane.SyncLayer

	log *slog.Logger
}

func CreateFlowService(cfg *config.ControllerConfig, s syncplane.SyncLayer) (*FlowService, error) {
	slog := logging.ForService("flow")

	runtimeRepo, err := repo.CreateRuntimeRepo(cfg.RuntimeRepo)
	if err != nil {
		return nil, err
	}

	f := FlowService{
		runtimeRepo: runtimeRepo,

		syncLayer: s,

		log: slog,
	}

	s.RegisterFlowStepHandler(f.FlowStepHandler)
	s.RegisterStepFinishedHandler(f.StepFinishedHandler)
	s.RegisterStepFailedHandler(f.StepFailedHandler)

	s.Start()

	return &f, nil
}

func (f *FlowService) Close(ctx context.Context) {
	f.runtimeRepo.Close(ctx)
}

func (f *FlowService) RunFlow(flow models.Flow, tasks []models.Task) (models.FlowRun, error) {

	ctx, span := telemetry.Tracer("pupload.controller").Start(context.Background(), "RunFlow")
	defer span.End()

	flow.Normalize()
	for i := range tasks {
		tasks[i].Normalize()
		f.log.Debug("task tier", "task", tasks[i].Name, "tier", tasks[i].Tier)
	}

	res := validation.Validate(flow, tasks)

	if res.HasError() {
		f.log.Warn("invalid flow", "errors", res.Errors, "warnings", res.Warnings)
		return models.FlowRun{}, fmt.Errorf("invalid flow")
	}

	runtime, err := runtime.CreateRuntimeFlow(ctx, flow, tasks)
	if err != nil {
		return models.FlowRun{}, err
	}

	span.SetAttributes(attribute.String("run_id", runtime.FlowRun.ID))

	runtime.Start(f.syncLayer)
	f.runtimeRepo.SaveRuntime(runtime)
	f.syncLayer.AddRunToScheduler(runtime.FlowRun.ID)

	return runtime.FlowRun, nil
}

func (f *FlowService) Status(runID string) (models.FlowRun, error) {
	runtime, err := f.runtimeRepo.LoadRuntime(runID)
	if err != nil {
		return models.FlowRun{}, err
	}

	return runtime.FlowRun, nil
}

func (f *FlowService) HandleFlowComplete(rt runtime.RuntimeFlow) error {
	f.runtimeRepo.SaveRuntimeWithTTL(rt, 10*time.Minute)
	f.syncLayer.RemoveRunFromScheduler(rt.FlowRun.ID)
	return nil
}
