package server

import (
	"pupload/internal/syncplane"
	"pupload/internal/worker/container"
	"pupload/internal/worker/node"
)

func NewWorkerServer(s syncplane.SyncLayer, cs *container.ContainerService) {

	ns := node.CreateNodeService(cs, s)

	s.RegisterExecuteNodeHandler(ns.FinishedMiddleware)
	s.Start()
}
