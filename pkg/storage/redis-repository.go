package storage

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

type Response struct {
	Header map[string][]string
	Body   []byte
}

type redisRepository struct {
	*redis.Client
	expirationTime time.Duration
}

func NewRepository(redis *redis.Client, expirationTime time.Duration) *redisRepository {
	return &redisRepository{redis, expirationTime}
}

func (r *redisRepository) LookUp(ctx context.Context, key string) (*Response, error) {
	ctx, _ = context.WithTimeout(ctx, 200*time.Millisecond)
	result, err := r.Get(ctx, key).Result()

	if err == redis.Nil {
		return nil, nil
	}

	if err != nil {
		log.Printf("Redis-repository LookUp error: %v", err)
		return nil, err
	}

	response := Response{}
	json.Unmarshal([]byte(result), &response)

	return &response, nil
}

func (r *redisRepository) Store(ctx context.Context, key string, resp *Response) error {
	ctx, _ = context.WithTimeout(ctx, 200*time.Millisecond)

	storedResp, err := json.Marshal(resp)
	if err != nil {
		return err
	}

	err = r.Set(ctx, key, storedResp, r.expirationTime).Err()
	if err != nil {
		return err
	}

	return nil
}

func (r *redisRepository) CheckConnection(ctx context.Context) error {
	return r.Ping(ctx).Err()
}
