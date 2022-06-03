package client

import (
	"github.com/go-redis/redis/v8"
	"neurocode.io/cache-offloader/config"
)

type redisClient struct {
	*redis.Client
}

func NewRedis(config config.RedisConfig) *redisClient {
	db := redis.NewClient(&redis.Options{
		Addr:     config.ConnectionString,
		Password: config.Password,
		DB:       config.Database,
	})

	return &redisClient{db}
}
