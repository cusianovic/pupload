package util

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

func AcquireLock(client *redis.Client, key, value string, duration time.Duration) (bool, error) {

	result, err := client.SetNX(context.TODO(), key, "lock", duration).Result()
	if err != nil {
		return false, err
	}

	return result, nil
}

func ReleaseLock(client *redis.Client, key string) error {

	_, err := client.Del(context.TODO(), key).Result()
	return err
}
