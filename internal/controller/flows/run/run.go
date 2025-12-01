package run

import (
	"time"

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
