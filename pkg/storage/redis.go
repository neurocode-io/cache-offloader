package storage

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog/log"
	"neurocode.io/cache-offloader/pkg/model"
)

const (
	expirationTime = 0 * time.Second
	commandTimeout = time.Millisecond * 100
)

type (
	IRedis interface {
		Get(ctx context.Context, key string) *redis.StringCmd
		Ping(ctx context.Context) *redis.StatusCmd
		Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	}
	RedisStorage struct {
		db             IRedis
		commandTimeout time.Duration
	}
)

func NewRedisStorage(db IRedis) RedisStorage {
	return RedisStorage{db: db, commandTimeout: commandTimeout}
}

func (r RedisStorage) LookUp(ctx context.Context, requestID string) (*model.Response, error) {
	logger := log.Ctx(ctx)
	ctx, cancel := context.WithTimeout(ctx, r.commandTimeout)
	defer cancel()
	result, err := r.db.Get(ctx, requestID).Result()

	if err == redis.Nil {
		logger.Debug().Msg("Redis-repository: key not found")

		return nil, nil
	}

	if err != nil {
		logger.Error().Err(err).Msg("Redis-repository: LookUp error.")

		return nil, err
	}

	response := &model.Response{}
	err = json.Unmarshal([]byte(result), response)

	return response, err
}

func (r RedisStorage) Store(ctx context.Context, requestID string, resp *model.Response) error {
	logger := log.Ctx(ctx)

	storedResp, err := json.Marshal(resp)
	if err != nil {
		logger.Error().Err(err).Msg("Redis-repository: Store error; failed to json encode the http response")

		return err
	}

	err = r.db.Set(ctx, requestID, storedResp, expirationTime).Err()
	if err != nil {
		logger.Error().Err(err).Msg("Redis-repository: Store error.")

		return err
	}

	return nil
}

func (r RedisStorage) CheckConnection(ctx context.Context) error {
	return r.db.Ping(ctx).Err()
}
