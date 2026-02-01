package projects

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/pupload/pupload/internal/models"
	"github.com/redis/go-redis/v9"
)

type ProjectRepo interface {
	SaveProject(ctx context.Context, p models.Project) error
	LoadProject(ctx context.Context, id uuid.UUID) (models.Project, error)
	DeleteProject(ctx context.Context, id uuid.UUID) error
	Close()
}

type RedisProjectRepo struct {
	rdb *redis.Client
}

type RedisProjectRepoConfig struct {
	Address  string
	Password string
	DB       int
}

func CreateRedisProjectRepo(cfg RedisProjectRepoConfig) (*RedisProjectRepo, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Address,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	err := rdb.Ping(context.Background()).Err()
	if err != nil {
		return nil, fmt.Errorf("unable to connect to redis")
	}

	return &RedisProjectRepo{
		rdb: rdb,
	}, nil
}

func (r *RedisProjectRepo) SaveProject(ctx context.Context, p models.Project) error {
	blob, err := json.Marshal(p)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("pupload:project:%s", p.ID.String())

	return r.rdb.Set(ctx, key, blob, 0).Err()
}

var ErrProjectDoesNotExist = fmt.Errorf("project does not exist")

func (r *RedisProjectRepo) LoadProject(ctx context.Context, id uuid.UUID) (models.Project, error) {
	var project models.Project

	key := fmt.Sprintf("pupload:project:%s", id.String())
	blob, err := r.rdb.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return models.Project{}, ErrProjectDoesNotExist
		}

		return models.Project{}, err
	}

	if err := json.Unmarshal([]byte(blob), &project); err != nil {
		return models.Project{}, err
	}

	return project, nil
}

func (r *RedisProjectRepo) DeleteProject(ctx context.Context, id uuid.UUID) error {
	return r.rdb.Del(ctx, id.String()).Err()
}

func (r *RedisProjectRepo) Close() {
	r.rdb.Close()
}
