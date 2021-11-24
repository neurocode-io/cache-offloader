package storage

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/skerkour/rz"
	"neurocode.io/cache-offloader/pkg/metrics"
)

const expirationTime = 0 * time.Second

type RedisRepository struct {
	*redis.Client
	commandTimeout time.Duration
	metrics        *metrics.MetricCollector
}

func NewRepository(redis *redis.Client, commandTimeout time.Duration) *RedisRepository {
	return &RedisRepository{redis, commandTimeout, metrics.NewMetricCollector()}
}

func (r *RedisRepository) LookUp(ctx context.Context, requestId string) (*Response, error) {
	logger := rz.FromCtx(ctx)
	ctx, cancel := context.WithTimeout(ctx, r.commandTimeout)
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
	response.UnmarshalJSON([]byte(result))
	r.metrics.Success()

	return &response, nil
}

func (r *RedisRepository) Store(ctx context.Context, requestId string, resp *Response) error {
	logger := rz.FromCtx(ctx)

	ctx, cancel := context.WithTimeout(ctx, r.commandTimeout)
	defer cancel()

	storedResp, err := resp.MarshalJSON()
	if err != nil {
		logger.Error("Redis-repository: Store error; failed to json encode the http response.", rz.Err(err))
		r.metrics.StorageError()
		return err
	}

	err = r.Set(ctx, requestId, storedResp, expirationTime).Err()
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