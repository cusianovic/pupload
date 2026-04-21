package runtime

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/pupload/pupload/internal/logging"
	"github.com/pupload/pupload/internal/models"
	"github.com/pupload/pupload/internal/stores"
	"github.com/pupload/pupload/internal/syncplane"
	"github.com/pupload/pupload/internal/telemetry"

	"github.com/google/uuid"
)

type RuntimeFlow struct {
	Flow    models.Flow
	FlowRun models.FlowRun
	Tasks   []models.Task

	steps  map[string]RuntimeStep
	stores map[string]models.Store

	log *slog.Logger

	TraceParent string
}

type RuntimeStep struct {
	*models.Step
	Task models.Task
}

func CreateRuntimeFlow(ctx context.Context, flow models.Flow, tasks []models.Task) (RuntimeFlow, error) {
	// Unmarshal Stores

	runtimeFlow := RuntimeFlow{
		Flow:  flow,
		Tasks: tasks,

		stores: make(map[string]models.Store),
		steps:  make(map[string]RuntimeStep),

		TraceParent: telemetry.InjectContext(ctx),
	}

	runtimeFlow.constructLogger()
	runtimeFlow.constructStores()
	err := runtimeFlow.constructRuntimeStep()
	if err != nil {
		return runtimeFlow, err
	}

	runtimeFlow.createFlowRun()

	if err := runtimeFlow.initialDatawellInput(); err != nil {
		return runtimeFlow, err
	}

	return runtimeFlow, nil
}

func (rt *RuntimeFlow) RebuildRuntimeFlow() {

	rt.steps = make(map[string]RuntimeStep)
	rt.stores = make(map[string]models.Store)

	rt.constructLogger()
	rt.constructStores()
	rt.constructRuntimeStep()
}

func (rt *RuntimeFlow) createFlowRun() {

	id := uuid.Must(uuid.NewV7())

	waitingUrls := make([]models.WaitingURL, 0)
	artifacts := make(map[string]models.Artifact)
	stepStates := make(map[string]models.StepState)

	for _, step := range rt.steps {
		stepStates[step.ID] = models.StepState{Status: models.STEPRUN_IDLE, Logs: []models.LogRecord{}}
	}

	value := models.FlowRun{
		ID:          id.String(),
		StepState:   stepStates,
		Status:      models.FLOWRUN_STOPPED,
		WaitingURLs: waitingUrls,
		Artifacts:   artifacts,
	}

	rt.FlowRun = value
}

func (rt *RuntimeFlow) constructRuntimeStep() error {
	for _, step := range rt.Flow.Steps {
		found := false
		defName := step.Uses

		for _, task := range rt.Tasks {
			if defName == fmt.Sprintf("%s/%s", task.Publisher, task.Name) {
				found = true
				rt.steps[step.ID] = RuntimeStep{Step: &step, Task: task}
				break
			}
		}

		if !found {
			return fmt.Errorf("unable to find task with defName %s", defName)
		}
	}

	return nil
}

func (rt *RuntimeFlow) constructStores() {
	for _, storeInput := range rt.Flow.Stores {
		store, err := stores.UnmarshalStore(storeInput)
		if err != nil {
			rt.log.Warn("Invalid store definition", "flow", rt.Flow.Name, "store", storeInput.Name, "error", err.Error())
			continue
		}

		rt.stores[storeInput.Name] = store
	}
}

func (rt *RuntimeFlow) constructLogger() {
	rt.log = logging.ForService("flow_runtime").With("run_id", rt.FlowRun.ID)
}

func (rt *RuntimeFlow) initialDatawellInput() error {
	for _, dw := range rt.Flow.DataWells {

		var err error

		if dw.Source == nil {
			continue
		}

		switch *dw.Source {
		case "upload":
			err = rt.handleUploadDatawell(dw)

		case "static":
			err = rt.handleStaticDatawell(dw)

		case "webhook":
			// TODO: implement webhook data source
		}

		if err != nil {
			rt.log.Error("error processing datawells", "err", err)
			rt.FlowRun.Status = models.FLOWRUN_ERROR
			return err
		}
	}

	return nil
}

func (rt *RuntimeFlow) handleUploadDatawell(dw models.DataWell) error {

	key := rt.processDatawellKey(dw)
	store, ok := rt.stores[dw.Store]

	if !ok {
		return fmt.Errorf("store %s does not exist", dw.Store)
	}

	artifact := models.Artifact{
		StoreName:  dw.Store,
		ObjectName: key,
		EdgeName:   dw.Edge,
	}

	url, err := store.PutURL(context.TODO(), artifact.ObjectName, 1*time.Hour)
	if err != nil {
		return fmt.Errorf("datawell for edge %q using store %q: %w", dw.Edge, dw.Store, err)
	}

	waitingURL := models.WaitingURL{
		Artifact: artifact,
		PutURL:   url.String(),
		TTL:      time.Now().Add(1 * time.Hour),
	}

	rt.FlowRun.WaitingURLs = append(rt.FlowRun.WaitingURLs, waitingURL)
	return nil
}

func (rt *RuntimeFlow) handleStaticDatawell(dw models.DataWell) error {

	if dw.Key == nil {
		return fmt.Errorf("handleStaticDatawell: datawell with static source must have key")
	}

	store, ok := rt.stores[dw.Store]
	if !ok {
		return fmt.Errorf("handleStaticDatawell: datawell has invalid store")
	}

	exists := store.Exists(*dw.Key)
	if !exists {
		return fmt.Errorf("datawell for edge %q using store %q: object %q not found or store is not reachable", dw.Edge, dw.Store, *dw.Key)
	}

	artifact := models.Artifact{
		StoreName:  dw.Store,
		ObjectName: *dw.Key,
		EdgeName:   dw.Edge,
	}

	rt.FlowRun.Artifacts[dw.Edge] = artifact

	return nil
}

func (rt *RuntimeFlow) processDatawellKey(dw models.DataWell) string {
	if dw.Key == nil {
		return fmt.Sprintf("%s_%s", dw.Edge, rt.FlowRun.ID)
	}

	runTime := time.Now()

	keyVars := map[string]string{
		"RUN_ID": rt.FlowRun.ID, "FLOW_NAME": rt.Flow.Name, "EDGE": dw.Edge,
		"TIMESTAMP": runTime.Format(time.RFC3339),
		"DATE":      runTime.Format("2006-01-02"),
		"YEAR":      runTime.Format("2006"),
		"MONTH":     runTime.Format("01"),
		"DAY":       runTime.Format("02"),
		"UUID":      uuid.NewString(),
	}

	return os.Expand(*dw.Key, func(s string) string {
		return keyVars[s]
	})
}

// Returns nil if flow is already running.
// Returns error if flow is in error state or already complete
func (rt *RuntimeFlow) Start(s syncplane.SyncLayer) error {
	switch rt.FlowRun.Status {

	case models.FLOWRUN_STOPPED:
		rt.FlowRun.Status = models.FLOWRUN_WAITING
		rt.FlowRun.StartedAt = time.Now()

	case models.FLOWRUN_COMPLETE:
		return fmt.Errorf("runtime already complete")

	case models.FLOWRUN_ERROR:
		return fmt.Errorf("runtime in error state")

	case models.FLOWRUN_WAITING:

	case models.FLOWRUN_RUNNING:

	default:
		return fmt.Errorf("no case")
	}

	rt.Step(s)
	return nil
}
