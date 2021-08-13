package storage

import (
	"context"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

type redisRepository struct {
	*redis.Client
}

func NewRepository(redis *redis.Client) *redisRepository {
	return &redisRepository{redis}
}

func (r *redisRepository) LookUp(ctx context.Context, key string) ([]byte, error) {
	ctx, _ = context.WithTimeout(ctx, 200*time.Millisecond)
	result, err := r.Get(ctx, key).Result()

	if err == redis.Nil {
		return nil, nil
	}

	if err != nil {
		log.Printf("Redis-repository LookUp error: %v", err)
		return nil, err
	}

	return []byte(result), nil
}

func (r *redisRepository) Store(ctx context.Context, key string, resp []byte) error {
	ctx, _ = context.WithTimeout(ctx, 200*time.Millisecond)

	// purge stored values after 12 hours
	expireationTime := 12 * time.Hour

	err := r.Set(ctx, key, resp, expireationTime).Err()
	if err != nil {
		return err
	}

	return nil
}

func (r *redisRepository) CheckConnection(ctx context.Context) error {
	return r.Ping(ctx).Err()
}
