package server

import (
	"fmt"

	"github.com/pupload/pupload/internal/resources"
	"github.com/pupload/pupload/internal/syncplane"
	"github.com/pupload/pupload/internal/worker/container"
	"github.com/pupload/pupload/internal/worker/step"
)

func NewWorkerServer(s syncplane.SyncLayer, cs *container.ContainerService, rm *resources.ResourceManager) {

	ss, err := step.CreateStepService(cs, s, rm)
	if err != nil {
		panic(fmt.Sprintf("Unable to create step service: %s", err))
	}

	s.RegisterExecuteStepHandler(ss.FinishedMiddleware)
	s.Start()
}
