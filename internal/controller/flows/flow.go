package flows

import (
	"context"
	"fmt"
	"pupload/internal/models"
	"pupload/internal/util"
	"time"
)

func (f *FlowService) ListFlows() map[string]models.Flow {
	return f.FlowList
}

func (f *FlowService) GetFlow(name string) (models.Flow, error) {

	flow, ok := f.FlowList[name]
	if !ok {
		return flow, fmt.Errorf("pooo %s does not exist", name)
	}

	return flow, nil
}

func (f *FlowService) StartFlow(ctx context.Context, name string) (string, error) {
	flow, exists := f.FlowList[name]
	if !exists {
		return "", fmt.Errorf("Flow %s does not exist", name)
	}

	if len(flow.Nodes) == 0 {
		return "", fmt.Errorf("Flow %s does not contain any nodes", name)
	}

	flowRun, err := f.CreateFlowRun(name)

	if err != nil {
		return "", err
	}

	err = f.HandleStepFlow(ctx, flowRun.ID)

	if err != nil {
		return "", err
	}

	return flowRun.ID, nil
}

func (f *FlowService) HandleStepFlow(ctx context.Context, id string) error {
	mutex_key := fmt.Sprintf("flowrunlock:%s", id)
	ok, err := util.AcquireLock(f.RedisClient, mutex_key, time.Second*10)
	if err != nil {
		return err
	}

	if !ok {
		return fmt.Errorf("flowrun %s currently being stepped", id)
	}

	defer util.ReleaseLock(f.RedisClient, mutex_key)

	run, err := f.GetFlowRun(id)
	if err != nil {
		return fmt.Errorf("Can't retrieve flow status %s. Is Redis running? %s", id, err)
	}

	newRun, err := f.stepFlow(context.TODO(), run)
	if err != nil {
		return fmt.Errorf("error stepping flow")
	}

	// prettyStruct, _ := json.MarshalIndent(newRun, "", "    ")
	// f.log.Info("updated run info", "data", string(prettyStruct))

	f.updateFlowRun(newRun)

	return nil
}

func (f *FlowService) stepFlow(ctx context.Context, run FlowRun) (FlowRun, error) {
	for {

		f.log.Info("stepFlow state", "runID", run.ID, "state", run.Status)

		if len(run.NodesLeft) == 0 {
			run.Status = FLOWRUN_COMPLETE
			return run, nil

		}

		nodesToRun := f.nodesAvailableToRun(run)

		switch run.Status {
		case FLOWRUN_IDLE:
			if len(nodesToRun) == 0 {
				run.Status = FLOWRUN_WAITING
			}

			run.Status = FLOWRUN_RUNNING

		case FLOWRUN_WAITING:

			shouldChangeToRunning := false
			for i, url := range run.WaitingURLs {
				result := f.checkWaitingURL(&run, url)
				f.log.Info("URL waiting state", "state", result)

				switch result {
				case WaitNoChange:

				case WaitReady:
					run.Status = FLOWRUN_RUNNING
					shouldChangeToRunning = true
					run.WaitingURLs = append(run.WaitingURLs[:i], run.WaitingURLs[i+1:]...)
					run.Artifacts[url.Artifact.EdgeName] = url.Artifact

				case WaitURLExpired:

				case WaitFailed:
					return run, fmt.Errorf("Checking WaitURL failed")
				}

			}

			if !shouldChangeToRunning {
				return run, nil
			}

		case FLOWRUN_RUNNING:
			for _, nodeIndex := range nodesToRun {
				f.HandleExecuteNode(&run, nodeIndex)
			}

			run.Status = FLOWRUN_WAITING
		}
	}

}

func (f *FlowService) GetStore(flowName string, storeName string) (store models.Store, ok bool) {

	// prefer local store for the given flow, fall back to global store
	stores, ok := f.LocalStoreMap[LocalStoreKey{flowName, storeName}]
	if !ok {
		stores, ok = f.GlobalStoreMap[storeName]
	}

	return stores, ok
}

func (f *FlowService) NodeLength(flowName string) int {
	return len(f.FlowList[flowName].Nodes)
}

func (f *FlowService) nodesAvailableToRun(flowRun FlowRun) []int {

	nodes := make([]int, 0)

	for i := range flowRun.NodesLeft {

		if _, running := flowRun.RunningNodes[i]; running {
			continue
		}

		runnable := true

		node := f.FlowList[flowRun.FlowName].Nodes[i]
		for _, input := range node.Inputs {

			_, ok := flowRun.Artifacts[input.Edge]
			if !ok {
				runnable = false
				break
			}
		}

		if runnable {
			nodes = append(nodes, i)
		}
	}

	return nodes
}
