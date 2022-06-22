package client

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/neurocode-io/cache-offloader/config"
	"github.com/rs/zerolog/log"
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

func (r *RedisClient) configureLRU(size string) error {
	ctx := context.TODO()
	if err := r.Client.ConfigSet(ctx, "maxmemory", size).Err(); err != nil {
		return err
	}
	if err := r.Client.ConfigSet(ctx, "maxmemory-policy", "allkeys-lru").Err(); err != nil {
		return err
	}

	return nil
}

func (r *RedisClient) configureLFU(size string) error {
	ctx := context.TODO()
	if err := r.Client.ConfigSet(ctx, "maxmemory", size).Err(); err != nil {
		return err
	}

	if err := r.Client.ConfigSet(ctx, "maxmemory-policy", "allkeys-lfu").Err(); err != nil {
		return err
	}

	if err := r.Client.ConfigSet(ctx, "maxmemory-samples", "5").Err(); err != nil {
		return err
	}

	if err := r.Client.ConfigSet(ctx, "maxmemory-eviction-tenacity", "10").Err(); err != nil {
		return err
	}

	return nil
}

func (r *RedisClient) ConfigureLRU(cacheSize float64) {
	err := r.configureLRU(fmt.Sprintf("%.0fmb", cacheSize))
	if err != nil {
		log.Warn().Err(err).Msg("Failed to configure redis for LRU cache")
		log.Warn().Msg("Please consider setting the redis maxmemory and maxmemory-policy options yourself")
	}
}

func (r *RedisClient) ConfigureLFU(cacheSize float64) {
	err := r.configureLFU(fmt.Sprintf("%.0fmb", cacheSize))
	if err != nil {
		log.Warn().Err(err).Msg("Failed to configure redis for LFU cache")
		log.Warn().Msg("Please consider setting the redis maxmemory and maxmemory-policy options yourself")
	}
}
