package storage

import (
	"context"
	"encoding/json"
	"time"

	"dpd.de/idempotency-offloader/pkg/metrics"
	"github.com/bloom42/rz-go"
	"github.com/bloom42/rz-go/log"
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
	ctx, _ = context.WithTimeout(ctx, r.commandTimeout.Value)
	result, err := r.Get(ctx, requestId).Result()

	if err == redis.Nil {
		log.Info("Redis-repository: key not found", rz.String("RequestId", requestId))
		r.metrics.Success()
		return nil, nil
	}

	if err != nil {
		log.Error("Redis-repository: LookUp error.", rz.String("Error", err.Error()), rz.String("RequestId", requestId))
		r.metrics.LookUpError()
		return nil, err
	}

	response := Response{}
	json.Unmarshal([]byte(result), &response)
	r.metrics.Success()

	return &response, nil
}

func (r *RedisRepository) Store(ctx context.Context, requestId string, resp *Response) error {
	ctx, _ = context.WithTimeout(ctx, r.commandTimeout.Value)

	storedResp, err := json.Marshal(resp)
	if err != nil {
		log.Error("Redis-repository: Store error; failed to json encode the http response.", rz.String("Error", err.Error()), rz.String("RequestId", requestId))
		r.metrics.StorageError()
		return err
	}

	err = r.Set(ctx, requestId, storedResp, r.expirationTime.Value).Err()
	if err != nil {
		log.Error("Redis-repository: Store error.", rz.String("Error", err.Error()), rz.String("RequestId", requestId))
		r.metrics.StorageError()
		return err
	}
	r.metrics.Success()

	return nil
}

func (r *RedisRepository) CheckConnection(ctx context.Context) error {
	return r.Ping(ctx).Err()
}
