package scheduler

import (
	"time"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

type Scheduler struct {
	mgr *asynq.PeriodicTaskManager
}

func NewScheduler(provider asynq.PeriodicTaskConfigProvider, redisClient *redis.Client) (*Scheduler, error) {

	mgr, err := asynq.NewPeriodicTaskManager(asynq.PeriodicTaskManagerOpts{
		RedisUniversalClient:       redisClient,
		PeriodicTaskConfigProvider: provider,
		SyncInterval:               10 * time.Second,
	})

	if err != nil {
		return nil, err
	}

	return &Scheduler{mgr: mgr}, nil

}

func (s *Scheduler) Start() {
	s.mgr.Start()
}

func (s *Scheduler) Close() {
	s.mgr.Shutdown()
}
