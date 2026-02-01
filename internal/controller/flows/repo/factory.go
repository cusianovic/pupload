package repo

import (
	"fmt"

	runtime_repo "github.com/pupload/pupload/internal/controller/flows/repo/runtime"
	"github.com/redis/go-redis/v9"
)

type RuntimeRepoType string

const (
	RedisRuntimeRepo RuntimeRepoType = "redis"
)

type RuntimeRepoSettings struct {
	Type RuntimeRepoType

	Redis RedisSettings
}

type RedisSettings struct {
	Address  string
	Password string
	DB       int
}

func CreateRuntimeRepo(cfg RuntimeRepoSettings) (RuntimeRepo, error) {
	switch cfg.Type {
	case RedisRuntimeRepo:
		rdb := redis.NewClient(&redis.Options{
			Addr:     cfg.Redis.Address,
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		})

		return runtime_repo.CreateRedisRuntimeRepo(rdb), nil
	}

	return nil, fmt.Errorf("invalid runtime repo config")
}
