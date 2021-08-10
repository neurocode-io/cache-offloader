package client

import (
	"sync"

	"dpd.de/indempotency-offloader/config"
	"github.com/go-redis/redis/v8"
)

type redisClient struct {
	*redis.Client
}

var (
	RedisInstance *redisClient
	redisOnce     sync.Once
)

func NewRedis() *redisClient {
	redisOnce.Do(func() {
		config := config.New()
		db := redis.NewClient(&redis.Options{
			Addr:     config.RedisConfig.ConnectionString,
			Password: config.RedisConfig.Password,
			DB:       config.RedisConfig.Database,
		})

		RedisInstance = &redisClient{db}
	})

	return RedisInstance
}
