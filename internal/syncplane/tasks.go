package syncplane

import (
	"context"

	"github.com/pupload/pupload/internal/models"
)

const (
	TypeFlowStep        = "flow:step"
	TypeStepExecute     = "step:execute"
	TypeStepFinished    = "step:finished"
	TypeStepFailed      = "step:failed"
	TypeControllerClean = "controller:clean"
)

type ExecuteStepHandler func(ctx context.Context, payload StepExecutePayload) error
type StepExecutePayload struct {
	RunID      string
	Task       models.Task
	Step       models.Step
	InputURLs  map[string]string
	OutputURLs map[string]string

	MaxAttempts int
	Attempt     int

	TraceParent string
}

type FlowStepHandler func(ctx context.Context, payload FlowStepPayload) error
type FlowStepPayload struct {
	RunID string
}

type StepFinishedHandler func(ctx context.Context, payload StepFinishedPayload) error
type StepFinishedPayload struct {
	RunID  string
	StepID string
	Logs   []models.LogRecord

	TraceParent string
}

type StepFailedHandler func(ctx context.Context, payload StepFailedPayload) error
type StepFailedPayload struct {
	RunID       string
	StepID      string
	Attempt     int
	MaxAttempts int
	Error       string
	Logs        []models.LogRecord

	TraceParent string
}
