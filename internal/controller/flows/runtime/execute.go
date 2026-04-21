package runtime

import (
	"context"
	"fmt"
	"time"

	"github.com/pupload/pupload/internal/models"
	"github.com/pupload/pupload/internal/syncplane"
	"github.com/pupload/pupload/internal/telemetry"
)

func (rt *RuntimeFlow) handleExecuteStep(ctx context.Context, stepID string, s syncplane.SyncLayer) error {
	step := rt.steps[stepID]
	inputs := make(map[string]string)

	for _, edge := range step.Inputs {
		artifact := rt.FlowRun.Artifacts[edge.Edge]
		store := rt.stores[artifact.StoreName]

		url, err := store.GetURL(context.TODO(), artifact.ObjectName, 1*time.Hour)
		if err != nil {
			rt.log.Error("unable to generate store get url", "message", err)
			return err
		}

		inputs[edge.Name] = url.String()
	}

	outputs := make(map[string]string)
	for _, edge := range step.Outputs {

		artifact, err := rt.makeOutputArtifact(edge)
		if err != nil {
			return err
		}

		store, ok := rt.stores[artifact.StoreName]
		if !ok {
			rt.log.Error("unable to acquire store", "store_name", artifact.StoreName)
			return fmt.Errorf("unable to acquire store described in artifact")
		}

		url, err := store.PutURL(context.TODO(), artifact.ObjectName, 15*time.Minute)
		if err != nil {
			rt.log.Error("could not generate put url", "err", err)
			return err
		}

		outputs[edge.Name] = url.String()
		WaitingURL := models.WaitingURL{
			Artifact: *artifact,
			PutURL:   url.String(),
			TTL:      time.Now().Add(15 * time.Minute),
		}

		rt.FlowRun.WaitingURLs = append(rt.FlowRun.WaitingURLs, WaitingURL)
	}

	err := step.executeStep(ctx, s, rt.FlowRun.ID, inputs, outputs)
	if err != nil {
		return err
	}

	return nil
}

func (rt *RuntimeFlow) makeOutputArtifact(edge models.StepEdge) (*models.Artifact, error) {
	for _, well := range rt.Flow.DataWells {
		if well.Edge != edge.Edge {
			continue
		}

		artifact := models.Artifact{
			StoreName:  well.Store,
			EdgeName:   well.Edge,
			ObjectName: rt.processDatawellKey(well),
		}

		return &artifact, nil
	}

	if rt.Flow.DefaultDataWell == nil {
		return nil, fmt.Errorf("default datawell is nil")
	}

	artifact := models.Artifact{
		StoreName:  rt.Flow.DefaultDataWell.Store,
		EdgeName:   edge.Edge,
		ObjectName: fmt.Sprintf("%s-%s", edge.Edge, rt.FlowRun.ID),
	}

	return &artifact, nil

}

func (rt *RuntimeFlow) HandleStepFinished(stepID string, logs []models.LogRecord) error {
	_, ok := rt.steps[stepID]
	if !ok {
		return fmt.Errorf("HandleStepFinished: step does not exist")
	}

	curr_state := rt.FlowRun.StepState[stepID]
	new_logs := append(curr_state.Logs, logs...)
	rt.FlowRun.StepState[stepID] = models.StepState{Status: models.STEPRUN_COMPLETE, Logs: new_logs}

	return nil
}

func (rt *RuntimeFlow) HandleStepFailed(stepID string, logs []models.LogRecord, err string, attempt, maxAttempt int, isFinal bool) error {
	_, ok := rt.steps[stepID]
	if !ok {
		return fmt.Errorf("HandleStepFailed: step does not exist")
	}

	curr_state := rt.FlowRun.StepState[stepID]
	new_logs := append(curr_state.Logs, logs...)

	status := models.STEPRUN_RETRYING
	if isFinal {
		rt.FlowRun.Status = models.FLOWRUN_ERROR
		status = models.STEPRUN_ERROR
	}

	rt.FlowRun.StepState[stepID] = models.StepState{
		Status:      status,
		Logs:        new_logs,
		Error:       err,
		Attempt:     attempt,
		MaxAttempts: maxAttempt,
	}

	return nil

}

func (rs *RuntimeStep) executeStep(ctx context.Context, s syncplane.SyncLayer, runID string, input, output map[string]string) error {
	payload := syncplane.StepExecutePayload{
		RunID:      runID,
		Step:       *rs.Step,
		Task:       rs.Task,
		InputURLs:  input,
		OutputURLs: output,

		MaxAttempts: rs.Task.MaxAttempts,

		TraceParent: telemetry.InjectContext(ctx),
	}

	return s.EnqueueExecuteStep(payload)
}

func (rt *RuntimeFlow) shouldStepReady(stepID string) {
	step := rt.steps[stepID]
	curr_state := rt.FlowRun.StepState[stepID].Status

	if curr_state != models.STEPRUN_IDLE {
		return
	}

	for _, input := range step.Inputs {
		_, ok := rt.FlowRun.Artifacts[input.Edge]
		if !ok {
			return
		}
	}

	rt.FlowRun.StepState[stepID] = models.StepState{Status: models.STEPRUN_READY, Logs: rt.FlowRun.StepState[stepID].Logs}
}
