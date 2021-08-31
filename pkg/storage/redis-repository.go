package storage

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"dpd.de/idempotency-offloader/pkg/metrics"
	"github.com/go-redis/redis/v8"
)

type ExpirationTime struct {
	Value time.Duration
}

type CommandTimeout struct {
	Value time.Duration
}

type RedisRepository struct {
	*redis.Client
	expirationTime *ExpirationTime
	commandTimeout *CommandTimeout
	metrics        *metrics.MetricCollector
}

func NewRepository(redis *redis.Client, expirationTime *ExpirationTime, commandTimeout *CommandTimeout) *RedisRepository {
	return &RedisRepository{redis, expirationTime, commandTimeout, metrics.NewMetricCollector()}
}

func (r *RedisRepository) LookUp(ctx context.Context, key string) (*Response, error) {
	ctx, _ = context.WithTimeout(ctx, r.commandTimeout.Value)
	result, err := r.Get(ctx, key).Result()

	if err == redis.Nil {
		r.metrics.Success()
		return nil, nil
	}

	if err != nil {
		log.Printf("Redis-repository LookUp error: %v", err)
		r.metrics.LookUpError()
		return nil, err
	}

	response := Response{}
	json.Unmarshal([]byte(result), &response)
	r.metrics.Success()

	return &response, nil
}

func (r *RedisRepository) Store(ctx context.Context, key string, resp *Response) error {
	ctx, _ = context.WithTimeout(ctx, r.commandTimeout.Value)

	storedResp, err := json.Marshal(resp)
	if err != nil {
		r.metrics.StorageError()
		return err
	}

	err = r.Set(ctx, key, storedResp, r.expirationTime.Value).Err()
	if err != nil {
		r.metrics.StorageError()
		return err
	}
	r.metrics.Success()

	return nil
}

func (r *RedisRepository) CheckConnection(ctx context.Context) error {
	return r.Ping(ctx).Err()
}
