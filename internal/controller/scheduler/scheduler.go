package scheduler

import (
	"context"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

type Scheduler struct {
	mgr   *asynq.PeriodicTaskManager
	mutex *redsync.Mutex
	ttl   time.Duration
}

func NewScheduler(provider asynq.PeriodicTaskConfigProvider, redisClient *redis.Client) (*Scheduler, error) {

	pool := goredis.NewPool(redisClient)
	rd := redsync.New(pool)
	ttl := 30 * time.Second

	mutex := rd.NewMutex("asynq:periodic:leader", redsync.WithExpiry(ttl), redsync.WithTries(1))

	mgr, err := asynq.NewPeriodicTaskManager(asynq.PeriodicTaskManagerOpts{
		RedisUniversalClient:       redisClient,
		PeriodicTaskConfigProvider: provider,
		SyncInterval:               10 * time.Second,
	})

	if err != nil {
		return nil, err
	}

	return &Scheduler{mgr: mgr, mutex: mutex, ttl: ttl}, nil

}

func (s *Scheduler) Start(ctx context.Context) {

	ticker := time.NewTicker(s.ttl / 2)
	defer ticker.Stop()

	leader := false

	tryBecomeLeader := func(ctx context.Context) {
		if leader {
			return
		}

		if err := s.mutex.LockContext(ctx); err != nil {
			// couldn't get the lock now; will retry on next tick
			return
		}

		if err := s.mgr.Start(); err != nil {
			// failed to start manager; release lock and don't mark leader
			s.mutex.UnlockContext(ctx)
			return
		}

		leader = true
	}

	tryBecomeLeader(ctx)

	for {
		select {
		case <-ctx.Done():
			s.Close()
			s.mutex.UnlockContext(ctx)
			return

		case <-ticker.C:
		}

		if !leader {
			tryBecomeLeader(ctx)
			continue

		}

		ok, err := s.mutex.ExtendContext(ctx)
		if err != nil || !ok {
			s.mgr.Shutdown()
			s.mutex.UnlockContext(ctx)
			leader = false
		}

	}
}

func (s *Scheduler) Close() {
	s.mgr.Shutdown()
}
