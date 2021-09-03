package storage

import (
	"context"
	"encoding/json"
	"time"

	"dpd.de/idempotency-offloader/pkg/metrics"
	"github.com/bloom42/rz-go"
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

func (r *RedisRepository) LookUp(ctx context.Context, requestId string) (*Response, error) {
	logger := rz.FromCtx(ctx)
	ctx, cancel := context.WithTimeout(ctx, r.commandTimeout.Value)
	defer cancel()
	result, err := r.Get(ctx, requestId).Result()

	if err == redis.Nil {
		logger.Debug("Redis-repository: key not found")
		r.metrics.Success()
		return nil, nil
	}

	if err != nil {
		logger.Error("Redis-repository: LookUp error.", rz.Err(err))
		r.metrics.LookUpError()
		return nil, err
	}

	response := Response{}
	json.Unmarshal([]byte(result), &response)
	r.metrics.Success()

	return &response, nil
}

func (r *RedisRepository) Store(ctx context.Context, requestId string, resp *Response) error {
	logger := rz.FromCtx(ctx)

	ctx, cancel := context.WithTimeout(ctx, r.commandTimeout.Value)
	defer cancel()

	storedResp, err := json.Marshal(resp)
	if err != nil {
		logger.Error("Redis-repository: Store error; failed to json encode the http response.", rz.Err(err))
		r.metrics.StorageError()
		return err
	}

	err = r.Set(ctx, requestId, storedResp, r.expirationTime.Value).Err()
	if err != nil {
		logger.Error("Redis-repository: Store error.", rz.Err(err))
		r.metrics.StorageError()
		return err
	}
	r.metrics.Success()

	return nil
}

func (r *RedisRepository) CheckConnection(ctx context.Context) error {
	return r.Ping(ctx).Err()
}
