package controller

import (
	"context"
	"log/slog"
	"net/http"
	"pupload/internal/controller/config"
	flow "pupload/internal/controller/flows/service"
	controllerserver "pupload/internal/controller/server"
	"pupload/internal/logging"
	"pupload/internal/syncplane"
	"pupload/internal/telemetry"
)

func RunWithConfig(cfg *config.ControllerSettings) error {
	ctx := context.Background()

	logging.Init(logging.Config{
		AppName: "pupload.controller",
		Level:   slog.LevelInfo,
		Format:  "text",
	})
	log := logging.Root()

	if err := telemetry.Init(cfg.Telemetry, "pupload.controller"); err != nil {
		log.Error("error initalizing telemetry", "err", err)
	}
	defer telemetry.Shutdown(context.Background())

	// SyncPlane
	s, err := syncplane.CreateControllerSyncLayer(cfg.SyncPlane)
	if err != nil {
		log.Error("error intalizing sync plane", "err", err)
		return err
	}
	defer s.Close()

	// Services

	f, err := flow.CreateFlowService(cfg, s)
	if err != nil {
		return err
	}
	defer f.Close(ctx)

	// Handlers

	handler := controllerserver.NewServer(*cfg, f)
	srv := &http.Server{
		Addr:    ":1234",
		Handler: handler,
	}

	return srv.ListenAndServe()
}

func Run() error {
	cfg := config.DefaultConfig()
	return RunWithConfig(cfg)
}
