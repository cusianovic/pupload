package worker

import (
	"log/slog"
	"os"
	"os/signal"
	"pupload/internal/logging"
	"pupload/internal/syncplane"
	"pupload/internal/telemetry"
	"pupload/internal/worker/config"
	"pupload/internal/worker/container"
	"pupload/internal/worker/server"
	"syscall"
)

func Run() error {
	cfg := config.DefaultConfig()
	return RunWithConfig(cfg)
}

func RunWithConfig(cfg *config.WorkerConfig) error {

	logging.Init(logging.Config{
		AppName: "worker",
		Level:   slog.LevelInfo,
		Format:  "json",
	})

	telemetry.Init(cfg.Telemetry, "pupload.worker")

	s, err := syncplane.CreateWorkerSyncLayer(cfg.SyncPlane, cfg.Resources)
	if err != nil {
		return err
	}

	cs := container.CreateContainerService()
	server.NewWorkerServer(s, &cs)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	sig := <-quit
	_ = sig

	return s.Close()

}
