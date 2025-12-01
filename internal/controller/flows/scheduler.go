package flows

import (
	"context"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

type FlowRunSchedulerConfigProvider struct {
	rc       *redis.Client
	Cronspec string
}

func CreateFlowRunScheduleProvider(rc *redis.Client, cronspec string) FlowRunSchedulerConfigProvider {
	return FlowRunSchedulerConfigProvider{
		rc:       rc,
		Cronspec: cronspec,
	}
}

func (p *FlowRunSchedulerConfigProvider) GetConfigs() ([]*asynq.PeriodicTaskConfig, error) {
	var configs []*asynq.PeriodicTaskConfig

	keys, _, err := p.rc.Scan(context.TODO(), 0, "flowrun:*", 1000).Result()
	if err != nil {
		return nil, err
	}

	for _, key := range keys {
		id := key[8:]
		task, err := NewFlowStepTask(id)
		if err != nil {
			continue
		}

		configs = append(configs, &asynq.PeriodicTaskConfig{Task: task, Cronspec: p.Cronspec, Opts: []asynq.Option{asynq.Queue("controller")}})
	}

	return configs, nil
}
