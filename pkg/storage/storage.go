package storage

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"dpd.de/indempotency-offloader/config"

	"github.com/go-redis/redis/v8"
)

var (
	ctx    = context.Background()
	client = redis.NewClient(&redis.Options{
		Addr:     strings.Join([]string{config.GetEnv("REDIS_HOST", "localhost"), config.GetEnv("REDIS_PORT", "6379")}, ":"),
		Password: config.GetEnv("REDIS_PASSWORD", ""),
		DB:       0,
	})
)

type Repository interface {
	LookUp(string) (*http.Response, error)
	Store(string, *http.Response) error
}

type RedisCache struct {
	expirationSeconds time.Duration
}

func NewRedisCache(expirationSeconds time.Duration) Repository {
	return &RedisCache{
		expirationSeconds: expirationSeconds,
	}
}

func (c *RedisCache) Store(key string, resp *http.Response) error {
	serializedResp, err := httputil.DumpResponse(resp, true)

	if err != nil {
		return err
	}

	err = client.Set(ctx, key, serializedResp, c.expirationSeconds*time.Second).Err()

	if err != nil {
		return err
	}

	return nil
}

func (c *RedisCache) LookUp(key string) (*http.Response, error) {
	result, err := client.Get(ctx, key).Result()

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
