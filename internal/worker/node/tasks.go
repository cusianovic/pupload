package node

import (
	"encoding/json"
	"pupload/internal/models"

	"github.com/hibiken/asynq"
)

func NewNodeFinishedTask(RunID, NodeID string, logs []models.LogRecord) *asynq.Task {
	payload, _ := json.Marshal(models.NodeFinishedPayload{
		RunID:  RunID,
		NodeID: NodeID,
		Logs:   logs,
	})

	return asynq.NewTask(models.TypeNodeFinished, payload, asynq.Queue("controller"))
}
