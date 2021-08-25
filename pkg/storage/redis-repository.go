package storage

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"dpd.de/idempotency-offloader/pkg/metrics"
	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	counter *prometheus.CounterVec
)

type ExpirationTime struct {
	Value time.Duration
}

type CommandTimeout struct {
	Value time.Duration
}

type Response struct {
	Header map[string][]string
	Body   []byte
}

type redisRepository struct {
	*redis.Client
	expirationTime *ExpirationTime
	commandTimeout *CommandTimeout
}

func NewRepository(redis *redis.Client, expirationTime *ExpirationTime, commandTimeout *CommandTimeout) *redisRepository {
	opts := metrics.CounterVecOpts{
		Name:       "cached_http_requests",
		Help:       "Number of cached http requests by status.",
		LabelNames: "status",
	}

	counter = metrics.AddCounterVec(&opts)
	return &redisRepository{redis, expirationTime, commandTimeout}
}

func (r *redisRepository) LookUp(ctx context.Context, key string) (*Response, error) {
	ctx, _ = context.WithTimeout(ctx, r.commandTimeout.Value)
	result, err := r.Get(ctx, key).Result()

	if err == redis.Nil {
		return nil, nil
	}

	if err != nil {
		counter.WithLabelValues(LookUpError).Inc()
		log.Printf("Redis-repository LookUp error: %v", err)
		return nil, err
	}

	response := Response{}
	json.Unmarshal([]byte(result), &response)

	return &response, nil
}

func (r *redisRepository) Store(ctx context.Context, key string, resp *Response) error {
	ctx, _ = context.WithTimeout(ctx, r.commandTimeout.Value)

	storedResp, err := json.Marshal(resp)
	if err != nil {
		return err
	}

	err = r.Set(ctx, key, storedResp, r.expirationTime.Value).Err()
	if err != nil {
		counter.WithLabelValues(StorageError).Inc()
		return err
	}

	counter.WithLabelValues(Success).Inc()

	return nil
}

func (r *redisRepository) CheckConnection(ctx context.Context) error {
	return r.Ping(ctx).Err()
}
