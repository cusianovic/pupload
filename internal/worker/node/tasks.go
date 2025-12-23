package node

import (
	"encoding/json"
	"pupload/internal/models"
	"pupload/internal/syncplane"

	"github.com/hibiken/asynq"
)

func NewNodeFinishedTask(RunID, NodeID string, logs []models.LogRecord) *asynq.Task {
	payload, _ := json.Marshal(syncplane.NodeFinishedPayload{
		RunID:  RunID,
		NodeID: NodeID,
		Logs:   logs,
	})

	return asynq.NewTask(syncplane.TypeNodeFinished, payload, asynq.Queue("controller"))
}
