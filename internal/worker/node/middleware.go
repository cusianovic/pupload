package node

import (
	"context"
	"log/slog"
	"os"
	"pupload/internal/logging"
	"pupload/internal/models"

	"github.com/hibiken/asynq"
)

func (ns *NodeService) FinishedMiddleware(h asynq.Handler) asynq.Handler {
	return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {

		logs := make([]models.LogRecord, 0, 64)
		ch := &logging.CollectHandler{
			Inner:   slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}),
			Records: &logs,
		}

		jobLog := slog.New(ch)

		ctx = logging.CtxWithLogger(ctx, jobLog)

		err := h.ProcessTask(ctx, t)
		if err == nil {
			task := NewNodeFinishedTask(logs)
			ns.AsynqClient.Enqueue(task)

		}

		return err
	})
}
