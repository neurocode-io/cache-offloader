package client

import (
	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog/log"
	"neurocode.io/cache-offloader/config"
)

type RedisClient struct {
	*redis.Client
}

func NewRedis(config config.RedisConfig) *RedisClient {
	opt, err := redis.ParseURL(config.ConnectionString)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse redis connection string")
	}

	return &RedisClient{redis.NewClient(opt)}
}
