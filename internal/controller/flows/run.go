package flows

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

type FlowRunStatus string

const (
	FLOWRUN_IDLE     FlowRunStatus = "IDLE"
	FLOWRUN_WAITING  FlowRunStatus = "WAITING"
	FLOWRUN_RUNNING  FlowRunStatus = "RUNNING"
	FLOWRUN_STOPPED  FlowRunStatus = "STOPPED"
	FLOWRUN_COMPLETE FlowRunStatus = "COMPLETE"
	FLOWRUN_ERROR    FlowRunStatus = "ERROR"
)

type Artifact struct {
	StoreName  string
	ObjectName string
	EdgeName   string
}

type WaitingURL struct {
	Artifact Artifact
	PutURL   string
	TTL      time.Time
}

type FlowRun struct {
	ID           string
	FlowName     string
	NodesLeft    map[int]struct{}
	RunningNodes map[int]*asynq.TaskInfo
	Status       FlowRunStatus
	Artifacts    map[string]Artifact // Maps given edge ID to Artifact
	WaitingURLs  []WaitingURL
}

func (f *FlowService) CreateFlowRun(flowName string) (FlowRun, error) {

	id := uuid.Must(uuid.NewV7())
	key := fmt.Sprintf("flowrun:%s", id.String())

	nodeCount := len(f.FlowList[flowName].Nodes)
	nodesLeft := make(map[int]struct{})
	for i := range nodeCount {
		nodesLeft[i] = struct{}{}
	}

	waitingUrls := make([]WaitingURL, 0)
	artifacts := make(map[string]Artifact)

	value := FlowRun{
		ID:           id.String(),
		FlowName:     flowName,
		NodesLeft:    nodesLeft,
		RunningNodes: make(map[int]*asynq.TaskInfo),
		Status:       FLOWRUN_IDLE,
		WaitingURLs:  waitingUrls,
		Artifacts:    artifacts,
	}

	f.initalizeWaitingURLs(&value)

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(value); err != nil {
		return value, err
	}

	if err := f.RedisClient.Set(context.TODO(), key, buf.Bytes(), 0).Err(); err != nil {
		return value, err
	}

	return value, nil

}

func (f *FlowService) GetFlowRun(id string) (FlowRun, error) {
	key := fmt.Sprintf("flowrun:%s", id)
	raw, err := f.RedisClient.Get(context.TODO(), key).Bytes()
	if err != nil {
		return FlowRun{}, fmt.Errorf("FlowRun %s does not exist", id)
	}

	var val FlowRun
	dec := gob.NewDecoder(bytes.NewReader(raw))
	if err := dec.Decode(&val); err != nil {
		return FlowRun{}, err
	}

	return val, nil

}

func (f *FlowService) updateFlowRun(flowRun FlowRun) error {
	key := fmt.Sprintf("flowrun:%s", flowRun.ID)

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(flowRun); err != nil {
		return err
	}

	if err := f.RedisClient.Set(context.TODO(), key, buf.Bytes(), 0).Err(); err != nil {
		return err
	}

	return nil

}

func (f *FlowService) initalizeWaitingURLs(flowrun *FlowRun) error {

	flow, err := f.GetFlow(flowrun.FlowName)
	if err != nil {
		return err
	}

	for _, datawell := range flow.DataWells {
		if datawell.Type != "dynamic" {
			continue
		}

		var key string
		if datawell.Key == nil {
			key = fmt.Sprintf("%s_%s", datawell.Edge, flowrun.ID)
		} else {
			key = f.getDataWellKey("beepboop", *flowrun)
		}

		store, ok := f.GetStore(flowrun.FlowName, datawell.Store)
		if !ok {
			return fmt.Errorf("Store %s does not exists", datawell.Store)
		}

		artifact := Artifact{
			StoreName:  datawell.Store,
			ObjectName: key,
			EdgeName:   datawell.Edge,
		}

		url, err := store.PutURL(context.TODO(), artifact.ObjectName, 10*time.Second)
		if err != nil {
			return fmt.Errorf("Store %s could not generate put url: %s", datawell.Store, err)
		}

		WaitingURL := WaitingURL{
			Artifact: artifact,
			PutURL:   url.String(),
			TTL:      time.Now().Add(10 * time.Second),
		}

		flowrun.WaitingURLs = append(flowrun.WaitingURLs, WaitingURL)

	}

	err = f.updateFlowRun(*flowrun)
	if err != nil {
		return err
	}

	return nil
}

func (f *FlowService) checkWaitingUrls(flowrun *FlowRun) (bool, error) {

	fileExists := false

	for i, waitingUrl := range flowrun.WaitingURLs {
		store, ok := f.GetStore(flowrun.FlowName, waitingUrl.Artifact.StoreName)
		if !ok {
			return false, fmt.Errorf("store referenced in waitingURL does not exists. this should never happen")
		}

		exists := store.Exists(waitingUrl.Artifact.ObjectName)
		if exists {

			flowrun.Artifacts[waitingUrl.Artifact.EdgeName] = waitingUrl.Artifact

			// TODO: Handle oob's edge case
			flowrun.WaitingURLs = append(flowrun.WaitingURLs[:i], flowrun.WaitingURLs[i+1:]...)
			fileExists = true
		}
	}

	return fileExists, nil
}

type WaitingURLResult int

const (
	WaitNoChange   WaitingURLResult = iota
	WaitReady                       // Object Exists
	WaitURLExpired                  // URL Expired
	WaitFailed                      // Non retryable error
)

func (f *FlowService) checkWaitingURL(flowrun *FlowRun, w WaitingURL) WaitingURLResult {
	/*
		if time.Now().After(w.TTL) {
			return WaitURLExpired
		}
	*/
	store, ok := f.GetStore(flowrun.FlowName, w.Artifact.StoreName)
	if !ok {
		return WaitFailed
	}
	exists := store.Exists(w.Artifact.ObjectName)
	if exists {
		return WaitReady
	}

	return WaitNoChange
}

func (f *FlowService) getDataWellKey(key string, flowrun FlowRun) string {
	// TODO: implement
	return key
}
