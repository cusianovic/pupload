package main

import (
	"log/slog"
	"pupload/internal/logging"
	"pupload/internal/worker/server"
)

func main() {

	logging.Init(logging.Config{
		AppName: "worker",
		Level:   slog.LevelInfo,
		Format:  "json",
	})

	server.NewWorkerServer()
}
