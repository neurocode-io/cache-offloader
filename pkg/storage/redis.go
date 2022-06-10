package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog/log"
	"neurocode.io/cache-offloader/pkg/model"
)

//go:generate mockgen -source=./redis.go -destination=./redis-mock_test.go -package=storage
type (
	IRedis interface {
		Ping(ctx context.Context) *redis.StatusCmd
		TxPipeline() redis.Pipeliner
	}
	RedisStorage struct {
		db             IRedis
		staleInSeconds int
		lookupTimeout  time.Duration
	}
)

func NewRedisStorage(db IRedis, staleInSeconds int) RedisStorage {
	return RedisStorage{db: db, staleInSeconds: staleInSeconds, lookupTimeout: lookupTimeout}
}

func (r RedisStorage) aliveKey(key string) string {
	return key + ":alive"
}

func (r RedisStorage) LookUp(ctx context.Context, requestID string) (*model.Response, error) {
	logger := log.Ctx(ctx)
	response := &model.Response{}
	ctx, cancel := context.WithTimeout(ctx, r.lookupTimeout)
	defer cancel()

	pipe := r.db.TxPipeline()
	stale := pipe.Exists(ctx, r.aliveKey(requestID))
	cachedResponse := pipe.Get(ctx, requestID)
	_, err := pipe.Exec(ctx)

	if cachedResponse.Err() == redis.Nil {
		logger.Debug().Msg("Redis-repository: key not found")

		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("redis-repository: LookUp error: %w", err)
	}

	err = json.Unmarshal([]byte(cachedResponse.Val()), response)
	response.StaleValue = uint8(stale.Val())

	return response, err
}

func (r RedisStorage) Store(ctx context.Context, requestID string, resp *model.Response) error {
	storedResp, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("redis-repository: Store error: %w", err)
	}

	pipe := r.db.TxPipeline()
	pipe.Set(ctx, requestID, storedResp, expirationTime)
	pipe.Set(ctx, r.aliveKey(requestID), model.FreshValue, time.Second*time.Duration(r.staleInSeconds))
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("redis-repository: Store error: %w", err)
	}

	return nil
}

func (r RedisStorage) CheckConnection(ctx context.Context) error {
	return r.db.Ping(ctx).Err()
}
