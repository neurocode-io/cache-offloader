package storage

import (
	"context"
	"errors"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"neurocode.io/cache-offloader/config"
	"neurocode.io/cache-offloader/pkg/client"
	"neurocode.io/cache-offloader/pkg/model"
)

func getConnString() string {
	connString := os.Getenv("REDIS_CONNECTION_STRING")
	if connString == "" {
		return "redis://localhost:6379"
	}

	return connString
}

func TestRedisLookup(t *testing.T) {
	ctx := context.Background()
	lookupTimeout := time.Millisecond * 500
	r := client.NewRedis(config.RedisConfig{ConnectionString: getConnString()})
	err := NewRedisStorage(r, 100).Store(ctx, "test", &model.Response{})
	assert.Nil(t, err)

	tests := []struct {
		name          string
		requestID     string
		redisClient   *client.RedisClient
		lookupTimeout time.Duration
		want          *model.Response
		expErr        error
	}{
		{
			name:          "key found",
			redisClient:   r,
			lookupTimeout: lookupTimeout,
			requestID:     "test",
			want:          &model.Response{StaleValue: model.FreshValue},
			expErr:        nil,
		},
		{
			name:          "key not found",
			redisClient:   r,
			lookupTimeout: lookupTimeout,
			requestID:     "not-found",
			want:          nil,
			expErr:        nil,
		},
		{
			name:          "redis error",
			redisClient:   client.NewRedis(config.RedisConfig{ConnectionString: "redis://localhost:6378"}),
			lookupTimeout: lookupTimeout,
			requestID:     "not-found",
			want:          nil,
			expErr:        errors.New("redis-repository: LookUp error: dial tcp [::1]:6378: connect: connection refused"),
		},
		{
			name:          "lookup timeout",
			redisClient:   r,
			lookupTimeout: 3 * time.Millisecond,
			requestID:     "test",
			want:          nil,
			expErr:        errors.New("redis-repository: LookUp error: context deadline exceeded"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			redis := NewRedisStorage(tt.redisClient, 100)
			redis.lookupTimeout = tt.lookupTimeout
			got, err := redis.LookUp(ctx, tt.requestID)
			assert.Equal(t, tt.want, got)
			if tt.expErr != nil {
				assert.EqualError(t, err, tt.expErr.Error())
			}
		})
	}
}

func TestRedisStore(t *testing.T) {
	ctx := context.Background()
	r := client.NewRedis(config.RedisConfig{ConnectionString: getConnString()})

	tests := []struct {
		name        string
		requestID   string
		response    *model.Response
		redisClient *client.RedisClient
		expErr      error
	}{
		{
			name:        "store response successfully",
			redisClient: r,
			requestID:   "test",
			response:    &model.Response{Status: http.StatusAccepted},
			expErr:      nil,
		},
		{
			name:        "redis error",
			redisClient: client.NewRedis(config.RedisConfig{ConnectionString: "redis://localhost:6378"}),
			requestID:   "test",
			response:    nil,
			expErr:      errors.New("redis-repository: Store error: dial tcp [::1]:6378: connect: connection refused"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			redis := NewRedisStorage(tt.redisClient, 100)
			err := redis.Store(ctx, tt.requestID, tt.response)
			if tt.expErr != nil {
				assert.EqualError(t, err, tt.expErr.Error())
			}
		})
	}
}
