package storage

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
)

type redisRepository struct {
	*redis.Client
}

func NewRepository(redis *redis.Client) *redisRepository {
	return &redisRepository{redis}
}

func (r *redisRepository) LookUp(ctx context.Context, key string) (*http.Response, error) {
	ctx, _ = context.WithTimeout(ctx, 50*time.Millisecond)
	result, err := r.Get(ctx, key).Result()

	if err == redis.Nil {
		return nil, nil
	}

	if err != nil {
		log.Printf("Redis-repository LookUp error: %v", err)
		return nil, err
	}

	response := http.Response{}
	err = json.Unmarshal([]byte(result), &response)

	if err != nil {
		return nil, err
	}

	return &response, nil

}

func (r *redisRepository) Store(ctx context.Context, key string, obj []byte) error {
	ctx, _ = context.WithTimeout(ctx, 50*time.Millisecond)
	// purge stored values after 12 hours
	expireationTime := 12 * time.Hour

	err := r.Set(ctx, key, obj, expireationTime).Err()
	if err != nil {
		return err
	}

	return nil
}

func (r *redisRepository) CheckConnection(ctx context.Context) error {
	return r.Ping(ctx).Err()
}