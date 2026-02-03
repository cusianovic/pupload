package controller

import (
	"context"
	"io"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/pupload/pupload/internal/controller/config"
	flow "github.com/pupload/pupload/internal/controller/flows/service"
	"github.com/pupload/pupload/internal/controller/projects"
	controllerserver "github.com/pupload/pupload/internal/controller/server"
	"github.com/pupload/pupload/internal/logging"
	"github.com/pupload/pupload/internal/syncplane"
	"github.com/pupload/pupload/internal/telemetry"
)

func Run() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	ctx := context.Background()
	return RunWithConfig(ctx, cfg)
}

func RunWithConfigSilent(ctx context.Context, cfg *config.ControllerConfig) error {

	log.SetOutput(io.Discard)
	logging.Init(logging.Config{
		AppName: "pupload.controller",
		Level:   slog.LevelInfo,
		Format:  "text",
		Out:     io.Discard,
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

	p, err := projects.CreateProjectService(cfg.ProjectRepo)
	if err != nil {
		return err
	}
	defer p.Close()

	// Handlers

	handler := controllerserver.NewServer(*cfg, f, p)
	srv := &http.Server{
		Addr:    ":1234",
		Handler: handler,
	}

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Error("server forced to shutdown", "err", err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	log.Info("Controller shutting down...")
	return srv.Shutdown(shutdownCtx)

}

func RunWithConfig(ctx context.Context, cfg *config.ControllerConfig) error {

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

	p, err := projects.CreateProjectService(cfg.ProjectRepo)
	if err != nil {
		return err
	}

	defer p.Close()

	// Handlers

	handler := controllerserver.NewServer(*cfg, f, p)
	srv := &http.Server{
		Addr:    ":1234",
		Handler: handler,
	}

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Error("server forced to shutdown", "err", err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	log.Info("Controller shutting down...")
	return srv.Shutdown(shutdownCtx)
}
