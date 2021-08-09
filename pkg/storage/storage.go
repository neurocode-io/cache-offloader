package storage

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"dpd.de/indempotency-offloader/config"

	"github.com/go-redis/redis/v8"
)

type Repository interface {
	LookUp(string) (*http.Response, error)
	Store(string, *http.Response) error
}

type RedisCache struct {
	expirationSeconds time.Duration
	ctx               context.Context
	client            *redis.Client
}

func NewRedisCache(expirationSeconds time.Duration) Repository {
	return &RedisCache{
		expirationSeconds: expirationSeconds,
		ctx:               context.Background(),
		client: redis.NewClient(&redis.Options{
			Addr:     strings.Join([]string{config.GetEnv("REDIS_HOST", "localhost"), config.GetEnv("REDIS_PORT", "6379")}, ":"),
			Password: config.GetEnv("REDIS_PASSWORD", "development"),
			DB:       0,
		}),
	}
}

func (c *RedisCache) Store(key string, resp *http.Response) error {
	jsonVal, err := json.Marshal(resp)

	if err != nil {
		return err
	}

	err = c.client.Set(c.ctx, key, jsonVal, c.expirationSeconds*time.Second).Err()

	if err != nil {
		return err
	}

	return nil
}

func (c *RedisCache) LookUp(key string) (*http.Response, error) {
	result, err := c.client.Get(c.ctx, key).Result()

	if err != nil {
		return nil, err
	}

	response := http.Response{}
	err = json.Unmarshal([]byte(result), &response)

	if err != nil {
		return nil, err
	}

	return &response, nil
}
