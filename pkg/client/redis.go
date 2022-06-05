package client

import (
	"github.com/go-redis/redis/v8"
	"neurocode.io/cache-offloader/config"
)

type RedisClient struct {
	*redis.Client
}

func NewRedis(config config.RedisConfig) *RedisClient {
	db := redis.NewClient(&redis.Options{
		Addr:     config.ConnectionString,
		Password: config.Password,
		DB:       config.Database,
	})

	return &RedisClient{db}
}
