package models

import (
	"time"
)

type FlowRunStatus string

const (
	FLOWRUN_STOPPED  FlowRunStatus = "STOPPED"
	FLOWRUN_WAITING  FlowRunStatus = "WAITING"
	FLOWRUN_RUNNING  FlowRunStatus = "RUNNING"
	FLOWRUN_COMPLETE FlowRunStatus = "COMPLETE"
	FLOWRUN_ERROR    FlowRunStatus = "ERROR"
)

type StepRunStatus string

const (
	STEPRUN_IDLE     StepRunStatus = "IDLE"
	STEPRUN_READY    StepRunStatus = "READY"
	STEPRUN_RUNNING  StepRunStatus = "RUNNING"
	STEPRUN_RETRYING StepRunStatus = "RETRYING"
	STEPRUN_COMPLETE StepRunStatus = "COMPLETE"
	STEPRUN_ERROR    StepRunStatus = "ERROR"
)

type StepState struct {
	Status      StepRunStatus
	Logs        []LogRecord
	Error       string
	Attempt     int
	MaxAttempts int
}

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
	ID string

	StepState   map[string]StepState
	Status      FlowRunStatus
	Artifacts   map[string]Artifact // Maps given edge ID to Artifact
	WaitingURLs []WaitingURL
	StartedAt   time.Time
}
