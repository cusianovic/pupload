package node

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/pupload/pupload/internal/logging"
	"github.com/pupload/pupload/internal/models"
	"github.com/pupload/pupload/internal/syncplane"
	"github.com/pupload/pupload/internal/telemetry"
)

func (ns *NodeService) FinishedMiddleware(ctx context.Context, payload syncplane.NodeExecutePayload) error {

	ctx = telemetry.ExtractContext(ctx, payload.TraceParent)

	logs := make([]models.LogRecord, 0, 64)
	ch := &logging.CollectHandler{
		Inner:   logging.Root().Handler(),
		Records: &logs,
	}

	jobLog := slog.New(ch)
	jobLog.With(
		"run_id", payload.RunID,
		"node_id", payload.Node.ID,
		"nodedef_publisher", payload.NodeDef.Publisher,
		"nodedef_name", payload.NodeDef.Name,
		"container_image", payload.NodeDef.Image,
	)
	jobLog.Debug("inside worker")

	ctx = logging.CtxWithLogger(ctx, jobLog)

	if err := ns.tryReserve(payload.NodeDef.Tier); err != nil {
		jobLog.Error("error attempting to reserve resources", "err", err)
		return err
	}

	res, genErr := ns.ResourceManger.GenerateContainerResource(payload.NodeDef.Tier)
	if genErr != nil {
		jobLog.Error("error attempting to generate ContainerResource", "err", genErr)
		return genErr
	}

	err := ns.NodeExecute(ctx, payload, res)
	if err == nil {
		if err := ns.SyncLayer.EnqueueNodeFinished(syncplane.NodeFinishedPayload{
			RunID:  payload.RunID,
			NodeID: payload.Node.ID,
			Logs:   logs,
		}); err != nil {
			jobLog.Error("error send node finished message", "err", err)

		}
	} else {
		if enqueueErr := ns.SyncLayer.EnqueueNodeFailed(syncplane.NodeFailedPayload{
			RunID:       payload.RunID,
			NodeID:      payload.Node.ID,
			Logs:        logs,
			Attempt:     payload.Attempt,
			MaxAttempts: payload.MaxAttempts,
			Error:       err.Error(),
			TraceParent: payload.TraceParent,
		}); enqueueErr != nil {
			jobLog.Error("error send node error message", "err", err)
		}

	}

	if err != nil {
		jobLog.Error(err.Error())
	}

	if err := ns.tryRelease(payload.NodeDef.Tier); err != nil {
		jobLog.Error("error attempting to release resources", "err", err)

	}

	return err
}

func (ns *NodeService) tryReserve(s string) error {
	ns.mu.Lock()
	defer ns.mu.Unlock()
	if err := ns.ResourceManger.Reserve(s); err != nil {
		return fmt.Errorf("ExecuteNodeHandler: Could not reserve resource %s: %w", s, err)
	}

	queueMap := ns.ResourceManger.GetValidTierMap()

	ns.SyncLayer.UpdateSubscribedQueues(queueMap)

	return nil
}

func (ns *NodeService) tryRelease(s string) error {
	ns.mu.Lock()
	defer ns.mu.Unlock()
	if err := ns.ResourceManger.Release(s); err != nil {
		return fmt.Errorf("ExecuteNodeHandler: Could not release resource %s: %w", s, err)
	}

	queueMap := ns.ResourceManger.GetValidTierMap()

	ns.SyncLayer.UpdateSubscribedQueues(queueMap)

	return nil
}
