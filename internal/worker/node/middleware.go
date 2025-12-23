package node

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"pupload/internal/logging"
	"pupload/internal/models"
	"pupload/internal/syncplane"

	"github.com/hibiken/asynq"
)

func (ns *NodeService) FinishedMiddleware(h asynq.Handler) asynq.Handler {
	return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {

		var p syncplane.NodeExecutePayload

		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return err
		}

		logs := make([]models.LogRecord, 0, 64)
		ch := &logging.CollectHandler{
			Inner:   slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}),
			Records: &logs,
		}

		jobLog := slog.New(ch)

		ctx = logging.CtxWithLogger(ctx, jobLog)

		err := h.ProcessTask(ctx, t)
		if err == nil {
			task := NewNodeFinishedTask(p.RunID, p.Node.ID, logs)
			ns.AsynqClient.Enqueue(task)

		}

		return err
	})
}
