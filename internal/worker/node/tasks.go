package node

import (
	"encoding/json"
	"pupload/internal/models"

	"github.com/hibiken/asynq"
)

func NewNodeFinishedTask(logs []models.LogRecord) *asynq.Task {
	payload, _ := json.Marshal(models.NodeFinishedPayload{
		Logs: logs,
	})

	return asynq.NewTask(models.TypeNodeFinished, payload, asynq.Queue("controller"))
}
