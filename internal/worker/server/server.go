package server

import (
	"log"
	"pupload/internal/models"
	"pupload/internal/worker/container"
	"pupload/internal/worker/node"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

func NewWorkerServer() {

	rc := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	srv := asynq.NewServerFromRedisClient(rc, asynq.Config{
		Concurrency: 1,
		Queues: map[string]int{
			"worker": 1,
		},
	})

	cli := asynq.NewClientFromRedisClient(rc)

	ds := container.CreateContainerService()
	ns := node.CreateNodeService(&ds, cli)

	mux := asynq.NewServeMux()
	mux.Use(ns.FinishedMiddleware)
	mux.HandleFunc(models.TypeNodeExecute, ns.HandleNodeExecuteTask)

	if err := srv.Run(mux); err != nil {
		log.Fatalf("Error starting worker: %s", err.Error())
	}

}
